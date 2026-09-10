package server

import (
	"encoding/json"
	"net/http"

	"github.com/1solomonwakhungu/kfleet/pkg/api"
)

// apiJSONFallback returns a handler for /api requests that no registered route
// matched. It replays the request against a shadow mux holding the same routes
// but no catch-all, so a known path requested with the wrong method produces
// the mux's method-not-allowed response (including its Allow header) and an
// unknown path produces a plain 404. Both are rewritten to the JSON error
// contract shared by the rest of /api, so clients never receive plain-text
// errors from the API surface. The shadow mux never dispatches to real
// handlers: any request it can route is routed by the primary mux first.
func apiJSONFallback(registerRoutes func(*http.ServeMux)) http.Handler {
	shadow := http.NewServeMux()
	registerRoutes(shadow)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		shadow.ServeHTTP(&jsonErrorWriter{ResponseWriter: w}, r)
	})
}

// jsonErrorWriter rewrites the net/http mux's built-in plain-text 404 and 405
// fallbacks into the JSON error contract defined by api.ErrorResponse. It is
// only ever used behind apiJSONFallback, so real handler responses are
// unaffected.
type jsonErrorWriter struct {
	http.ResponseWriter
	wroteHeader bool
	replaced    bool
}

func (w *jsonErrorWriter) WriteHeader(status int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true
	switch status {
	case http.StatusNotFound:
		w.writeJSONError(http.StatusNotFound, "not found")
	case http.StatusMethodNotAllowed:
		w.writeJSONError(http.StatusMethodNotAllowed, "method not allowed")
	default:
		w.ResponseWriter.WriteHeader(status)
	}
}

func (w *jsonErrorWriter) writeJSONError(status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.ResponseWriter.WriteHeader(status)
	w.replaced = true
	_ = json.NewEncoder(w.ResponseWriter).Encode(api.ErrorResponse{Error: message, Code: status})
}

func (w *jsonErrorWriter) Write(p []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	if w.replaced {
		return len(p), nil
	}
	return w.ResponseWriter.Write(p)
}
