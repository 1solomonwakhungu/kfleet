package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/1solomonwakhungu/kfleet/internal/store"
	"github.com/1solomonwakhungu/kfleet/pkg/api"
	"github.com/1solomonwakhungu/kfleet/pkg/types"
)

const maxAlertRequestBodyBytes = 64 << 10

type alertListResponse struct {
	Alerts []types.Alert `json:"alerts"`
}

type alertRuleListResponse struct {
	Rules []types.AlertRule `json:"rules"`
}

type acknowledgeAlertRequest struct {
	AcknowledgedBy string `json:"acknowledgedBy"`
}

func (s *Server) registerAlertRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/alerts", s.requireAuth(s.handleListAlerts))
	mux.HandleFunc("POST /api/v1/alerts/{id}/acknowledge", s.requireRole(types.RoleOperator, s.handleAcknowledgeAlert))
	mux.HandleFunc("GET /api/v1/alert-rules", s.requireAuth(s.handleListAlertRules))
	mux.HandleFunc("PUT /api/v1/alert-rules/{id}", s.requireRole(types.RoleAdmin, s.handlePutAlertRule))
}

func (s *Server) handleListAlerts(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := tenantIDFromRequest(w, r)
	if !ok {
		return
	}
	status := types.AlertStatus(strings.TrimSpace(r.URL.Query().Get("status")))
	if status != "" && !validAlertStatus(status) {
		api.WriteError(w, http.StatusBadRequest, "status must be firing, acknowledged, or resolved")
		return
	}
	limit := 100
	if value := r.URL.Query().Get("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed <= 0 || parsed > 500 {
			api.WriteError(w, http.StatusBadRequest, "limit must be between 1 and 500")
			return
		}
		limit = parsed
	}
	alerts, err := s.store.ListAlertsForTenant(r.Context(), tenantID, status, limit)
	if err != nil {
		s.logger.Error("failed to list alerts", "error", err)
		api.WriteError(w, http.StatusInternalServerError, "failed to list alerts")
		return
	}
	if err := api.WriteJSON(w, http.StatusOK, alertListResponse{Alerts: alerts}); err != nil {
		s.logger.Error("failed to write alert list", "error", err)
	}
}

func (s *Server) handleAcknowledgeAlert(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := tenantIDFromRequest(w, r)
	if !ok {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxAlertRequestBodyBytes)
	request := acknowledgeAlertRequest{AcknowledgedBy: "operator"}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	actor, _ := authenticatedUser(r.Context())

	alertID := r.PathValue("id")
	if err := s.store.AcknowledgeAlertForTenant(r.Context(), tenantID, alertID, actor.Username, time.Now().UTC()); err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			api.WriteError(w, http.StatusNotFound, "alert not found")
		case errors.Is(err, store.ErrInvalidState):
			api.WriteError(w, http.StatusConflict, "only firing alerts can be acknowledged")
		default:
			s.logger.Error("failed to acknowledge alert", "alert_id", alertID, "error", err)
			api.WriteError(w, http.StatusInternalServerError, "failed to acknowledge alert")
		}
		return
	}
	s.recordAudit(
		r.Context(),
		r,
		auditActorFromUser(actor),
		"alert.acknowledge",
		"alert",
		alertID,
		types.AuditSuccess,
		"tenant_id="+tenantID,
	)
	alert, err := s.store.GetAlertForTenant(r.Context(), tenantID, alertID)
	if err != nil {
		s.logger.Error("failed to read acknowledged alert", "alert_id", alertID, "error", err)
		api.WriteError(w, http.StatusInternalServerError, "failed to acknowledge alert")
		return
	}
	if err := api.WriteJSON(w, http.StatusOK, alert); err != nil {
		s.logger.Error("failed to write acknowledged alert", "alert_id", alertID, "error", err)
	}
}

func (s *Server) handleListAlertRules(w http.ResponseWriter, r *http.Request) {
	rules, err := s.store.ListAlertRules(r.Context())
	if err != nil {
		s.logger.Error("failed to list alert rules", "error", err)
		api.WriteError(w, http.StatusInternalServerError, "failed to list alert rules")
		return
	}
	if err := api.WriteJSON(w, http.StatusOK, alertRuleListResponse{Rules: rules}); err != nil {
		s.logger.Error("failed to write alert rules", "error", err)
	}
}

func (s *Server) handlePutAlertRule(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxAlertRequestBodyBytes)
	var rule types.AlertRule
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&rule); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	rule.ID = strings.TrimSpace(r.PathValue("id"))
	rule.Name = strings.TrimSpace(rule.Name)
	if rule.ID == "" || len(rule.ID) > 128 || rule.Name == "" || len(rule.Name) > 200 {
		api.WriteError(w, http.StatusBadRequest, "rule id and name are required and must be within length limits")
		return
	}
	if rule.Health != types.HealthDegraded && rule.Health != types.HealthUnreachable {
		api.WriteError(w, http.StatusBadRequest, "health must be degraded or unreachable")
		return
	}
	if rule.Severity != types.AlertSeverityWarning && rule.Severity != types.AlertSeverityCritical {
		api.WriteError(w, http.StatusBadRequest, "severity must be warning or critical")
		return
	}
	if rule.CooldownSeconds < 0 || rule.CooldownSeconds > int64((30*24*time.Hour)/time.Second) {
		api.WriteError(w, http.StatusBadRequest, "cooldownSeconds must be between 0 and 2592000")
		return
	}
	now := time.Now().UTC()
	rule.CreatedAt = now
	rule.UpdatedAt = now
	if err := s.store.UpsertAlertRule(r.Context(), rule); err != nil {
		s.logger.Error("failed to save alert rule", "rule_id", rule.ID, "error", err)
		api.WriteError(w, http.StatusInternalServerError, "failed to save alert rule")
		return
	}
	actor, _ := authenticatedUser(r.Context())
	s.recordAudit(
		r.Context(),
		r,
		auditActorFromUser(actor),
		"alert_rule.update",
		"alert_rule",
		rule.ID,
		types.AuditSuccess,
		"",
	)
	rules, err := s.store.ListAlertRules(r.Context())
	if err != nil {
		s.logger.Error("failed to read saved alert rule", "rule_id", rule.ID, "error", err)
		api.WriteError(w, http.StatusInternalServerError, "failed to save alert rule")
		return
	}
	for _, saved := range rules {
		if saved.ID == rule.ID {
			if err := api.WriteJSON(w, http.StatusOK, saved); err != nil {
				s.logger.Error("failed to write saved alert rule", "rule_id", rule.ID, "error", err)
			}
			return
		}
	}
	api.WriteError(w, http.StatusInternalServerError, "failed to save alert rule")
}

func validAlertStatus(status types.AlertStatus) bool {
	return status == types.AlertStatusFiring ||
		status == types.AlertStatusAcknowledged ||
		status == types.AlertStatusResolved
}
