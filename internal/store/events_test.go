package store

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/1solomonwakhungu/kfleet/pkg/types"
)

func newEventTestStore(t *testing.T) (Store, string) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "kfleet.db")
	st, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := st.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})
	return st, dbPath
}

func TestSQLiteStoreAppendEventOrderingAndDedup(t *testing.T) {
	t.Parallel()

	st, _ := newEventTestStore(t)
	ctx := context.Background()

	cluster := types.Cluster{
		ID: "cluster-1", Name: "production", Health: types.HealthUnknown,
		RegisteredAt: time.Now().UTC(), Labels: map[string]string{},
	}
	if err := st.CreateCluster(ctx, cluster); err != nil {
		t.Fatalf("CreateCluster() error = %v", err)
	}

	base := time.Date(2026, time.July, 20, 0, 0, 0, 0, time.UTC)
	events := []types.OperationalEvent{
		{ClusterID: cluster.ID, Kind: types.EventClusterRegistered, Message: "registered", OccurredAt: base, DedupeKey: "registered"},
		{ClusterID: cluster.ID, Kind: types.EventAgentApproved, Message: "approved", OccurredAt: base.Add(time.Minute), DedupeKey: "approved"},
		{ClusterID: cluster.ID, Kind: types.EventHeartbeatStateChange, Message: "unknown->healthy", OccurredAt: base.Add(2 * time.Minute), DedupeKey: "unknown->healthy:1"},
	}
	for _, event := range events {
		inserted, err := st.AppendEvent(ctx, event)
		if err != nil {
			t.Fatalf("AppendEvent(%s) error = %v", event.Kind, err)
		}
		if !inserted {
			t.Fatalf("AppendEvent(%s) inserted = false, want true", event.Kind)
		}
	}

	// Ordering must be deterministic: newest first.
	page, err := st.ListTimelineEvents(ctx, EventFilter{ClusterID: cluster.ID, Limit: 10})
	if err != nil {
		t.Fatalf("ListTimelineEvents() error = %v", err)
	}
	wantOrder := []types.OperationalEventKind{
		types.EventHeartbeatStateChange, types.EventAgentApproved, types.EventClusterRegistered,
	}
	if len(page.Events) != len(wantOrder) {
		t.Fatalf("ListTimelineEvents() = %#v, want %d events", page.Events, len(wantOrder))
	}
	for i, kind := range wantOrder {
		if page.Events[i].Kind != kind {
			t.Fatalf("event[%d].Kind = %q, want %q", i, page.Events[i].Kind, kind)
		}
	}
	if page.NextCursor != 0 {
		t.Fatalf("NextCursor = %d, want 0 (last page)", page.NextCursor)
	}

	// Duplicate suppression: identical cluster/kind/dedupe key is a no-op.
	duplicate := events[1]
	inserted, err := st.AppendEvent(ctx, duplicate)
	if err != nil {
		t.Fatalf("AppendEvent(duplicate) error = %v", err)
	}
	if inserted {
		t.Fatal("AppendEvent(duplicate) inserted = true, want false")
	}
	page, err = st.ListTimelineEvents(ctx, EventFilter{ClusterID: cluster.ID, Limit: 10})
	if err != nil {
		t.Fatalf("ListTimelineEvents() after duplicate error = %v", err)
	}
	if len(page.Events) != 3 {
		t.Fatalf("ListTimelineEvents() after duplicate = %d events, want 3 (duplicate suppressed)", len(page.Events))
	}

	// A different dedupe key for the same kind is a genuinely new event, not
	// globally suppressed.
	inserted, err = st.AppendEvent(ctx, types.OperationalEvent{
		ClusterID: cluster.ID, Kind: types.EventAgentApproved, Message: "approved again",
		OccurredAt: base.Add(3 * time.Minute), DedupeKey: "approved-again",
	})
	if err != nil {
		t.Fatalf("AppendEvent(distinct dedupe key) error = %v", err)
	}
	if !inserted {
		t.Fatal("AppendEvent(distinct dedupe key) inserted = false, want true")
	}

	// Events without a dedupe key never collide, even when identical.
	noKey := types.OperationalEvent{ClusterID: cluster.ID, Kind: types.EventPolicyFinding, Message: "finding", OccurredAt: base}
	for i := 0; i < 2; i++ {
		inserted, err := st.AppendEvent(ctx, noKey)
		if err != nil {
			t.Fatalf("AppendEvent(no dedupe key, iteration %d) error = %v", i, err)
		}
		if !inserted {
			t.Fatalf("AppendEvent(no dedupe key, iteration %d) inserted = false, want true", i)
		}
	}
}

