package registrar

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/1solomonwakhungu/kfleet/internal/agent/config"
)

func TestRegisterKeepsBootstrapTokenAfterRuntimeTokenRotation(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if got := r.Header.Get("Authorization"); got != "Bearer bootstrap-token" {
			t.Errorf("registration %d Authorization = %q, want bootstrap token", requests, got)
		}
		if got := r.Header.Get("X-Kfleet-Tenant-ID"); got != "tenant-a" {
			t.Errorf("registration %d tenant = %q, want tenant-a", requests, got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"clusterId":"cluster-a","token":"runtime-token"}`))
	}))
	t.Cleanup(server.Close)

	registrar := New(&config.Config{
		HubURL:      server.URL,
		ClusterName: "cluster-a",
		HubToken:    "bootstrap-token",
		TenantID:    "tenant-a",
	}, nil)
	if _, err := registrar.Register(context.Background(), "v1.32.3"); err != nil {
		t.Fatalf("first Register() error = %v", err)
	}
	if _, err := registrar.Register(context.Background(), "v1.32.3"); err != nil {
		t.Fatalf("second Register() error = %v", err)
	}
	if requests != 2 || registrar.Token() != "runtime-token" {
		t.Fatalf("requests/token = (%d, %q), want (2, runtime-token)", requests, registrar.Token())
	}
}

// TestHeartbeatSignalsApprovalOnce simulates a pending agent whose operator
// approves it in the hub: heartbeats first report approved=false, then
// approved=true, and exactly one re-registration must follow.
func TestHeartbeatSignalsApprovalOnce(t *testing.T) {
	var mu sync.Mutex
	registrations, heartbeats := 0, 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		switch {
		case strings.HasSuffix(r.URL.Path, "/register"):
			registrations++
			w.Header().Set("Content-Type", "application/json")
			if registrations == 1 {
				w.WriteHeader(http.StatusAccepted)
				_, _ = w.Write([]byte(`{"clusterId":"cluster-a","token":"pending-token"}`))
				return
			}
			_, _ = w.Write([]byte(`{"clusterId":"cluster-a","token":"approved-token"}`))
		case strings.HasSuffix(r.URL.Path, "/heartbeat"):
			heartbeats++
			w.Header().Set("Content-Type", "application/json")
			approved := "false"
			if heartbeats >= 3 {
				approved = "true"
			}
			_, _ = w.Write([]byte(`{"clusterId":"cluster-a","approved":` + approved + `}`))
		default:
			t.Errorf("unexpected request to %s", r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)

	registrar := New(&config.Config{
		HubURL:      server.URL,
		ClusterName: "cluster-a",
		HubToken:    "bootstrap-token",
	}, nil)

	first, err := registrar.Register(context.Background(), "v1.32.3")
	if err != nil {
		t.Fatalf("first Register() error = %v", err)
	}
	if first.Approved {
		t.Fatal("first Register() reported approved, want pending")
	}

	for i := 0; i < 2; i++ {
		reregister, err := registrar.Heartbeat(context.Background())
		if err != nil {
			t.Fatalf("pending Heartbeat() %d error = %v", i+1, err)
		}
		if reregister {
			t.Fatalf("pending Heartbeat() %d signaled re-registration, want no signal before approval", i+1)
		}
	}

	reregister, err := registrar.Heartbeat(context.Background())
	if err != nil {
		t.Fatalf("approved Heartbeat() error = %v", err)
	}
	if !reregister {
		t.Fatal("approved Heartbeat() did not signal re-registration")
	}

	second, err := registrar.Register(context.Background(), "v1.32.3")
	if err != nil {
		t.Fatalf("re-registration Register() error = %v", err)
	}
	if !second.Approved {
		t.Fatal("re-registration Register() reported pending, want approved")
	}

	reregister, err = registrar.Heartbeat(context.Background())
	if err != nil {
		t.Fatalf("post-approval Heartbeat() error = %v", err)
	}
	if reregister {
		t.Fatal("post-approval Heartbeat() signaled a second re-registration")
	}

	mu.Lock()
	defer mu.Unlock()
	if registrations != 2 {
		t.Errorf("registrations = %d, want exactly 2", registrations)
	}
	if heartbeats != 4 {
		t.Errorf("heartbeats = %d, want 4", heartbeats)
	}
	if got := registrar.Token(); got != "approved-token" {
		t.Errorf("token after re-registration = %q, want approved-token", got)
	}
}

// TestHeartbeatAfterApprovedRegistrationNeverSignals proves an agent that
// registered as approved keeps heartbeating without re-registering.
func TestHeartbeatAfterApprovedRegistrationNeverSignals(t *testing.T) {
	var mu sync.Mutex
	registrations := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/heartbeat") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"clusterId":"cluster-a","approved":true}`))
			return
		}
		mu.Lock()
		defer mu.Unlock()
		registrations++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"clusterId":"cluster-a","token":"approved-token"}`))
	}))
	t.Cleanup(server.Close)

	registrar := New(&config.Config{
		HubURL:      server.URL,
		ClusterName: "cluster-a",
		HubToken:    "bootstrap-token",
	}, nil)
	if _, err := registrar.Register(context.Background(), "v1.32.3"); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	for i := 0; i < 3; i++ {
		reregister, err := registrar.Heartbeat(context.Background())
		if err != nil {
			t.Fatalf("Heartbeat() %d error = %v", i+1, err)
		}
		if reregister {
			t.Fatalf("Heartbeat() %d signaled re-registration for an approved agent", i+1)
		}
	}
	mu.Lock()
	defer mu.Unlock()
	if registrations != 1 {
		t.Errorf("registrations = %d, want 1", registrations)
	}
}

// TestHeartbeatServerErrorReturnsError proves a failed heartbeat is reported
// as an error without signaling approval.
func TestHeartbeatServerErrorReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "hub unavailable", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	registrar := New(&config.Config{
		HubURL:      server.URL,
		ClusterName: "cluster-a",
		HubToken:    "bootstrap-token",
	}, nil)
	reregister, err := registrar.Heartbeat(context.Background())
	if err == nil {
		t.Fatal("Heartbeat() error = nil, want error for server failure")
	}
	if reregister {
		t.Error("Heartbeat() signaled re-registration on server failure")
	}
}
