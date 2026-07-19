package health

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerReportsLivenessAndReadiness(t *testing.T) {
	handler := Handler()
	for _, path := range []string{"/healthz", "/readyz"} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		if recorder.Code != http.StatusOK {
			t.Fatalf("GET %s status = %d, want %d", path, recorder.Code, http.StatusOK)
		}
	}
}
