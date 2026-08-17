package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/1solomonwakhungu/kfleet/pkg/api"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

const (
	// defaultLogTailLines bounds the backlog replayed before live output so
	// a busy pod cannot flood the browser on connect.
	defaultLogTailLines = 200
	maxLogTailLines     = 5000
	// sseKeepAliveInterval keeps intermediaries from dropping idle streams.
	sseKeepAliveInterval = 25 * time.Second
	// demoLogInterval paces synthetic follow output in demo mode.
	demoLogInterval = time.Second
	// demoLogBacklog is how many synthetic lines are replayed on connect.
	demoLogBacklog = 40
	// sseStreamErrorEvent carries a stream-level failure. It deliberately
	// avoids the name "error" because EventSource dispatches server-sent
	// "error" events through the same handler as transport failures.
	sseStreamErrorEvent = "stream-error"
)

// handleAgentLogStream accepts the agent's outbound reverse channel. The
// hub has no kubeconfig of its own, so this connection is the only path to
// pod logs. It authenticates with the same per-agent bearer token scheme as
// the snapshot and heartbeat routes.
func (s *Server) handleAgentLogStream(w http.ResponseWriter, r *http.Request) {
	cluster, approved, ok := s.authenticateAgentPath(w, r)
	if !ok {
		return
	}
	if !approved {
		api.WriteError(w, http.StatusForbidden, "agent is pending approval")
		return
	}

	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		s.logger.Error("failed to accept agent log channel", "cluster_id", cluster.ID, "error", err)
		return
	}
	defer func() {
		_ = conn.Close(websocket.StatusNormalClosure, "")
	}()
	conn.SetReadLimit(maxLogFrameBytes)

	agentConn := newAgentLogConn(cluster.ID)
	s.logs.register(agentConn)
	defer s.logs.unregister(agentConn)

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	readErr := make(chan error, 1)
	writeErr := make(chan error, 1)
	go func() {
		readErr <- readAgentLogFrames(ctx, conn, agentConn)
	}()
	go func() {
		writeErr <- writeAgentLogFrames(ctx, conn, agentConn.send)
	}()

	select {
	case <-ctx.Done():
	case err := <-readErr:
		s.logger.Debug("agent log channel reader stopped", "cluster_id", cluster.ID, "error", err)
	case err := <-writeErr:
		s.logger.Debug("agent log channel writer stopped", "cluster_id", cluster.ID, "error", err)
	}
}

// maxLogFrameBytes bounds a single agent frame. Log lines are truncated by
// the agent well below this, so anything larger is malformed.
const maxLogFrameBytes = 1 << 20

func readAgentLogFrames(ctx context.Context, conn *websocket.Conn, agentConn *agentLogConn) error {
	for {
		var message api.LogStreamMessage
		if err := wsjson.Read(ctx, conn, &message); err != nil {
			return err
		}
		switch message.Type {
		case api.LogStreamData, api.LogStreamEnd:
			agentConn.deliver(message)
		default:
			// Agents only ever answer with data/end frames; ignore the rest
			// rather than tearing down a channel serving other streams.
		}
	}
}

func writeAgentLogFrames(ctx context.Context, conn *websocket.Conn, messages <-chan api.LogStreamMessage) error {
	ticker := time.NewTicker(websocketPingInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case message, ok := <-messages:
			if !ok {
				return nil
			}
			writeCtx, cancel := context.WithTimeout(ctx, websocketWriteTimeout)
			err := wsjson.Write(writeCtx, conn, message)
			cancel()
			if err != nil {
				return err
			}
		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, websocketWriteTimeout)
			err := conn.Ping(pingCtx)
			cancel()
			if err != nil {
				return err
			}
		}
	}
}

