package reporter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/1solomonwakhungu/kfleet/internal/agent/collector"
	"github.com/1solomonwakhungu/kfleet/internal/agent/config"
	"github.com/1solomonwakhungu/kfleet/internal/agent/hubcontact"
)

func TestReport(t *testing.T) {
	var received collector.ClusterState
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/clusters/production/status" {
			t.Errorf("request = %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer agent-token" {
			t.Errorf("Authorization = %q", got)
		}
		if got := r.Header.Get("X-Kfleet-Tenant-ID"); got != "tenant-a" {
			t.Errorf("X-Kfleet-Tenant-ID = %q, want tenant-a", got)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	r := New(&config.Config{HubURL: server.URL, ClusterName: "production", HubToken: "agent-token", TenantID: "tenant-a"}, nil)
	want := &collector.ClusterState{NodeCount: 2, PodCount: 10, K8sVersion: "v1.32.0"}
	if err := r.Report(context.Background(), want); err != nil {
		t.Fatalf("Report() error = %v", err)
	}
	if received.NodeCount != want.NodeCount || received.PodCount != want.PodCount {
		t.Fatalf("reported state = %#v, want %#v", received, want)
	}
}

func TestReportReturnsStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()
	r := New(&config.Config{HubURL: server.URL, ClusterName: "production"}, nil)
	if err := r.Report(context.Background(), &collector.ClusterState{}); err == nil {
		t.Fatal("Report() error = nil, want status error")
	}
}

// TestReportSurfacesHubErrorBody proves a rejected report carries the hub's
// explanation so operators can diagnose it from agent logs.
func TestReportSurfacesHubErrorBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"agent is pending approval","code":403}`))
	}))
	defer server.Close()
	r := New(&config.Config{HubURL: server.URL, ClusterName: "production"}, nil)
	err := r.Report(context.Background(), &collector.ClusterState{})
	if err == nil {
		t.Fatal("Report() error = nil, want status error")
	}
	if !strings.Contains(err.Error(), "hub returned status 403 Forbidden") {
		t.Errorf("Report() error = %q, want it to describe the status", err)
	}
	if !strings.Contains(err.Error(), "agent is pending approval") {
		t.Errorf("Report() error = %q, want it to carry the hub's error body", err)
	}
}

// TestReportRefreshesHubContact proves a successful report refreshes the
// readiness tracker while a failed report leaves it stale, so the readiness
// probe reflects real hub reachability.
func TestReportRefreshesHubContact(t *testing.T) {
	tracker := hubcontact.NewTracker(time.Minute)
	clock := time.Now()
	tracker.SetClockForTesting(func() time.Time { return clock })
	tracker.MarkRegistered()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/status") && r.Header.Get("Authorization") == "Bearer accepted-token" {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Error(w, "hub unavailable", http.StatusInternalServerError)
	}))
	defer server.Close()

	r := New(&config.Config{HubURL: server.URL, ClusterName: "production", HubToken: "accepted-token"}, tracker)

	clock = clock.Add(2 * time.Minute)
	if tracker.Ready() {
		t.Fatal("tracker ready before the report, want stale")
	}
	if err := r.Report(context.Background(), &collector.ClusterState{}); err != nil {
		t.Fatalf("Report() error = %v", err)
	}
	if !tracker.Ready() {
		t.Fatal("tracker not ready after a successful report, want ready")
	}

	rejecting := New(&config.Config{HubURL: server.URL, ClusterName: "production", HubToken: "stale-token"}, tracker)
	if err := rejecting.Report(context.Background(), &collector.ClusterState{}); err == nil {
		t.Fatal("Report() error = nil, want status error for the rejected token")
	}
	if !tracker.Ready() {
		t.Fatal("tracker lost freshness after a failed report, want the earlier successful contact kept")
	}
}
