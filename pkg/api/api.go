// Package api defines the HTTP contracts shared by kfleet services.
package api

import (
	"encoding/json"
	"net/http"

	"github.com/1solomonwakhungu/kfleet/pkg/types"
)

// APIVersion is the current kfleet API version.
const APIVersion = "v1"

// RegisterClusterRequest is sent when an agent registers a cluster.
type RegisterClusterRequest struct {
	Name         string            `json:"name"`
	ClusterName  string            `json:"clusterName,omitempty"`
	Labels       map[string]string `json:"labels"`
	AgentVersion string            `json:"agentVersion,omitempty"`
	K8sVersion   string            `json:"k8sVersion,omitempty"`
}

// RegisterClusterResponse contains the registered cluster identity and token.
type RegisterClusterResponse struct {
	ClusterID string `json:"clusterId"`
	Token     string `json:"token"`
}

// ClusterStatusResponse contains cluster metadata and its nodes.
type ClusterStatusResponse struct {
	Cluster types.Cluster `json:"cluster"`
	Nodes   []types.Node  `json:"nodes"`
}

// ListClustersResponse contains all registered clusters.
type ListClustersResponse struct {
	Clusters []types.Cluster `json:"clusters"`
}

// HeartbeatRequest reports a cluster's current state.
type HeartbeatRequest struct {
	ClusterID    string `json:"clusterId"`
	NodeCount    int    `json:"nodeCount"`
	PodCount     int    `json:"podCount"`
	HealthyNodes int    `json:"healthyNodes"`
	Version      string `json:"version"`
}

// ErrorResponse is returned for API errors.
type ErrorResponse struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

// WriteJSON writes v as a JSON response with the supplied HTTP status.
func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

// WriteError writes a JSON error response. Encoding errors cannot be usefully
// reported after the HTTP status has been written, so they are intentionally ignored.
func WriteError(w http.ResponseWriter, status int, msg string) {
	_ = WriteJSON(w, status, ErrorResponse{Error: msg, Code: status})
}
