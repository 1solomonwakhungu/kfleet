package logs

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/1solomonwakhungu/kfleet/pkg/api"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// fakeHub accepts one agent log channel and hands the connection back.
func fakeHub(t *testing.T, handle func(ctx context.Context, conn *websocket.Conn)) (*httptest.Server, chan http.Header) {
	t.Helper()
	headers := make(chan http.Header, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case headers <- r.Header.Clone():
		default:
		}
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			t.Errorf("websocket.Accept() error = %v", err)
			return
		}
		defer func() { _ = conn.Close(websocket.StatusNormalClosure, "") }()
		handle(r.Context(), conn)
	}))
	t.Cleanup(server.Close)
	return server, headers
}

func TestClientStreamsLogsBackToHub(t *testing.T) {
	frames := make(chan api.LogStreamMessage, 16)
	server, headers := fakeHub(t, func(ctx context.Context, conn *websocket.Conn) {
		if err := wsjson.Write(ctx, conn, api.LogStreamMessage{
			Type: api.LogStreamStart, StreamID: "s1", Namespace: "apps", Pod: "api", TailLines: 5,
		}); err != nil {
			t.Errorf("write start error = %v", err)
			return
		}
		for {
			var message api.LogStreamMessage
			if err := wsjson.Read(ctx, conn, &message); err != nil {
				return
			}
			frames <- message
			if message.Type == api.LogStreamEnd {
				return
			}
		}
	})

	opener := &stubOpener{reader: nopCloser{strings.NewReader("alpha\nbravo\n")}}
	client := NewClient(server.URL, "prod", "agent-token", "acme", NewStreamerWithOpener(opener), testLogger())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = client.connect(ctx)
	}()

	var lines []string
	for {
		select {
		case message := <-frames:
			switch message.Type {
			case api.LogStreamData:
				lines = append(lines, message.Line)
				continue
			case api.LogStreamEnd:
				if message.Error != "" {
					t.Fatalf("end error = %q, want none", message.Error)
				}
				if strings.Join(lines, "|") != "alpha|bravo" {
					t.Fatalf("lines = %#v, want alpha, bravo", lines)
				}
				cancel()
				<-done
				got := <-headers
				if got.Get("Authorization") != "Bearer agent-token" {
					t.Fatalf("authorization header = %q", got.Get("Authorization"))
				}
				if got.Get("X-Kfleet-Tenant-ID") != "acme" {
					t.Fatalf("tenant header = %q", got.Get("X-Kfleet-Tenant-ID"))
				}
				return
			}
		case <-ctx.Done():
			t.Fatal("timed out waiting for log frames")
		}
	}
}

func TestClientStopsStreamOnStopFrame(t *testing.T) {
	reader := newBlockingReader()
	stopped := make(chan struct{})
	server, _ := fakeHub(t, func(ctx context.Context, conn *websocket.Conn) {
		if err := wsjson.Write(ctx, conn, api.LogStreamMessage{
			Type: api.LogStreamStart, StreamID: "s1", Namespace: "apps", Pod: "api", Follow: true,
		}); err != nil {
			return
		}
		// Give the agent time to open the stream before cancelling it.
		time.Sleep(50 * time.Millisecond)
		if err := wsjson.Write(ctx, conn, api.LogStreamMessage{Type: api.LogStreamStop, StreamID: "s1"}); err != nil {
			return
		}
		<-ctx.Done()
	})

	client := NewClient(server.URL, "prod", "agent-token", "", NewStreamerWithOpener(&stubOpener{reader: reader}), testLogger())
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	go func() {
		_ = client.connect(ctx)
	}()

	go func() {
		<-reader.closed
		close(stopped)
	}()
	select {
	case <-stopped:
	case <-ctx.Done():
		t.Fatal("stop frame did not terminate the Kubernetes log reader")
	}
}

func TestClientReportsStreamFailureToHub(t *testing.T) {
	frames := make(chan api.LogStreamMessage, 4)
	server, _ := fakeHub(t, func(ctx context.Context, conn *websocket.Conn) {
		if err := wsjson.Write(ctx, conn, api.LogStreamMessage{
			Type: api.LogStreamStart, StreamID: "s1", Namespace: "apps", Pod: "missing",
		}); err != nil {
			return
		}
		for {
			var message api.LogStreamMessage
			if err := wsjson.Read(ctx, conn, &message); err != nil {
				return
			}
			frames <- message
			if message.Type == api.LogStreamEnd {
				return
			}
		}
	})

	client := NewClient(server.URL, "prod", "token", "",
		NewStreamerWithOpener(&stubOpener{err: http.ErrNotSupported}), testLogger())
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	go func() { _ = client.connect(ctx) }()

	select {
	case message := <-frames:
		if message.Type != api.LogStreamEnd || message.Error == "" {
			t.Fatalf("frame = %#v, want terminal error frame", message)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for the error frame")
	}
}
