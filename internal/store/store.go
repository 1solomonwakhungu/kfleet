// Package store defines persistence for hub state.
package store

import (
	"context"
	"errors"
	"time"

	"github.com/1solomonwakhungu/kfleet/pkg/types"
)

// ErrNotFound is returned when a requested cluster does not exist.
var ErrNotFound = errors.New("cluster not found")

// Store defines the persistence operations used by the hub.
type Store interface {
	CreateCluster(ctx context.Context, cluster types.Cluster) error
	GetCluster(ctx context.Context, id string) (types.Cluster, error)
	ListClusters(ctx context.Context) ([]types.Cluster, error)
	DeleteCluster(ctx context.Context, id string) error
	UpdateHealth(ctx context.Context, id string, health types.ClusterHealth, lastHeartbeat time.Time) error
	UpdateSnapshot(ctx context.Context, id string, nodeCount, podCount int, version string) error
	ReplaceSnapshot(ctx context.Context, id string, snapshot types.ClusterSnapshot, version, agentVersion string, health types.ClusterHealth, lastHeartbeat time.Time) error
	ListNodes(ctx context.Context, clusterID string) ([]types.Node, error)
	ListPods(ctx context.Context, clusterID, namespace string) ([]types.Pod, error)
	ListServices(ctx context.Context, clusterID, namespace string) ([]types.Service, error)
	ListDeployments(ctx context.Context, clusterID, namespace string) ([]types.Deployment, error)
	ListEvents(ctx context.Context, clusterID, namespace string) ([]types.Event, error)
	ListNamespaces(ctx context.Context, clusterID string) ([]string, error)
	IssueAgentToken(ctx context.Context, clusterID, tokenHash string) error
	ValidateAgentToken(ctx context.Context, clusterID, tokenHash string) (approved bool, err error)
	ApproveAgent(ctx context.Context, clusterID string) error
	ListPendingAgents(ctx context.Context) ([]types.Cluster, error)
	Close() error
}
