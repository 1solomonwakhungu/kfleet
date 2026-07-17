// Package web serves the hub's embedded React application.
package web

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// dist contains the production frontend bundle.
//
//go:embed all:dist
var dist embed.FS

// Handler returns an HTTP handler for static frontend assets with an index.html
// fallback for client-side routes. API and WebSocket paths never use the SPA fallback.
func Handler() http.Handler {
	root, err := fs.Sub(dist, "dist")
	if err != nil {
		return http.NotFoundHandler()
	}
	files := http.FileServer(http.FS(root))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/api" ||
			strings.HasPrefix(r.URL.Path, "/ws/") || r.URL.Path == "/ws" {
			http.NotFound(w, r)
			return
		}

		assetPath := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if assetPath != "." && assetPath != "" {
			if info, statErr := fs.Stat(root, assetPath); statErr == nil && !info.IsDir() {
				files.ServeHTTP(w, r)
				return
			}
		}

		index, readErr := fs.ReadFile(root, "index.html")
		if readErr != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(index)
	})
}
