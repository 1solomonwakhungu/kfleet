package collector

import "github.com/1solomonwakhungu/kfleet/pkg/api"

// ClusterState is a point-in-time view of Kubernetes resources collected by an agent.
type ClusterState = api.ClusterSnapshotRequest

// NodeInfo contains the node fields needed by the hub.
type NodeInfo = api.SnapshotNode

// PodInfo contains the pod fields needed by the hub.
type PodInfo = api.SnapshotPod

// ServiceInfo contains the service fields needed by the hub.
type ServiceInfo = api.SnapshotService

// DeploymentInfo contains deployment replica status.
type DeploymentInfo = api.SnapshotDeployment

// NamespaceInfo is the compatibility name for collected namespace configuration.
type NamespaceInfo = api.SnapshotNamespace

// EventInfo contains a recent Kubernetes event.
type EventInfo = api.SnapshotEvent
