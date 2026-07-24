package server

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
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

// registerAgentRoutes wires the agent-facing endpoints. These continue to
// authenticate with the shared registration token and per-agent bearer
// tokens established during registration; they intentionally do not use
// user session cookies, since agents are not hub users. Endpoints that
// expose or approve pending agents to a human operator require an
// authenticated hub session with sufficient role.
func (s *Server) registerAgentRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/agents/register", s.handleAgentRegister)
	mux.HandleFunc("POST /api/v1/agents/{id}/approve", s.requireRole(types.RoleOperator, s.handleAgentApprove))
	mux.HandleFunc("GET /api/v1/agents/pending", s.requireAuth(s.handleListPendingAgents))
	mux.HandleFunc("POST /api/v1/agents/heartbeat", s.handleHeartbeat)
	mux.HandleFunc("POST /api/v1/agents/{id}/heartbeat", s.handleAgentLiveness)
	mux.HandleFunc("POST /api/v1/agents/{id}/deregister", s.handleAgentDeregister)
}

func (s *Server) handleAgentLiveness(w http.ResponseWriter, r *http.Request) {
	cluster, approved, ok := s.authenticateAgentPath(w, r)
	if !ok {
		return
	}
	if err := s.store.UpdateHealth(r.Context(), cluster.ID, cluster.Health, time.Now().UTC()); err != nil {
		handleHeartbeatStoreError(w, err)
		return
	}
	if err := api.WriteJSON(w, http.StatusOK, api.AgentRegistrationStatus{
		ClusterID: cluster.ID,
		Approved:  approved,
	}); err != nil {
		s.logger.Error("failed to write agent heartbeat response", "error", err)
	}
}

