package logs

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/1solomonwakhungu/kfleet/pkg/api"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

const (
	// sendQueueSize bounds pending agent -> hub frames. Log readers block
	// on a full queue, which applies backpressure instead of buffering an
	// unbounded amount of output in the agent.
	sendQueueSize = 256
	// writeTimeout bounds a single WebSocket write.
	writeTimeout = 10 * time.Second
	// maxFrameBytes matches the hub side read limit.
	maxFrameBytes = 1 << 20

	minReconnectDelay = time.Second
	maxReconnectDelay = 30 * time.Second
)

// Dialer opens the reverse channel to the hub. It is injectable so tests can
// avoid a real network dial.
type Dialer func(ctx context.Context, endpoint string, options *websocket.DialOptions) (*websocket.Conn, *http.Response, error)

// Client maintains the agent's outbound log channel to the hub.
type Client struct {
	endpoint string
	token    string
	tenantID string
	streamer *Streamer
	logger   *slog.Logger
	dial     Dialer
}

// NewClient builds a log channel client. hubURL is the hub base URL (http or
// https); it is rewritten to the matching WebSocket scheme.
func NewClient(hubURL, clusterName, token, tenantID string, streamer *Streamer, logger *slog.Logger) *Client {
	return &Client{
		endpoint: logChannelEndpoint(hubURL, clusterName),
		token:    token,
		tenantID: tenantID,
		streamer: streamer,
		logger:   logger,
		dial:     websocket.Dial,
	}
}

func logChannelEndpoint(hubURL, clusterName string) string {
	endpoint := strings.TrimRight(strings.TrimSpace(hubURL), "/")
	switch {
	case strings.HasPrefix(endpoint, "https://"):
		endpoint = "wss://" + strings.TrimPrefix(endpoint, "https://")
	case strings.HasPrefix(endpoint, "http://"):
		endpoint = "ws://" + strings.TrimPrefix(endpoint, "http://")
	}
	return endpoint + "/api/v1/agents/" + url.PathEscape(clusterName) + "/logs"
}

// Run keeps the channel connected until ctx is cancelled, reconnecting with
// bounded exponential backoff.
func (c *Client) Run(ctx context.Context) {
	delay := minReconnectDelay
	for ctx.Err() == nil {
		err := c.connect(ctx)
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			c.logger.Warn("agent log channel disconnected", "error", err)
		}
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
		delay *= 2
		if delay > maxReconnectDelay {
			delay = maxReconnectDelay
		}
	}
}

func (c *Client) connect(ctx context.Context) error {
	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+c.token)
	if c.tenantID != "" {
		headers.Set("X-Kfleet-Tenant-ID", c.tenantID)
	}
	conn, response, err := c.dial(ctx, c.endpoint, &websocket.DialOptions{HTTPHeader: headers})
	if response != nil && response.Body != nil {
		_ = response.Body.Close()
	}
	if err != nil {
		return err
	}
	defer func() {
		_ = conn.Close(websocket.StatusNormalClosure, "")
	}()
	conn.SetReadLimit(maxFrameBytes)
	c.logger.Info("agent log channel connected")
	return c.serve(ctx, conn)
}

// serve multiplexes hub requests over one connection until it fails.
func (c *Client) serve(ctx context.Context, conn *websocket.Conn) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	outbound := make(chan api.LogStreamMessage, sendQueueSize)
	var streams sync.Map
	var workers sync.WaitGroup
	defer func() {
		cancel()
		workers.Wait()
	}()

	writeErr := make(chan error, 1)
	go func() {
		writeErr <- writeFrames(ctx, conn, outbound)
	}()

	readErr := make(chan error, 1)
	go func() {
		readErr <- func() error {
			for {
				var message api.LogStreamMessage
				if err := wsjson.Read(ctx, conn, &message); err != nil {
					return err
				}
				switch message.Type {
				case api.LogStreamStart:
					c.startStream(ctx, message, outbound, &streams, &workers)
				case api.LogStreamStop:
					if cancelStream, ok := streams.LoadAndDelete(message.StreamID); ok {
						cancelStream.(context.CancelFunc)()
					}
				}
			}
		}()
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-readErr:
		return err
	case err := <-writeErr:
		return err
	}
}

func (c *Client) startStream(
	ctx context.Context,
	message api.LogStreamMessage,
	outbound chan api.LogStreamMessage,
	streams *sync.Map,
	workers *sync.WaitGroup,
) {
	if message.StreamID == "" {
		return
	}
	streamCtx, cancel := context.WithCancel(ctx)
	if _, loaded := streams.LoadOrStore(message.StreamID, context.CancelFunc(cancel)); loaded {
		cancel()
		return
	}
	workers.Add(1)
	go func() {
		defer workers.Done()
		defer cancel()
		defer streams.Delete(message.StreamID)
		c.runStream(streamCtx, message, outbound)
	}()
}

func (c *Client) runStream(ctx context.Context, message api.LogStreamMessage, outbound chan api.LogStreamMessage) {
	request := Request{
		Namespace: message.Namespace,
		Pod:       message.Pod,
		Container: message.Container,
		Follow:    message.Follow,
		TailLines: message.TailLines,
	}
	err := c.streamer.Stream(ctx, request, func(line string) error {
		return sendFrame(ctx, outbound, api.LogStreamMessage{
			Type:     api.LogStreamData,
			StreamID: message.StreamID,
			Line:     line,
		})
	})
	if ctx.Err() != nil {
		// The hub cancelled the stream or the channel closed; the hub is
		// no longer interested in a terminal frame.
		return
	}
	end := api.LogStreamMessage{Type: api.LogStreamEnd, StreamID: message.StreamID}
	if err != nil && !errors.Is(err, context.Canceled) {
		end.Error = err.Error()
		c.logger.Warn("pod log stream failed",
			"namespace", message.Namespace, "pod", message.Pod, "error", err)
	}
	_ = sendFrame(ctx, outbound, end)
}

func sendFrame(ctx context.Context, outbound chan api.LogStreamMessage, message api.LogStreamMessage) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case outbound <- message:
		return nil
	}
}

func writeFrames(ctx context.Context, conn *websocket.Conn, outbound <-chan api.LogStreamMessage) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case message := <-outbound:
			writeCtx, cancel := context.WithTimeout(ctx, writeTimeout)
			err := wsjson.Write(writeCtx, conn, message)
			cancel()
			if err != nil {
				return err
			}
		}
	}
}
