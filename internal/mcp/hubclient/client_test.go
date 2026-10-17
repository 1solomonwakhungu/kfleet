package hubclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/1solomonwakhungu/kfleet/pkg/api"
	"github.com/1solomonwakhungu/kfleet/pkg/types"
)

func TestClientRequests(t *testing.T) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization = %q, want bearer token", got)
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/clusters":
			_ = json.NewEncoder(w).Encode(api.ListClustersResponse{Clusters: []types.Cluster{{ID: "cluster-1"}}})
		case "/api/v1/clusters/cluster-1/status":
			_ = json.NewEncoder(w).Encode(api.ClusterStatusResponse{Cluster: types.Cluster{ID: "cluster-1"}})
		case "/api/v1/clusters/cluster-1/pods":
			if got := r.URL.Query().Get("namespace"); got != "apps" {
				t.Errorf("pod namespace = %q, want apps", got)
			}
			_ = json.NewEncoder(w).Encode([]types.Pod{{Name: "api"}})
		case "/api/v1/clusters/cluster-1/events":
			if got := r.URL.Query().Get("namespace"); got != "apps" {
				t.Errorf("event namespace = %q, want apps", got)
			}
			_ = json.NewEncoder(w).Encode([]types.Event{{Reason: "Started"}})
		case "/api/v1/clusters/cluster-1/namespaces":
			_ = json.NewEncoder(w).Encode([]string{"apps", "default"})
		case "/api/v1/clusters/cluster-1/timeline":
			if got := r.URL.Query().Get("limit"); got != "10" {
				t.Errorf("timeline limit = %q, want 10", got)
			}
			_ = json.NewEncoder(w).Encode(api.ListTimelineEventsResponse{
				Events:     []types.OperationalEvent{{ID: 2, ClusterID: "cluster-1", Kind: types.EventAgentApproved}},
				NextCursor: 1,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, err := New(server.URL, "test-token")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	ctx := context.Background()
	clusters, err := client.ListClusters(ctx)
	if err != nil || len(clusters) != 1 {
		t.Fatalf("ListClusters() = %#v, %v", clusters, err)
	}
	status, err := client.GetClusterStatus(ctx, "cluster-1")
	if err != nil || status.Cluster.ID != "cluster-1" {
		t.Fatalf("GetClusterStatus() = %#v, %v", status, err)
	}
	pods, err := client.GetPods(ctx, "cluster-1", "apps")
	if err != nil || len(pods) != 1 {
		t.Fatalf("GetPods() = %#v, %v", pods, err)
	}
	events, err := client.GetEvents(ctx, "cluster-1", "apps")
	if err != nil || len(events) != 1 {
		t.Fatalf("GetEvents() = %#v, %v", events, err)
	}
	namespaces, err := client.ListNamespaces(ctx, "cluster-1")
	if err != nil || len(namespaces) != 2 || namespaces[0] != "apps" {
		t.Fatalf("ListNamespaces() = %#v, %v", namespaces, err)
	}
	timeline, err := client.GetTimeline(ctx, TimelineQuery{ClusterID: "cluster-1", Limit: 10})
	if err != nil || len(timeline.Events) != 1 || timeline.Events[0].Kind != types.EventAgentApproved || timeline.NextCursor != 1 {
		t.Fatalf("GetTimeline() = %#v, %v", timeline, err)
	}
}

func TestNewRejectsInvalidURL(t *testing.T) {
	for _, value := range []string{"", "localhost:8080", "://bad"} {
		if _, err := New(value, ""); err == nil {
			t.Errorf("New(%q) error = nil, want error", value)
		}
	}
}

func TestClientAuthenticatesWithCredentials(t *testing.T) {
	const sessionCookie = "session-token"
	var loginCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/auth/login":
			loginCount++
			var login api.LoginRequest
			if err := json.NewDecoder(r.Body).Decode(&login); err != nil {
				t.Fatalf("decode login request error = %v", err)
			}
			if login.Username != "reader" || login.Password != "correct-password" {
				t.Errorf("login = %+v, want configured credentials", login)
			}
			http.SetCookie(w, &http.Cookie{Name: "kfleet_session", Value: sessionCookie, Path: "/"})
			_ = json.NewEncoder(w).Encode(api.UserResponse{Username: "reader", Role: types.RoleReadOnly})
		case "/api/v1/clusters":
			cookie, err := r.Cookie("kfleet_session")
			if err != nil || cookie.Value != sessionCookie {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			_ = json.NewEncoder(w).Encode(api.ListClustersResponse{Clusters: []types.Cluster{{ID: "cluster-1"}}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, err := NewWithCredentials(server.URL, "reader", "correct-password")
	if err != nil {
		t.Fatalf("NewWithCredentials() error = %v", err)
	}
	for i := 0; i < 2; i++ {
		clusters, err := client.ListClusters(context.Background())
		if err != nil || len(clusters) != 1 {
			t.Fatalf("ListClusters() = %#v, %v", clusters, err)
		}
	}
	if loginCount != 1 {
		t.Fatalf("login requests = %d, want one reused session", loginCount)
	}
}

func TestNewWithCredentialsRequiresBothValues(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()
	for _, credentials := range [][2]string{{"", "password"}, {"reader", ""}} {
		if _, err := NewWithCredentials(server.URL, credentials[0], credentials[1]); err == nil {
			t.Errorf("NewWithCredentials(%q, %q) error = nil, want error", credentials[0], credentials[1])
		}
	}
}

func TestGetPodLogsReadsSSEStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/clusters/cluster-1/pods/apps/api-7f9/logs" {
			http.NotFound(w, r)
			return
		}
		if got := r.Header.Get("Accept"); got != "text/event-stream" {
			t.Errorf("Accept = %q, want text/event-stream", got)
		}
		query := r.URL.Query()
		want := map[string]string{"follow": "false", "container": "app", "tailLines": "50", "previous": "true", "sinceSeconds": "120"}
		for key, value := range want {
			if got := query.Get(key); got != value {
				t.Errorf("query %s = %q, want %q", key, got, value)
			}
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = fmt.Fprint(w, ": keep-alive\n\ndata: starting\n\ndata: panic: boom\ndata:   at main.go:12\n\ndata:\n\nevent: end\ndata: {}\n\n")
	}))
	defer server.Close()

	client, err := New(server.URL, "")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	lines, err := client.GetPodLogs(context.Background(), PodLogsQuery{
		ClusterID: "cluster-1", Namespace: "apps", Pod: "api-7f9", Container: "app",
		TailLines: 50, Previous: true, SinceSeconds: 120,
	})
	if err != nil {
		t.Fatalf("GetPodLogs() error = %v", err)
	}
	want := []string{"starting", "panic: boom\n  at main.go:12", ""}
	if strings.Join(lines, "|") != strings.Join(want, "|") || len(lines) != len(want) {
		t.Fatalf("GetPodLogs() = %#v, want %#v", lines, want)
	}
}

func TestGetPodLogsOmitsUnsetOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		for _, key := range []string{"container", "tailLines", "previous", "sinceSeconds"} {
			if query.Has(key) {
				t.Errorf("query has %s = %q, want it omitted", key, query.Get(key))
			}
		}
		_, _ = fmt.Fprint(w, "event: end\ndata: {}\n\n")
	}))
	defer server.Close()

	client, err := New(server.URL, "")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	lines, err := client.GetPodLogs(context.Background(), PodLogsQuery{ClusterID: "cluster-1", Namespace: "apps", Pod: "api"})
	if err != nil || len(lines) != 0 {
		t.Fatalf("GetPodLogs() = %#v, %v, want no lines", lines, err)
	}
}

func TestGetPodLogsErrors(t *testing.T) {
	tests := map[string]struct {
		handler http.HandlerFunc
		want    string
	}{
		"stream error event": {
			handler: func(w http.ResponseWriter, _ *http.Request) {
				_, _ = fmt.Fprint(w, "data: partial\n\nevent: stream-error\ndata: {\"message\":\"previous terminated container \\\"app\\\" not found\"}\n\n")
			},
			want: `pod log stream failed: previous terminated container "app" not found`,
		},
		"no agent connected": {
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusServiceUnavailable)
				_ = json.NewEncoder(w).Encode(api.ErrorResponse{Error: "no agent is connected for this cluster"})
			},
			want: "no agent is connected for this cluster",
		},
		"stream cut short": {
			handler: func(w http.ResponseWriter, _ *http.Request) {
				_, _ = fmt.Fprint(w, "data: partial\n\n")
			},
			want: "ended before the hub finished it",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()
			client, err := New(server.URL, "")
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}
			_, err = client.GetPodLogs(context.Background(), PodLogsQuery{ClusterID: "cluster-1", Namespace: "apps", Pod: "api"})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("GetPodLogs() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestGetPodLogsRequiresNamespaceAndPod(t *testing.T) {
	client, err := New("http://hub.invalid", "")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if _, err := client.GetPodLogs(context.Background(), PodLogsQuery{ClusterID: "cluster-1", Namespace: " "}); err == nil {
		t.Fatal("GetPodLogs() error = nil, want missing namespace/pod error")
	}
}
