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
	"strings"
	"testing"
	"time"

	"github.com/1solomonwakhungu/kfleet/internal/config"
	"github.com/1solomonwakhungu/kfleet/internal/demo"
	"github.com/1solomonwakhungu/kfleet/internal/store"
	"github.com/1solomonwakhungu/kfleet/pkg/types"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

func TestDemoModeBlocksMutationsAndSetsSecurityHeaders(t *testing.T) {
	st, err := demo.Open(context.Background())
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := New(&config.Config{ListenAddr: ":0", DemoMode: true}, logger, st)
	httpServer := httptest.NewServer(srv.httpServer.Handler)
	t.Cleanup(httpServer.Close)

	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		req, err := http.NewRequest(method, httpServer.URL+"/api/v1/clusters/register", bytes.NewBufferString(`{"name":"must-not-exist"}`))
		if err != nil {
			t.Fatalf("http.NewRequest(%s) error = %v", method, err)
		}
		response, err := httpServer.Client().Do(req)
		if err != nil {
			t.Fatalf("%s request error = %v", method, err)
		}
		_ = response.Body.Close()
		if response.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("%s status = %d, want %d", method, response.StatusCode, http.StatusMethodNotAllowed)
		}
		if got := response.Header.Get("Allow"); got != "GET, HEAD, OPTIONS" {
			t.Errorf("%s Allow = %q", method, got)
		}
	}

	response, err := http.Get(httpServer.URL + "/api/v1/meta")
	if err != nil {
		t.Fatalf("GET /api/v1/meta error = %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/v1/meta status = %d", response.StatusCode)
	}
	var metadata struct {
		DemoMode      bool `json:"demoMode"`
		ReadOnly      bool `json:"readOnly"`
		SyntheticData bool `json:"syntheticData"`
	}
	if err := json.NewDecoder(response.Body).Decode(&metadata); err != nil {
		t.Fatalf("decode metadata: %v", err)
	}
	if !metadata.DemoMode || !metadata.ReadOnly || !metadata.SyntheticData {
		t.Fatalf("metadata = %#v, want all safety flags true", metadata)
	}

	clusterResponse, err := http.Get(httpServer.URL + "/api/v1/clusters")
	if err != nil {
		t.Fatalf("GET /api/v1/clusters error = %v", err)
	}
	var clusters struct {
		Clusters []types.Cluster `json:"clusters"`
	}
	if err := json.NewDecoder(clusterResponse.Body).Decode(&clusters); err != nil {
		t.Fatalf("decode clusters: %v", err)
	}
	clusterResponse.Body.Close()
	if clusterResponse.StatusCode != http.StatusOK || len(clusters.Clusters) != 3 {
		t.Fatalf("public demo clusters status=%d count=%d, want 200 and 3", clusterResponse.StatusCode, len(clusters.Clusters))
	}
	for _, cluster := range clusters.Clusters {
		if cluster.Labels["data"] != "synthetic" || cluster.Labels["environment"] != "demo" {
			t.Fatalf("public demo leaked a non-synthetic cluster: %#v", cluster)
		}
	}

	meResponse, err := http.Get(httpServer.URL + "/api/v1/auth/me")
	if err != nil {
		t.Fatalf("GET /api/v1/auth/me error = %v", err)
	}
	var me struct {
		Username string     `json:"username"`
		Role     types.Role `json:"role"`
	}
	if err := json.NewDecoder(meResponse.Body).Decode(&me); err != nil {
		t.Fatalf("decode demo user: %v", err)
	}
	meResponse.Body.Close()
	if meResponse.StatusCode != http.StatusOK || me.Username != "public-demo" || me.Role != types.RoleReadOnly {
		t.Fatalf("public demo user status=%d user=%#v", meResponse.StatusCode, me)
	}

	usersResponse, err := http.Get(httpServer.URL + "/api/v1/users")
	if err != nil {
		t.Fatalf("GET /api/v1/users error = %v", err)
	}
	usersResponse.Body.Close()
	if usersResponse.StatusCode != http.StatusForbidden {
		t.Fatalf("public demo users status = %d, want %d", usersResponse.StatusCode, http.StatusForbidden)
	}
	for _, header := range []string{
		"Content-Security-Policy",
		"Permissions-Policy",
		"Referrer-Policy",
		"Strict-Transport-Security",
		"X-Content-Type-Options",
		"X-Frame-Options",
	} {
		if response.Header.Get(header) == "" {
			t.Errorf("%s header is missing", header)
		}
	}
}

