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
