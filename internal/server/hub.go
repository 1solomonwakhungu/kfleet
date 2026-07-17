package server

import (
	"context"
	"log/slog"

	"github.com/1solomonwakhungu/kfleet/pkg/types"
	"github.com/coder/websocket"
)

const (
	broadcastQueueSize = 64
	clientQueueSize    = 16
)

// ClusterUpdate is a real-time update about a registered cluster.
type ClusterUpdate struct {
	Type    string        `json:"type"`
	Cluster types.Cluster `json:"cluster"`
}

type wsClient struct {
	conn       *websocket.Conn
	send       chan ClusterUpdate
	registered chan struct{}
	closed     chan struct{}
}

// BroadcastHub coordinates WebSocket clients and cluster updates.
type BroadcastHub struct {
	logger     *slog.Logger
	clients    map[*wsClient]struct{}
	register   chan *wsClient
	unregister chan *wsClient
	broadcast  chan ClusterUpdate
	done       chan struct{}
}

// NewBroadcastHub creates an empty broadcast hub.
func NewBroadcastHub(logger *slog.Logger) *BroadcastHub {
	return &BroadcastHub{
		logger:     logger,
		clients:    make(map[*wsClient]struct{}),
		register:   make(chan *wsClient),
		unregister: make(chan *wsClient, broadcastQueueSize),
		broadcast:  make(chan ClusterUpdate, broadcastQueueSize),
		done:       make(chan struct{}),
	}
}

// Run processes client registration and broadcasts until ctx is cancelled.
func (h *BroadcastHub) Run(ctx context.Context) {
	defer close(h.done)
	for {
		select {
		case <-ctx.Done():
			for client := range h.clients {
				h.removeClient(client, true)
			}
			return
		case client := <-h.register:
			h.clients[client] = struct{}{}
			close(client.registered)
		case client := <-h.unregister:
			h.removeClient(client, false)
		case update := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- update:
				default:
					h.logger.Warn("dropping slow WebSocket client")
					h.removeClient(client, true)
				}
			}
		}
	}
}

// Broadcast queues an update without waiting for clients to consume it.
func (h *BroadcastHub) Broadcast(update ClusterUpdate) {
	select {
	case h.broadcast <- update:
	default:
		h.logger.Warn("dropping cluster update because the broadcast queue is full", "type", update.Type)
	}
}

func (h *BroadcastHub) registerClient(client *wsClient) bool {
	select {
	case h.register <- client:
	case <-h.done:
		return false
	}

	select {
	case <-client.registered:
		return true
	case <-h.done:
		return false
	}
}

func (h *BroadcastHub) unregisterClient(client *wsClient) {
	select {
	case h.unregister <- client:
	case <-h.done:
	}
}

func (h *BroadcastHub) removeClient(client *wsClient, closeConnection bool) {
	if _, ok := h.clients[client]; !ok {
		return
	}
	delete(h.clients, client)
	close(client.send)
	close(client.closed)
	if closeConnection && client.conn != nil {
		_ = client.conn.CloseNow()
	}
}
