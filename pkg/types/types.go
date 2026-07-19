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
