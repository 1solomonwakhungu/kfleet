package demo

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/1solomonwakhungu/kfleet/internal/store"
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

	alerts, err := st.ListAlertsForTenant(context.Background(), store.DefaultTenantID, "", 100)
	if err != nil {
		t.Fatalf("ListAlertsForTenant() error = %v", err)
	}
	if len(alerts) != 2 {
		t.Fatalf("len(alerts) = %d, want 2", len(alerts))
	}
	for _, alert := range alerts {
		if !strings.HasPrefix(alert.ID, "demo-alert-demo-cluster-") ||
			!strings.Contains(alert.Summary, "synthetic demo") {
			t.Fatalf("alert is not clearly synthetic: %#v", alert)
		}
	}

	timeline, err := st.ListTimelineEvents(context.Background(), store.EventFilter{
		TenantID: store.DefaultTenantID,
		Limit:    100,
	})
	if err != nil {
		t.Fatalf("ListTimelineEvents() error = %v", err)
	}
	if len(timeline.Events) != 3 {
		t.Fatalf("len(timeline.Events) = %d, want 3", len(timeline.Events))
	}
	for _, event := range timeline.Events {
		if event.Details["data"] != "synthetic" || event.Details["environment"] != "demo" {
			t.Fatalf("timeline event is not clearly synthetic: %#v", event)
		}
	}
}
