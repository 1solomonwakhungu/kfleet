// Package types defines the core domain models shared by the hub and agents.
package types

import "time"

// ClusterHealth describes the observed health of a managed cluster.
type ClusterHealth string

// Supported cluster health states.
const (
	HealthUnknown     ClusterHealth = "unknown"
	HealthHealthy     ClusterHealth = "healthy"
	HealthDegraded    ClusterHealth = "degraded"
	HealthUnreachable ClusterHealth = "unreachable"
)

// Cluster describes a Kubernetes cluster registered with kfleet.
type Cluster struct {
	ID            string            `json:"id"`
	TenantID      string            `json:"-"`
	Name          string            `json:"name"`
	Health        ClusterHealth     `json:"health"`
	Version       string            `json:"version"`
	AgentVersion  string            `json:"agentVersion"`
	NodeCount     int               `json:"nodeCount"`
	PodCount      int               `json:"podCount"`
	RegisteredAt  time.Time         `json:"registeredAt"`
	LastHeartbeat time.Time         `json:"lastHeartbeat"`
	Labels        map[string]string `json:"labels"`
}

// Node describes the state and capacity of a Kubernetes node.
type Node struct {
	Name           string   `json:"name"`
	Status         string   `json:"status"`
	Roles          []string `json:"roles"`
	Version        string   `json:"version"`
	CPUCapacity    string   `json:"cpuCapacity"`
	MemoryCapacity string   `json:"memoryCapacity"`
	Ready          bool     `json:"ready"`
}

// Pod describes the runtime state of a Kubernetes pod.
type Pod struct {
	Name                      string    `json:"name"`
	Namespace                 string    `json:"namespace"`
	Phase                     string    `json:"phase"`
	NodeName                  string    `json:"nodeName"`
	RestartCount              int32     `json:"restartCount"`
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

// Event describes an event observed in a managed cluster.
type Event struct {
	ClusterID     string    `json:"clusterId"`
	Namespace     string    `json:"namespace"`
	Reason        string    `json:"reason"`
	Message       string    `json:"message"`
	Type          string    `json:"type"`
	Count         int32     `json:"count"`
	LastTimestamp time.Time `json:"lastTimestamp"`
}

// ServicePort describes a port exposed by a Kubernetes service.
type ServicePort struct {
	Name       string `json:"name"`
	Port       int32  `json:"port"`
	TargetPort int32  `json:"targetPort"`
	Protocol   string `json:"protocol"`
}

// Service describes a Kubernetes service in a cluster snapshot.
type Service struct {
	Name        string        `json:"name"`
	Namespace   string        `json:"namespace"`
	Type        string        `json:"type"`
	ClusterIP   string        `json:"clusterIP"`
	ExternalIPs []string      `json:"externalIPs"`
	Ports       []ServicePort `json:"ports"`
	Age         string        `json:"age"`
}

// Deployment describes the replica state of a Kubernetes deployment.
type Deployment struct {
	Name              string   `json:"name"`
	Namespace         string   `json:"namespace"`
	ReadyReplicas     int32    `json:"readyReplicas"`
	DesiredReplicas   int32    `json:"desiredReplicas"`
	UpdatedReplicas   int32    `json:"updatedReplicas"`
	AvailableReplicas int32    `json:"availableReplicas"`
	Age               string   `json:"age"`
	ConfigHash        string   `json:"configHash"`
	Images            []string `json:"images"`
}

// Namespace describes configuration attached to a Kubernetes namespace.
type Namespace struct {
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels"`
}

// ClusterSnapshot is the normalized, durable resource state for one cluster.
type ClusterSnapshot struct {
	Nodes       []Node
	Pods        []Pod
	Services    []Service
	Deployments []Deployment
	Namespaces  []Namespace
	Events      []Event
}

// OperationalEventKind categorizes an entry in the fleet timeline.
type OperationalEventKind string

const (
	EventClusterRegistered    OperationalEventKind = "cluster_registered"
	EventAgentApproved        OperationalEventKind = "agent_approved"
	EventHeartbeatStateChange OperationalEventKind = "heartbeat_state_change"
	EventVersionChanged       OperationalEventKind = "version_changed"
	EventAgentReconnected     OperationalEventKind = "agent_reconnected"
	EventAgentDisconnected    OperationalEventKind = "agent_disconnected"
	EventPolicyFinding        OperationalEventKind = "policy_finding"
)

// OperationalEvent is a durable, append-only fleet lifecycle record.
type OperationalEvent struct {
	ID         int64                `json:"id"`
	TenantID   string               `json:"-"`
	ClusterID  string               `json:"clusterId"`
	Kind       OperationalEventKind `json:"kind"`
	Message    string               `json:"message"`
	Details    map[string]string    `json:"details,omitempty"`
	OccurredAt time.Time            `json:"occurredAt"`
	DedupeKey  string               `json:"-"`
}

// Role identifies a user's permission level in the hub.
type Role string

// Supported user roles, ordered from least to most privileged.
const (
	// RoleReadOnly can view fleet state but cannot perform mutations.
	RoleReadOnly Role = "read_only"
	// RoleOperator can perform day-to-day fleet operations such as
	// approving agents and registering or removing clusters.
	RoleOperator Role = "operator"
	// RoleAdmin can perform every operator action plus user management
	// and hub configuration changes.
	RoleAdmin Role = "admin"
)

// ValidRole reports whether role is one of the known roles.
func ValidRole(role Role) bool {
	switch role {
	case RoleReadOnly, RoleOperator, RoleAdmin:
		return true
	default:
		return false
	}
}

// User is a hub operator account used to authenticate to the REST API and web UI.
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	Role         Role      `json:"role"`
	Disabled     bool      `json:"disabled"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	PasswordHash string    `json:"-"`
}

// AuditOutcome describes whether an audited action succeeded or failed.
type AuditOutcome string

// Supported audit outcomes.
const (
	AuditSuccess AuditOutcome = "success"
	AuditFailure AuditOutcome = "failure"
)

// AuditEvent is an immutable record of a security-relevant action.
type AuditEvent struct {
	ID            string       `json:"id"`
	OccurredAt    time.Time    `json:"occurredAt"`
	ActorUserID   string       `json:"actorUserId,omitempty"`
	ActorUsername string       `json:"actorUsername"`
	ActorRole     Role         `json:"actorRole,omitempty"`
	Action        string       `json:"action"`
	TargetType    string       `json:"targetType"`
	TargetID      string       `json:"targetId"`
	Outcome       AuditOutcome `json:"outcome"`
	Details       string       `json:"details,omitempty"`
	SourceIP      string       `json:"sourceIp,omitempty"`
}