func TestSQLiteStoreListTimelineEventsFilteringAndPagination(t *testing.T) {
	t.Parallel()

	st, _ := newEventTestStore(t)
	ctx := context.Background()

	clusterA := types.Cluster{ID: "cluster-a", Name: "a", RegisteredAt: time.Now().UTC(), Labels: map[string]string{}}
	clusterB := types.Cluster{ID: "cluster-b", Name: "b", RegisteredAt: time.Now().UTC(), Labels: map[string]string{}}
	if err := st.CreateCluster(ctx, clusterA); err != nil {
		t.Fatalf("CreateCluster(a) error = %v", err)
	}
	if err := st.CreateCluster(ctx, clusterB); err != nil {
		t.Fatalf("CreateCluster(b) error = %v", err)
	}

	base := time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 5; i++ {
		clusterID := clusterA.ID
		if i%2 == 0 {
			clusterID = clusterB.ID
		}
		event := types.OperationalEvent{
			ClusterID:  clusterID,
			Kind:       types.EventHeartbeatStateChange,
			Message:    fmt.Sprintf("event-%d", i),
			OccurredAt: base.Add(time.Duration(i) * time.Hour),
			DedupeKey:  fmt.Sprintf("event-%d", i),
		}
		if _, err := st.AppendEvent(ctx, event); err != nil {
			t.Fatalf("AppendEvent(%d) error = %v", i, err)
		}
	}

	// Time-range filter: events at hour 1, 2, 3 (since inclusive, until exclusive).
	since := base.Add(time.Hour)
	until := base.Add(4 * time.Hour)
	page, err := st.ListTimelineEvents(ctx, EventFilter{Since: &since, Until: &until, Limit: 10})
	if err != nil {
		t.Fatalf("ListTimelineEvents(time range) error = %v", err)
	}
	if len(page.Events) != 3 {
		t.Fatalf("ListTimelineEvents(time range) = %d events, want 3", len(page.Events))
	}

	// Cluster-scoped filter: clusterA got the odd-indexed events (1, 3) => 2 events.
	pageA, err := st.ListTimelineEvents(ctx, EventFilter{ClusterID: clusterA.ID, Limit: 10})
	if err != nil {
		t.Fatalf("ListTimelineEvents(clusterA) error = %v", err)
	}
	if len(pageA.Events) != 2 {
		t.Fatalf("ListTimelineEvents(clusterA) = %d events, want 2", len(pageA.Events))
	}

	// Cursor-based pagination across all 5 fleet-wide events, two per page.
	seen := make(map[int64]bool)
	var cursor int64
	total := 0
	for pageIndex := 0; ; pageIndex++ {
		if pageIndex > 10 {
			t.Fatal("pagination did not terminate")
		}
		got, err := st.ListTimelineEvents(ctx, EventFilter{Before: cursor, Limit: 2})
		if err != nil {
			t.Fatalf("ListTimelineEvents(page %d) error = %v", pageIndex, err)
		}
		for _, event := range got.Events {
			if seen[event.ID] {
				t.Fatalf("event id %d returned on more than one page", event.ID)
			}
			seen[event.ID] = true
		}
		total += len(got.Events)
		if got.NextCursor == 0 {
			break
		}
		cursor = got.NextCursor
	}
	if total != 5 {
		t.Fatalf("paged through %d events, want 5", total)
	}
}

