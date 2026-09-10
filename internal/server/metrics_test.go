package server

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/1solomonwakhungu/kfleet/internal/config"
	"github.com/1solomonwakhungu/kfleet/internal/store"
	"github.com/1solomonwakhungu/kfleet/internal/version"
	"github.com/1solomonwakhungu/kfleet/pkg/api"
	"github.com/1solomonwakhungu/kfleet/pkg/types"
	"github.com/google/uuid"
)

func parseMetrics(t *testing.T, body string) map[string]float64 {
	t.Helper()
	samples := map[string]float64{}
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) != 2 {
			t.Fatalf("malformed metrics sample line %q", line)
		}
		value, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			t.Fatalf("metric %q value %q is not numeric: %v", parts[0], parts[1], err)
		}
		samples[parts[0]] = value
	}
	return samples
}

func requireMetric(t *testing.T, samples map[string]float64, name string) float64 {
	t.Helper()
	value, ok := samples[name]
	if !ok {
		t.Fatalf("metric %q missing from payload %v", name, samples)
	}
	return value
}

func TestMetricsPayloadShape(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "kfleet.db")
	st, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := st.Close(); err != nil {
			t.Errorf("store.Close() error = %v", err)
		}
	})
	now := time.Now().UTC()
	for _, cluster := range []types.Cluster{
		{ID: "cluster-1", Name: "alpha", Health: types.HealthHealthy, RegisteredAt: now},
		{ID: "cluster-2", Name: "bravo", Health: types.HealthDegraded, RegisteredAt: now},
	} {
		if err := st.CreateCluster(context.Background(), cluster); err != nil {
			t.Fatalf("CreateCluster(%s) error = %v", cluster.ID, err)
		}
	}

	srv := New(&config.Config{ListenAddr: ":0", DBPath: dbPath}, slog.New(slog.NewTextHandler(io.Discard, nil)), st)
	httpServer := httptest.NewServer(srv.httpServer.Handler)
	t.Cleanup(httpServer.Close)

	response, err := http.Get(httpServer.URL + "/metrics")
	if err != nil {
		t.Fatalf("GET /metrics error = %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("GET /metrics status = %d, want %d", response.StatusCode, http.StatusOK)
	}
	if contentType := response.Header.Get("Content-Type"); contentType != metricsContentType {
		t.Fatalf("GET /metrics Content-Type = %q, want %q", contentType, metricsContentType)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read metrics body error = %v", err)
	}

	for _, name := range []string{
		"kfleet_agents_registered",
		"kfleet_log_relay_connected",
		"kfleet_ws_clients",
		"kfleet_log_streams_active",
		"kfleet_alerts_dead_letter",
		"kfleet_db_size_bytes",
	} {
		if !strings.Contains(string(body), "# TYPE "+name+" gauge") {
			t.Errorf("metrics payload missing TYPE declaration for %q:\n%s", name, body)
		}
	}

	samples := parseMetrics(t, string(body))
	if got := requireMetric(t, samples, "kfleet_agents_registered"); got != 2 {
		t.Errorf("kfleet_agents_registered = %v, want 2", got)
	}
	if got := requireMetric(t, samples, "kfleet_log_relay_connected"); got != 0 {
		t.Errorf("kfleet_log_relay_connected = %v, want 0 with no agents connected", got)
	}
	if got := requireMetric(t, samples, "kfleet_ws_clients"); got != 0 {
		t.Errorf("kfleet_ws_clients = %v, want 0 with no dashboard clients", got)
	}
	if got := requireMetric(t, samples, "kfleet_log_streams_active"); got != 0 {
		t.Errorf("kfleet_log_streams_active = %v, want 0 with no streams", got)
	}
	if got := requireMetric(t, samples, "kfleet_alerts_dead_letter"); got != 0 {
		t.Errorf("kfleet_alerts_dead_letter = %v, want 0", got)
	}
	if got := requireMetric(t, samples, "kfleet_db_size_bytes"); got <= 0 {
		t.Errorf("kfleet_db_size_bytes = %v, want the size of the opened database file", got)
	}
}

func TestMetricsEndpointDisabledByConfig(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "kfleet.db"))
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := st.Close(); err != nil {
			t.Errorf("store.Close() error = %v", err)
		}
	})

	srv := New(&config.Config{ListenAddr: ":0", MetricsDisabled: true}, slog.New(slog.NewTextHandler(io.Discard, nil)), st)
	httpServer := httptest.NewServer(srv.httpServer.Handler)
	t.Cleanup(httpServer.Close)

	response, err := http.Get(httpServer.URL + "/metrics")
	if err != nil {
		t.Fatalf("GET /metrics error = %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("GET /metrics with KFLEET_METRICS_ENABLED=false status = %d, want %d", response.StatusCode, http.StatusNotFound)
	}
}

