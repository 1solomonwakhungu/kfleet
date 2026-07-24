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

// DefaultTenantID is used by single-tenant installations and legacy records.
const DefaultTenantID = "default"

// ErrConflict is returned when a create or update violates a uniqueness
// constraint, such as a duplicate username or email.
var ErrConflict = errors.New("conflict")

// ErrLastAdmin is returned when an operation would leave the hub without at
// least one enabled admin account.
var ErrLastAdmin = errors.New("at least one enabled admin is required")

// Store defines the persistence operations used by the hub.
type Store interface {
	CreateCluster(ctx context.Context, cluster types.Cluster) error
	GetCluster(ctx context.Context, id string) (types.Cluster, error)
	GetClusterForTenant(ctx context.Context, tenantID, id string) (types.Cluster, error)
	ListClusters(ctx context.Context) ([]types.Cluster, error)
	ListClustersForTenant(ctx context.Context, tenantID string) ([]types.Cluster, error)
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
	ListNamespaceConfigs(ctx context.Context, clusterID string) ([]types.Namespace, error)
	IssueAgentToken(ctx context.Context, clusterID, tokenHash string) error
	ValidateAgentToken(ctx context.Context, clusterID, tokenHash string) (approved bool, err error)
	ApproveAgent(ctx context.Context, clusterID string) error
	ListPendingAgents(ctx context.Context) ([]types.Cluster, error)
	ListPendingAgentsForTenant(ctx context.Context, tenantID string) ([]types.Cluster, error)

	// AppendEvent durably records an operational timeline event and suppresses
	// retries with the same cluster, kind, and dedupe key.
	AppendEvent(ctx context.Context, event types.OperationalEvent) (inserted bool, err error)
	ListTimelineEvents(ctx context.Context, filter EventFilter) (EventPage, error)
	PruneEventsBefore(ctx context.Context, cutoff time.Time) (int64, error)

	// User accounts and RBAC.
	CreateUser(ctx context.Context, user types.User) error
	GetUserByID(ctx context.Context, id string) (types.User, error)
	GetUserByUsername(ctx context.Context, username string) (types.User, error)
	ListUsers(ctx context.Context) ([]types.User, error)
	// UpdateUser and DeleteUser return ErrLastAdmin instead of applying a
	// change that would leave the hub with zero enabled admins.
	UpdateUser(ctx context.Context, id string, role types.Role, disabled bool) error
	DeleteUser(ctx context.Context, id string) error

	// Sessions.
	CreateSession(ctx context.Context, tokenHash, userID string, expiresAt time.Time) error
	GetSessionUser(ctx context.Context, tokenHash string, now time.Time) (types.User, error)
	DeleteSession(ctx context.Context, tokenHash string) error
	DeleteExpiredSessions(ctx context.Context, now time.Time) error

	// Append-only audit log.
	RecordAuditEvent(ctx context.Context, event types.AuditEvent) error
	ListAuditEvents(ctx context.Context, limit int) ([]types.AuditEvent, error)

	// Runtime settings (e.g. rotated registration token hash).
	GetSetting(ctx context.Context, key string) (value string, ok bool, err error)
	SetSetting(ctx context.Context, key, value string) error

	Close() error
}

// EventFilter narrows ListTimelineEvents results.
type EventFilter struct {
	// TenantID restricts results to one tenant; empty is reserved for internal
	// unscoped operations and tests.
	TenantID string
	// ClusterID restricts results to one cluster; empty means fleet-wide.
	ClusterID string
	Since     *time.Time
	Until     *time.Time
	// Before is the ID of the last event from the previous page.
	Before int64
	Limit  int
}

// EventPage is one page of operational events.
type EventPage struct {
	Events     []types.OperationalEvent
	NextCursor int64
}
