// Package health provides the agent's local Kubernetes probe endpoints.
package health

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/1solomonwakhungu/kfleet/internal/agent/hubcontact"
)

const shutdownTimeout = 5 * time.Second

// Handler returns the agent's liveness and readiness endpoints. /healthz is
// plain liveness and always answers 200 while the process can serve. /readyz
// answers 200 only when the agent has registered with the hub and a
// registration, heartbeat, or status report succeeded within the tracker's
// staleness window; otherwise it answers 503.
func Handler(tracker *hubcontact.Tracker) http.Handler {
	mux := http.NewServeMux()
	live := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}
	mux.HandleFunc("GET /healthz", live)
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, _ *http.Request) {
		if !tracker.Ready() {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(readyzError{Error: "no successful hub contact"})
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	return mux
}

type readyzError struct {
	Error string `json:"error"`
}

// Serve runs the probe server until ctx is cancelled.
func Serve(ctx context.Context, address string, tracker *hubcontact.Tracker) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	server := &http.Server{
		Handler:           Handler(tracker),
		ReadHeaderTimeout: 2 * time.Second,
	}
	shutdownDone := make(chan struct{})
	go func() {
		defer close(shutdownDone)
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	err = server.Serve(listener)
	if errors.Is(err, http.ErrServerClosed) {
		<-shutdownDone
		return nil
	}
	return err
}
