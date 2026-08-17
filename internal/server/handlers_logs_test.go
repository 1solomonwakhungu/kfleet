package server

import (
	"bufio"
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
	"github.com/1solomonwakhungu/kfleet/pkg/api"
	"github.com/1solomonwakhungu/kfleet/pkg/types"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

// newLogTestServer builds a hub with cfg and exposes the underlying Server
// so log tests can inspect the relay directly.
func newLogTestServer(t *testing.T, cfg *config.Config) (*httptest.Server, *Server) {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "kfleet.db"))
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := New(cfg, logger, st)
	httpServer := httptest.NewServer(srv.httpServer.Handler)
	t.Cleanup(func() {
		httpServer.Close()
		if err := st.Close(); err != nil {
			t.Errorf("store.Close() error = %v", err)
		}
	})
	registerDefaultSession(httpServer, st, sessionCookieFor(t, st, types.RoleAdmin))
	return httpServer, srv
}

// registeredLogAgent registers and approves an agent so its token can open
// the reverse log channel.
func registeredLogAgent(t *testing.T, server *httptest.Server) api.RegisterClusterResponse {
	t.Helper()
	response := agentRequest(t, server, http.MethodPost, "/api/v1/agents/register", "", `{"name":"production"}`)
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("register agent status = %d, want %d", response.StatusCode, http.StatusCreated)
	}
	var registration api.RegisterClusterResponse
	decodeResponse(t, response, &registration)
	approveAgent(t, server, registration.ClusterID)
	return registration
}

// dialAgentLogChannel opens the agent reverse channel and waits until the
// relay has registered it.
func dialAgentLogChannel(t *testing.T, server *httptest.Server, srv *Server, clusterID, token string) *websocket.Conn {
	t.Helper()
	endpoint := "ws" + strings.TrimPrefix(server.URL, "http") + "/api/v1/agents/" + clusterID + "/logs"
	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+token)
	conn, response, err := websocket.Dial(context.Background(), endpoint, &websocket.DialOptions{HTTPHeader: headers})
	if response != nil && response.Body != nil {
		_ = response.Body.Close()
	}
	if err != nil {
		t.Fatalf("dial agent log channel error = %v", err)
	}
	t.Cleanup(func() { _ = conn.CloseNow() })

	deadline := time.Now().Add(3 * time.Second)
	for !srv.logs.Connected(clusterID) {
		if time.Now().After(deadline) {
			t.Fatal("relay never registered the agent log channel")
		}
		time.Sleep(5 * time.Millisecond)
	}
	return conn
}

// openLogStream issues the browser-facing SSE request.
func openLogStream(t *testing.T, server *httptest.Server, path, sessionCookie, tenantID string) *http.Response {
	t.Helper()
	request, err := http.NewRequest(http.MethodGet, server.URL+path, nil)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	if sessionCookie != "" {
		request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionCookie})
	}
	if tenantID != "" {
		request.Header.Set(tenantHeader, tenantID)
	}
	response, err := server.Client().Do(request)
	if err != nil {
		t.Fatalf("log stream request error = %v", err)
	}
	t.Cleanup(func() { _ = response.Body.Close() })
	return response
}

// readSSEEvents collects SSE events until count events have arrived or the
// stream ends.
func readSSEEvents(t *testing.T, body io.Reader, count int) []sseEvent {
	t.Helper()
	reader := bufio.NewReader(body)
	events := make([]sseEvent, 0, count)
	current := sseEvent{Name: "message"}
	var data []string
	for len(events) < count {
		line, err := reader.ReadString('\n')
		if line == "" && err != nil {
			t.Fatalf("read SSE stream error = %v (events so far: %#v)", err, events)
		}
		line = strings.TrimRight(line, "\n")
		switch {
		case line == "":
			if len(data) > 0 || current.Name != "message" {
				current.Data = strings.Join(data, "\n")
				events = append(events, current)
			}
			current = sseEvent{Name: "message"}
			data = nil
		case strings.HasPrefix(line, ":"):
		case strings.HasPrefix(line, "event: "):
			current.Name = strings.TrimPrefix(line, "event: ")
		case strings.HasPrefix(line, "data: "):
			data = append(data, strings.TrimPrefix(line, "data: "))
		}
		if err != nil {
			break
		}
	}
	return events
}

