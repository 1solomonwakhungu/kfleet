package main

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/1solomonwakhungu/kfleet/internal/agent/config"
	"github.com/1solomonwakhungu/kfleet/internal/agent/registrar"
)

func TestRegisterWithRetryThenSuccess(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		requests++
		if request.Method != http.MethodPost || request.URL.Path != "/api/v1/agents/register" {
			t.Errorf("request = %s %s", request.Method, request.URL.Path)
		}
		var payload registrar.RegisterRequest
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Errorf("decode registration request: %v", err)
		}
		if payload.ClusterName != "production" || payload.K8sVersion != "v1.32.0" {
			t.Errorf("registration request = %#v", payload)
		}
		if requests == 1 {
			http.Error(w, "try again", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"clusterId":"cluster-1","token":"rotated-token"}`)
	}))
	defer server.Close()

	client := registrar.New(&config.Config{
		HubURL:      server.URL,
		HubToken:    "bootstrap-token",
		ClusterName: "production",
	}, nil)
	backoff := &recordingBackoff{delay: 25 * time.Millisecond}
	waits := 0

	registration, ok := registerWithRetry(
		context.Background(),
		client,
		"v1.32.0",
		false,
		backoff,
		func(_ context.Context, delay time.Duration) bool {
			waits++
			if delay != backoff.delay {
				t.Errorf("retry delay = %v, want %v", delay, backoff.delay)
			}
			return true
		},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)

	if !ok {
		t.Fatal("registerWithRetry() ok = false, want true")
	}
	if registration.ClusterID != "cluster-1" || !registration.Approved {
		t.Fatalf("registerWithRetry() response = %#v", registration)
	}
	if requests != 2 || waits != 1 || backoff.nextCalls != 1 || backoff.resetCalls != 1 {
		t.Fatalf(
			"requests = %d, waits = %d, Next() = %d, Reset() = %d; want 2, 1, 1, 1",
			requests,
			waits,
			backoff.nextCalls,
			backoff.resetCalls,
		)
	}
	if got := client.Token(); got != "rotated-token" {
		t.Fatalf("Token() = %q, want rotated token", got)
	}
}

func TestRunAgentLoopSendsPeriodicHeartbeatsAndStopsOnCancellation(t *testing.T) {
	var heartbeatCount atomic.Int32
	heartbeatReceived := make(chan struct{}, 4)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.URL.Path != "/api/v1/agents/production/heartbeat" {
			t.Errorf("request = %s %s", request.Method, request.URL.Path)
		}
		heartbeatCount.Add(1)
		heartbeatReceived <- struct{}{}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	agentRegistrar := registrar.New(&config.Config{
		HubURL:      server.URL,
		HubToken:    "agent-token",
		ClusterName: "production",
	}, nil)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan bool, 1)
	go func() {
		done <- runAgentLoop(
			ctx,
			10*time.Millisecond,
			time.Hour,
			nil,
			nil,
			agentRegistrar,
			slog.New(slog.NewTextHandler(io.Discard, nil)),
		)
	}()

	for range 2 {
		select {
		case <-heartbeatReceived:
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for periodic heartbeat")
		}
	}
	cancel()

	select {
	case cancelled := <-done:
		if !cancelled {
			t.Fatal("runAgentLoop() = false after cancellation, want true")
		}
	case <-time.After(time.Second):
		t.Fatal("runAgentLoop() did not stop after cancellation")
	}

	countAfterCancellation := heartbeatCount.Load()
	time.Sleep(30 * time.Millisecond)
	if got := heartbeatCount.Load(); got != countAfterCancellation {
		t.Fatalf("heartbeat count after cancellation = %d, want %d", got, countAfterCancellation)
	}
}

type recordingBackoff struct {
	delay      time.Duration
	nextCalls  int
	resetCalls int
}

func (b *recordingBackoff) Next() time.Duration {
	b.nextCalls++
	return b.delay
}

func (b *recordingBackoff) Reset() {
	b.resetCalls++
}
