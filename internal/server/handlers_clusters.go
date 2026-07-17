package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/1solomonwakhungu/kfleet/internal/store"
	"github.com/1solomonwakhungu/kfleet/pkg/api"
	"github.com/1solomonwakhungu/kfleet/pkg/types"
	"github.com/google/uuid"
)

const placeholderClusterToken = "placeholder-token"

func (s *Server) registerClusterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/clusters", s.handleListClusters)
	mux.HandleFunc("GET /api/v1/clusters/{id}", s.handleGetCluster)
	mux.HandleFunc("POST /api/v1/clusters/register", s.handleRegisterCluster)
	mux.HandleFunc("DELETE /api/v1/clusters/{id}", s.handleDeleteCluster)
	mux.HandleFunc("GET /api/v1/clusters/{id}/status", s.handleClusterStatus)
	mux.HandleFunc("GET /api/v1/clusters/{id}/events", s.handleClusterEvents)
}

func (s *Server) handleListClusters(w http.ResponseWriter, r *http.Request) {
	clusters, err := s.store.ListClusters(r.Context())
	if err != nil {
		s.logger.Error("failed to list clusters", "error", err)
		api.WriteError(w, http.StatusInternalServerError, "failed to list clusters")
		return
	}
	if err := api.WriteJSON(w, http.StatusOK, api.ListClustersResponse{Clusters: clusters}); err != nil {
		s.logger.Error("failed to write cluster list", "error", err)
	}
}

func (s *Server) handleGetCluster(w http.ResponseWriter, r *http.Request) {
	cluster, err := s.store.GetCluster(r.Context(), r.PathValue("id"))
	if handleStoreError(w, err) {
		return
	}
	if err := api.WriteJSON(w, http.StatusOK, cluster); err != nil {
		s.logger.Error("failed to write cluster", "error", err)
	}
}

func (s *Server) handleRegisterCluster(w http.ResponseWriter, r *http.Request) {
	var request api.RegisterClusterRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	request.Name = strings.TrimSpace(request.Name)
	if request.Name == "" {
		api.WriteError(w, http.StatusBadRequest, "cluster name is required")
		return
	}

	cluster := types.Cluster{
		ID:           uuid.NewString(),
		Name:         request.Name,
		Health:       types.HealthUnknown,
		RegisteredAt: time.Now().UTC(),
		Labels:       request.Labels,
	}
	if err := s.store.CreateCluster(r.Context(), cluster); err != nil {
		s.logger.Error("failed to register cluster", "error", err)
		api.WriteError(w, http.StatusInternalServerError, "failed to register cluster")
		return
	}
	s.broadcast.Broadcast(ClusterUpdate{Type: "registered", Cluster: cluster})

	response := api.RegisterClusterResponse{
		ClusterID: cluster.ID,
		Token:     placeholderClusterToken,
	}
	if err := api.WriteJSON(w, http.StatusCreated, response); err != nil {
		s.logger.Error("failed to write registration response", "error", err)
	}
}

func (s *Server) handleDeleteCluster(w http.ResponseWriter, r *http.Request) {
	cluster, err := s.store.GetCluster(r.Context(), r.PathValue("id"))
	if handleStoreError(w, err) {
		return
	}
	err = s.store.DeleteCluster(r.Context(), cluster.ID)
	if handleStoreError(w, err) {
		return
	}
	s.broadcast.Broadcast(ClusterUpdate{Type: "deleted", Cluster: cluster})
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleClusterStatus(w http.ResponseWriter, r *http.Request) {
	cluster, err := s.store.GetCluster(r.Context(), r.PathValue("id"))
	if handleStoreError(w, err) {
		return
	}
	response := api.ClusterStatusResponse{
		Cluster: cluster,
		Nodes:   make([]types.Node, 0),
	}
	if err := api.WriteJSON(w, http.StatusOK, response); err != nil {
		s.logger.Error("failed to write cluster status", "error", err)
	}
}

func (s *Server) handleClusterEvents(w http.ResponseWriter, r *http.Request) {
	if _, err := s.store.GetCluster(r.Context(), r.PathValue("id")); handleStoreError(w, err) {
		return
	}
	if err := api.WriteJSON(w, http.StatusOK, make([]types.Event, 0)); err != nil {
		s.logger.Error("failed to write cluster events", "error", err)
	}
}

func handleStoreError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, store.ErrNotFound) {
		api.WriteError(w, http.StatusNotFound, "cluster not found")
		return true
	}
	api.WriteError(w, http.StatusInternalServerError, "internal server error")
	return true
}
