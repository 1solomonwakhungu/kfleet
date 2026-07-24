package server

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/1solomonwakhungu/kfleet/internal/config"
	"github.com/1solomonwakhungu/kfleet/internal/store"
	"github.com/1solomonwakhungu/kfleet/pkg/api"
	"github.com/1solomonwakhungu/kfleet/pkg/types"
)

func TestClusterAndFleetTimelineEndpoints(t *testing.T) {
	server, registration := registeredAgent(t)
	approveAgent(t, server, registration.ClusterID)

	timelineResp := request(t, server, http.MethodGet, "/api/v1/clusters/"+registration.ClusterID+"/timeline", "")
	if timelineResp.StatusCode != http.StatusOK {
		t.Fatalf("timeline status = %d, want %d", timelineResp.StatusCode, http.StatusOK)
	}
	var page api.ListTimelineEventsResponse
	decodeResponse(t, timelineResp, &page)
	if len(page.Events) != 2 {
		t.Fatalf("timeline events = %#v, want 2 (registered + approved)", page.Events)
	}
	if page.Events[0].Kind != types.EventAgentApproved || page.Events[1].Kind != types.EventClusterRegistered {
		t.Fatalf("timeline order = %#v, want [approved, registered] (newest first)", page.Events)
	}
	for _, event := range page.Events {
		if event.ClusterID != registration.ClusterID {
			t.Fatalf("event.ClusterID = %q, want %q", event.ClusterID, registration.ClusterID)
		}
	}

	// A snapshot with a degraded node triggers a heartbeat state change.
	posted := agentRequest(t, server, http.MethodPost, "/api/v1/clusters/"+registration.ClusterID+"/status", registration.Token, snapshotPayload())
	if posted.StatusCode != http.StatusOK {
		t.Fatalf("snapshot status = %d, want %d", posted.StatusCode, http.StatusOK)
	}
	posted.Body.Close()

	fleetResp := request(t, server, http.MethodGet, "/api/v1/timeline", "")
	if fleetResp.StatusCode != http.StatusOK {
		t.Fatalf("fleet timeline status = %d, want %d", fleetResp.StatusCode, http.StatusOK)
	}
	var fleetPage api.ListTimelineEventsResponse
	decodeResponse(t, fleetResp, &fleetPage)
	if len(fleetPage.Events) != 3 {
		t.Fatalf("fleet timeline events = %#v, want 3", fleetPage.Events)
	}
	if fleetPage.Events[0].Kind != types.EventHeartbeatStateChange {
		t.Fatalf("fleet timeline events[0].Kind = %q, want %q", fleetPage.Events[0].Kind, types.EventHeartbeatStateChange)
	}

	// A later snapshot with a new Kubernetes version records one version
	// transition without duplicating the unchanged health state.
	versionPayload := strings.ReplaceAll(snapshotPayload(), "v1.31.1", "v1.32.0")
	versioned := agentRequest(t, server, http.MethodPost, "/api/v1/clusters/"+registration.ClusterID+"/status", registration.Token, versionPayload)
	if versioned.StatusCode != http.StatusOK {
		t.Fatalf("version snapshot status = %d, want %d", versioned.StatusCode, http.StatusOK)
	}
	versioned.Body.Close()
	versionTimeline := request(t, server, http.MethodGet, "/api/v1/clusters/"+registration.ClusterID+"/timeline", "")
	var versionPage api.ListTimelineEventsResponse
	decodeResponse(t, versionTimeline, &versionPage)
	if versionPage.Events[0].Kind != types.EventVersionChanged ||
		versionPage.Events[0].Details["from"] != "v1.31.1" ||
		versionPage.Events[0].Details["to"] != "v1.32.0" {
		t.Fatalf("latest version event = %#v, want v1.31.1 to v1.32.0", versionPage.Events[0])
	}

	// Pagination: limit=1 leaves a next cursor.
	limitedResp := request(t, server, http.MethodGet, "/api/v1/timeline?limit=1", "")
	var limitedPage api.ListTimelineEventsResponse
	decodeResponse(t, limitedResp, &limitedPage)
	if len(limitedPage.Events) != 1 || limitedPage.NextCursor == 0 {
		t.Fatalf("limited page = %#v, want 1 event with a next cursor", limitedPage)
	}
	nextResp := request(t, server, http.MethodGet, "/api/v1/timeline?limit=1&before="+strconv.FormatInt(limitedPage.NextCursor, 10), "")
	var nextPage api.ListTimelineEventsResponse
	decodeResponse(t, nextResp, &nextPage)
	if len(nextPage.Events) != 1 || nextPage.Events[0].ID == limitedPage.Events[0].ID {
		t.Fatalf("next page = %#v, want a distinct event from %#v", nextPage, limitedPage)
	}

	// since/until rejects malformed timestamps.
	badTime := request(t, server, http.MethodGet, "/api/v1/timeline?since=not-a-time", "")
	defer badTime.Body.Close()
	if badTime.StatusCode != http.StatusBadRequest {
		t.Fatalf("bad since status = %d, want %d", badTime.StatusCode, http.StatusBadRequest)
	}
}

