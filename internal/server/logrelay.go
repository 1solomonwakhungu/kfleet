package server

import (
	"errors"
	"log/slog"
	"sync"

	"github.com/1solomonwakhungu/kfleet/pkg/api"
	"github.com/google/uuid"
)

// errNoAgentConnected is returned when a log stream is requested for a
// cluster whose agent has no live reverse channel. The hub has no direct
// Kubernetes access, so this is a terminal condition rather than something
// the browser should retry in a tight loop.
var errNoAgentConnected = errors.New("no agent is connected for this cluster")

// errAgentChannelClosed is returned when the reverse channel goes away
// between picking the connection and queueing the start frame.
var errAgentChannelClosed = errors.New("agent log channel closed")

const (
	// agentSendQueueSize bounds pending hub -> agent control frames.
	agentSendQueueSize = 64
	// logStreamQueueSize bounds buffered log lines per browser stream. A
	// browser that falls further behind than this has its stream ended
	// rather than being allowed to stall the shared agent connection.
	logStreamQueueSize = 512
)

// agentLogConn is one agent's reverse channel plus the streams multiplexed
// over it. All mutable state is guarded by mu so the WebSocket reader
// goroutine and HTTP handler goroutines stay race-free.
type agentLogConn struct {
	clusterID string
	send      chan api.LogStreamMessage

	mu      sync.Mutex
	streams map[string]chan api.LogStreamMessage
	closed  bool
}

func newAgentLogConn(clusterID string) *agentLogConn {
	return &agentLogConn{
		clusterID: clusterID,
		send:      make(chan api.LogStreamMessage, agentSendQueueSize),
		streams:   make(map[string]chan api.LogStreamMessage),
	}
}

// deliver routes an agent frame to the stream that requested it.
func (c *agentLogConn) deliver(message api.LogStreamMessage) {
	c.mu.Lock()
	defer c.mu.Unlock()
	queue, ok := c.streams[message.StreamID]
	if !ok {
		return
	}
	select {
	case queue <- message:
		if message.Type == api.LogStreamEnd {
			delete(c.streams, message.StreamID)
			close(queue)
		}
	default:
		// The consumer fell behind and the buffer is full, so there is no
		// room left even for a terminal frame. Closing the queue signals
		// the SSE handler to report the stream as broken.
		delete(c.streams, message.StreamID)
		close(queue)
	}
}

// closeAll terminates every stream on the connection, which unblocks the
// SSE handlers waiting on them.
func (c *agentLogConn) closeAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}
	c.closed = true
	for id, queue := range c.streams {
		delete(c.streams, id)
		close(queue)
	}
}

func (c *agentLogConn) addStream(id string) (chan api.LogStreamMessage, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil, false
	}
	queue := make(chan api.LogStreamMessage, logStreamQueueSize)
	c.streams[id] = queue
	return queue, true
}

func (c *agentLogConn) removeStream(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	queue, ok := c.streams[id]
	if !ok {
		return
	}
	delete(c.streams, id)
	close(queue)
}

// enqueue queues a control frame for the writer goroutine. It reports
// false when the connection is gone or hopelessly backed up.
func (c *agentLogConn) enqueue(message api.LogStreamMessage) bool {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return false
	}
	c.mu.Unlock()
	select {
	case c.send <- message:
		return true
	default:
		return false
	}
}

// logStreamRequest describes a single browser log request.
type logStreamRequest struct {
	Namespace string
	Pod       string
	Container string
	Follow    bool
	TailLines int64
}

// logStream is the hub side handle for one in-flight browser log stream.
type logStream struct {
	id       string
	conn     *agentLogConn
	messages chan api.LogStreamMessage
	once     sync.Once
}

// Messages returns the frames relayed from the agent. The channel is closed
// when the stream ends or the agent connection is lost.
func (s *logStream) Messages() <-chan api.LogStreamMessage {
	return s.messages
}

// Close tells the agent to stop streaming and releases hub side state. It
// is safe to call more than once.
func (s *logStream) Close() {
	s.once.Do(func() {
		s.conn.enqueue(api.LogStreamMessage{Type: api.LogStreamStop, StreamID: s.id})
		s.conn.removeStream(s.id)
	})
}

// LogRelay tracks the agent reverse channels available to the hub and
// multiplexes browser log requests onto them.
type LogRelay struct {
	logger *slog.Logger

	mu     sync.RWMutex
	agents map[string]*agentLogConn
}

// NewLogRelay creates an empty relay.
func NewLogRelay(logger *slog.Logger) *LogRelay {
	return &LogRelay{logger: logger, agents: make(map[string]*agentLogConn)}
}

// register attaches a reverse channel for clusterID, replacing and closing
// any previous channel so a reconnecting agent never leaves a stale entry.
func (r *LogRelay) register(conn *agentLogConn) {
	r.mu.Lock()
	previous, ok := r.agents[conn.clusterID]
	r.agents[conn.clusterID] = conn
	r.mu.Unlock()
	if ok && previous != conn {
		previous.closeAll()
	}
}

// unregister detaches conn if it is still the active channel for its
// cluster and terminates the streams it was serving.
func (r *LogRelay) unregister(conn *agentLogConn) {
	r.mu.Lock()
	if current, ok := r.agents[conn.clusterID]; ok && current == conn {
		delete(r.agents, conn.clusterID)
	}
	r.mu.Unlock()
	conn.closeAll()
}

// Connected reports whether an agent reverse channel exists for clusterID.
func (r *LogRelay) Connected(clusterID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.agents[clusterID]
	return ok
}

func (r *LogRelay) connection(clusterID string) (*agentLogConn, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	conn, ok := r.agents[clusterID]
	return conn, ok
}

// open asks the cluster's agent to start streaming pod logs. The caller
// must Close the returned stream to release the agent side reader.
func (r *LogRelay) open(clusterID string, request logStreamRequest) (*logStream, error) {
	conn, ok := r.connection(clusterID)
	if !ok {
		return nil, errNoAgentConnected
	}
	id := uuid.NewString()
	queue, ok := conn.addStream(id)
	if !ok {
		return nil, errNoAgentConnected
	}
	start := api.LogStreamMessage{
		Type:      api.LogStreamStart,
		StreamID:  id,
		Namespace: request.Namespace,
		Pod:       request.Pod,
		Container: request.Container,
		Follow:    request.Follow,
		TailLines: request.TailLines,
	}
	if !conn.enqueue(start) {
		conn.removeStream(id)
		return nil, errAgentChannelClosed
	}
	return &logStream{id: id, conn: conn, messages: queue}, nil
}
