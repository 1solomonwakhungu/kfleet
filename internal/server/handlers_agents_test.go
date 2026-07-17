package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/1solomonwakhungu/kfleet/internal/config"
	"github.com/1solomonwakhungu/kfleet/internal/store"
	"github.com/1solomonwakhungu/kfleet/pkg/api"
	"github.com/1solomonwakhungu/kfleet/pkg/types"
)

func TestAgentRegistrationApprovalAndHeartbeat(t *testing.T) {
	httpServer, _, _ := newAgentTestServer(t, 30*time.Second)

	register := agentRequest(t, httpServer, http.MethodPost, "/api/v1/agents/register", "", `{
		"name":"production","labels":{"region":"us-central1"}
	}`)
	if register.StatusCode != http.StatusCreated {
		t.Fatalf("register status = %d, want %d", register.StatusCode, http.StatusCreated)
	}
	var registration api.RegisterClusterResponse
	decodeResponse(t, register, &registration)
	if registration.ClusterID == "" || len(registration.Token) != 64 {
		t.Fatalf("registration = %#v, want ID and 32-byte hex token", registration)
	}

	pending := agentRequest(t, httpServer, http.MethodGet, "/api/v1/agents/pending", "", "")
	if pending.StatusCode != http.StatusOK {
		t.Fatalf("pending status = %d, want %d", pending.StatusCode, http.StatusOK)
	}
	var pendingList api.ListClustersResponse
	decodeResponse(t, pending, &pendingList)
	if len(pendingList.Clusters) != 1 || pendingList.Clusters[0].ID != registration.ClusterID {
		t.Fatalf("pending agents = %#v, want registered agent", pendingList)
	}

	heartbeatBody := `{"clusterId":"` + registration.ClusterID + `","nodeCount":3,"healthyNodes":3,"podCount":12,"version":"v1.31.1"}`
	beforeApproval := agentRequest(t, httpServer, http.MethodPost, "/api/v1/agents/heartbeat", registration.Token, heartbeatBody)
	if beforeApproval.StatusCode != http.StatusForbidden {
		t.Fatalf("heartbeat before approval status = %d, want %d", beforeApproval.StatusCode, http.StatusForbidden)
	}
	beforeApproval.Body.Close()

	approve := agentRequest(t, httpServer, http.MethodPost, "/api/v1/agents/"+registration.ClusterID+"/approve", "", "")
	if approve.StatusCode != http.StatusOK {
		t.Fatalf("approve status = %d, want %d", approve.StatusCode, http.StatusOK)
	}
	approve.Body.Close()

	heartbeat := agentRequest(t, httpServer, http.MethodPost, "/api/v1/agents/heartbeat", registration.Token, heartbeatBody)
	if heartbeat.StatusCode != http.StatusOK {
		t.Fatalf("heartbeat status = %d, want %d", heartbeat.StatusCode, http.StatusOK)
	}
	var cluster types.Cluster
	decodeResponse(t, heartbeat, &cluster)
	if cluster.Health != types.HealthHealthy || cluster.NodeCount != 3 || cluster.PodCount != 12 {
		t.Fatalf("heartbeat cluster = %#v, want healthy snapshot", cluster)
	}

	wrongToken := agentRequest(t, httpServer, http.MethodPost, "/api/v1/agents/heartbeat", "wrong", heartbeatBody)
	if wrongToken.StatusCode != http.StatusUnauthorized {
		t.Fatalf("heartbeat wrong token status = %d, want %d", wrongToken.StatusCode, http.StatusUnauthorized)
	}
	wrongToken.Body.Close()

	missingApprove := agentRequest(t, httpServer, http.MethodPost, "/api/v1/agents/missing/approve", "", "")
	if missingApprove.StatusCode != http.StatusNotFound {
		t.Fatalf("missing approve status = %d, want %d", missingApprove.StatusCode, http.StatusNotFound)
	}
	missingApprove.Body.Close()
}

func TestStalenessMarksClusterUnreachable(t *testing.T) {
	_, srv, st := newAgentTestServer(t, time.Second)
	ctx := context.Background()
	cluster := types.Cluster{
		ID:            "stale-cluster",
		Name:          "stale",
		Health:        types.HealthHealthy,
		RegisteredAt:  time.Now().UTC().Add(-time.Hour),
		LastHeartbeat: time.Now().UTC().Add(-4 * time.Second),
		Labels:        map[string]string{},
	}
	if err := st.CreateCluster(ctx, cluster); err != nil {
		t.Fatalf("CreateCluster() error = %v", err)
	}

	srv.markStaleClusters(ctx, time.Now().UTC())
	got, err := st.GetCluster(ctx, cluster.ID)
	if err != nil {
		t.Fatalf("GetCluster() error = %v", err)
	}
	if got.Health != types.HealthUnreachable {
		t.Fatalf("cluster health = %q, want %q", got.Health, types.HealthUnreachable)
	}
}

func newAgentTestServer(t *testing.T, interval time.Duration) (*httptest.Server, *Server, store.Store) {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "kfleet.db"))
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := New(&config.Config{ListenAddr: ":0", HeartbeatInterval: interval}, logger, st)
	httpServer := httptest.NewServer(srv.httpServer.Handler)
	t.Cleanup(func() {
		httpServer.Close()
		if err := st.Close(); err != nil {
			t.Errorf("store.Close() error = %v", err)
		}
	})
	return httpServer, srv, st
}

func agentRequest(t *testing.T, server *httptest.Server, method, path, token, body string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(method, server.URL+path, bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	response, err := server.Client().Do(req)
	if err != nil {
		t.Fatalf("request error = %v", err)
	}
	return response
}

func TestGenerateToken(t *testing.T) {
	raw, hash := generateToken()
	if len(raw) != 64 || len(hash) != 64 || hash != hashToken(raw) {
		t.Fatalf("generateToken() returned invalid raw/hash pair")
	}
	if _, err := json.Marshal(raw); err != nil {
		t.Fatalf("generated token is not JSON-safe: %v", err)
	}
}