func TestClusterTimelineUnknownCluster(t *testing.T) {
	server := newTestHTTPServer(t)
	response := request(t, server, http.MethodGet, "/api/v1/clusters/missing/timeline", "")
	defer response.Body.Close()
	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("unknown cluster timeline status = %d, want %d", response.StatusCode, http.StatusNotFound)
	}
}

func TestRecordPolicyFindingIsDeduped(t *testing.T) {
	server, registration := registeredAgent(t)
	approveAgent(t, server, registration.ClusterID)

	body := `{
		"ruleId": "no-priv-escalation",
		"resource": "pod/default/api",
		"severity": "high",
		"message": "privilege escalation allowed",
		"occurredAt": "2026-07-20T00:00:00Z"
	}`
	created := agentRequest(t, server, http.MethodPost, "/api/v1/clusters/"+registration.ClusterID+"/policy-findings", registration.Token, body)
	if created.StatusCode != http.StatusCreated {
		t.Fatalf("policy finding status = %d, want %d", created.StatusCode, http.StatusCreated)
	}
	created.Body.Close()

	// The identical finding, resubmitted (e.g. a retried scan), is suppressed
	// but still reports success.
	duplicate := agentRequest(t, server, http.MethodPost, "/api/v1/clusters/"+registration.ClusterID+"/policy-findings", registration.Token, body)
	if duplicate.StatusCode != http.StatusOK {
		t.Fatalf("duplicate policy finding status = %d, want %d", duplicate.StatusCode, http.StatusOK)
	}
	duplicate.Body.Close()

	missingFields := agentRequest(t, server, http.MethodPost, "/api/v1/clusters/"+registration.ClusterID+"/policy-findings", registration.Token, `{"ruleId":"x"}`)
	defer missingFields.Body.Close()
	if missingFields.StatusCode != http.StatusBadRequest {
		t.Fatalf("missing fields status = %d, want %d", missingFields.StatusCode, http.StatusBadRequest)
	}

	timelineResp := request(t, server, http.MethodGet, "/api/v1/clusters/"+registration.ClusterID+"/timeline", "")
	var page api.ListTimelineEventsResponse
	decodeResponse(t, timelineResp, &page)
	findings := 0
	for _, event := range page.Events {
		if event.Kind == types.EventPolicyFinding {
			findings++
			if event.Details["ruleId"] != "no-priv-escalation" || event.Details["resource"] != "pod/default/api" {
				t.Fatalf("finding details = %#v, want ruleId/resource populated", event.Details)
			}
		}
	}
	if findings != 1 {
		t.Fatalf("policy finding events = %d, want 1 (duplicate suppressed)", findings)
	}
}

func TestPolicyFindingRequiresApprovedAgentToken(t *testing.T) {
	server, registration := registeredAgent(t)
	path := "/api/v1/clusters/" + registration.ClusterID + "/policy-findings"
	body := `{"ruleId":"rule","resource":"pod/default/api","message":"finding"}`

	for name, testCase := range map[string]struct {
		token string
		want  int
	}{
		"missing": {"", http.StatusUnauthorized},
		"invalid": {"wrong", http.StatusUnauthorized},
		"pending": {registration.Token, http.StatusForbidden},
	} {
		t.Run(name, func(t *testing.T) {
			response := agentRequest(t, server, http.MethodPost, path, testCase.token, body)
			defer response.Body.Close()
			if response.StatusCode != testCase.want {
				t.Fatalf("status = %d, want %d", response.StatusCode, testCase.want)
			}
		})
	}
}

func TestTimelineValidatesTimeWindowAndLimit(t *testing.T) {
	server := newTestHTTPServer(t)
	cases := []string{
		"/api/v1/timeline?limit=501",
		"/api/v1/timeline?since=2026-07-21T00:00:00Z&until=2026-07-20T00:00:00Z",
		"/api/v1/timeline?since=2026-07-20T00:00:00Z&until=2026-07-20T00:00:00Z",
	}
	for _, path := range cases {
		response := request(t, server, http.MethodGet, path, "")
		response.Body.Close()
		if response.StatusCode != http.StatusBadRequest {
			t.Errorf("GET %s status = %d, want %d", path, response.StatusCode, http.StatusBadRequest)
		}
	}
}

func TestPolicyFindingRejectsOversizedBody(t *testing.T) {
	server, registration := registeredAgent(t)
	approveAgent(t, server, registration.ClusterID)
	body := `{"ruleId":"rule","resource":"pod/default/api","message":"` + strings.Repeat("x", maxPolicyFindingBodyBytes) + `"}`
	response := agentRequest(t, server, http.MethodPost, "/api/v1/clusters/"+registration.ClusterID+"/policy-findings", registration.Token, body)
	defer response.Body.Close()
	if response.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusRequestEntityTooLarge)
	}
}