func (s *Server) handleAgentDeregister(w http.ResponseWriter, r *http.Request) {
	cluster, _, ok := s.authenticateAgentPath(w, r)
	if !ok {
		return
	}
	now := time.Now().UTC()
	if err := s.store.UpdateHealth(r.Context(), cluster.ID, types.HealthUnreachable, now); err != nil {
		handleHeartbeatStoreError(w, err)
		return
	}
	previousHealth := cluster.Health
	cluster.Health = types.HealthUnreachable
	cluster.LastHeartbeat = now
	s.broadcast.Broadcast(ClusterUpdate{Type: "health_changed", Cluster: cluster})
	if previousHealth != types.HealthUnreachable {
		s.recordAgentDisconnected(r.Context(), cluster, "deregistered", now)
		s.recordHeartbeatTransition(r.Context(), cluster, previousHealth, types.HealthUnreachable, now)
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) authenticateAgentPath(w http.ResponseWriter, r *http.Request) (types.Cluster, bool, bool) {
	cluster, err := s.clusterByIDOrName(r, r.PathValue("id"))
	if err != nil {
		api.WriteError(w, http.StatusUnauthorized, "invalid agent token")
		return types.Cluster{}, false, false
	}
	token, ok := bearerToken(r.Header.Get("Authorization"))
	if !ok {
		api.WriteError(w, http.StatusUnauthorized, "invalid agent token")
		return types.Cluster{}, false, false
	}
	approved, err := s.store.ValidateAgentToken(r.Context(), cluster.ID, hashToken(token))
	if err != nil {
		api.WriteError(w, http.StatusUnauthorized, "invalid agent token")
		return types.Cluster{}, false, false
	}
	return cluster, approved, true
}

func (s *Server) clusterByIDOrName(r *http.Request, idOrName string) (types.Cluster, error) {
	cluster, err := s.store.GetCluster(r.Context(), idOrName)
	if err == nil {
		return cluster, nil
	}
	if !errors.Is(err, store.ErrNotFound) {
		return types.Cluster{}, err
	}
	return s.findClusterByName(r, idOrName)
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
	if !s.validRegistrationToken(r.Context(), r.Header.Get("Authorization")) {
		api.WriteError(w, http.StatusUnauthorized, "invalid registration token")
		return
	}
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
	tenantID, ok := tenantIDFromRequest(w, r)
	if !ok {
		return
	}

	cluster, err := s.findClusterByName(r, name)
	newlyRegistered := false
	if errors.Is(err, store.ErrNotFound) {
		cluster = types.Cluster{
			ID:           uuid.NewString(),
			TenantID:     tenantID,
			Name:         name,
			Health:       types.HealthUnknown,
			Version:      request.K8sVersion,
			AgentVersion: request.AgentVersion,
			RegisteredAt: time.Now().UTC(),
			Labels:       request.Labels,
		}
		if err := s.store.CreateCluster(r.Context(), cluster); err != nil {
			s.logger.Error("failed to create agent cluster", "error", err)
			api.WriteError(w, http.StatusInternalServerError, "failed to register agent")
			return
		}
		newlyRegistered = true
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
	if newlyRegistered {
		s.recordClusterRegistered(r.Context(), cluster)
	}
	response := api.RegisterClusterResponse{ClusterID: cluster.ID, Token: rawToken}
	approved, err := s.store.ValidateAgentToken(r.Context(), cluster.ID, tokenHash)
	if err != nil {
		s.logger.Error("failed to read agent approval", "cluster_id", cluster.ID, "error", err)
		api.WriteError(w, http.StatusInternalServerError, "failed to register agent")
		return
	}
	status := http.StatusCreated
	if approved {
		status = http.StatusOK
	}
	if err := api.WriteJSON(w, status, response); err != nil {
		s.logger.Error("failed to write agent registration response", "error", err)
	}
}

// validRegistrationToken checks the bearer token against a rotated
// registration token stored in the settings table, if one has ever been
// issued via handleRotateRegistrationToken; otherwise it falls back to the
// static KFLEET_REGISTRATION_TOKEN configured at startup. This preserves
// the original env-var-based registration flow for installations that
// never rotate the token.
func (s *Server) validRegistrationToken(ctx context.Context, authorization string) bool {
	rotatedHash, ok, err := s.store.GetSetting(ctx, settingRegistrationTokenHash)
	if err != nil {
		s.logger.Error("failed to read rotated registration token setting", "error", err)
		return false
	}
	if ok {
		token, tokenOK := bearerToken(authorization)
		if !tokenOK {
			return false
		}
		return subtle.ConstantTimeCompare([]byte(hashToken(token)), []byte(rotatedHash)) == 1
	}

	if s.cfg.RegistrationToken == "" {
		return true
	}
	token, ok := bearerToken(authorization)
	if !ok || len(token) != len(s.cfg.RegistrationToken) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(token), []byte(s.cfg.RegistrationToken)) == 1
}

func (s *Server) findClusterByName(r *http.Request, name string) (types.Cluster, error) {
	tenantID := store.DefaultTenantID
	if values := r.Header.Values(tenantHeader); len(values) == 1 && tenantIDPattern.MatchString(strings.TrimSpace(values[0])) {
		tenantID = strings.TrimSpace(values[0])
	}
	clusters, err := s.store.ListClustersForTenant(r.Context(), tenantID)
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
	actor, _ := authenticatedUser(r.Context())
	clusterID := r.PathValue("id")
	tenantID, ok := tenantIDFromRequest(w, r)
	if !ok {
		return
	}
	_, err := s.store.GetClusterForTenant(r.Context(), tenantID, clusterID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			api.WriteError(w, http.StatusNotFound, "agent not found")
			return
		}
		api.WriteError(w, http.StatusInternalServerError, "failed to approve agent")
		return
	}
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
	s.recordAgentApproved(r.Context(), cluster, time.Now().UTC())
	s.recordAudit(r.Context(), r, auditActorFromUser(actor), "agent.approve", "cluster", cluster.ID, types.AuditSuccess, "name="+cluster.Name)
	if err := api.WriteJSON(w, http.StatusOK, cluster); err != nil {
		s.logger.Error("failed to write approved agent", "error", err)
	}
}

func (s *Server) handleListPendingAgents(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := tenantIDFromRequest(w, r)
	if !ok {
		return
	}
	clusters, err := s.store.ListPendingAgentsForTenant(r.Context(), tenantID)
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

	previous, err := s.store.GetCluster(r.Context(), request.ClusterID)
	if err != nil {
		handleHeartbeatStoreError(w, err)
		return
	}

	health := types.HealthDegraded
	if request.NodeCount > 0 && request.HealthyNodes == request.NodeCount {
		health = types.HealthHealthy
	}
	now := time.Now().UTC()
	if err := s.store.UpdateSnapshot(r.Context(), request.ClusterID, request.NodeCount, request.PodCount, request.Version); err != nil {
		handleHeartbeatStoreError(w, err)
		return
	}
	if err := s.store.UpdateHealth(r.Context(), request.ClusterID, health, now); err != nil {
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
	s.recordHeartbeatTransition(r.Context(), cluster, previous.Health, cluster.Health, now)
	s.recordVersionChanged(r.Context(), cluster, previous.Version, cluster.Version, now)
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