type sseEvent struct {
	Name string
	Data string
}

func TestPodLogsRequiresAuthentication(t *testing.T) {
	server, _ := newLogTestServer(t, &config.Config{ListenAddr: ":0"})
	registration := registeredLogAgent(t, server)

	response := openLogStream(t, server, "/api/v1/clusters/"+registration.ClusterID+"/pods/apps/api/logs", "", "")
	if response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusUnauthorized)
	}
}

func TestPodLogsEnforcesTenantScope(t *testing.T) {
	server, _ := newLogTestServer(t, &config.Config{ListenAddr: ":0"})
	registration := registeredLogAgent(t, server)

	path := "/api/v1/clusters/" + registration.ClusterID + "/pods/apps/api/logs"
	response := openLogStream(t, server, path, defaultSessionFor(server), "other-tenant")
	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("cross-tenant status = %d, want %d", response.StatusCode, http.StatusNotFound)
	}
}

func TestPodLogsWithoutConnectedAgent(t *testing.T) {
	server, _ := newLogTestServer(t, &config.Config{ListenAddr: ":0"})
	registration := registeredLogAgent(t, server)

	path := "/api/v1/clusters/" + registration.ClusterID + "/pods/apps/api/logs"
	response := openLogStream(t, server, path, defaultSessionFor(server), "")
	if response.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusServiceUnavailable)
	}
	if contentType := response.Header.Get("Content-Type"); strings.HasPrefix(contentType, "text/event-stream") {
		// A non-SSE content type makes the browser fail the EventSource
		// permanently instead of reconnecting in a loop.
		t.Fatalf("content type = %q, want a non event-stream type", contentType)
	}
	body, _ := io.ReadAll(response.Body)
	if !strings.Contains(string(body), "no agent is connected") {
		t.Fatalf("body = %q, want an agent-disconnected message", body)
	}
}

func TestPodLogsStreamsAgentOutput(t *testing.T) {
	server, srv := newLogTestServer(t, &config.Config{ListenAddr: ":0"})
	registration := registeredLogAgent(t, server)
	conn := dialAgentLogChannel(t, server, srv, registration.ClusterID, registration.Token)

	path := "/api/v1/clusters/" + registration.ClusterID + "/pods/apps/api/logs?container=sidecar&follow=false&tailLines=42"
	response := openLogStream(t, server, path, defaultSessionFor(server), "")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusOK)
	}
	if got := response.Header.Get("Content-Type"); !strings.HasPrefix(got, "text/event-stream") {
		t.Fatalf("content type = %q, want text/event-stream", got)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var start api.LogStreamMessage
	if err := wsjson.Read(ctx, conn, &start); err != nil {
		t.Fatalf("read start frame error = %v", err)
	}
	if start.Type != api.LogStreamStart || start.Namespace != "apps" || start.Pod != "api" {
		t.Fatalf("start frame = %#v", start)
	}
	if start.Container != "sidecar" || start.Follow || start.TailLines != 42 {
		t.Fatalf("start frame options = %#v, want container/follow/tail passthrough", start)
	}

	for _, line := range []string{"alpha", "bravo"} {
		if err := wsjson.Write(ctx, conn, api.LogStreamMessage{
			Type: api.LogStreamData, StreamID: start.StreamID, Line: line,
		}); err != nil {
			t.Fatalf("write data frame error = %v", err)
		}
	}
	if err := wsjson.Write(ctx, conn, api.LogStreamMessage{Type: api.LogStreamEnd, StreamID: start.StreamID}); err != nil {
		t.Fatalf("write end frame error = %v", err)
	}

	events := readSSEEvents(t, response.Body, 3)
	if events[0].Name != "message" || events[0].Data != "alpha" {
		t.Fatalf("first event = %#v, want data alpha", events[0])
	}
	if events[1].Data != "bravo" {
		t.Fatalf("second event = %#v, want data bravo", events[1])
	}
	if events[2].Name != "end" {
		t.Fatalf("third event = %#v, want end event", events[2])
	}
}

