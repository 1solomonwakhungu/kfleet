package store

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/1solomonwakhungu/kfleet/pkg/types"
)

func TestCountClusters(t *testing.T) {
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
	count, err := st.CountClusters(ctx)
	if err != nil {
		t.Fatalf("CountClusters() error = %v", err)
	}
	if count != 0 {
		t.Fatalf("CountClusters() on an empty store = %d, want 0", count)
	}

	now := time.Now().UTC()
	for _, cluster := range []types.Cluster{
		{ID: "cluster-1", Name: "alpha", Health: types.HealthUnknown, RegisteredAt: now},
		{ID: "cluster-2", Name: "bravo", Health: types.HealthUnknown, RegisteredAt: now},
	} {
		if err := st.CreateCluster(ctx, cluster); err != nil {
			t.Fatalf("CreateCluster(%s) error = %v", cluster.ID, err)
		}
	}
	count, err = st.CountClusters(ctx)
	if err != nil {
		t.Fatalf("CountClusters() error = %v", err)
	}
	if count != 2 {
		t.Fatalf("CountClusters() = %d, want 2", count)
	}

	if err := st.DeleteCluster(ctx, "cluster-1"); err != nil {
		t.Fatalf("DeleteCluster() error = %v", err)
	}
	count, err = st.CountClusters(ctx)
	if err != nil {
		t.Fatalf("CountClusters() error = %v", err)
	}
	if count != 1 {
		t.Fatalf("CountClusters() after delete = %d, want 1", count)
	}
}

func TestCountDeadLetteredAlerts(t *testing.T) {
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
	count, err := st.CountDeadLetteredAlerts(ctx)
	if err != nil {
		t.Fatalf("CountDeadLetteredAlerts() error = %v", err)
	}
	if count != 0 {
		t.Fatalf("CountDeadLetteredAlerts() on an empty store = %d, want 0", count)
	}

	now := time.Now().UTC()
	alert := types.Alert{
		ID:             "alert-dead-letter-1",
		RuleID:         "fleet-health-unreachable",
		RuleName:       "Cluster unreachable",
		ClusterID:      "cluster-1",
		ClusterName:    "alpha",
		DedupeKey:      "unreachable:cluster-1",
		Health:         types.HealthUnreachable,
		Severity:       types.AlertSeverityCritical,
		Summary:        "alpha is unreachable",
		Status:         types.AlertStatusFiring,
		TriggeredAt:    now,
		UpdatedAt:      now,
		DeliveryStatus: types.AlertDeliveryPending,
	}
	created, err := st.CreateAlertIfDue(ctx, alert, 0)
	if err != nil || !created {
		t.Fatalf("CreateAlertIfDue() = %v, %v, want true, nil", created, err)
	}

	// A failure with a next attempt stays in the retrying state and must not
	// count as dead-lettered.
	if err := st.RecordAlertDeliveryFailure(ctx, alert.ID, 1, ptrTime(now.Add(time.Minute)), "webhook timeout", now); err != nil {
		t.Fatalf("RecordAlertDeliveryFailure() error = %v", err)
	}
	count, err = st.CountDeadLetteredAlerts(ctx)
	if err != nil {
		t.Fatalf("CountDeadLetteredAlerts() error = %v", err)
	}
	if count != 0 {
		t.Fatalf("CountDeadLetteredAlerts() with a retrying alert = %d, want 0", count)
	}

	// A terminal failure with no next attempt dead-letters the alert.
	if err := st.RecordAlertDeliveryFailure(ctx, alert.ID, 5, nil, "webhook unavailable", now); err != nil {
		t.Fatalf("RecordAlertDeliveryFailure() error = %v", err)
	}
	count, err = st.CountDeadLetteredAlerts(ctx)
	if err != nil {
		t.Fatalf("CountDeadLetteredAlerts() error = %v", err)
	}
	if count != 1 {
		t.Fatalf("CountDeadLetteredAlerts() = %d, want 1", count)
	}
}

func ptrTime(value time.Time) *time.Time {
	return &value
}