// handleClusterPodLogs streams pod logs to the browser as SSE. Tenant
// scoping matches the other cluster resource routes: the cluster must
// resolve within the caller's tenant before anything is streamed.
func (s *Server) handleClusterPodLogs(w http.ResponseWriter, r *http.Request) {
	cluster, ok := s.resourceCluster(w, r)
	if !ok {
		return
	}
	namespace := strings.TrimSpace(r.PathValue("namespace"))
	pod := strings.TrimSpace(r.PathValue("pod"))
	if namespace == "" || pod == "" {
		api.WriteError(w, http.StatusBadRequest, "namespace and pod are required")
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		api.WriteError(w, http.StatusInternalServerError, "streaming is not supported")
		return
	}
	request := logStreamRequest{
		Namespace: namespace,
		Pod:       pod,
		Container: strings.TrimSpace(r.URL.Query().Get("container")),
		Follow:    parseFollow(r.URL.Query().Get("follow")),
		TailLines: parseTailLines(r.URL.Query().Get("tailLines")),
	}

	stream, err := s.logs.open(cluster.ID, request)
	if err != nil {
		if s.cfg.DemoMode {
			// Demo clusters are synthetic and have no real agent, so the
			// demo keeps a working log experience with synthetic output.
			s.streamDemoPodLogs(w, r, flusher, request)
			return
		}
		if errors.Is(err, errNoAgentConnected) || errors.Is(err, errAgentChannelClosed) {
			api.WriteError(w, http.StatusServiceUnavailable, "no agent is connected for this cluster")
			return
		}
		s.logger.Error("failed to open pod log stream", "cluster_id", cluster.ID, "error", err)
		api.WriteError(w, http.StatusInternalServerError, "failed to open log stream")
		return
	}
	defer stream.Close()

	writeSSEHeaders(w)
	flusher.Flush()

	keepAlive := time.NewTicker(sseKeepAliveInterval)
	defer keepAlive.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-keepAlive.C:
			if _, err := fmt.Fprint(w, ": keep-alive\n\n"); err != nil {
				return
			}
			flusher.Flush()
		case message, open := <-stream.Messages():
			if !open {
				writeSSEError(w, "the agent log channel closed unexpectedly")
				flusher.Flush()
				return
			}
			switch message.Type {
			case api.LogStreamData:
				if err := writeSSEData(w, message.Line); err != nil {
					return
				}
				flusher.Flush()
			case api.LogStreamEnd:
				if message.Error != "" {
					writeSSEError(w, message.Error)
				} else {
					writeSSEEvent(w, "end", "{}")
				}
				flusher.Flush()
				return
			}
		}
	}
}

// streamDemoPodLogs synthesizes plausible output for the public demo, which
// runs without any agent attached.
func (s *Server) streamDemoPodLogs(w http.ResponseWriter, r *http.Request, flusher http.Flusher, request logStreamRequest) {
	writeSSEHeaders(w)
	flusher.Flush()

	backlog := int(request.TailLines)
	if backlog > demoLogBacklog {
		backlog = demoLogBacklog
	}
	started := time.Now().UTC().Add(-time.Duration(backlog) * demoLogInterval)
	for index := range backlog {
		line := demoLogLine(request, started.Add(time.Duration(index)*demoLogInterval), index)
		if err := writeSSEData(w, line); err != nil {
			return
		}
	}
	flusher.Flush()
	if !request.Follow {
		writeSSEEvent(w, "end", "{}")
		flusher.Flush()
		return
	}

	ticker := time.NewTicker(demoLogInterval)
	defer ticker.Stop()
	for index := backlog; ; index++ {
		select {
		case <-r.Context().Done():
			return
		case now := <-ticker.C:
			if err := writeSSEData(w, demoLogLine(request, now.UTC(), index)); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

// demoLogMessages are deliberately generic so the demo never implies real
// workloads or customer data.
var demoLogMessages = []string{
	"handled request method=GET path=/healthz status=200 duration=3ms",
	"synthetic sample data only; no live cluster is attached",
	"cache warm complete entries=128",
	"reconcile loop finished changed=0",
	"published metrics sample points=42",
	"heartbeat ok peer=demo-control-plane",
}

var demoLogLevels = []string{"info", "info", "info", "debug", "warn", "info"}

func demoLogLine(request logStreamRequest, at time.Time, index int) string {
	container := request.Container
	if container == "" {
		container = "app"
	}
	slot := index % len(demoLogMessages)
	return fmt.Sprintf(
		"%s %-5s [%s/%s:%s] %s",
		at.Format(time.RFC3339),
		demoLogLevels[slot],
		request.Namespace,
		request.Pod,
		container,
		demoLogMessages[slot],
	)
}

func writeSSEHeaders(w http.ResponseWriter) {
	headers := w.Header()
	headers.Set("Content-Type", "text/event-stream")
	headers.Set("Cache-Control", "no-store")
	headers.Set("Connection", "keep-alive")
	headers.Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
}

// writeSSEData emits one log line. Embedded newlines are expanded into
// multiple data fields so a single line always arrives as a single event.
func writeSSEData(w http.ResponseWriter, line string) error {
	line = strings.ReplaceAll(line, "\r\n", "\n")
	for _, part := range strings.Split(line, "\n") {
		if _, err := fmt.Fprintf(w, "data: %s\n", part); err != nil {
			return err
		}
	}
	_, err := fmt.Fprint(w, "\n")
	return err
}

func writeSSEEvent(w http.ResponseWriter, name, payload string) {
	_, _ = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", name, payload)
}

func writeSSEError(w http.ResponseWriter, message string) {
	payload, err := json.Marshal(map[string]string{"message": message})
	if err != nil {
		payload = []byte(`{"message":"log stream failed"}`)
	}
	writeSSEEvent(w, sseStreamErrorEvent, string(payload))
}

func parseFollow(raw string) bool {
	if strings.TrimSpace(raw) == "" {
		return true
	}
	follow, err := strconv.ParseBool(strings.TrimSpace(raw))
	if err != nil {
		return true
	}
	return follow
}

func parseTailLines(raw string) int64 {
	value, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || value <= 0 {
		return defaultLogTailLines
	}
	if value > maxLogTailLines {
		return maxLogTailLines
	}
	return value
}