func TestPodLogsRelaysAgentError(t *testing.T) {
	server, srv := newLogTestServer(t, &config.Config{ListenAddr: ":0"})
	registration := registeredLogAgent(t, server)
	conn := dialAgentLogChannel(t, server, srv, registration.ClusterID, registration.Token)

	path := "/api/v1/clusters/" + registration.ClusterID + "/pods/apps/api/logs"
	response := openLogStream(t, server, path, defaultSessionFor(server), "")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var start api.LogStreamMessage
	if err := wsjson.Read(ctx, conn, &start); err != nil {
		t.Fatalf("read start frame error = %v", err)
	}
	if err := wsjson.Write(ctx, conn, api.LogStreamMessage{
		Type: api.LogStreamEnd, StreamID: start.StreamID, Error: "pods \"api\" is forbidden",
	}); err != nil {
		t.Fatalf("write end frame error = %v", err)
	}

	events := readSSEEvents(t, response.Body, 1)
	if events[0].Name != sseStreamErrorEvent || !strings.Contains(events[0].Data, "forbidden") {
		t.Fatalf("event = %#v, want an error event", events[0])
	}
}

func TestPodLogsStopsAgentStreamOnDisconnect(t *testing.T) {
	server, srv := newLogTestServer(t, &config.Config{ListenAddr: ":0"})
	registration := registeredLogAgent(t, server)
	conn := dialAgentLogChannel(t, server, srv, registration.ClusterID, registration.Token)

	path := "/api/v1/clusters/" + registration.ClusterID + "/pods/apps/api/logs"
	response := openLogStream(t, server, path, defaultSessionFor(server), "")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var start api.LogStreamMessage
	if err := wsjson.Read(ctx, conn, &start); err != nil {
		t.Fatalf("read start frame error = %v", err)
	}
	// Closing the browser side must terminate the agent side stream so the
	// Kubernetes reader is released.
	_ = response.Body.Close()

	var stop api.LogStreamMessage
	if err := wsjson.Read(ctx, conn, &stop); err != nil {
		t.Fatalf("read stop frame error = %v", err)
	}
	if stop.Type != api.LogStreamStop || stop.StreamID != start.StreamID {
		t.Fatalf("frame = %#v, want stop for %s", stop, start.StreamID)
	}
}

func TestPodLogsEndsWhenAgentChannelDrops(t *testing.T) {
	server, srv := newLogTestServer(t, &config.Config{ListenAddr: ":0"})
	registration := registeredLogAgent(t, server)
	conn := dialAgentLogChannel(t, server, srv, registration.ClusterID, registration.Token)

	path := "/api/v1/clusters/" + registration.ClusterID + "/pods/apps/api/logs"
	response := openLogStream(t, server, path, defaultSessionFor(server), "")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var start api.LogStreamMessage
	if err := wsjson.Read(ctx, conn, &start); err != nil {
		t.Fatalf("read start frame error = %v", err)
	}
	_ = conn.CloseNow()

	events := readSSEEvents(t, response.Body, 1)
	if events[0].Name != sseStreamErrorEvent {
		t.Fatalf("event = %#v, want an error event after the agent disconnects", events[0])
	}
}

