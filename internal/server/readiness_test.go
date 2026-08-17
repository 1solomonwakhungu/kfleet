package server

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/1solomonwakhungu/kfleet/internal/config"
	"github.com/1solomonwakhungu/kfleet/internal/store"
)

// TestReadyzReflectsStoreAvailability proves readiness probes the database
// instead of unconditionally reporting healthy, while liveness stays a pure
// process check.
func TestReadyzReflectsStoreAvailability(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "kfleet.db"))
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := New(&config.Config{ListenAddr: ":0"}, logger, st)
	httpServer := httptest.NewServer(srv.httpServer.Handler)
	t.Cleanup(httpServer.Close)

	healthy := getStatus(t, httpServer.URL+"/readyz")
	if healthy != http.StatusOK {
		t.Fatalf("GET /readyz status = %d, want %d", healthy, http.StatusOK)
	}

	if err := st.Close(); err != nil {
		t.Fatalf("store.Close() error = %v", err)
	}

	response, err := http.Get(httpServer.URL + "/readyz")
	if err != nil {
		t.Fatalf("GET /readyz error = %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("GET /readyz with a closed store status = %d, want %d", response.StatusCode, http.StatusServiceUnavailable)
	}
	var body struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode readiness body: %v", err)
	}
	if body.Status == "" || body.Reason == "" {
		t.Fatalf("readiness body = %#v, want a short JSON explanation", body)
	}

	// Liveness must stay independent of the database.
	if live := getStatus(t, httpServer.URL+"/healthz"); live != http.StatusOK {
		t.Fatalf("GET /healthz status = %d, want %d", live, http.StatusOK)
	}
}

// TestStorePingFailsAfterClose covers the store probe used by readiness.
func TestStorePingFailsAfterClose(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "kfleet.db"))
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	if err := st.Ping(context.Background()); err != nil {
		t.Fatalf("Ping() on an open store error = %v", err)
	}
	if err := st.Close(); err != nil {
		t.Fatalf("store.Close() error = %v", err)
	}
	if err := st.Ping(context.Background()); err == nil {
		t.Fatal("Ping() on a closed store returned nil error")
	}
}

func getStatus(t *testing.T, url string) int {
	t.Helper()
	response, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s error = %v", url, err)
	}
	defer response.Body.Close()
	return response.StatusCode
}