func TestSQLiteStoreOrdersTimelineByOccurrenceTime(t *testing.T) {
	t.Parallel()

	st, _ := newEventTestStore(t)
	ctx := context.Background()
	cluster := types.Cluster{
		ID: "cluster-1", Name: "production",
		RegisteredAt: time.Now().UTC(), Labels: map[string]string{},
	}
	if err := st.CreateCluster(ctx, cluster); err != nil {
		t.Fatalf("CreateCluster() error = %v", err)
	}
	base := time.Date(2026, time.July, 23, 12, 0, 0, 0, time.UTC)
	// Insert the later occurrence first to prove ordering is based on the
	// event timestamp rather than its ingestion ID.
	for _, event := range []types.OperationalEvent{
		{ClusterID: cluster.ID, Kind: types.EventPolicyFinding, Message: "newer occurrence", OccurredAt: base.Add(time.Hour), DedupeKey: "newer"},
		{ClusterID: cluster.ID, Kind: types.EventClusterRegistered, Message: "older occurrence", OccurredAt: base, DedupeKey: "older"},
	} {
		if _, err := st.AppendEvent(ctx, event); err != nil {
			t.Fatalf("AppendEvent(%s) error = %v", event.Message, err)
		}
	}

	first, err := st.ListTimelineEvents(ctx, EventFilter{ClusterID: cluster.ID, Limit: 1})
	if err != nil {
		t.Fatalf("ListTimelineEvents(first) error = %v", err)
	}
	if len(first.Events) != 1 || first.Events[0].Message != "newer occurrence" || first.NextCursor == 0 {
		t.Fatalf("first page = %#v, want newer occurrence and cursor", first)
	}
	second, err := st.ListTimelineEvents(ctx, EventFilter{
		ClusterID: cluster.ID, Before: first.NextCursor, Limit: 1,
	})
	if err != nil {
		t.Fatalf("ListTimelineEvents(second) error = %v", err)
	}
	if len(second.Events) != 1 || second.Events[0].Message != "older occurrence" {
		t.Fatalf("second page = %#v, want older occurrence", second)
	}
}

func TestSQLiteStorePruneEventsBefore(t *testing.T) {
	t.Parallel()

	st, _ := newEventTestStore(t)
	ctx := context.Background()

	cluster := types.Cluster{ID: "cluster-1", Name: "production", RegisteredAt: time.Now().UTC(), Labels: map[string]string{}}
	if err := st.CreateCluster(ctx, cluster); err != nil {
		t.Fatalf("CreateCluster() error = %v", err)
	}

	base := time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC)
	old := types.OperationalEvent{ClusterID: cluster.ID, Kind: types.EventHeartbeatStateChange, Message: "old", OccurredAt: base, DedupeKey: "old"}
	recent := types.OperationalEvent{ClusterID: cluster.ID, Kind: types.EventHeartbeatStateChange, Message: "recent", OccurredAt: base.Add(48 * time.Hour), DedupeKey: "recent"}
	for _, event := range []types.OperationalEvent{old, recent} {
		if _, err := st.AppendEvent(ctx, event); err != nil {
			t.Fatalf("AppendEvent(%s) error = %v", event.Message, err)
		}
	}

	cutoff := base.Add(24 * time.Hour)
	removed, err := st.PruneEventsBefore(ctx, cutoff)
	if err != nil {
		t.Fatalf("PruneEventsBefore() error = %v", err)
	}
	if removed != 1 {
		t.Fatalf("PruneEventsBefore() removed = %d, want 1", removed)
	}

	page, err := st.ListTimelineEvents(ctx, EventFilter{ClusterID: cluster.ID, Limit: 10})
	if err != nil {
		t.Fatalf("ListTimelineEvents() after prune error = %v", err)
	}
	if len(page.Events) != 1 || page.Events[0].Message != "recent" {
		t.Fatalf("ListTimelineEvents() after prune = %#v, want only the recent event", page.Events)
	}

	// Pruning again removes nothing further.
	removed, err = st.PruneEventsBefore(ctx, cutoff)
	if err != nil {
		t.Fatalf("PruneEventsBefore() second call error = %v", err)
	}
	if removed != 0 {
		t.Fatalf("PruneEventsBefore() second call removed = %d, want 0", removed)
	}
}

