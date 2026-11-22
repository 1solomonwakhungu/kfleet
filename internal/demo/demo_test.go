package demo

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/1solomonwakhungu/kfleet/internal/policy"
	"github.com/1solomonwakhungu/kfleet/internal/store"
	"github.com/1solomonwakhungu/kfleet/pkg/types"
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

func TestOpenHeartbeatsStayFreshAfterClockAdvances(t *testing.T) {
	ctx := context.Background()
	current := time.Now().UTC().Add(-2 * time.Hour)
	st, err := openWithClock(ctx, func() time.Time { return current })
	if err != nil {
		t.Fatalf("openWithClock() error = %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	// Advance the clock by 2 hours past seeding; current now matches wall-clock
	// time.Now() so policy.Engine evaluates against the same timestamp.
	current = current.Add(2 * time.Hour)
	wantNow := current.UTC().Truncate(time.Second)

	// Verify underlying SQLite rows were not mutated.
	wrapped, ok := st.(*demoStore)
	if !ok {
		t.Fatalf("openWithClock() returned %T, want *demoStore", st)
	}
	rawAlpha, err := wrapped.Store.GetCluster(ctx, ClusterAlphaID)
	if err != nil {
		t.Fatalf("raw GetCluster(alpha) error = %v", err)
	}
	if wantNow.Sub(rawAlpha.LastHeartbeat) < time.Hour {
		t.Fatalf("underlying store was mutated: raw LastHeartbeat = %v, want ~2h before %v", rawAlpha.LastHeartbeat, wantNow)
	}

	wantHeartbeats := map[string]time.Time{
		ClusterAlphaID:   wantNow.Add(alphaHeartbeatOffset),
		ClusterBravoID:   wantNow.Add(bravoHeartbeatOffset),
		ClusterCharlieID: wantNow.Add(charlieHeartbeatOffset),
	}

	checkList := func(label string, clusters []types.Cluster) {
		t.Helper()
		if len(clusters) != 3 {
			t.Fatalf("%s returned %d clusters, want 3", label, len(clusters))
		}
		for _, cluster := range clusters {
			want := wantHeartbeats[cluster.ID]
			if !cluster.LastHeartbeat.Equal(want) {
				t.Errorf("%s cluster %q LastHeartbeat = %v, want %v", label, cluster.ID, cluster.LastHeartbeat, want)
			}
		}
	}

	list, err := st.ListClusters(ctx)
	if err != nil {
		t.Fatalf("ListClusters() error = %v", err)
	}
	checkList("ListClusters", list)

	tenantList, err := st.ListClustersForTenant(ctx, store.DefaultTenantID)
	if err != nil {
		t.Fatalf("ListClustersForTenant() error = %v", err)
	}
	checkList("ListClustersForTenant", tenantList)

	for id, want := range wantHeartbeats {
		cluster, err := st.GetCluster(ctx, id)
		if err != nil {
			t.Fatalf("GetCluster(%q) error = %v", id, err)
		}
		if !cluster.LastHeartbeat.Equal(want) {
			t.Errorf("GetCluster(%q) LastHeartbeat = %v, want %v", id, cluster.LastHeartbeat, want)
		}
		tenantCluster, err := st.GetClusterForTenant(ctx, store.DefaultTenantID, id)
		if err != nil {
			t.Fatalf("GetClusterForTenant(%q) error = %v", id, err)
		}
		if !tenantCluster.LastHeartbeat.Equal(want) {
			t.Errorf("GetClusterForTenant(%q) LastHeartbeat = %v, want %v", id, tenantCluster.LastHeartbeat, want)
		}
	}

	engine := policy.NewEngine(st, 90*time.Second)
	results, err := engine.Evaluate(ctx, store.DefaultTenantID)
	if err != nil {
		t.Fatalf("policy.Engine.Evaluate() error = %v", err)
	}

	counts := map[string]map[types.PolicyStatus]int{
		ClusterAlphaID:   {},
		ClusterBravoID:   {},
		ClusterCharlieID: {},
	}
	for _, result := range results {
		counts[result.Subject.ClusterID][result.Status]++
	}

	for _, freshID := range []string{ClusterAlphaID, ClusterBravoID} {
		if stale := counts[freshID][types.PolicyStale]; stale != 0 {
			t.Errorf("cluster %q has %d stale policy results after 2h, want 0 (counts=%v)", freshID, stale, counts[freshID])
		}
		if counts[freshID][types.PolicyPass] == 0 || counts[freshID][types.PolicyFail] == 0 {
			t.Errorf("cluster %q expected both pass and fail policy results, got %v", freshID, counts[freshID])
		}
	}
	if counts[ClusterCharlieID][types.PolicyStale] == 0 {
		t.Errorf("cluster %q expected stale policy results, got %v", ClusterCharlieID, counts[ClusterCharlieID])
	}
	for status, count := range counts[ClusterCharlieID] {
		if status != types.PolicyStale && count != 0 {
			t.Errorf("cluster %q has %d %q policy results, want only stale", ClusterCharlieID, count, status)
		}
	}
}

func TestOpenSeedsCrashLoopLedgerPod(t *testing.T) {
	ctx := context.Background()
	st, err := Open(ctx)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	bravoPods, err := st.ListPods(ctx, ClusterBravoID, "payments")
	if err != nil {
		t.Fatalf("ListPods(bravo, payments) error = %v", err)
	}
	var foundLedger bool
	for _, pod := range bravoPods {
		if pod.Name == "ledger-api-6f75-demo" {
			foundLedger = true
			if pod.Phase != "CrashLoopBackOff" || pod.RestartCount != 7 || pod.Ready {
				t.Fatalf("ledger-api-6f75-demo = %+v, want CrashLoopBackOff with 7 restarts and Ready=false", pod)
			}
		}
	}
	if !foundLedger {
		t.Fatalf("ledger-api-6f75-demo not found in %+v", bravoPods)
	}

	charliePods, err := st.ListPods(ctx, ClusterCharlieID, "edge-sample")
	if err != nil {
		t.Fatalf("ListPods(charlie, edge-sample) error = %v", err)
	}
	var foundTelemetry bool
	for _, pod := range charliePods {
		if pod.Name == "telemetry-demo" {
			foundTelemetry = true
			if pod.Phase != "Pending" || pod.RestartCount != 2 || pod.Ready {
				t.Fatalf("telemetry-demo = %+v, want Pending with 2 restarts and Ready=false", pod)
			}
		}
	}
	if !foundTelemetry {
		t.Fatalf("telemetry-demo not found in %+v", charliePods)
	}
}