func TestAgentLogChannelRejectsUnauthorizedAgents(t *testing.T) {
	server, _ := newLogTestServer(t, &config.Config{ListenAddr: ":0"})
	registration := registeredLogAgent(t, server)

	tests := []struct {
		name       string
		token      string
		wantStatus int
	}{
		{name: "wrong token", token: "not-the-token", wantStatus: http.StatusUnauthorized},
		{name: "missing token", token: "", wantStatus: http.StatusUnauthorized},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoint := "ws" + strings.TrimPrefix(server.URL, "http") +
				"/api/v1/agents/" + registration.ClusterID + "/logs"
			headers := http.Header{}
			if tt.token != "" {
				headers.Set("Authorization", "Bearer "+tt.token)
			}
			conn, response, err := websocket.Dial(context.Background(), endpoint, &websocket.DialOptions{HTTPHeader: headers})
			if err == nil {
				_ = conn.CloseNow()
				t.Fatal("dial succeeded, want rejection")
			}
			if response == nil {
				t.Fatalf("dial error = %v, want an HTTP response", err)
			}
			defer func() { _ = response.Body.Close() }()
			if response.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", response.StatusCode, tt.wantStatus)
			}
		})
	}
}

func TestAgentLogChannelRejectsPendingAgents(t *testing.T) {
	server, _ := newLogTestServer(t, &config.Config{ListenAddr: ":0"})
	response := agentRequest(t, server, http.MethodPost, "/api/v1/agents/register", "", `{"name":"pending"}`)
	var registration api.RegisterClusterResponse
	decodeResponse(t, response, &registration)

	endpoint := "ws" + strings.TrimPrefix(server.URL, "http") + "/api/v1/agents/" + registration.ClusterID + "/logs"
	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+registration.Token)
	conn, dialResponse, err := websocket.Dial(context.Background(), endpoint, &websocket.DialOptions{HTTPHeader: headers})
	if err == nil {
		_ = conn.CloseNow()
		t.Fatal("dial succeeded, want rejection for a pending agent")
	}
	if dialResponse == nil {
		t.Fatalf("dial error = %v, want an HTTP response", err)
	}
	defer func() { _ = dialResponse.Body.Close() }()
	if dialResponse.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", dialResponse.StatusCode, http.StatusForbidden)
	}
}

func TestDemoModeSynthesizesPodLogs(t *testing.T) {
	server, srv := newLogTestServer(t, &config.Config{ListenAddr: ":0", DemoMode: true})
	cluster := types.Cluster{
		ID:       "demo-cluster-alpha",
		TenantID: store.DefaultTenantID,
		Name:     "demo-us-central",
		Health:   types.HealthHealthy,
	}
	if err := srv.store.CreateCluster(context.Background(), cluster); err != nil {
		t.Fatalf("CreateCluster() error = %v", err)
	}

	response := openLogStream(t, server,
		"/api/v1/clusters/"+cluster.ID+"/pods/storefront/catalog-api/logs?follow=false&tailLines=3", "", "")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusOK)
	}
	events := readSSEEvents(t, response.Body, 4)
	for index := range 3 {
		if events[index].Name != "message" || !strings.Contains(events[index].Data, "storefront/catalog-api") {
			t.Fatalf("event %d = %#v, want a synthetic log line", index, events[index])
		}
	}
	if events[3].Name != "end" {
		t.Fatalf("final event = %#v, want end", events[3])
	}
}

func TestParseLogStreamQueryParams(t *testing.T) {
	if !parseFollow("") || !parseFollow("true") || !parseFollow("nonsense") || parseFollow("false") {
		t.Fatal("parseFollow() did not default to following")
	}
	if got := parseTailLines(""); got != defaultLogTailLines {
		t.Fatalf("parseTailLines(\"\") = %d, want %d", got, defaultLogTailLines)
	}
	if got := parseTailLines("-5"); got != defaultLogTailLines {
		t.Fatalf("parseTailLines(\"-5\") = %d, want %d", got, defaultLogTailLines)
	}
	if got := parseTailLines("100000"); got != maxLogTailLines {
		t.Fatalf("parseTailLines(\"100000\") = %d, want %d", got, maxLogTailLines)
	}
	if got := parseTailLines("77"); got != 77 {
		t.Fatalf("parseTailLines(\"77\") = %d, want 77", got)
	}
}
