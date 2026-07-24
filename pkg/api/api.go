// Package api defines the HTTP contracts shared by kfleet services.
package api

import (
	"encoding/json"
	"net/http"
	"time"

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

// AgentRegistrationStatus describes whether a registered agent is approved.
type AgentRegistrationStatus struct {
	ClusterID string `json:"clusterId"`
	Approved  bool   `json:"approved"`
}

// ClusterStatusResponse contains cluster metadata and its nodes.
type ClusterStatusResponse struct {
	Cluster types.Cluster `json:"cluster"`
	Nodes   []types.Node  `json:"nodes"`
}

// ClusterSnapshotRequest is the collector wire format accepted by the hub.
// It deliberately lives outside internal/agent so the hub does not depend on
// agent implementation packages.
type ClusterSnapshotRequest struct {
	Nodes        []SnapshotNode       `json:"nodes"`
	Pods         []SnapshotPod        `json:"pods"`
	Services     []SnapshotService    `json:"services"`
	Deployments  []SnapshotDeployment `json:"deployments"`
	Namespaces   []SnapshotNamespace  `json:"namespaces"`
	Events       []SnapshotEvent      `json:"events"`
	NodeCount    int                  `json:"nodeCount"`
	PodCount     int                  `json:"podCount"`
	K8sVersion   string               `json:"k8sVersion"`
	CollectedAt  time.Time            `json:"collectedAt"`
	AgentVersion string               `json:"agentVersion"`
}

// SnapshotNode is the node shape emitted by the agent collector.
type SnapshotNode struct {
	Name           string   `json:"name"`
	Status         string   `json:"status"`
	K8sVersion     string   `json:"k8sVersion"`
	Roles          []string `json:"roles"`
	CPUCapacity    string   `json:"cpuCapacity"`
	MemoryCapacity string   `json:"memoryCapacity"`
}

// SnapshotPod is the pod shape emitted by the agent collector.
type SnapshotPod struct {
	Namespace                 string    `json:"namespace"`
	Name                      string    `json:"name"`
	Phase                     string    `json:"phase"`
	Restarts                  int32     `json:"restarts"`
	Node                      string    `json:"node"`
	Ready                     bool      `json:"ready"`
	StartTime                 time.Time `json:"startTime"`
	SecurityContextKnown      bool      `json:"securityContextKnown"`
	Privileged                bool      `json:"privileged"`
	RunAsNonRoot              bool      `json:"runAsNonRoot"`
	ReadOnlyRootFilesystem    bool      `json:"readOnlyRootFilesystem"`
	AllowsPrivilegeEscalation bool      `json:"allowsPrivilegeEscalation"`
	CapabilitiesDroppedAll    bool      `json:"capabilitiesDroppedAll"`
	HostNetwork               bool      `json:"hostNetwork"`
	HostPID                   bool      `json:"hostPID"`
	HostIPC                   bool      `json:"hostIPC"`
}

// SnapshotService is the service shape emitted by the agent collector.
type SnapshotService struct {
	Namespace   string   `json:"namespace"`
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	ClusterIP   string   `json:"clusterIP"`
	ExternalIPs []string `json:"externalIPs"`
	Ports       []string `json:"ports"`
	Age         string   `json:"age"`
}

// SnapshotDeployment is the deployment shape emitted by the agent collector.
type SnapshotDeployment struct {
	Namespace         string   `json:"namespace"`
	Name              string   `json:"name"`
	DesiredReplicas   int32    `json:"desiredReplicas"`
	ReadyReplicas     int32    `json:"readyReplicas"`
	AvailableReplicas int32    `json:"availableReplicas"`
	UpdatedReplicas   int32    `json:"updatedReplicas"`
	Age               string   `json:"age"`
	ConfigHash        string   `json:"configHash"`
	Images            []string `json:"images"`
}

// SnapshotNamespace is namespace configuration emitted by the agent.
type SnapshotNamespace struct {
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels"`
}

// PolicyListResponse contains the immutable policy catalog.
type PolicyListResponse struct {
	Policies []types.Policy `json:"policies"`
}

// PolicyResultsResponse contains a tenant-scoped evaluation.
type PolicyResultsResponse struct {
	Results []types.PolicyResult `json:"results"`
	Summary types.PolicySummary  `json:"summary"`
}

// SnapshotEvent is the event shape emitted by the agent collector.
type SnapshotEvent struct {
	Namespace      string    `json:"namespace"`
	Name           string    `json:"name"`
	Type           string    `json:"type"`
	Reason         string    `json:"reason"`
	Message        string    `json:"message"`
	InvolvedObject string    `json:"involvedObject"`
	Count          int32     `json:"count"`
	LastTimestamp  time.Time `json:"lastTimestamp"`
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

// LoginRequest authenticates a user and starts a session.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// UserResponse is the public representation of a user account. It never
// includes the password hash.
type UserResponse struct {
	ID        string     `json:"id"`
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	Role      types.Role `json:"role"`
	Disabled  bool       `json:"disabled"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

// ListUsersResponse contains every user account.
type ListUsersResponse struct {
	Users []UserResponse `json:"users"`
}

// CreateUserRequest creates a new user account.
type CreateUserRequest struct {
	Username string     `json:"username"`
	Email    string     `json:"email"`
	Password string     `json:"password"`
	Role     types.Role `json:"role"`
}

// UpdateUserRequest changes a user's role or enabled status.
type UpdateUserRequest struct {
	Role     types.Role `json:"role"`
	Disabled bool       `json:"disabled"`
}

// ListAuditEventsResponse contains recent audit log entries, newest first.
type ListAuditEventsResponse struct {
	Events []types.AuditEvent `json:"events"`
}

// RotateRegistrationTokenResponse contains the new agent registration token.
// The raw token is returned exactly once; only its hash is persisted.
type RotateRegistrationTokenResponse struct {
	Token string `json:"token"`
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
