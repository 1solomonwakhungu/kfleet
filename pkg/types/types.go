// Package types defines the core domain models shared by the hub and agents.
package types

import "time"

// ClusterHealth describes the observed health of a managed cluster.
type ClusterHealth string

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
