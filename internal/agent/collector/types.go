package collector

import "time"

// ClusterState is a point-in-time view of Kubernetes resources collected by an agent.
type ClusterState struct {
	Nodes        []NodeInfo       `json:"nodes"`
	Pods         []PodInfo        `json:"pods"`
	Services     []ServiceInfo    `json:"services"`
	Deployments  []DeploymentInfo `json:"deployments"`
	Events       []EventInfo      `json:"events"`
	NodeCount    int              `json:"nodeCount"`
	PodCount     int              `json:"podCount"`
	K8sVersion   string           `json:"k8sVersion"`
	CollectedAt  time.Time        `json:"collectedAt"`
	AgentVersion string           `json:"agentVersion"`
}

// NodeInfo contains the node fields needed by the hub.
type NodeInfo struct {
	Name           string   `json:"name"`
	Status         string   `json:"status"`
	K8sVersion     string   `json:"k8sVersion"`
	Roles          []string `json:"roles"`
	CPUCapacity    string   `json:"cpuCapacity"`
	MemoryCapacity string   `json:"memoryCapacity"`
}

// PodInfo contains the pod fields needed by the hub.
type PodInfo struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Phase     string `json:"phase"`
	Restarts  int32  `json:"restarts"`
	Node      string `json:"node"`
}

// ServiceInfo contains the service fields needed by the hub.
type ServiceInfo struct {
	Namespace   string   `json:"namespace"`
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	ClusterIP   string   `json:"clusterIP"`
	ExternalIPs []string `json:"externalIPs"`
	Ports       []string `json:"ports"`
}

// DeploymentInfo contains deployment replica status.
type DeploymentInfo struct {
	Namespace         string `json:"namespace"`
	Name              string `json:"name"`
	DesiredReplicas   int32  `json:"desiredReplicas"`
	ReadyReplicas     int32  `json:"readyReplicas"`
	AvailableReplicas int32  `json:"availableReplicas"`
}

// EventInfo contains a recent Kubernetes event.
type EventInfo struct {
	Namespace      string    `json:"namespace"`
	Name           string    `json:"name"`
	Type           string    `json:"type"`
	Reason         string    `json:"reason"`
	Message        string    `json:"message"`
	InvolvedObject string    `json:"involvedObject"`
	Count          int32     `json:"count"`
	LastTimestamp  time.Time `json:"lastTimestamp"`
}
