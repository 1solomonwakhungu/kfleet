// Package store defines persistence for hub state.
package store

import (
	"context"
	"errors"
	"time"

	"github.com/1solomonwakhungu/kfleet/pkg/types"
)

var (
	// ErrNotFound is returned when a requested record does not exist.
	ErrNotFound = errors.New("record not found")
	// ErrInvalidState is returned when a lifecycle transition is not allowed.
	ErrInvalidState = errors.New("invalid record state")
)

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
	ListAlertRules(ctx context.Context) ([]types.AlertRule, error)
	UpsertAlertRule(ctx context.Context, rule types.AlertRule) error
	CreateAlertIfDue(ctx context.Context, alert types.Alert, cooldown time.Duration) (bool, error)
	ResolveClusterAlerts(ctx context.Context, clusterID string, health types.ClusterHealth, resolvedAt time.Time) error
	ListAlerts(ctx context.Context, status types.AlertStatus, limit int) ([]types.Alert, error)
	GetAlert(ctx context.Context, id string) (types.Alert, error)
	AcknowledgeAlert(ctx context.Context, id, acknowledgedBy string, acknowledgedAt time.Time) error
	ListDueAlertDeliveries(ctx context.Context, now time.Time, limit int) ([]types.Alert, error)
	RecordAlertDelivered(ctx context.Context, id string, attempts int, deliveredAt time.Time) error
	RecordAlertDeliveryFailure(ctx context.Context, id string, attempts int, nextAttemptAt *time.Time, deliveryError string, failedAt time.Time) error
	Close() error
}
