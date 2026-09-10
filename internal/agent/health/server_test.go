package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/1solomonwakhungu/kfleet/internal/agent/hubcontact"
)

func TestHandlerLivenessAlwaysReportsOK(t *testing.T) {
	handler := Handler(hubcontact.NewTracker(45 * time.Second))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("GET /healthz status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if got := recorder.Body.String(); got != "ok" {
		t.Fatalf("GET /healthz body = %q, want ok", got)
	}
}

// TestHandlerReadinessTracksHubContact proves the readiness probe reports
// honest hub reachability: 503 before the first successful contact, 200
// while contact is fresh, and 503 again once contact goes stale.
func TestHandlerReadinessTracksHubContact(t *testing.T) {
	tracker := hubcontact.NewTracker(45 * time.Second)
	clock := time.Now()
	tracker.SetClockForTesting(func() time.Time { return clock })
	handler := Handler(tracker)

	status, body := getReadiness(t, handler)
	if status != http.StatusServiceUnavailable {
		t.Fatalf("GET /readyz before first contact status = %d, want %d", status, http.StatusServiceUnavailable)
	}
	var decoded struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("decode readiness error body %q: %v", body, err)
	}
	if decoded.Error != "no successful hub contact" {
		t.Fatalf("readiness error = %q, want %q", decoded.Error, "no successful hub contact")
	}

	tracker.MarkRegistered()
	status, _ = getReadiness(t, handler)
	if status != http.StatusOK {
		t.Fatalf("GET /readyz after fresh registration status = %d, want %d", status, http.StatusOK)
	}

	// A contact inside the window keeps the agent ready...
	clock = clock.Add(44 * time.Second)
	tracker.MarkContact()
	clock = clock.Add(40 * time.Second)
	status, _ = getReadiness(t, handler)
	if status != http.StatusOK {
		t.Fatalf("GET /readyz 40s after heartbeat status = %d, want %d", status, http.StatusOK)
	}

	// ...but contact that goes stale flips readiness back to 503.
	clock = clock.Add(6 * time.Second)
	status, body = getReadiness(t, handler)
	if status != http.StatusServiceUnavailable {
		t.Fatalf("GET /readyz after stale contact status = %d, want %d", status, http.StatusServiceUnavailable)
	}
	if string(body) == "" {
		t.Fatal("GET /readyz after stale contact returned an empty body, want an error message")
	}
}

func TestHandlerNilTrackerIsNeverReady(t *testing.T) {
	handler := Handler(nil)
	status, _ := getReadiness(t, handler)
	if status != http.StatusServiceUnavailable {
		t.Fatalf("GET /readyz with no tracker status = %d, want %d", status, http.StatusServiceUnavailable)
	}
}

func getReadiness(t *testing.T, handler http.Handler) (int, []byte) {
	t.Helper()
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	return recorder.Code, recorder.Body.Bytes()
}