func TestStaleAndReconnectEvents(t *testing.T) {
	httpServer, srv, st := newAgentTestServer(t, time.Second)
	ctx := context.Background()
	registered := agentRequest(t, httpServer, http.MethodPost, "/api/v1/agents/register", "", `{"name":"stale"}`)
	var registration api.RegisterClusterResponse
	decodeResponse(t, registered, &registration)
	approveAgent(t, httpServer, registration.ClusterID)
	body := `{"clusterId":"` + registration.ClusterID + `","nodeCount":1,"healthyNodes":1,"podCount":2,"version":"v1.31.1"}`
	heartbeat := agentRequest(t, httpServer, http.MethodPost, "/api/v1/agents/heartbeat", registration.Token, body)
	heartbeat.Body.Close()
	if heartbeat.StatusCode != http.StatusOK {
		t.Fatalf("initial heartbeat status = %d, want %d", heartbeat.StatusCode, http.StatusOK)
	}

	if err := st.UpdateHealth(ctx, registration.ClusterID, types.HealthHealthy, time.Now().UTC().Add(-4*time.Second)); err != nil {
		t.Fatalf("age heartbeat error = %v", err)
	}
	cluster, err := st.GetCluster(ctx, registration.ClusterID)
	if err != nil {
		t.Fatalf("GetCluster() error = %v", err)
	}
	now := time.Now().UTC()
	srv.markStaleClusters(ctx, now)
	page, err := st.ListTimelineEvents(ctx, store.EventFilter{ClusterID: cluster.ID})
	if err != nil {
		t.Fatalf("ListTimelineEvents() error = %v", err)
	}
	if page.Events[0].Kind != types.EventHeartbeatStateChange || page.Events[1].Kind != types.EventAgentDisconnected {
		t.Fatalf("stale event kinds = [%s, %s], want heartbeat_state_change then agent_disconnected", page.Events[0].Kind, page.Events[1].Kind)
	}
	if page.Events[1].Details["reason"] != "heartbeat_timeout" {
		t.Fatalf("disconnect details = %#v, want heartbeat_timeout reason", page.Events[1].Details)
	}

	reconnected := agentRequest(t, httpServer, http.MethodPost, "/api/v1/agents/heartbeat", registration.Token, body)
	reconnected.Body.Close()
	if reconnected.StatusCode != http.StatusOK {
		t.Fatalf("reconnect heartbeat status = %d, want %d", reconnected.StatusCode, http.StatusOK)
	}
	page, err = st.ListTimelineEvents(ctx, store.EventFilter{ClusterID: cluster.ID})
	if err != nil {
		t.Fatalf("ListTimelineEvents() after reconnect error = %v", err)
	}
	if page.Events[0].Kind != types.EventAgentReconnected {
		t.Fatalf("latest event kind = %s, want %s", page.Events[0].Kind, types.EventAgentReconnected)
	}
}

func TestConfiguredTimelineRetentionPrunesExpiredEvents(t *testing.T) {
	_, srv, st := newAgentTestServerWithConfig(t, &config.Config{
		ListenAddr:     ":0",
		EventRetention: 24 * time.Hour,
	})
	ctx := context.Background()
	now := time.Date(2026, time.July, 23, 12, 0, 0, 0, time.UTC)
	cluster := types.Cluster{
		ID: "retention-cluster", Name: "retention",
		RegisteredAt: now.Add(-48 * time.Hour), Labels: map[string]string{},
	}
	if err := st.CreateCluster(ctx, cluster); err != nil {
		t.Fatalf("CreateCluster() error = %v", err)
	}
	for _, event := range []types.OperationalEvent{
		{ClusterID: cluster.ID, Kind: types.EventClusterRegistered, Message: "expired", OccurredAt: now.Add(-25 * time.Hour), DedupeKey: "expired"},
		{ClusterID: cluster.ID, Kind: types.EventAgentApproved, Message: "retained", OccurredAt: now.Add(-23 * time.Hour), DedupeKey: "retained"},
	} {
		if _, err := st.AppendEvent(ctx, event); err != nil {
			t.Fatalf("AppendEvent(%s) error = %v", event.Message, err)
		}
	}

	srv.pruneExpiredEvents(ctx, now)
	page, err := st.ListTimelineEvents(ctx, store.EventFilter{ClusterID: cluster.ID})
	if err != nil {
		t.Fatalf("ListTimelineEvents() error = %v", err)
	}
	if len(page.Events) != 1 || page.Events[0].Message != "retained" {
		t.Fatalf("events after configured retention = %#v, want retained event only", page.Events)
	}
}
