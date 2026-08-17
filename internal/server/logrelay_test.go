package server

import (
	"io"
	"log/slog"
	"testing"

	"github.com/1solomonwakhungu/kfleet/pkg/api"
)

func newTestRelay() *LogRelay {
	return NewLogRelay(slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func TestLogRelayOpenWithoutAgent(t *testing.T) {
	relay := newTestRelay()
	if relay.Connected("cluster-1") {
		t.Fatal("Connected() = true for an unknown cluster")
	}
	if _, err := relay.open("cluster-1", logStreamRequest{Namespace: "apps", Pod: "api"}); err != errNoAgentConnected {
		t.Fatalf("open() error = %v, want errNoAgentConnected", err)
	}
}

func TestLogRelayRoutesFramesToStream(t *testing.T) {
	relay := newTestRelay()
	conn := newAgentLogConn("cluster-1")
	relay.register(conn)
	defer relay.unregister(conn)

	stream, err := relay.open("cluster-1", logStreamRequest{Namespace: "apps", Pod: "api", TailLines: 10})
	if err != nil {
		t.Fatalf("open() error = %v", err)
	}
	start := <-conn.send
	if start.Type != api.LogStreamStart || start.Namespace != "apps" || start.TailLines != 10 {
		t.Fatalf("start frame = %#v", start)
	}

	conn.deliver(api.LogStreamMessage{Type: api.LogStreamData, StreamID: start.StreamID, Line: "hello"})
	conn.deliver(api.LogStreamMessage{Type: api.LogStreamEnd, StreamID: start.StreamID})
	if message := <-stream.Messages(); message.Line != "hello" {
		t.Fatalf("data frame = %#v", message)
	}
	if message := <-stream.Messages(); message.Type != api.LogStreamEnd {
		t.Fatalf("terminal frame = %#v", message)
	}
	if _, open := <-stream.Messages(); open {
		t.Fatal("stream channel stayed open after the end frame")
	}

	// Frames for an unknown stream are dropped rather than panicking.
	conn.deliver(api.LogStreamMessage{Type: api.LogStreamData, StreamID: "gone", Line: "ignored"})
}

func TestLogRelayCloseSendsStop(t *testing.T) {
	relay := newTestRelay()
	conn := newAgentLogConn("cluster-1")
	relay.register(conn)
	defer relay.unregister(conn)

	stream, err := relay.open("cluster-1", logStreamRequest{Namespace: "apps", Pod: "api"})
	if err != nil {
		t.Fatalf("open() error = %v", err)
	}
	<-conn.send
	stream.Close()
	stop := <-conn.send
	if stop.Type != api.LogStreamStop || stop.StreamID != stream.id {
		t.Fatalf("stop frame = %#v", stop)
	}
	if _, open := <-stream.Messages(); open {
		t.Fatal("Close() left the stream channel open")
	}
	// A second Close must not queue another stop frame or panic.
	stream.Close()
	select {
	case extra := <-conn.send:
		t.Fatalf("second Close() queued %#v", extra)
	default:
	}
}

func TestLogRelayEndsStreamThatFallsBehind(t *testing.T) {
	relay := newTestRelay()
	conn := newAgentLogConn("cluster-1")
	relay.register(conn)
	defer relay.unregister(conn)

	stream, err := relay.open("cluster-1", logStreamRequest{Namespace: "apps", Pod: "api"})
	if err != nil {
		t.Fatalf("open() error = %v", err)
	}
	<-conn.send
	for range logStreamQueueSize + 5 {
		conn.deliver(api.LogStreamMessage{Type: api.LogStreamData, StreamID: stream.id, Line: "line"})
	}
	drained := 0
	for range stream.Messages() {
		drained++
	}
	if drained != logStreamQueueSize {
		t.Fatalf("drained %d frames, want %d before the stream was cut", drained, logStreamQueueSize)
	}
}

func TestLogRelayReplacesReconnectingAgent(t *testing.T) {
	relay := newTestRelay()
	first := newAgentLogConn("cluster-1")
	relay.register(first)
	stream, err := relay.open("cluster-1", logStreamRequest{Namespace: "apps", Pod: "api"})
	if err != nil {
		t.Fatalf("open() error = %v", err)
	}
	<-first.send

	second := newAgentLogConn("cluster-1")
	relay.register(second)
	if _, open := <-stream.Messages(); open {
		t.Fatal("streams on the replaced connection were not terminated")
	}
	if !relay.Connected("cluster-1") {
		t.Fatal("Connected() = false after reconnect")
	}

	// Unregistering the stale connection must not detach the live one.
	relay.unregister(first)
	if !relay.Connected("cluster-1") {
		t.Fatal("unregistering a replaced connection detached the live agent")
	}
	relay.unregister(second)
	if relay.Connected("cluster-1") {
		t.Fatal("Connected() = true after the agent disconnected")
	}
	if _, err := relay.open("cluster-1", logStreamRequest{Namespace: "apps", Pod: "api"}); err != errNoAgentConnected {
		t.Fatalf("open() error = %v, want errNoAgentConnected", err)
	}
}

func TestLogRelayOpenOnClosedConnection(t *testing.T) {
	relay := newTestRelay()
	conn := newAgentLogConn("cluster-1")
	relay.register(conn)
	conn.closeAll()
	if _, err := relay.open("cluster-1", logStreamRequest{Namespace: "apps", Pod: "api"}); err != errNoAgentConnected {
		t.Fatalf("open() error = %v, want errNoAgentConnected", err)
	}
}
