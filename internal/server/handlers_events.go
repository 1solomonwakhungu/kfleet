package server

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/1solomonwakhungu/kfleet/internal/store"
	"github.com/1solomonwakhungu/kfleet/pkg/api"
	"github.com/1solomonwakhungu/kfleet/pkg/types"
)

const maxPolicyFindingBodyBytes = 64 << 10

func (s *Server) registerEventRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/timeline", s.handleFleetTimeline)
	mux.HandleFunc("GET /api/v1/clusters/{id}/timeline", s.handleClusterTimeline)
	mux.HandleFunc("POST /api/v1/clusters/{id}/policy-findings", s.handleRecordPolicyFinding)
}

func (s *Server) handleFleetTimeline(w http.ResponseWriter, r *http.Request) {
	s.serveTimeline(w, r, "")
}

func (s *Server) handleClusterTimeline(w http.ResponseWriter, r *http.Request) {
	cluster, err := s.clusterByIDOrName(r, r.PathValue("id"))
	if handleStoreError(w, err) {
		return
	}
	s.serveTimeline(w, r, cluster.ID)
}

func (s *Server) serveTimeline(w http.ResponseWriter, r *http.Request, clusterID string) {
	filter := store.EventFilter{ClusterID: clusterID}
	query := r.URL.Query()

	if raw := strings.TrimSpace(query.Get("since")); raw != "" {
		since, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			api.WriteError(w, http.StatusBadRequest, "since must be an RFC3339 timestamp")
			return
		}
		filter.Since = &since
	}
	if raw := strings.TrimSpace(query.Get("until")); raw != "" {
		until, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			api.WriteError(w, http.StatusBadRequest, "until must be an RFC3339 timestamp")
			return
		}
		filter.Until = &until
	}
	if raw := strings.TrimSpace(query.Get("before")); raw != "" {
		before, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || before <= 0 {
			api.WriteError(w, http.StatusBadRequest, "before must be a positive integer cursor")
			return
		}
		filter.Before = before
	}
	if raw := strings.TrimSpace(query.Get("limit")); raw != "" {
		limit, err := strconv.Atoi(raw)
		if err != nil || limit <= 0 || limit > 500 {
			api.WriteError(w, http.StatusBadRequest, "limit must be an integer between 1 and 500")
			return
		}
		filter.Limit = limit
	}
	if filter.Since != nil && filter.Until != nil && !filter.Since.Before(*filter.Until) {
		api.WriteError(w, http.StatusBadRequest, "since must be before until")
		return
	}

	page, err := s.store.ListTimelineEvents(r.Context(), filter)
	if err != nil {
		s.logger.Error("failed to list timeline events", "cluster_id", clusterID, "error", err)
		api.WriteError(w, http.StatusInternalServerError, "failed to list timeline events")
		return
	}
	response := api.ListTimelineEventsResponse{Events: page.Events, NextCursor: page.NextCursor}
	if err := api.WriteJSON(w, http.StatusOK, response); err != nil {
		s.logger.Error("failed to write timeline response", "error", err)
	}
}

func (s *Server) handleRecordPolicyFinding(w http.ResponseWriter, r *http.Request) {
	cluster, approved, ok := s.authenticateAgentPath(w, r)
	if !ok {
		return
	}
	if !approved {
		api.WriteError(w, http.StatusForbidden, "agent is pending approval")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxPolicyFindingBodyBytes)
	var request api.RecordPolicyFindingRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			api.WriteError(w, http.StatusRequestEntityTooLarge, "request body is too large")
		} else {
			api.WriteError(w, http.StatusBadRequest, "invalid request body")
		}
		return
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		api.WriteError(w, http.StatusBadRequest, "request body must contain one JSON object")
		return
	}
	request.RuleID = strings.TrimSpace(request.RuleID)
	request.Resource = strings.TrimSpace(request.Resource)
	request.Severity = strings.TrimSpace(request.Severity)
	request.Message = strings.TrimSpace(request.Message)
	if request.RuleID == "" || request.Resource == "" || request.Message == "" {
		api.WriteError(w, http.StatusBadRequest, "ruleId, resource, and message are required")
		return
	}
	occurredAt := request.OccurredAt.UTC()
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}

	details := make(map[string]string, len(request.Details)+3)
	for key, value := range request.Details {
		details[key] = value
	}
	details["ruleId"] = request.RuleID
	details["resource"] = request.Resource
	if request.Severity != "" {
		details["severity"] = request.Severity
	}
	dedupePayload, _ := json.Marshal(struct {
		RuleID     string
		Resource   string
		Severity   string
		Message    string
		Details    map[string]string
		OccurredAt time.Time
	}{
		RuleID: request.RuleID, Resource: request.Resource,
		Severity: request.Severity, Message: request.Message,
		Details: details, OccurredAt: occurredAt,
	})

	inserted, err := s.store.AppendEvent(r.Context(), types.OperationalEvent{
		ClusterID:  cluster.ID,
		Kind:       types.EventPolicyFinding,
		Message:    request.Message,
		Details:    details,
		OccurredAt: occurredAt,
		DedupeKey:  fmt.Sprintf("%x", sha256.Sum256(dedupePayload)),
	})
	if err != nil {
		s.logger.Error("failed to record policy finding", "cluster_id", cluster.ID, "error", err)
		api.WriteError(w, http.StatusInternalServerError, "failed to record policy finding")
		return
	}
	status := http.StatusCreated
	if !inserted {
		status = http.StatusOK
	}
	w.WriteHeader(status)
}
