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
