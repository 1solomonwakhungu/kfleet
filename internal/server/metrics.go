package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

// storeMetricsCacheTTL bounds how long COUNT query results back /metrics, so
// scraper traffic never becomes per-scrape database load.
const storeMetricsCacheTTL = 30 * time.Second

// metricsContentType is the Prometheus text exposition format (version 0.0.4)
// content type. Hand-rolled output keeps the hub dependency-free while staying
// directly scrapable by Prometheus, VictoriaMetrics, and similar scrapers.
const metricsContentType = "text/plain; version=0.0.4; charset=utf-8"

// metric is one gauge rendered into the Prometheus text format. A metric
// whose value could not be determined is omitted from the payload rather
// than misreported as zero.
type metric struct {
	name  string
	help  string
	value int64
	valid bool
}

// cachedCount holds the most recent successful store count. An errored query
// invalidates the value until the next refresh succeeds.
type cachedCount struct {
	value     int64
	valid     bool
	expiresAt time.Time
}

// storeMetrics holds the store-backed counters behind a small shared cache so
// concurrent scrapes share one pair of COUNT queries per TTL window.
type storeMetrics struct {
	mu           sync.Mutex
	agents       cachedCount
	deadLettered cachedCount
}

// get returns the cached values, refreshing any that expired. A failed query
// invalidates that value until the next refresh instead of serving a stale
// number.
func (m *storeMetrics) get(ctx context.Context, s *Server) (agents, deadLettered metric) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	if !m.agents.valid || !m.agents.expiresAt.After(now) {
		if value, err := s.store.CountClusters(ctx); err == nil {
			m.agents = cachedCount{value: value, valid: true, expiresAt: now.Add(storeMetricsCacheTTL)}
		} else {
			s.logger.Warn("metrics count query failed", "metric", "kfleet_agents_registered", "error", err)
			m.agents.valid = false
		}
	}
	if !m.deadLettered.valid || !m.deadLettered.expiresAt.After(now) {
		if count, err := s.store.CountDeadLetteredAlerts(ctx); err == nil {
			m.deadLettered = cachedCount{value: count, valid: true, expiresAt: now.Add(storeMetricsCacheTTL)}
		} else {
			s.logger.Warn("metrics count query failed", "metric", "kfleet_alerts_dead_letter", "error", err)
			m.deadLettered.valid = false
		}
	}
	return metric{name: "kfleet_agents_registered", help: "Clusters with a registered agent.", value: m.agents.value, valid: m.agents.valid},
		metric{name: "kfleet_alerts_dead_letter", help: "Alerts in the dead-letter delivery state.", value: m.deadLettered.value, valid: m.deadLettered.valid}
}

// handleMetrics serves the hub's operational metrics in the Prometheus text
// exposition format. It is registered at the root (not under /api) because
// Prometheus scrapers expect /metrics, and it is deliberately outside the
// authenticated API surface: every value is a non-sensitive aggregate count
// or file size (no cluster names, IDs, tenants, or credentials), the endpoint
// is read-only, and it can be disabled with KFLEET_METRICS_ENABLED=false.
// The API JSON fallback's shadow mux is not involved because /metrics is not
// an /api route.
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if s.cfg.MetricsDisabled {
		http.NotFound(w, r)
		return
	}
	agents, deadLettered := s.storeMetrics.get(r.Context(), s)
	rendered := renderMetrics([]metric{
		agents,
		{name: "kfleet_log_relay_connected", help: "Agents with a live log relay reverse channel.", value: int64(s.logs.ConnectedAgents()), valid: true},
		{name: "kfleet_ws_clients", help: "Connected WebSocket dashboard clients.", value: int64(s.broadcast.ClientCount()), valid: true},
		{name: "kfleet_log_streams_active", help: "In-flight pod log streams relayed from agents.", value: int64(s.logs.ActiveStreams()), valid: true},
		deadLettered,
		{name: "kfleet_db_size_bytes", help: "Hub SQLite database file size in bytes; 0 in demo mode or when unknown.", value: s.dbSizeBytes(), valid: true},
	})
	w.Header().Set("Content-Type", metricsContentType)
	w.Header().Set("Cache-Control", "no-store")
	for _, line := range rendered {
		if _, err := fmt.Fprintln(w, line); err != nil {
			return
		}
	}
}

func renderMetrics(metrics []metric) []string {
	lines := make([]string, 0, len(metrics)*3)
	for _, m := range metrics {
		if !m.valid {
			continue
		}
		lines = append(lines,
			fmt.Sprintf("# HELP %s %s", m.name, m.help),
			fmt.Sprintf("# TYPE %s gauge", m.name),
			fmt.Sprintf("%s %d", m.name, m.value),
		)
	}
	return lines
}

// dbSizeBytes reports the size of the configured SQLite database file. Demo
// mode runs on an in-memory store, so it always reports zero there rather
// than leaking the size of an unrelated file at the configured path.
func (s *Server) dbSizeBytes() int64 {
	if s.cfg.DemoMode {
		return 0
	}
	info, err := os.Stat(s.cfg.DBPath)
	if err != nil {
		return 0
	}
	return info.Size()
}
