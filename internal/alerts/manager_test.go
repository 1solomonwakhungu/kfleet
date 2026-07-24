package alerts

import (
	"context"
	"crypto/subtle"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/1solomonwakhungu/kfleet/internal/store"
	"github.com/1solomonwakhungu/kfleet/pkg/types"
)

func TestManagerCooldownSignedRetryAndResolution(t *testing.T) {
	st := openTestStore(t)
	ctx := context.Background()
	cluster := types.Cluster{
		ID:           "cluster-a",
		Name:         "production",
		Health:       types.HealthDegraded,
		RegisteredAt: time.Now().UTC(),
		Labels:       map[string]string{},
	}
	if err := st.CreateCluster(ctx, cluster); err != nil {
		t.Fatalf("CreateCluster() error = %v", err)
	}

	const secret = "test-secret"
	var mu sync.Mutex
	requests := 0
	signaturesValid := true
	receiver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read webhook body: %v", err)
		}
		timestamp := r.Header.Get("X-Kfleet-Timestamp")
		want := Signature(secret, timestamp, body)
		got := r.Header.Get("X-Kfleet-Signature")
		if len(want) != len(got) || subtle.ConstantTimeCompare([]byte(want), []byte(got)) != 1 {
			mu.Lock()
			signaturesValid = false
			mu.Unlock()
		}
		if r.Header.Get("X-Kfleet-Event") != EventType ||
			r.Header.Get("X-Kfleet-Delivery") == "" ||
			r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("webhook headers = %#v", r.Header)
		}
		mu.Lock()
		requests++
		currentRequest := requests
		mu.Unlock()
		if currentRequest == 1 {
			http.Error(w, "injected failure", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(receiver.Close)

	now := time.Date(2026, time.July, 23, 12, 0, 0, 0, time.UTC)
	manager := New(st, discardLogger(), Config{
		WebhookURL:  receiver.URL,
		Secret:      secret,
		MaxAttempts: 3,
		RetryBase:   time.Second,
		HTTPClient:  receiver.Client(),
	})
	manager.now = func() time.Time { return now }

	manager.Evaluate(ctx, cluster)
	manager.Evaluate(ctx, cluster)
	history, err := st.ListAlerts(ctx, "", 100)
	if err != nil {
		t.Fatalf("ListAlerts() error = %v", err)
	}
	if len(history) != 1 {
		t.Fatalf("alerts after duplicate evaluations = %d, want 1", len(history))
	}

	manager.ProcessDue(ctx)
	afterFailure, err := st.GetAlert(ctx, history[0].ID)
	if err != nil {
		t.Fatalf("GetAlert() after failure error = %v", err)
	}
	if afterFailure.DeliveryStatus != types.AlertDeliveryRetrying ||
		afterFailure.DeliveryAttempts != 1 ||
		afterFailure.NextDeliveryAt == nil ||
		afterFailure.LastDeliveryError == "" {
		t.Fatalf("alert after failure = %#v", afterFailure)
	}

	now = now.Add(time.Second)
	manager.ProcessDue(ctx)
	delivered, err := st.GetAlert(ctx, history[0].ID)
	if err != nil {
		t.Fatalf("GetAlert() after retry error = %v", err)
	}
	if delivered.DeliveryStatus != types.AlertDeliveryDelivered ||
		delivered.DeliveryAttempts != 2 ||
		delivered.DeliveredAt == nil ||
		delivered.NextDeliveryAt != nil {
		t.Fatalf("delivered alert = %#v", delivered)
	}
	mu.Lock()
	valid := signaturesValid
	mu.Unlock()
	if !valid {
		t.Fatal("webhook signature was invalid")
	}

	cluster.Health = types.HealthHealthy
	now = now.Add(time.Minute)
	manager.Evaluate(ctx, cluster)
	resolved, err := st.GetAlert(ctx, history[0].ID)
	if err != nil {
		t.Fatalf("GetAlert() after resolution error = %v", err)
	}
	if resolved.Status != types.AlertStatusResolved || resolved.ResolvedAt == nil {
		t.Fatalf("resolved alert = %#v", resolved)
	}
	if err := st.DeleteCluster(ctx, cluster.ID); err != nil {
		t.Fatalf("DeleteCluster() error = %v", err)
	}
	if _, err := st.GetAlert(ctx, history[0].ID); err != nil {
		t.Fatalf("GetAlert() after cluster deletion error = %v, want retained history", err)
	}
}

func TestManagerDeadLettersAfterMaximumAttempts(t *testing.T) {
	st := openTestStore(t)
	ctx := context.Background()
	cluster := types.Cluster{
		ID:           "cluster-b",
		Name:         "edge",
		Health:       types.HealthUnreachable,
		RegisteredAt: time.Now().UTC(),
		Labels:       map[string]string{},
	}
	if err := st.CreateCluster(ctx, cluster); err != nil {
		t.Fatalf("CreateCluster() error = %v", err)
	}

	receiver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "still failing", http.StatusBadGateway)
	}))
	t.Cleanup(receiver.Close)
	now := time.Date(2026, time.July, 23, 12, 0, 0, 0, time.UTC)
	manager := New(st, discardLogger(), Config{
		WebhookURL:  receiver.URL,
		Secret:      "secret",
		MaxAttempts: 2,
		RetryBase:   10 * time.Millisecond,
		HTTPClient:  receiver.Client(),
	})
	manager.now = func() time.Time { return now }

	manager.Evaluate(ctx, cluster)
	manager.ProcessDue(ctx)
	now = now.Add(10 * time.Millisecond)
	manager.ProcessDue(ctx)

	history, err := st.ListAlerts(ctx, "", 100)
	if err != nil {
		t.Fatalf("ListAlerts() error = %v", err)
	}
	if len(history) != 1 ||
		history[0].DeliveryStatus != types.AlertDeliveryDeadLetter ||
		history[0].DeliveryAttempts != 2 ||
		history[0].DeadLetteredAt == nil ||
		history[0].NextDeliveryAt != nil {
		t.Fatalf("dead-letter alert = %#v", history)
	}
}

func TestSignatureKnownValue(t *testing.T) {
	got := Signature("secret", strconv.FormatInt(1_721_736_000, 10), []byte(`{"event":"test"}`))
	const want = "v1=98a75d8b3da6f12e693ffd7882f84e4c8eba7079db32d9fe53f0df4d3c6bc1e0"
	if got != want {
		t.Fatalf("Signature() = %q, want %q", got, want)
	}
}

func openTestStore(t *testing.T) store.Store {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "kfleet.db"))
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := st.Close(); err != nil {
			t.Errorf("store.Close() error = %v", err)
		}
	})
	return st
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
