package registrar

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/1solomonwakhungu/kfleet/internal/agent/config"
	"github.com/1solomonwakhungu/kfleet/internal/version"
)

// TestRegisterReportsStampedVersion proves the agent reports the version
// stamped at build time rather than a hard-coded constant.
func TestRegisterReportsStampedVersion(t *testing.T) {
	original := version.Version
	t.Cleanup(func() { version.Version = original })
	version.Version = "v9.9.9"

	var reported string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read registration body: %v", err)
			return
		}
		var payload RegisterRequest
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Errorf("decode registration body: %v", err)
			return
		}
		reported = payload.AgentVersion
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"clusterId":"cluster-a","token":"runtime-token"}`))
	}))
	t.Cleanup(server.Close)

	registrar := New(&config.Config{
		HubURL:      server.URL,
		ClusterName: "cluster-a",
		HubToken:    "bootstrap-token",
		TenantID:    "default",
	}, nil)
	if _, err := registrar.Register(context.Background(), "v1.32.3"); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if reported != "v9.9.9" {
		t.Fatalf("reported agent version = %q, want the stamped build version", reported)
	}
}
