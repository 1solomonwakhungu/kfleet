package registrar

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/1solomonwakhungu/kfleet/internal/agent/config"
)

func TestRegisterKeepsBootstrapTokenAfterRuntimeTokenRotation(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if got := r.Header.Get("Authorization"); got != "Bearer bootstrap-token" {
			t.Errorf("registration %d Authorization = %q, want bootstrap token", requests, got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"clusterId":"cluster-a","token":"runtime-token"}`))
	}))
	t.Cleanup(server.Close)

	registrar := New(&config.Config{
		HubURL:      server.URL,
		ClusterName: "cluster-a",
		HubToken:    "bootstrap-token",
	}, nil)
	if _, err := registrar.Register(context.Background(), "v1.32.3"); err != nil {
		t.Fatalf("first Register() error = %v", err)
	}
	if _, err := registrar.Register(context.Background(), "v1.32.3"); err != nil {
		t.Fatalf("second Register() error = %v", err)
	}
	if requests != 2 || registrar.Token() != "runtime-token" {
		t.Fatalf("requests/token = (%d, %q), want (2, runtime-token)", requests, registrar.Token())
	}
}
