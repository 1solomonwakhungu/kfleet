package server

import (
	"context"
	"fmt"
	"time"

	"github.com/1solomonwakhungu/kfleet/pkg/types"
)

// recordEvent appends an operational timeline event. Failures are logged but
// never fail the calling request: the timeline is an observability aid, not
// part of the request's correctness contract.
func (s *Server) recordEvent(ctx context.Context, event types.OperationalEvent) {
	if _, err := s.store.AppendEvent(ctx, event); err != nil {
		s.logger.Error("failed to record operational event",
			"cluster_id", event.ClusterID, "kind", event.Kind, "error", err)
	}
}

func (s *Server) recordClusterRegistered(ctx context.Context, cluster types.Cluster) {
	s.recordEvent(ctx, types.OperationalEvent{
		TenantID:   cluster.TenantID,
		ClusterID:  cluster.ID,
		Kind:       types.EventClusterRegistered,
		Message:    fmt.Sprintf("cluster %q registered", cluster.Name),
		OccurredAt: cluster.RegisteredAt,
		DedupeKey:  "registered",
	})
}

func (s *Server) recordAgentApproved(ctx context.Context, cluster types.Cluster, occurredAt time.Time) {
	s.recordEvent(ctx, types.OperationalEvent{
		TenantID:   cluster.TenantID,
		ClusterID:  cluster.ID,
		Kind:       types.EventAgentApproved,
		Message:    fmt.Sprintf("agent for cluster %q approved", cluster.Name),
		OccurredAt: occurredAt,
		DedupeKey:  "approved",
	})
}

// recordHeartbeatTransition records a health transition, distinguishing a
// reconnect (unreachable -> reachable) from any other transition. It is a
// no-op when health did not actually change, so callers can invoke it
// unconditionally on every heartbeat/snapshot.
func (s *Server) recordHeartbeatTransition(ctx context.Context, cluster types.Cluster, oldHealth, newHealth types.ClusterHealth, occurredAt time.Time) {
	if oldHealth == newHealth {
		return
	}
	kind := types.EventHeartbeatStateChange
	message := fmt.Sprintf("cluster %q health changed from %s to %s", cluster.Name, oldHealth, newHealth)
	if oldHealth == types.HealthUnreachable && newHealth != types.HealthUnreachable {
		kind = types.EventAgentReconnected
		message = fmt.Sprintf("cluster %q reconnected (now %s)", cluster.Name, newHealth)
	}
	s.recordEvent(ctx, types.OperationalEvent{
		TenantID:   cluster.TenantID,
		ClusterID:  cluster.ID,
		Kind:       kind,
		Message:    message,
		Details:    map[string]string{"from": string(oldHealth), "to": string(newHealth)},
		OccurredAt: occurredAt,
		DedupeKey:  fmt.Sprintf("%s->%s:%d", oldHealth, newHealth, occurredAt.UnixNano()),
	})
}

func (s *Server) recordAgentDisconnected(ctx context.Context, cluster types.Cluster, reason string, occurredAt time.Time) {
	details := map[string]string{"reason": reason}
	if !cluster.LastHeartbeat.IsZero() {
		details["lastHeartbeat"] = cluster.LastHeartbeat.UTC().Format(time.RFC3339Nano)
	}
	message := fmt.Sprintf("agent for cluster %q disconnected", cluster.Name)
	if reason == "heartbeat_timeout" {
		message = fmt.Sprintf("agent for cluster %q became stale after missing heartbeats", cluster.Name)
	}
	s.recordEvent(ctx, types.OperationalEvent{
		TenantID:   cluster.TenantID,
		ClusterID:  cluster.ID,
		Kind:       types.EventAgentDisconnected,
		Message:    message,
		Details:    details,
		OccurredAt: occurredAt,
		DedupeKey:  fmt.Sprintf("%s:%d", reason, occurredAt.UnixNano()),
	})
}

// recordVersionChanged is a no-op when there is no previous version (initial
// registration) or the version is unchanged, so it only reports genuine
// transitions.
func (s *Server) recordVersionChanged(ctx context.Context, cluster types.Cluster, oldVersion, newVersion string, occurredAt time.Time) {
	if oldVersion == "" || newVersion == "" || oldVersion == newVersion {
		return
	}
	s.recordEvent(ctx, types.OperationalEvent{
		TenantID:   cluster.TenantID,
		ClusterID:  cluster.ID,
		Kind:       types.EventVersionChanged,
		Message:    fmt.Sprintf("cluster %q version changed from %s to %s", cluster.Name, oldVersion, newVersion),
		Details:    map[string]string{"from": oldVersion, "to": newVersion},
		OccurredAt: occurredAt,
		DedupeKey:  fmt.Sprintf("%s->%s:%d", oldVersion, newVersion, occurredAt.UnixNano()),
	})
}
