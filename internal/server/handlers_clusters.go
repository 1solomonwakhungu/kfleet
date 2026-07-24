package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/1solomonwakhungu/kfleet/internal/store"
	"github.com/1solomonwakhungu/kfleet/pkg/api"
	"github.com/1solomonwakhungu/kfleet/pkg/types"
	"github.com/google/uuid"
)

const placeholderClusterToken = "placeholder-token"

const maxSnapshotBodyBytes = 4 << 20

func (s *Server) registerClusterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/clusters", s.handleListClusters)
	mux.HandleFunc("GET /api/v1/clusters/{id}", s.handleGetCluster)
	mux.HandleFunc("POST /api/v1/clusters/register", s.handleRegisterCluster)
	mux.HandleFunc("DELETE /api/v1/clusters/{id}", s.handleDeleteCluster)
	mux.HandleFunc("GET /api/v1/clusters/{id}/status", s.handleClusterStatus)
	mux.HandleFunc("POST /api/v1/clusters/{id}/status", s.handleClusterSnapshot)
	mux.HandleFunc("GET /api/v1/clusters/{id}/pods", s.handleClusterPods)
	mux.HandleFunc("GET /api/v1/clusters/{id}/services", s.handleClusterServices)
	mux.HandleFunc("GET /api/v1/clusters/{id}/deployments", s.handleClusterDeployments)
	mux.HandleFunc("GET /api/v1/clusters/{id}/namespaces", s.handleClusterNamespaces)
	mux.HandleFunc("GET /api/v1/clusters/{id}/events", s.handleClusterEvents)
}

func (s *Server) handleListClusters(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := tenantIDFromRequest(w, r)
	if !ok {
		return
	}
	clusters, err := s.store.ListClustersForTenant(r.Context(), tenantID)
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
	tenantID, ok := tenantIDFromRequest(w, r)
	if !ok {
		return
	}
	cluster, err := s.store.GetClusterForTenant(r.Context(), tenantID, r.PathValue("id"))
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

	tenantID, ok := tenantIDFromRequest(w, r)
	if !ok {
		return
	}
	cluster := types.Cluster{
		ID:           uuid.NewString(),
		TenantID:     tenantID,
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
	tenantID, ok := tenantIDFromRequest(w, r)
	if !ok {
		return
	}
	cluster, err := s.store.GetClusterForTenant(r.Context(), tenantID, r.PathValue("id"))
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
	tenantID, ok := tenantIDFromRequest(w, r)
	if !ok {
		return
	}
	cluster, err := s.store.GetClusterForTenant(r.Context(), tenantID, r.PathValue("id"))
	if handleStoreError(w, err) {
		return
	}
	nodes, err := s.store.ListNodes(r.Context(), cluster.ID)
	if handleStoreError(w, err) {
		return
	}
	response := api.ClusterStatusResponse{
		Cluster: cluster,
		Nodes:   nodes,
	}
	if err := api.WriteJSON(w, http.StatusOK, response); err != nil {
		s.logger.Error("failed to write cluster status", "error", err)
	}
}

func (s *Server) handleClusterSnapshot(w http.ResponseWriter, r *http.Request) {
	cluster, err := s.clusterByIDOrName(r, r.PathValue("id"))
	if handleStoreError(w, err) {
		return
	}
	token, ok := bearerToken(r.Header.Get("Authorization"))
	if !ok {
		api.WriteError(w, http.StatusUnauthorized, "invalid agent token")
		return
	}
	approved, err := s.store.ValidateAgentToken(r.Context(), cluster.ID, hashToken(token))
	if err != nil {
		if !errors.Is(err, store.ErrNotFound) {
			s.logger.Error("failed to validate snapshot token", "cluster_id", cluster.ID, "error", err)
		}
		api.WriteError(w, http.StatusUnauthorized, "invalid agent token")
		return
	}
	if !approved {
		api.WriteError(w, http.StatusForbidden, "agent is pending approval")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxSnapshotBodyBytes)
	var request api.ClusterSnapshotRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		writeSnapshotDecodeError(w, err)
		return
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			api.WriteError(w, http.StatusBadRequest, "invalid request body")
		} else {
			writeSnapshotDecodeError(w, err)
		}
		return
	}

	snapshot := normalizeSnapshot(cluster.ID, request)
	health := snapshotHealth(snapshot.Nodes)
	now := time.Now().UTC()
	if err := s.store.ReplaceSnapshot(r.Context(), cluster.ID, snapshot, request.K8sVersion, request.AgentVersion, health, now); err != nil {
		s.logger.Error("failed to persist cluster snapshot", "cluster_id", cluster.ID, "error", err)
		if handleStoreError(w, err) {
			return
		}
	}
	updated, err := s.store.GetCluster(r.Context(), cluster.ID)
	if handleStoreError(w, err) {
		return
	}
	if cluster.Health != updated.Health {
		s.broadcast.Broadcast(ClusterUpdate{Type: "health_changed", Cluster: updated})
	}
	s.broadcast.Broadcast(ClusterUpdate{Type: "snapshot", Cluster: updated})
	if err := api.WriteJSON(w, http.StatusOK, updated); err != nil {
		s.logger.Error("failed to write snapshot response", "error", err)
	}
}

