package server

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/1solomonwakhungu/kfleet/internal/config"
	"github.com/1solomonwakhungu/kfleet/internal/store"
	"github.com/1solomonwakhungu/kfleet/pkg/api"
	"github.com/1solomonwakhungu/kfleet/pkg/types"
)

func TestClusterHandlersLifecycle(t *testing.T) {
	server := newTestHTTPServer(t)

	registerResponse := request(t, server, http.MethodPost, "/api/v1/clusters/register", `{
		"name": "production",
		"labels": {"region": "us-central1"}
	}`)
	if registerResponse.StatusCode != http.StatusCreated {
		t.Fatalf("register status = %d, want %d", registerResponse.StatusCode, http.StatusCreated)
	}
	var registration api.RegisterClusterResponse
	decodeResponse(t, registerResponse, &registration)
	if registration.ClusterID == "" || registration.Token == "" {
		t.Fatalf("register response = %#v, want ID and token", registration)
	}

	listResponse := request(t, server, http.MethodGet, "/api/v1/clusters", "")
	if listResponse.StatusCode != http.StatusOK {
		t.Fatalf("list status = %d, want %d", listResponse.StatusCode, http.StatusOK)
	}
	var list api.ListClustersResponse
	decodeResponse(t, listResponse, &list)
	if len(list.Clusters) != 1 || list.Clusters[0].ID != registration.ClusterID {
		t.Fatalf("list response = %#v, want registered cluster", list)
	}

	getResponse := request(t, server, http.MethodGet, "/api/v1/clusters/"+registration.ClusterID, "")
	if getResponse.StatusCode != http.StatusOK {
		t.Fatalf("get status = %d, want %d", getResponse.StatusCode, http.StatusOK)
	}
	var cluster types.Cluster
	decodeResponse(t, getResponse, &cluster)
	if cluster.ID != registration.ClusterID || cluster.Name != "production" {
		t.Fatalf("get response = %#v, want registered cluster", cluster)
	}

	unknownResponse := request(t, server, http.MethodGet, "/api/v1/clusters/unknown", "")
	if unknownResponse.StatusCode != http.StatusNotFound {
		t.Fatalf("unknown get status = %d, want %d", unknownResponse.StatusCode, http.StatusNotFound)
	}
	unknownResponse.Body.Close()

	deleteResponse := request(t, server, http.MethodDelete, "/api/v1/clusters/"+registration.ClusterID, "")
	if deleteResponse.StatusCode != http.StatusNoContent {
		t.Fatalf("delete status = %d, want %d", deleteResponse.StatusCode, http.StatusNoContent)
	}
	deleteResponse.Body.Close()

	deletedResponse := request(t, server, http.MethodGet, "/api/v1/clusters/"+registration.ClusterID, "")
	if deletedResponse.StatusCode != http.StatusNotFound {
		t.Fatalf("deleted get status = %d, want %d", deletedResponse.StatusCode, http.StatusNotFound)
	}
	deletedResponse.Body.Close()
}

func TestRegisterClusterRejectsEmptyName(t *testing.T) {
	server := newTestHTTPServer(t)

	response := request(t, server, http.MethodPost, "/api/v1/clusters/register", `{"name":"  "}`)
	defer response.Body.Close()
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusBadRequest)
	}
}

func TestClusterSnapshotPostAuthenticationAndApproval(t *testing.T) {
	tests := []struct {
		name       string
		path       func(api.RegisterClusterResponse) string
		token      func(api.RegisterClusterResponse) string
		approve    bool
		wantStatus int
	}{
		{
			name:       "approved by ID",
			path:       func(reg api.RegisterClusterResponse) string { return "/api/v1/clusters/" + reg.ClusterID + "/status" },
			token:      func(reg api.RegisterClusterResponse) string { return reg.Token },
			approve:    true,
			wantStatus: http.StatusOK,
		},
		{
			name:       "approved by exact name",
			path:       func(api.RegisterClusterResponse) string { return "/api/v1/clusters/production/status" },
			token:      func(reg api.RegisterClusterResponse) string { return reg.Token },
			approve:    true,
			wantStatus: http.StatusOK,
		},
		{
			name:       "pending agent",
			path:       func(reg api.RegisterClusterResponse) string { return "/api/v1/clusters/" + reg.ClusterID + "/status" },
			token:      func(reg api.RegisterClusterResponse) string { return reg.Token },
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "invalid token",
			path:       func(reg api.RegisterClusterResponse) string { return "/api/v1/clusters/" + reg.ClusterID + "/status" },
			token:      func(api.RegisterClusterResponse) string { return "wrong" },
			approve:    true,
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, registration := registeredAgent(t)
			if tt.approve {
				approveAgent(t, server, registration.ClusterID)
			}
			response := agentRequest(t, server, http.MethodPost, tt.path(registration), tt.token(registration), snapshotPayload())
			defer response.Body.Close()
			if response.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", response.StatusCode, tt.wantStatus)
			}
		})
	}

	server := newTestHTTPServer(t)
	unknown := agentRequest(t, server, http.MethodPost, "/api/v1/clusters/missing/status", "token", snapshotPayload())
	defer unknown.Body.Close()
	if unknown.StatusCode != http.StatusNotFound {
		t.Fatalf("unknown cluster status = %d, want %d", unknown.StatusCode, http.StatusNotFound)
	}
}

