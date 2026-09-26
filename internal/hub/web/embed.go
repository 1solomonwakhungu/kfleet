// Package web serves the hub's embedded React application.
package web

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"regexp"
	"strings"
)

// dist contains the production frontend bundle.
//
//go:embed all:dist
var dist embed.FS

// Empty reports whether the embedded frontend bundle has no index.html, which
// happens when the web UI was not built before the binary was compiled. The
// hub still serves API and MCP clients in that state.
func Empty() bool {
	_, err := fs.ReadFile(dist, "dist/index.html")
	return err != nil
}

// Handler returns an HTTP handler for static frontend assets with an index.html
// fallback for client-side routes. API and WebSocket paths never use the SPA fallback.
func Handler() http.Handler {
	root, err := fs.Sub(dist, "dist")
	if err != nil {
		return http.NotFoundHandler()
	}
	return HandlerFor(root)
}

// HandlerFor serves static frontend assets from root with an index.html
// fallback for client-side routes. It applies caching headers that keep the
// document fresh: content-hashed files under assets/ are immutable, while the
// HTML document (including SPA fallbacks) and non-hashed files must be
// revalidated on every load so deploys take effect without a hard refresh.
func HandlerFor(root fs.FS) http.Handler {
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
				w.Header().Set("Cache-Control", cacheControl(assetPath))
				files.ServeHTTP(w, r)
				return
			}
			// Bundled assets are individually addressable files; a missing one
			// must 404 instead of falling back to the HTML document.
			if strings.HasPrefix(assetPath, "assets/") {
				w.Header().Set("Cache-Control", "no-cache")
				http.NotFound(w, r)
				return
			}
		}

		index, readErr := fs.ReadFile(root, "index.html")
		if readErr != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		// The HTML document must always be revalidated: it references the
		// content-hashed asset files, so a stale document would keep pointing
		// users at a previous deployment's assets.
		w.Header().Set("Cache-Control", "no-cache")
		_, _ = w.Write(index)
	})
}

// contentHashPattern matches Vite's default content-hashed asset names:
// <name>-<8 hash chars>.<ext> (e.g. index-DgAqilCc.css).
var contentHashPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.]*-[A-Za-z0-9_-]{8}\.[A-Za-z0-9]+$`)

// cacheControl returns the Cache-Control value for a served static file.
// Only content-hashed files under assets/ are safe to cache forever; other
// asset files (e.g. runtime helpers) and non-hashed root files are
// revalidated on every load.
func cacheControl(assetPath string) string {
	if strings.HasPrefix(assetPath, "assets/") &&
		contentHashPattern.MatchString(path.Base(assetPath)) {
		return "public, max-age=31536000, immutable"
	}
	return "no-cache"
}