func TestBroadcastHubDeliversClusterUpdate(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "kfleet.db"))
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := st.Close(); err != nil {
			t.Errorf("store.Close() error = %v", err)
		}
	})

	seed := types.Cluster{ID: "cluster-1", Name: "production", Health: types.HealthUnknown, RegisteredAt: time.Now().UTC()}
	if err := st.CreateCluster(context.Background(), seed); err != nil {
		t.Fatalf("CreateCluster() error = %v", err)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := New(&config.Config{ListenAddr: ":0"}, logger, st)
	hubCtx, stopHub := context.WithCancel(context.Background())
	go srv.broadcast.Run(hubCtx)
	t.Cleanup(stopHub)

	httpServer := httptest.NewServer(srv.httpServer.Handler)
	t.Cleanup(httpServer.Close)
	sessionCookie := sessionCookieFor(t, st, types.RoleReadOnly)
	readyResponse, err := http.Get(httpServer.URL + "/readyz")
	if err != nil {
		t.Fatalf("GET /readyz error = %v", err)
	}
	defer readyResponse.Body.Close()
	if readyResponse.StatusCode != http.StatusOK {
		t.Fatalf("GET /readyz status = %d, want %d", readyResponse.StatusCode, http.StatusOK)
	}
	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/ws/clusters"
	conn, _, err := websocket.Dial(context.Background(), wsURL, &websocket.DialOptions{
		HTTPHeader: http.Header{"Cookie": []string{sessionCookieName + "=" + sessionCookie}},
	})
	if err != nil {
		t.Fatalf("websocket.Dial() error = %v", err)
	}
	defer func() {
		if err := conn.CloseNow(); err != nil {
			t.Logf("websocket close error: %v", err)
		}
	}()

	readCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var snapshot ClusterUpdate
	if err := wsjson.Read(readCtx, conn, &snapshot); err != nil {
		t.Fatalf("read initial snapshot error = %v", err)
	}
	if snapshot.Type != "snapshot" || snapshot.Cluster.ID != seed.ID {
		t.Fatalf("initial snapshot = %#v, want cluster %q", snapshot, seed.ID)
	}

	want := ClusterUpdate{
		Type: "health_changed",
		Cluster: types.Cluster{
			ID:     seed.ID,
			Name:   seed.Name,
			Health: types.HealthHealthy,
		},
	}
	srv.broadcast.Broadcast(want)

	var got ClusterUpdate
	if err := wsjson.Read(readCtx, conn, &got); err != nil {
		t.Fatalf("read broadcast error = %v", err)
	}
	if got.Type != want.Type || got.Cluster.ID != want.Cluster.ID || got.Cluster.Health != want.Cluster.Health {
		t.Fatalf("broadcast = %#v, want %#v", got, want)
	}
}

func TestBroadcastHubDropsSlowClientWithoutBlocking(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	hub := NewBroadcastHub(logger)
	ctx, cancel := context.WithCancel(context.Background())
	go hub.Run(ctx)
	defer cancel()

	slowClient := &wsClient{
		send:       make(chan ClusterUpdate, 1),
		registered: make(chan struct{}),
		closed:     make(chan struct{}),
	}
	if !hub.registerClient(slowClient) {
		t.Fatal("registerClient() = false, want true")
	}
	slowClient.send <- ClusterUpdate{Type: "snapshot"}
	hub.Broadcast(ClusterUpdate{Type: "snapshot"})

	select {
	case <-slowClient.closed:
	case <-time.After(time.Second):
		t.Fatal("slow client was not dropped")
	}

	done := make(chan struct{})
	go func() {
		for range 1000 {
			hub.Broadcast(ClusterUpdate{Type: "snapshot"})
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Broadcast blocked after dropping a slow client")
	}
}

func TestBroadcastHubScopesUpdatesToTenant(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	hub := NewBroadcastHub(logger)
	ctx, cancel := context.WithCancel(context.Background())
	go hub.Run(ctx)
	defer cancel()

	clientA := &wsClient{
		tenantID: "tenant-a", send: make(chan ClusterUpdate, 1),
		registered: make(chan struct{}), closed: make(chan struct{}),
	}
	clientB := &wsClient{
		tenantID: "tenant-b", send: make(chan ClusterUpdate, 1),
		registered: make(chan struct{}), closed: make(chan struct{}),
	}
	if !hub.registerClient(clientA) || !hub.registerClient(clientB) {
		t.Fatal("failed to register tenant clients")
	}
	defer hub.unregisterClient(clientA)
	defer hub.unregisterClient(clientB)

	hub.Broadcast(ClusterUpdate{
		Type: "snapshot",
		Cluster: types.Cluster{
			ID: "a", TenantID: "tenant-a", Name: "alpha",
		},
	})

	select {
	case update := <-clientA.send:
		if update.Cluster.ID != "a" {
			t.Fatalf("tenant-a update = %#v, want cluster a", update)
		}
	case <-time.After(time.Second):
		t.Fatal("tenant-a did not receive its update")
	}
	select {
	case update := <-clientB.send:
		t.Fatalf("tenant-b received tenant-a update: %#v", update)
	case <-time.After(50 * time.Millisecond):
	}
}