func (s *Server) handleClusterPods(w http.ResponseWriter, r *http.Request) {
	cluster, ok := s.resourceCluster(w, r)
	if !ok {
		return
	}
	pods, err := s.store.ListPods(r.Context(), cluster.ID, strings.TrimSpace(r.URL.Query().Get("namespace")))
	if handleStoreError(w, err) {
		return
	}
	if err := api.WriteJSON(w, http.StatusOK, pods); err != nil {
		s.logger.Error("failed to write cluster pods", "error", err)
	}
}

func (s *Server) handleClusterServices(w http.ResponseWriter, r *http.Request) {
	cluster, ok := s.resourceCluster(w, r)
	if !ok {
		return
	}
	services, err := s.store.ListServices(r.Context(), cluster.ID, strings.TrimSpace(r.URL.Query().Get("namespace")))
	if handleStoreError(w, err) {
		return
	}
	if err := api.WriteJSON(w, http.StatusOK, services); err != nil {
		s.logger.Error("failed to write cluster services", "error", err)
	}
}

func (s *Server) handleClusterDeployments(w http.ResponseWriter, r *http.Request) {
	cluster, ok := s.resourceCluster(w, r)
	if !ok {
		return
	}
	deployments, err := s.store.ListDeployments(r.Context(), cluster.ID, strings.TrimSpace(r.URL.Query().Get("namespace")))
	if handleStoreError(w, err) {
		return
	}
	if err := api.WriteJSON(w, http.StatusOK, deployments); err != nil {
		s.logger.Error("failed to write cluster deployments", "error", err)
	}
}

func (s *Server) handleClusterNamespaces(w http.ResponseWriter, r *http.Request) {
	cluster, ok := s.resourceCluster(w, r)
	if !ok {
		return
	}
	namespaces, err := s.store.ListNamespaces(r.Context(), cluster.ID)
	if handleStoreError(w, err) {
		return
	}
	if err := api.WriteJSON(w, http.StatusOK, namespaces); err != nil {
		s.logger.Error("failed to write cluster namespaces", "error", err)
	}
}

func (s *Server) handleClusterEvents(w http.ResponseWriter, r *http.Request) {
	cluster, ok := s.resourceCluster(w, r)
	if !ok {
		return
	}
	events, err := s.store.ListEvents(r.Context(), cluster.ID, strings.TrimSpace(r.URL.Query().Get("namespace")))
	if handleStoreError(w, err) {
		return
	}
	if err := api.WriteJSON(w, http.StatusOK, events); err != nil {
		s.logger.Error("failed to write cluster events", "error", err)
	}
}

func (s *Server) resourceCluster(w http.ResponseWriter, r *http.Request) (types.Cluster, bool) {
	tenantID, ok := tenantIDFromRequest(w, r)
	if !ok {
		return types.Cluster{}, false
	}
	cluster, err := s.store.GetClusterForTenant(r.Context(), tenantID, r.PathValue("id"))
	if handleStoreError(w, err) {
		return types.Cluster{}, false
	}
	return cluster, true
}

func writeSnapshotDecodeError(w http.ResponseWriter, err error) {
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		api.WriteError(w, http.StatusRequestEntityTooLarge, "request body too large")
		return
	}
	api.WriteError(w, http.StatusBadRequest, "invalid request body")
}

