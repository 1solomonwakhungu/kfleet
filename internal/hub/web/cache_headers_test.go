package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func cacheTestFS() fstest.MapFS {
	return fstest.MapFS{
		"index.html":             {Data: []byte("<html><body>hub</body></html>")},
		"assets/index-abc12.css": {Data: []byte("body{}")},
		"assets/index-9xy89.js":  {Data: []byte("console.log(1)")},
		"favicon.ico":            {Data: []byte("ico")},
	}
}

func getHeader(t *testing.T, root fstest.MapFS, path string) *http.Response {
	t.Helper()
	srv := httptest.NewServer(HandlerFor(root))
	defer srv.Close()
	resp, err := http.Get(srv.URL + path)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	defer resp.Body.Close()
	return resp
}

func requireCacheControl(t *testing.T, resp *http.Response, want string) {
	t.Helper()
	if got := resp.Header.Get("Cache-Control"); got != want {
		t.Fatalf("Cache-Control = %q, want %q", got, want)
	}
}

func TestCacheHeaders(t *testing.T) {
	root := cacheTestFS()

	t.Run("root document is revalidated", func(t *testing.T) {
		resp := getHeader(t, root, "/")
		defer resp.Body.Close()
		requireCacheControl(t, resp, "no-cache")
	})

	t.Run("spa fallback route is revalidated", func(t *testing.T) {
		resp := getHeader(t, root, "/clusters/142fa774-249b-494f-ba8c-a5cceab6c4f5")
		defer resp.Body.Close()
		requireCacheControl(t, resp, "no-cache")
		if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "text/html") {
			t.Fatalf("Content-Type = %q, want text/html", ct)
		}
	})

	t.Run("hashed assets are immutable", func(t *testing.T) {
		for _, path := range []string{"/assets/index-abc12.css", "/assets/index-9xy89.js"} {
			resp := getHeader(t, root, path)
			requireCacheControl(t, resp, "public, max-age=31536000, immutable")
			resp.Body.Close()
		}
	})

	t.Run("missing hashed asset does not get immutable caching", func(t *testing.T) {
		resp := getHeader(t, root, "/assets/missing-1234.css")
		defer resp.Body.Close()
		cc := resp.Header.Get("Cache-Control")
		if strings.Contains(cc, "immutable") {
			t.Fatalf("missing asset Cache-Control = %q, must not be immutable", cc)
		}
	})

	t.Run("non-hashed static files are revalidated", func(t *testing.T) {
		resp := getHeader(t, root, "/favicon.ico")
		defer resp.Body.Close()
		requireCacheControl(t, resp, "no-cache")
	})
}
