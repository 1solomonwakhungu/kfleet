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

	// AppendEvent durably records an operational timeline event. It returns
	// inserted=false without error when an event with the same cluster, kind,
	// and dedupe key already exists, so retried callers stay idempotent.
	AppendEvent(ctx context.Context, event types.OperationalEvent) (inserted bool, err error)
	// ListTimelineEvents returns operational events matching filter, ordered by
	// occurrence time newest first, with cursor pagination via EventFilter.Before.
	ListTimelineEvents(ctx context.Context, filter EventFilter) (EventPage, error)
	// PruneEventsBefore deletes operational events older than cutoff and
	// returns the number of rows removed. It is the retention mechanism that
	// keeps the durable timeline bounded.
	PruneEventsBefore(ctx context.Context, cutoff time.Time) (int64, error)

	Close() error
}

// EventFilter narrows ListTimelineEvents results.
type EventFilter struct {
	// ClusterID restricts results to one cluster; empty means fleet-wide.
	ClusterID string
	// Since restricts to events at or after this time, if set.
	Since *time.Time
	// Until restricts to events strictly before this time, if set.
	Until *time.Time
	// Before is the ID of the last event from the previous page. The store uses
	// its occurrence time and ID as a stable pagination position. Zero starts
	// from the newest event.
	Before int64
	// Limit caps the number of events returned. Non-positive values fall back
	// to a package default.
	Limit int
}

// EventPage is one page of operational events.
type EventPage struct {
	Events []types.OperationalEvent
	// NextCursor is the value to pass as EventFilter.Before to fetch the next
	// page, or zero if there are no more events.
	NextCursor int64
}
