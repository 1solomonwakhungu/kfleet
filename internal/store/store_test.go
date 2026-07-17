package store

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/1solomonwakhungu/kfleet/pkg/types"
)

func TestSQLiteStoreClusterLifecycle(t *testing.T) {
	t.Parallel()

	st, err := Open(filepath.Join(t.TempDir(), "kfleet.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := st.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})

	ctx := context.Background()
	registeredAt := time.Date(2025, time.January, 2, 3, 4, 5, 0, time.UTC)
	cluster := types.Cluster{
		ID:           "cluster-1",
		Name:         "production",
		Health:       types.HealthUnknown,
		RegisteredAt: registeredAt,
		Labels:       map[string]string{"region": "us-central1"},
	}

	if err := st.CreateCluster(ctx, cluster); err != nil {
		t.Fatalf("CreateCluster() error = %v", err)
	}

	got, err := st.GetCluster(ctx, cluster.ID)
	if err != nil {
		t.Fatalf("GetCluster() error = %v", err)
	}
	if !reflect.DeepEqual(got, cluster) {
		t.Fatalf("GetCluster() = %#v, want %#v", got, cluster)
	}

	clusters, err := st.ListClusters(ctx)
	if err != nil {
		t.Fatalf("ListClusters() error = %v", err)
	}
	if len(clusters) != 1 || !reflect.DeepEqual(clusters[0], cluster) {
		t.Fatalf("ListClusters() = %#v, want [%#v]", clusters, cluster)
	}

	heartbeat := registeredAt.Add(time.Minute)
	if err := st.UpdateHealth(ctx, cluster.ID, types.HealthHealthy, heartbeat); err != nil {
		t.Fatalf("UpdateHealth() error = %v", err)
	}
	got, err = st.GetCluster(ctx, cluster.ID)
	if err != nil {
		t.Fatalf("GetCluster() after UpdateHealth error = %v", err)
	}
	if got.Health != types.HealthHealthy || !got.LastHeartbeat.Equal(heartbeat) {
		t.Fatalf("health update = (%q, %v), want (%q, %v)", got.Health, got.LastHeartbeat, types.HealthHealthy, heartbeat)
	}

	if err := st.UpdateSnapshot(ctx, cluster.ID, 3, 42, "v1.31.1"); err != nil {
		t.Fatalf("UpdateSnapshot() error = %v", err)
	}
	got, err = st.GetCluster(ctx, cluster.ID)
	if err != nil {
		t.Fatalf("GetCluster() after UpdateSnapshot error = %v", err)
	}
	if got.NodeCount != 3 || got.PodCount != 42 || got.Version != "v1.31.1" {
		t.Fatalf("snapshot update = (%d, %d, %q), want (3, 42, %q)", got.NodeCount, got.PodCount, got.Version, "v1.31.1")
	}

	if err := st.DeleteCluster(ctx, cluster.ID); err != nil {
		t.Fatalf("DeleteCluster() error = %v", err)
	}
	_, err = st.GetCluster(ctx, cluster.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetCluster() after delete error = %v, want ErrNotFound", err)
	}
}

func TestSQLiteStoreAgentApprovalLifecycle(t *testing.T) {
	t.Parallel()

	st, err := Open(filepath.Join(t.TempDir(), "kfleet.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := st.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})

	ctx := context.Background()
	cluster := types.Cluster{
		ID:           "cluster-agent",
		Name:         "agent-cluster",
		Health:       types.HealthUnknown,
		RegisteredAt: time.Now().UTC(),
		Labels:       map[string]string{},
	}
	if err := st.CreateCluster(ctx, cluster); err != nil {
		t.Fatalf("CreateCluster() error = %v", err)
	}
	if err := st.IssueAgentToken(ctx, cluster.ID, "hash-one"); err != nil {
		t.Fatalf("IssueAgentToken() error = %v", err)
	}

	approved, err := st.ValidateAgentToken(ctx, cluster.ID, "hash-one")
	if err != nil || approved {
		t.Fatalf("ValidateAgentToken() = (%v, %v), want (false, nil)", approved, err)
	}
	if _, err := st.ValidateAgentToken(ctx, cluster.ID, "wrong-hash"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("ValidateAgentToken() wrong hash error = %v, want ErrNotFound", err)
	}

	pending, err := st.ListPendingAgents(ctx)
	if err != nil {
		t.Fatalf("ListPendingAgents() error = %v", err)
	}
	if len(pending) != 1 || pending[0].ID != cluster.ID {
		t.Fatalf("ListPendingAgents() = %#v, want cluster %q", pending, cluster.ID)
	}

	if err := st.ApproveAgent(ctx, cluster.ID); err != nil {
		t.Fatalf("ApproveAgent() error = %v", err)
	}
	approved, err = st.ValidateAgentToken(ctx, cluster.ID, "hash-one")
	if err != nil || !approved {
		t.Fatalf("ValidateAgentToken() = (%v, %v), want (true, nil)", approved, err)
	}
	pending, err = st.ListPendingAgents(ctx)
	if err != nil || len(pending) != 0 {
		t.Fatalf("ListPendingAgents() after approval = (%#v, %v), want empty", pending, err)
	}

	if err := st.IssueAgentToken(ctx, cluster.ID, "hash-two"); err != nil {
		t.Fatalf("IssueAgentToken() rotation error = %v", err)
	}
	if _, err := st.ValidateAgentToken(ctx, cluster.ID, "hash-one"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("ValidateAgentToken() old token error = %v, want ErrNotFound", err)
	}
	approved, err = st.ValidateAgentToken(ctx, cluster.ID, "hash-two")
	if err != nil || approved {
		t.Fatalf("ValidateAgentToken() rotated = (%v, %v), want pending", approved, err)
	}

	if err := st.ApproveAgent(ctx, "missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("ApproveAgent() missing error = %v, want ErrNotFound", err)
	}
	if err := st.IssueAgentToken(ctx, "missing", "hash"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("IssueAgentToken() missing error = %v, want ErrNotFound", err)
	}
}
