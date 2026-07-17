package server

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/1solomonwakhungu/kfleet/internal/store"
	"github.com/1solomonwakhungu/kfleet/pkg/api"
	"github.com/1solomonwakhungu/kfleet/pkg/types"
	"github.com/google/uuid"
)

func (s *Server) registerAgentRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/agents/register", s.handleAgentRegister)
	mux.HandleFunc("POST /api/v1/agents/{id}/approve", s.handleAgentApprove)
	mux.HandleFunc("GET /api/v1/agents/pending", s.handleListPendingAgents)
	mux.HandleFunc("POST /api/v1/agents/heartbeat", s.handleHeartbeat)
}

func generateToken() (raw string, hash string) {
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return "", ""
	}
	raw = hex.EncodeToString(token)
	digest := sha256.Sum256([]byte(raw))
	return raw, hex.EncodeToString(digest[:])
}

func hashToken(raw string) string {
	digest := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(digest[:])
}

func (s *Server) handleAgentRegister(w http.ResponseWriter, r *http.Request) {
	var request api.RegisterClusterRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	name := strings.TrimSpace(request.Name)
	if name == "" {
		name = strings.TrimSpace(request.ClusterName)
	}
	if name == "" {
		api.WriteError(w, http.StatusBadRequest, "cluster name is required")
		return
	}

	cluster, err := s.findClusterByName(r, name)
	if errors.Is(err, store.ErrNotFound) {
		cluster = types.Cluster{
			ID:           uuid.NewString(),
			Name:         name,
			Health:       types.HealthUnknown,
			RegisteredAt: time.Now().UTC(),
			Labels:       request.Labels,
		}
		if err := s.store.CreateCluster(r.Context(), cluster); err != nil {
			s.logger.Error("failed to create agent cluster", "error", err)
			api.WriteError(w, http.StatusInternalServerError, "failed to register agent")
			return
		}
	} else if err != nil {
		s.logger.Error("failed to find agent cluster", "error", err)
		api.WriteError(w, http.StatusInternalServerError, "failed to register agent")
		return
	}

	rawToken, tokenHash := generateToken()
	if rawToken == "" {
		s.logger.Error("failed to generate agent token")
		api.WriteError(w, http.StatusInternalServerError, "failed to register agent")
		return
	}
	if err := s.store.IssueAgentToken(r.Context(), cluster.ID, tokenHash); err != nil {
		s.logger.Error("failed to issue agent token", "cluster_id", cluster.ID, "error", err)
		api.WriteError(w, http.StatusInternalServerError, "failed to register agent")
		return
	}

	s.broadcast.Broadcast(ClusterUpdate{Type: "registered", Cluster: cluster})
	response := api.RegisterClusterResponse{ClusterID: cluster.ID, Token: rawToken}
	if err := api.WriteJSON(w, http.StatusCreated, response); err != nil {
		s.logger.Error("failed to write agent registration response", "error", err)
	}
}

func (s *Server) findClusterByName(r *http.Request, name string) (types.Cluster, error) {
	clusters, err := s.store.ListClusters(r.Context())
	if err != nil {
		return types.Cluster{}, err
	}
	for _, cluster := range clusters {
		if cluster.Name == name {
			return cluster, nil
		}
	}
	return types.Cluster{}, store.ErrNotFound
}

func (s *Server) handleAgentApprove(w http.ResponseWriter, r *http.Request) {
	clusterID := r.PathValue("id")
	if err := s.store.ApproveAgent(r.Context(), clusterID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			api.WriteError(w, http.StatusNotFound, "agent not found")
			return
		}
		s.logger.Error("failed to approve agent", "cluster_id", clusterID, "error", err)
		api.WriteError(w, http.StatusInternalServerError, "failed to approve agent")
		return
	}
	cluster, err := s.store.GetCluster(r.Context(), clusterID)
	if err != nil {
		s.logger.Error("failed to get approved agent", "cluster_id", clusterID, "error", err)
		api.WriteError(w, http.StatusInternalServerError, "failed to approve agent")
		return
	}
	if err := api.WriteJSON(w, http.StatusOK, cluster); err != nil {
		s.logger.Error("failed to write approved agent", "error", err)
	}
}

func (s *Server) handleListPendingAgents(w http.ResponseWriter, r *http.Request) {
	clusters, err := s.store.ListPendingAgents(r.Context())
	if err != nil {
		s.logger.Error("failed to list pending agents", "error", err)
		api.WriteError(w, http.StatusInternalServerError, "failed to list pending agents")
		return
	}
	if err := api.WriteJSON(w, http.StatusOK, api.ListClustersResponse{Clusters: clusters}); err != nil {
		s.logger.Error("failed to write pending agents", "error", err)
	}
}

func (s *Server) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	var request api.HeartbeatRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	request.ClusterID = strings.TrimSpace(request.ClusterID)
	if request.ClusterID == "" {
		api.WriteError(w, http.StatusBadRequest, "clusterId is required")
		return
	}

	token, ok := bearerToken(r.Header.Get("Authorization"))
	if !ok {
		api.WriteError(w, http.StatusUnauthorized, "invalid agent token")
		return
	}
	approved, err := s.store.ValidateAgentToken(r.Context(), request.ClusterID, hashToken(token))
	if err != nil {
		if !errors.Is(err, store.ErrNotFound) {
			s.logger.Error("failed to validate agent token", "cluster_id", request.ClusterID, "error", err)
		}
		api.WriteError(w, http.StatusUnauthorized, "invalid agent token")
		return
	}
	if !approved {
		api.WriteError(w, http.StatusForbidden, "agent is pending approval")
		return
	}

	health := types.HealthDegraded
	if request.NodeCount > 0 && request.HealthyNodes == request.NodeCount {
		health = types.HealthHealthy
	}
	if err := s.store.UpdateSnapshot(r.Context(), request.ClusterID, request.NodeCount, request.PodCount, request.Version); err != nil {
		handleHeartbeatStoreError(w, err)
		return
	}
	if err := s.store.UpdateHealth(r.Context(), request.ClusterID, health, time.Now().UTC()); err != nil {
		handleHeartbeatStoreError(w, err)
		return
	}
	cluster, err := s.store.GetCluster(r.Context(), request.ClusterID)
	if err != nil {
		handleHeartbeatStoreError(w, err)
		return
	}
	s.broadcast.Broadcast(ClusterUpdate{Type: "health_changed", Cluster: cluster})
	s.broadcast.Broadcast(ClusterUpdate{Type: "snapshot", Cluster: cluster})
	if err := api.WriteJSON(w, http.StatusOK, cluster); err != nil {
		s.logger.Error("failed to write heartbeat response", "error", err)
	}
}

func bearerToken(authorization string) (string, bool) {
	scheme, token, ok := strings.Cut(strings.TrimSpace(authorization), " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") || strings.TrimSpace(token) == "" {
		return "", false
	}
	return strings.TrimSpace(token), true
}

func handleHeartbeatStoreError(w http.ResponseWriter, err error) {
	if errors.Is(err, store.ErrNotFound) {
		api.WriteError(w, http.StatusUnauthorized, "invalid agent token")
		return
	}
	api.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("failed to update heartbeat: %v", err))
}
