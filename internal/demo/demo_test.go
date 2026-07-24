package demo

import (
	"context"
	"testing"
	"time"
)

func TestOpenContainsOnlySyntheticFixtures(t *testing.T) {
	st, err := Open(context.Background())
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	clusters, err := st.ListClusters(context.Background())
	if err != nil {
		t.Fatalf("ListClusters() error = %v", err)
	}
	if len(clusters) != 3 {
		t.Fatalf("len(clusters) = %d, want 3", len(clusters))
	}
	for _, cluster := range clusters {
		if cluster.Labels["data"] != "synthetic" || cluster.Labels["environment"] != "demo" {
			t.Fatalf("cluster %q is not marked synthetic: %#v", cluster.ID, cluster.Labels)
		}
		if cluster.LastHeartbeat.Before(time.Now().Add(-24 * time.Hour)) {
			t.Fatalf("cluster %q heartbeat is unexpectedly stale", cluster.ID)
		}
	}
}
