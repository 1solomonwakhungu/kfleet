package server

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/1solomonwakhungu/kfleet/pkg/api"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

const (
	websocketWriteTimeout = 10 * time.Second
	websocketPingInterval = 30 * time.Second
)

func (s *Server) handleWSClusters(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := tenantIDFromRequest(w, r)
	if !ok {
		return
	}
	clusters, err := s.store.ListClustersForTenant(r.Context(), tenantID)
	if err != nil {
		s.logger.Error("failed to create WebSocket cluster snapshot", "error", err)
		api.WriteError(w, http.StatusInternalServerError, "failed to list clusters")
		return
	}

	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		s.logger.Error("failed to accept WebSocket connection", "error", err)
		return
	}
	defer func() {
		_ = conn.Close(websocket.StatusNormalClosure, "")
	}()

	queueSize := clientQueueSize
	if len(clusters) > queueSize {
		queueSize = len(clusters)
	}
	client := &wsClient{
		conn:       conn,
		tenantID:   tenantID,
		send:       make(chan ClusterUpdate, queueSize),
		registered: make(chan struct{}),
		closed:     make(chan struct{}),
	}
	if !s.broadcast.registerClient(client) {
		return
	}
	defer s.broadcast.unregisterClient(client)

	for _, cluster := range clusters {
		client.send <- ClusterUpdate{Type: "snapshot", Cluster: cluster}
	}

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	readErr := make(chan error, 1)
	writeErr := make(chan error, 1)
	go func() {
		readErr <- readWebSocket(ctx, conn)
	}()
	go func() {
		writeErr <- writeWebSocket(ctx, conn, client.send)
	}()

	select {
	case <-ctx.Done():
	case err := <-readErr:
		s.logger.Debug("WebSocket reader stopped", "error", err)
	case err := <-writeErr:
		s.logger.Debug("WebSocket writer stopped", "error", err)
	}
}

func readWebSocket(ctx context.Context, conn *websocket.Conn) error {
	for {
		_, reader, err := conn.Reader(ctx)
		if err != nil {
			return err
		}
		if _, err := io.Copy(io.Discard, reader); err != nil {
			return err
		}
	}
}

func writeWebSocket(ctx context.Context, conn *websocket.Conn, updates <-chan ClusterUpdate) error {
	ticker := time.NewTicker(websocketPingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case update, ok := <-updates:
			if !ok {
				return nil
			}
			writeCtx, cancel := context.WithTimeout(ctx, websocketWriteTimeout)
			err := wsjson.Write(writeCtx, conn, update)
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