func TestMetricsDisabledByDefaultConfig(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	if cfg.MetricsDisabled {
		t.Fatal("MetricsDisabled = true with no KFLEET_METRICS_ENABLED set, want the default enabled")
	}

	t.Setenv("KFLEET_METRICS_ENABLED", "false")
	cfg, err = config.Load()
	if err != nil {
		t.Fatalf("config.Load() with KFLEET_METRICS_ENABLED=false error = %v", err)
	}
	if !cfg.MetricsDisabled {
		t.Fatal("MetricsDisabled = false with KFLEET_METRICS_ENABLED=false, want metrics disabled")
	}
}

func TestMetaIncludesHubVersion(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "kfleet.db"))
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := st.Close(); err != nil {
			t.Errorf("store.Close() error = %v", err)
		}
	})

	srv := New(&config.Config{ListenAddr: ":0"}, slog.New(slog.NewTextHandler(io.Discard, nil)), st)
	httpServer := httptest.NewServer(srv.httpServer.Handler)
	t.Cleanup(httpServer.Close)

	response, err := http.Get(httpServer.URL + "/api/v1/meta")
	if err != nil {
		t.Fatalf("GET /api/v1/meta error = %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/v1/meta status = %d, want %d", response.StatusCode, http.StatusOK)
	}
	var meta api.RuntimeInfo
	if err := json.NewDecoder(response.Body).Decode(&meta); err != nil {
		t.Fatalf("decode meta error = %v", err)
	}
	if meta.Version != version.String() {
		t.Fatalf("meta version = %q, want %q", meta.Version, version.String())
	}
}

func TestRequestIDMiddleware(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "kfleet.db"))
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := st.Close(); err != nil {
			t.Errorf("store.Close() error = %v", err)
		}
	})

	srv := New(&config.Config{ListenAddr: ":0"}, slog.New(slog.NewTextHandler(io.Discard, nil)), st)
	httpServer := httptest.NewServer(srv.httpServer.Handler)
	t.Cleanup(httpServer.Close)

	t.Run("generates a UUID when none is supplied", func(t *testing.T) {
		response, err := http.Get(httpServer.URL + "/healthz")
		if err != nil {
			t.Fatalf("GET /healthz error = %v", err)
		}
		defer response.Body.Close()
		requestID := response.Header.Get(requestIDHeader)
		if requestID == "" {
			t.Fatal("X-Request-ID response header is empty for a request without one")
		}
		if _, err := uuid.Parse(requestID); err != nil {
			t.Fatalf("generated X-Request-ID %q is not a UUID: %v", requestID, err)
		}
	})

	t.Run("honors an incoming ID", func(t *testing.T) {
		request, err := http.NewRequest(http.MethodGet, httpServer.URL+"/healthz", nil)
		if err != nil {
			t.Fatalf("http.NewRequest() error = %v", err)
		}
		request.Header.Set(requestIDHeader, "correlation-abc-123")
		response, err := httpServer.Client().Do(request)
		if err != nil {
			t.Fatalf("GET /healthz error = %v", err)
		}
		defer response.Body.Close()
		if got := response.Header.Get(requestIDHeader); got != "correlation-abc-123" {
			t.Fatalf("X-Request-ID = %q, want the incoming value", got)
		}
	})

	t.Run("replaces an unsafe incoming ID", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		request.Header.Set(requestIDHeader, "bad id\nwith newline")
		srv.httpServer.Handler.ServeHTTP(recorder, request)
		got := recorder.Header().Get(requestIDHeader)
		if got == "bad id\nwith newline" {
			t.Fatal("unsafe incoming X-Request-ID was echoed back")
		}
		if _, err := uuid.Parse(got); err != nil {
			t.Fatalf("replacement X-Request-ID %q is not a UUID: %v", got, err)
		}
	})

	t.Run("replaces an overlong incoming ID", func(t *testing.T) {
		long := strings.Repeat("x", maxRequestIDLength+1)
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		request.Header.Set(requestIDHeader, long)
		srv.httpServer.Handler.ServeHTTP(recorder, request)
		if got := recorder.Header().Get(requestIDHeader); got == long {
			t.Fatal("overlong incoming X-Request-ID was echoed back")
		}
	})
}

