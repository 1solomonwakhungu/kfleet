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
	Name         string    `json:"name"`
	Namespace    string    `json:"namespace"`
	Phase        string    `json:"phase"`
	NodeName     string    `json:"nodeName"`
	RestartCount int32     `json:"restartCount"`
	Ready        bool      `json:"ready"`
	StartTime    time.Time `json:"startTime"`
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
	Name              string `json:"name"`
	Namespace         string `json:"namespace"`
	ReadyReplicas     int32  `json:"readyReplicas"`
	DesiredReplicas   int32  `json:"desiredReplicas"`
	UpdatedReplicas   int32  `json:"updatedReplicas"`
	AvailableReplicas int32  `json:"availableReplicas"`
	Age               string `json:"age"`
}

// ClusterSnapshot is the normalized, durable resource state for one cluster.
type ClusterSnapshot struct {
	Nodes       []Node
	Pods        []Pod
	Services    []Service
	Deployments []Deployment
	Events      []Event
}

// AlertSeverity describes the operational impact of a fleet health alert.
type AlertSeverity string

const (
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityCritical AlertSeverity = "critical"
)

// AlertStatus describes the operator lifecycle of an alert.
type AlertStatus string

const (
	AlertStatusFiring       AlertStatus = "firing"
	AlertStatusAcknowledged AlertStatus = "acknowledged"
	AlertStatusResolved     AlertStatus = "resolved"
)

// AlertDeliveryStatus describes webhook delivery state independently from the
// operator lifecycle.
type AlertDeliveryStatus string

const (
	AlertDeliveryPending    AlertDeliveryStatus = "pending"
	AlertDeliveryRetrying   AlertDeliveryStatus = "retrying"
	AlertDeliveryDelivered  AlertDeliveryStatus = "delivered"
	AlertDeliveryDeadLetter AlertDeliveryStatus = "dead_letter"
	AlertDeliveryDisabled   AlertDeliveryStatus = "disabled"
)

// AlertRule maps a cluster health state to an alert severity and cooldown.
type AlertRule struct {
	ID              string        `json:"id"`
	Name            string        `json:"name"`
	Health          ClusterHealth `json:"health"`
	Severity        AlertSeverity `json:"severity"`
	CooldownSeconds int64         `json:"cooldownSeconds"`
	Enabled         bool          `json:"enabled"`
	CreatedAt       time.Time     `json:"createdAt"`
	UpdatedAt       time.Time     `json:"updatedAt"`
}

// Alert is a durable fleet health alert and its webhook delivery record.
type Alert struct {
	ID                string              `json:"id"`
	RuleID            string              `json:"ruleId"`
	RuleName          string              `json:"ruleName"`
	ClusterID         string              `json:"clusterId"`
	ClusterName       string              `json:"clusterName"`
	DedupeKey         string              `json:"dedupeKey"`
	Health            ClusterHealth       `json:"health"`
	Severity          AlertSeverity       `json:"severity"`
	Summary           string              `json:"summary"`
	Status            AlertStatus         `json:"status"`
	TriggeredAt       time.Time           `json:"triggeredAt"`
	UpdatedAt         time.Time           `json:"updatedAt"`
	AcknowledgedAt    *time.Time          `json:"acknowledgedAt,omitempty"`
	AcknowledgedBy    string              `json:"acknowledgedBy,omitempty"`
	ResolvedAt        *time.Time          `json:"resolvedAt,omitempty"`
	DeliveryStatus    AlertDeliveryStatus `json:"deliveryStatus"`
	DeliveryAttempts  int                 `json:"deliveryAttempts"`
	NextDeliveryAt    *time.Time          `json:"nextDeliveryAt,omitempty"`
	LastDeliveryError string              `json:"lastDeliveryError,omitempty"`
	DeliveredAt       *time.Time          `json:"deliveredAt,omitempty"`
	DeadLetteredAt    *time.Time          `json:"deadLetteredAt,omitempty"`
}