func TestSQLiteStoreEventsPersistAcrossRestart(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "kfleet.db")
	st, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	ctx := context.Background()
	cluster := types.Cluster{ID: "cluster-1", Name: "production", RegisteredAt: time.Now().UTC(), Labels: map[string]string{}}
	if err := st.CreateCluster(ctx, cluster); err != nil {
		t.Fatalf("CreateCluster() error = %v", err)
	}
	inserted, err := st.AppendEvent(ctx, types.OperationalEvent{
		ClusterID: cluster.ID, Kind: types.EventClusterRegistered, Message: "registered",
		OccurredAt: time.Now().UTC(), DedupeKey: "registered",
	})
	if err != nil || !inserted {
		t.Fatalf("AppendEvent() = (%v, %v), want (true, nil)", inserted, err)
	}
	if err := st.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	// Simulate a hub restart: reopen the same database file.
	reopened, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() reopen error = %v", err)
	}
	t.Cleanup(func() {
		if err := reopened.Close(); err != nil {
			t.Errorf("Close() reopened error = %v", err)
		}
	})

	page, err := reopened.ListTimelineEvents(ctx, EventFilter{ClusterID: cluster.ID, Limit: 10})
	if err != nil {
		t.Fatalf("ListTimelineEvents() after restart error = %v", err)
	}
	if len(page.Events) != 1 || page.Events[0].Kind != types.EventClusterRegistered {
		t.Fatalf("ListTimelineEvents() after restart = %#v, want the persisted registration event", page.Events)
	}
}

func TestSQLiteStoreEventsSurviveClusterDeletionUntilRetention(t *testing.T) {
	t.Parallel()

	st, _ := newEventTestStore(t)
	ctx := context.Background()
	cluster := types.Cluster{
		ID: "cluster-1", Name: "production",
		RegisteredAt: time.Now().UTC(), Labels: map[string]string{},
	}
	if err := st.CreateCluster(ctx, cluster); err != nil {
		t.Fatalf("CreateCluster() error = %v", err)
	}
	if _, err := st.AppendEvent(ctx, types.OperationalEvent{
		ClusterID: cluster.ID, Kind: types.EventClusterRegistered, Message: "registered",
		OccurredAt: cluster.RegisteredAt, DedupeKey: "registered",
	}); err != nil {
		t.Fatalf("AppendEvent() error = %v", err)
	}
	if err := st.DeleteCluster(ctx, cluster.ID); err != nil {
		t.Fatalf("DeleteCluster() error = %v", err)
	}

	page, err := st.ListTimelineEvents(ctx, EventFilter{ClusterID: cluster.ID})
	if err != nil {
		t.Fatalf("ListTimelineEvents() error = %v", err)
	}
	if len(page.Events) != 1 {
		t.Fatalf("events after cluster deletion = %#v, want retained audit event", page.Events)
	}
}

func TestSQLiteStoreHighVolumeEvents(t *testing.T) {
	t.Parallel()

	st, _ := newEventTestStore(t)
	ctx := context.Background()

	cluster := types.Cluster{ID: "cluster-1", Name: "production", RegisteredAt: time.Now().UTC(), Labels: map[string]string{}}
	if err := st.CreateCluster(ctx, cluster); err != nil {
		t.Fatalf("CreateCluster() error = %v", err)
	}

	const total = 500
	base := time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < total; i++ {
		event := types.OperationalEvent{
			ClusterID:  cluster.ID,
			Kind:       types.EventHeartbeatStateChange,
			Message:    fmt.Sprintf("event-%d", i),
			OccurredAt: base.Add(time.Duration(i) * time.Second),
			DedupeKey:  fmt.Sprintf("event-%d", i),
		}
		inserted, err := st.AppendEvent(ctx, event)
		if err != nil {
			t.Fatalf("AppendEvent(%d) error = %v", i, err)
		}
		if !inserted {
			t.Fatalf("AppendEvent(%d) inserted = false, want true", i)
		}
	}

	seen := make(map[int64]bool, total)
	var cursor int64
	count := 0
	for pageIndex := 0; ; pageIndex++ {
		if pageIndex > total {
			t.Fatal("pagination did not terminate")
		}
		page, err := st.ListTimelineEvents(ctx, EventFilter{ClusterID: cluster.ID, Before: cursor, Limit: 100})
		if err != nil {
			t.Fatalf("ListTimelineEvents(page %d) error = %v", pageIndex, err)
		}
		for _, event := range page.Events {
			if seen[event.ID] {
				t.Fatalf("event id %d returned on more than one page", event.ID)
			}
			seen[event.ID] = true
		}
		count += len(page.Events)
		if page.NextCursor == 0 {
			break
		}
		cursor = page.NextCursor
	}
	if count != total {
		t.Fatalf("paged through %d events, want %d", count, total)
	}
}