func TestClusterSnapshotPersistsResourceEndpoints(t *testing.T) {
	server, registration := registeredAgent(t)
	approveAgent(t, server, registration.ClusterID)

	posted := agentRequest(t, server, http.MethodPost, "/api/v1/clusters/production/status", registration.Token, snapshotPayload())
	if posted.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(posted.Body)
		posted.Body.Close()
		t.Fatalf("snapshot status = %d, want %d: %s", posted.StatusCode, http.StatusOK, body)
	}
	var postedCluster types.Cluster
	decodeResponse(t, posted, &postedCluster)
	if postedCluster.NodeCount != 2 || postedCluster.PodCount != 2 || postedCluster.Version != "v1.31.1" || postedCluster.AgentVersion != "0.1.0" || postedCluster.Health != types.HealthDegraded || postedCluster.LastHeartbeat.IsZero() {
		t.Fatalf("posted cluster = %#v, want degraded snapshot metadata", postedCluster)
	}

	statusResponse := request(t, server, http.MethodGet, "/api/v1/clusters/"+registration.ClusterID+"/status", "")
	var status api.ClusterStatusResponse
	decodeResponse(t, statusResponse, &status)
	if len(status.Nodes) != 2 || status.Nodes[0].Name != "node-a" || !status.Nodes[0].Ready || status.Nodes[1].Ready {
		t.Fatalf("status nodes = %#v, want persisted nodes", status.Nodes)
	}

	assertResourceList(t, server, "/api/v1/clusters/"+registration.ClusterID+"/pods", []types.Pod{}, 2)
	assertResourceList(t, server, "/api/v1/clusters/"+registration.ClusterID+"/pods?namespace=apps", []types.Pod{}, 1)
	assertResourceList(t, server, "/api/v1/clusters/"+registration.ClusterID+"/services", []types.Service{}, 2)
	assertResourceList(t, server, "/api/v1/clusters/"+registration.ClusterID+"/services?namespace=apps", []types.Service{}, 1)
	assertResourceList(t, server, "/api/v1/clusters/"+registration.ClusterID+"/deployments", []types.Deployment{}, 2)
	assertResourceList(t, server, "/api/v1/clusters/"+registration.ClusterID+"/deployments?namespace=apps", []types.Deployment{}, 1)
	assertResourceList(t, server, "/api/v1/clusters/"+registration.ClusterID+"/events", []types.Event{}, 2)
	assertResourceList(t, server, "/api/v1/clusters/"+registration.ClusterID+"/events?namespace=apps", []types.Event{}, 1)

	namespacesResponse := request(t, server, http.MethodGet, "/api/v1/clusters/"+registration.ClusterID+"/namespaces", "")
	var namespaces []string
	decodeResponse(t, namespacesResponse, &namespaces)
	if got := strings.Join(namespaces, ","); got != "apps,kube-system" {
		t.Fatalf("namespaces = %q, want apps,kube-system", got)
	}
}

func TestClusterResourceEndpointsReturnEmptyArrays(t *testing.T) {
	server := newTestHTTPServer(t)
	registered := request(t, server, http.MethodPost, "/api/v1/clusters/register", `{"name":"empty"}`)
	var registration api.RegisterClusterResponse
	decodeResponse(t, registered, &registration)

	for _, suffix := range []string{"pods", "services", "deployments", "events", "namespaces"} {
		response := request(t, server, http.MethodGet, "/api/v1/clusters/"+registration.ClusterID+"/"+suffix, "")
		body, err := io.ReadAll(response.Body)
		response.Body.Close()
		if err != nil {
			t.Fatalf("read %s response: %v", suffix, err)
		}
		if response.StatusCode != http.StatusOK || strings.TrimSpace(string(body)) != "[]" {
			t.Fatalf("%s response = (%d, %q), want (200, [])", suffix, response.StatusCode, body)
		}
	}
}

func TestClusterSnapshotRejectsMalformedAndOversizedBodies(t *testing.T) {
	server, registration := registeredAgent(t)
	approveAgent(t, server, registration.ClusterID)
	path := "/api/v1/clusters/" + registration.ClusterID + "/status"

	malformed := agentRequest(t, server, http.MethodPost, path, registration.Token, `{"nodes":`)
	defer malformed.Body.Close()
	if malformed.StatusCode != http.StatusBadRequest {
		t.Fatalf("malformed status = %d, want %d", malformed.StatusCode, http.StatusBadRequest)
	}

	oversized := agentRequest(t, server, http.MethodPost, path, registration.Token, `{"padding":"`+strings.Repeat("x", maxSnapshotBodyBytes)+`"}`)
	defer oversized.Body.Close()
	if oversized.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversized status = %d, want %d", oversized.StatusCode, http.StatusRequestEntityTooLarge)
	}
}