func normalizeSnapshot(clusterID string, request api.ClusterSnapshotRequest) types.ClusterSnapshot {
	snapshot := types.ClusterSnapshot{
		Nodes:       make([]types.Node, 0, len(request.Nodes)),
		Pods:        make([]types.Pod, 0, len(request.Pods)),
		Services:    make([]types.Service, 0, len(request.Services)),
		Deployments: make([]types.Deployment, 0, len(request.Deployments)),
		Namespaces:  make([]types.Namespace, 0, len(request.Namespaces)),
		Events:      make([]types.Event, 0, len(request.Events)),
	}
	for _, node := range request.Nodes {
		snapshot.Nodes = append(snapshot.Nodes, types.Node{
			Name:           node.Name,
			Status:         node.Status,
			Roles:          nonNilSlice(node.Roles),
			Version:        node.K8sVersion,
			CPUCapacity:    node.CPUCapacity,
			MemoryCapacity: node.MemoryCapacity,
			Ready:          strings.EqualFold(node.Status, "Ready"),
		})
	}
	for _, pod := range request.Pods {
		snapshot.Pods = append(snapshot.Pods, types.Pod{
			Name:                      pod.Name,
			Namespace:                 pod.Namespace,
			Phase:                     pod.Phase,
			NodeName:                  pod.Node,
			RestartCount:              pod.Restarts,
			Ready:                     pod.Ready,
			StartTime:                 pod.StartTime,
			SecurityContextKnown:      pod.SecurityContextKnown,
			Privileged:                pod.Privileged,
			RunAsNonRoot:              pod.RunAsNonRoot,
			ReadOnlyRootFilesystem:    pod.ReadOnlyRootFilesystem,
			AllowsPrivilegeEscalation: pod.AllowsPrivilegeEscalation,
			CapabilitiesDroppedAll:    pod.CapabilitiesDroppedAll,
			HostNetwork:               pod.HostNetwork,
			HostPID:                   pod.HostPID,
			HostIPC:                   pod.HostIPC,
		})
	}
	for _, service := range request.Services {
		ports := make([]types.ServicePort, 0, len(service.Ports))
		for _, port := range service.Ports {
			ports = append(ports, parseServicePort(port))
		}
		snapshot.Services = append(snapshot.Services, types.Service{
			Name:        service.Name,
			Namespace:   service.Namespace,
			Type:        service.Type,
			ClusterIP:   service.ClusterIP,
			ExternalIPs: nonNilSlice(service.ExternalIPs),
			Ports:       ports,
			Age:         service.Age,
		})
	}
	for _, deployment := range request.Deployments {
		snapshot.Deployments = append(snapshot.Deployments, types.Deployment{
			Name:              deployment.Name,
			Namespace:         deployment.Namespace,
			ReadyReplicas:     deployment.ReadyReplicas,
			DesiredReplicas:   deployment.DesiredReplicas,
			AvailableReplicas: deployment.AvailableReplicas,
			UpdatedReplicas:   deployment.UpdatedReplicas,
			Age:               deployment.Age,
			ConfigHash:        deployment.ConfigHash,
			Images:            nonNilSlice(deployment.Images),
		})
	}
	for _, namespace := range request.Namespaces {
		snapshot.Namespaces = append(snapshot.Namespaces, types.Namespace{
			Name:   namespace.Name,
			Labels: nonNilMap(namespace.Labels),
		})
	}
	for _, event := range request.Events {
		snapshot.Events = append(snapshot.Events, types.Event{
			ClusterID:     clusterID,
			Namespace:     event.Namespace,
			Reason:        event.Reason,
			Message:       event.Message,
			Type:          event.Type,
			Count:         event.Count,
			LastTimestamp: event.LastTimestamp,
		})
	}
	return snapshot
}

func snapshotHealth(nodes []types.Node) types.ClusterHealth {
	if len(nodes) == 0 {
		return types.HealthDegraded
	}
	for _, node := range nodes {
		if !node.Ready {
			return types.HealthDegraded
		}
	}
	return types.HealthHealthy
}

func parseServicePort(value string) types.ServicePort {
	var result types.ServicePort
	portAndProtocol := value
	if before, after, ok := strings.Cut(value, "/"); ok {
		portAndProtocol = before
		result.Protocol = after
	}
	portValue := portAndProtocol
	if before, after, ok := strings.Cut(portAndProtocol, ":"); ok {
		result.Name = before
		portValue = after
	}
	if port, err := strconv.ParseInt(portValue, 10, 32); err == nil {
		result.Port = int32(port)
		// The current collector wire format omits targetPort. Using the service
		// port is the least surprising stable representation until it reports it.
		result.TargetPort = int32(port)
	}
	return result
}

func nonNilSlice[T any](values []T) []T {
	if values == nil {
		return make([]T, 0)
	}
	return values
}

func nonNilMap(values map[string]string) map[string]string {
	if values == nil {
		return make(map[string]string)
	}
	return values
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