// recordCollector is a slog handler that captures every record so tests can
// assert on levels and attributes.
type recordCollector struct {
	mu      sync.Mutex
	records []slog.Record
}

func (c *recordCollector) Enabled(_ context.Context, level slog.Level) bool {
	return level >= slog.LevelDebug
}

func (c *recordCollector) Handle(_ context.Context, record slog.Record) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.records = append(c.records, record.Clone())
	return nil
}

func (c *recordCollector) WithAttrs([]slog.Attr) slog.Handler { return c }

func (c *recordCollector) WithGroup(string) slog.Handler { return c }

func (c *recordCollector) snapshot() []slog.Record {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]slog.Record(nil), c.records...)
}

func recordAttr(t *testing.T, record slog.Record, key string) (slog.Value, bool) {
	t.Helper()
	var found slog.Value
	var ok bool
	record.Attrs(func(attr slog.Attr) bool {
		if attr.Key == key {
			found = attr.Value
			ok = true
			return false
		}
		return true
	})
	return found, ok
}

func TestLoggingIncludesRequestIDAndStatus(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "kfleet.db"))
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := st.Close(); err != nil {
			t.Errorf("store.Close() error = %v", err)
		}
	})

	collector := &recordCollector{}
	srv := New(&config.Config{ListenAddr: ":0"}, slog.New(collector), st)
	httpServer := httptest.NewServer(srv.httpServer.Handler)
	t.Cleanup(httpServer.Close)

	request, err := http.NewRequest(http.MethodGet, httpServer.URL+"/api/v1/meta", nil)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	request.Header.Set(requestIDHeader, "log-correlation-test")
	response, err := httpServer.Client().Do(request)
	if err != nil {
		t.Fatalf("GET /api/v1/meta error = %v", err)
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, response.Body)

	var requestLog *slog.Record
	for _, record := range collector.snapshot() {
		if path, ok := recordAttr(t, record, "path"); ok && path.String() == "/api/v1/meta" {
			requestLog = &record
			break
		}
	}
	if requestLog == nil {
		t.Fatal("no log record found for /api/v1/meta")
	}
	if requestLog.Level != slog.LevelInfo {
		t.Fatalf("request log level = %v, want Info", requestLog.Level)
	}
	if id, ok := recordAttr(t, *requestLog, "request_id"); !ok || id.String() != "log-correlation-test" {
		t.Fatalf("request log request_id = (%q, %v), want the incoming ID", id, ok)
	}
	if status, ok := recordAttr(t, *requestLog, "status"); !ok || status.Int64() != http.StatusOK {
		t.Fatalf("request log status = (%v, %v), want %d", status, ok, http.StatusOK)
	}
	if duration, ok := recordAttr(t, *requestLog, "duration"); !ok || duration.Kind() != slog.KindDuration {
		t.Fatalf("request log duration = (%v, %v), want a duration attribute", duration, ok)
	}
}

func TestLoggingQuietsProbes(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "kfleet.db"))
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := st.Close(); err != nil {
			t.Errorf("store.Close() error = %v", err)
		}
	})

	collector := &recordCollector{}
	srv := New(&config.Config{ListenAddr: ":0"}, slog.New(collector), st)
	httpServer := httptest.NewServer(srv.httpServer.Handler)
	t.Cleanup(httpServer.Close)

	for _, path := range []string{"/healthz", "/readyz"} {
		response, err := http.Get(httpServer.URL + path)
		if err != nil {
			t.Fatalf("GET %s error = %v", path, err)
		}
		if response.StatusCode != http.StatusOK {
			t.Fatalf("GET %s status = %d, want %d", path, response.StatusCode, http.StatusOK)
		}
		_, _ = io.Copy(io.Discard, response.Body)
		_ = response.Body.Close()
	}

	probeRecords := 0
	for _, record := range collector.snapshot() {
		path, ok := recordAttr(t, record, "path")
		if !ok {
			continue
		}
		if path.String() != "/healthz" && path.String() != "/readyz" {
			continue
		}
		probeRecords++
		if record.Level >= slog.LevelInfo {
			t.Errorf("probe request for %s logged at %v, want Debug only", path.String(), record.Level)
		}
		if status, ok := recordAttr(t, record, "status"); !ok || status.Int64() != http.StatusOK {
			t.Errorf("probe request for %s logged status = (%v, %v), want 200", path.String(), status, ok)
		}
	}
	if probeRecords != 2 {
		t.Fatalf("captured %d probe log records, want one Debug record per probe (2)", probeRecords)
	}
}
