package mcp

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/1solomonwakhungu/kfleet/internal/config"
	"github.com/1solomonwakhungu/kfleet/internal/demo"
	"github.com/1solomonwakhungu/kfleet/internal/server"
	"github.com/1solomonwakhungu/kfleet/pkg/api"
	"github.com/1solomonwakhungu/kfleet/pkg/types"
)

func TestConfigureHubClientDemoModeEndToEnd(t *testing.T) {
	ctx := context.Background()
	st, err := demo.Open(ctx)
	if err != nil {
		t.Fatalf("demo.Open() error = %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	hubServer := server.New(&config.Config{ListenAddr: ":0", DemoMode: true}, logger, st)
	ts := httptest.NewServer(hubServer.Handler())
	t.Cleanup(ts.Close)

	hub, err := configureHubClient(ctx, ts.URL, "", "")
	if err != nil {
		t.Fatalf("configureHubClient(demo) error = %v", err)
	}

	client := newTestClient(t, hub)

	clustersRes := decodeResult[struct {
		Clusters []types.Cluster `json:"clusters"`
		Count    int             `json:"count"`
	}](t, callTool(t, client, "list_clusters", nil))
	if clustersRes.Count != 3 || len(clustersRes.Clusters) != 3 {
		t.Fatalf("list_clusters = %+v, want 3 synthetic clusters", clustersRes)
	}

	crashRes := decodeResult[struct {
		Pods  []crashloopPod `json:"pods"`
		Count int            `json:"count"`
	}](t, callTool(t, client, "find_crashloop_pods", nil))
	if crashRes.Count != 1 || len(crashRes.Pods) != 1 {
		t.Fatalf("find_crashloop_pods = %+v, want 1 crashloop pod", crashRes)
	}
	pod := crashRes.Pods[0]
	if pod.Name != "ledger-api-6f75-demo" || pod.Namespace != "payments" || pod.ClusterName != "demo-eu-west" {
		t.Fatalf("crashloop pod = %+v, want ledger-api-6f75-demo in payments on demo-eu-west", pod)
	}
	if pod.Phase != "CrashLoopBackOff" || pod.RestartCount != 7 {
		t.Fatalf("crashloop pod state = phase:%q restarts:%d, want CrashLoopBackOff / 7", pod.Phase, pod.RestartCount)
	}
}

func TestConfigureHubClientRejectsUnauthenticatedNonDemoHub(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/meta" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(api.RuntimeInfo{DemoMode: false})
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	_, err := configureHubClient(context.Background(), ts.URL, "", "")
	if err == nil || !strings.Contains(err.Error(), "KFLEET_HUB_USERNAME and KFLEET_HUB_PASSWORD are required") {
		t.Fatalf("configureHubClient(non-demo) error = %v, want missing credentials error", err)
	}
}

func TestConfigureHubClientReportsProbeFailureWhenUnauthenticated(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(api.ErrorResponse{Error: "boom"})
	}))
	defer ts.Close()

	_, err := configureHubClient(context.Background(), ts.URL, "", "")
	if err == nil {
		t.Fatal("configureHubClient(probe error) error = nil, want error")
	}
	if !strings.Contains(err.Error(), "KFLEET_HUB_USERNAME and KFLEET_HUB_PASSWORD are required") ||
		!strings.Contains(err.Error(), "boom") {
		t.Fatalf("configureHubClient(probe error) error = %v, want credentials message with probe error", err)
	}
}

func TestConfigureHubClientRejectsPartialCredentials(t *testing.T) {
	metaProbed := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/meta" {
			metaProbed = true
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	for _, tc := range [][2]string{{"reader", ""}, {"", "secret"}} {
		_, err := configureHubClient(context.Background(), ts.URL, tc[0], tc[1])
		if err == nil {
			t.Errorf("configureHubClient(%q, %q) error = nil, want error", tc[0], tc[1])
		}
	}
	if metaProbed {
		t.Fatal("partial credentials unexpectedly probed /api/v1/meta")
	}
}

func TestConfigureHubClientWithFullCredentials(t *testing.T) {
	const cookieValue = "session-123"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/auth/login":
			var req api.LoginRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			if req.Username != "reader" || req.Password != "secret" {
				http.Error(w, "invalid", http.StatusUnauthorized)
				return
			}
			http.SetCookie(w, &http.Cookie{Name: "kfleet_session", Value: cookieValue, Path: "/"})
			_ = json.NewEncoder(w).Encode(api.UserResponse{Username: "reader", Role: types.RoleReadOnly})
		case "/api/v1/clusters":
			c, err := r.Cookie("kfleet_session")
			if err != nil || c.Value != cookieValue {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			_ = json.NewEncoder(w).Encode(api.ListClustersResponse{
				Clusters: []types.Cluster{{ID: "c1", Name: "prod"}},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	hub, err := configureHubClient(context.Background(), ts.URL, "reader", "secret")
	if err != nil {
		t.Fatalf("configureHubClient(full creds) error = %v", err)
	}
	clusters, err := hub.ListClusters(context.Background())
	if err != nil || len(clusters) != 1 || clusters[0].Name != "prod" {
		t.Fatalf("ListClusters() = %+v, %v, want prod", clusters, err)
	}
}

func TestConfigureHubClientRequiresURL(t *testing.T) {
	if _, err := configureHubClient(context.Background(), "   ", "", ""); err == nil {
		t.Fatal("configureHubClient(empty URL) error = nil, want error")
	}
}