func registeredAgent(t *testing.T) (*httptest.Server, api.RegisterClusterResponse) {
	t.Helper()
	server := newTestHTTPServer(t)
	response := agentRequest(t, server, http.MethodPost, "/api/v1/agents/register", "", `{"name":"production"}`)
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("register agent status = %d, want %d", response.StatusCode, http.StatusCreated)
	}
	var registration api.RegisterClusterResponse
	decodeResponse(t, response, &registration)
	return server, registration
}

func approveAgent(t *testing.T, server *httptest.Server, clusterID string) {
	t.Helper()
	response := agentRequest(t, server, http.MethodPost, "/api/v1/agents/"+clusterID+"/approve", "", "")
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("approve status = %d, want %d", response.StatusCode, http.StatusOK)
	}
}

func snapshotPayload() string {
	return `{
		"nodes":[
			{"name":"node-a","status":"Ready","k8sVersion":"v1.31.1","roles":["control-plane"],"cpuCapacity":"4","memoryCapacity":"8Gi"},
			{"name":"node-b","status":"NotReady","k8sVersion":"v1.31.1","roles":["worker"],"cpuCapacity":"8","memoryCapacity":"16Gi"}
		],
		"pods":[
			{"namespace":"apps","name":"api","phase":"Running","restarts":2,"node":"node-a"},
			{"namespace":"kube-system","name":"dns","phase":"Running","restarts":0,"node":"node-b"}
		],
		"services":[
			{"namespace":"apps","name":"api","type":"ClusterIP","clusterIP":"10.0.0.1","externalIPs":[],"ports":["http:80/TCP"]},
			{"namespace":"kube-system","name":"dns","type":"ClusterIP","clusterIP":"10.0.0.10","externalIPs":[],"ports":["dns:53/UDP"]}
		],
		"deployments":[
			{"namespace":"apps","name":"api","desiredReplicas":3,"readyReplicas":3,"availableReplicas":3},
			{"namespace":"kube-system","name":"dns","desiredReplicas":2,"readyReplicas":2,"availableReplicas":2}
		],
		"events":[
			{"namespace":"apps","name":"api-started","type":"Normal","reason":"Started","message":"Started container","involvedObject":"Pod/api","count":1,"lastTimestamp":"2026-07-19T12:00:00Z"},
			{"namespace":"kube-system","name":"dns-failed","type":"Warning","reason":"Unhealthy","message":"Readiness failed","involvedObject":"Pod/dns","count":2,"lastTimestamp":"2026-07-19T12:01:00Z"}
		],
		"nodeCount":2,"podCount":2,"k8sVersion":"v1.31.1","collectedAt":"2026-07-19T12:01:00Z","agentVersion":"0.1.0"
	}`
}

func assertResourceList[T any](t *testing.T, server *httptest.Server, path string, target []T, want int) {
	t.Helper()
	response := request(t, server, http.MethodGet, path, "")
	decodeResponse(t, response, &target)
	if len(target) != want {
		t.Fatalf("GET %s returned %d resources, want %d", path, len(target), want)
	}
}

func newTestHTTPServer(t *testing.T) *httptest.Server {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "kfleet.db"))
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := New(&config.Config{ListenAddr: ":0"}, logger, st)
	httpServer := httptest.NewServer(srv.httpServer.Handler)
	t.Cleanup(func() {
		httpServer.Close()
		if err := st.Close(); err != nil {
			t.Errorf("store.Close() error = %v", err)
		}
	})
	registerDefaultSession(httpServer, st, sessionCookieFor(t, st, types.RoleAdmin))
	return httpServer
}

// request issues an HTTP request against server using its default admin
// session cookie (see registerDefaultSession). Tests exercising a specific
// role's allowed/denied actions should use requestWithSession instead.
func request(t *testing.T, server *httptest.Server, method, path, body string) *http.Response {
	t.Helper()
	return requestWithSession(t, server, method, path, defaultSessionFor(server), body)
}

// requestWithSession issues an HTTP request against server, attaching
// sessionCookie as the kfleet_session cookie when non-empty. Pass an empty
// sessionCookie to exercise unauthenticated requests.
func requestWithSession(t *testing.T, server *httptest.Server, method, path, sessionCookie, body string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(method, server.URL+path, bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if sessionCookie != "" {
		req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionCookie})
		if mutationRequest(req) {
			req.Header.Set(csrfHeaderName, "1")
		}
	}
	response, err := server.Client().Do(req)
	if err != nil {
		t.Fatalf("request error = %v", err)
	}
	return response
}

func decodeResponse(t *testing.T, response *http.Response, target any) {
	t.Helper()
	defer response.Body.Close()
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		t.Fatalf("decode response error = %v", err)
	}
}
