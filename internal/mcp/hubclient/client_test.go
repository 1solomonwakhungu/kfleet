package hubclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
