package server

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/1solomonwakhungu/kfleet/internal/config"
	"github.com/1solomonwakhungu/kfleet/internal/store"
	"github.com/1solomonwakhungu/kfleet/pkg/types"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

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
