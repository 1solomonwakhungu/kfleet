// Package server provides the kfleet hub HTTP server.
package server

import (
	"bufio"
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/1solomonwakhungu/kfleet/internal/alerts"
	"github.com/1solomonwakhungu/kfleet/internal/config"
	hubweb "github.com/1solomonwakhungu/kfleet/internal/hub/web"
	"github.com/1solomonwakhungu/kfleet/internal/policy"
	"github.com/1solomonwakhungu/kfleet/internal/store"
	"github.com/1solomonwakhungu/kfleet/internal/version"
	"github.com/1solomonwakhungu/kfleet/pkg/api"
	"github.com/1solomonwakhungu/kfleet/pkg/types"
	"github.com/google/uuid"
)

const shutdownTimeout = 5 * time.Second

// eventPruneInterval controls how often the retention sweep runs. Hourly is
// frequent enough to keep the table bounded without adding meaningful load.
const eventPruneInterval = time.Hour
const defaultEventRetention = 90 * 24 * time.Hour

// Server is the kfleet hub HTTP server.
type Server struct {
	cfg          *config.Config
	logger       *slog.Logger
	store        store.Store
	alerts       *alerts.Manager
	policies     *policy.Engine
	broadcast    *BroadcastHub
	logs         *LogRelay
	storeMetrics *storeMetrics
	httpServer   *http.Server
}

// New constructs a hub server with its routes configured.
func New(cfg *config.Config, logger *slog.Logger, st store.Store) *Server {
	server := &Server{
		cfg:    cfg,
		logger: logger,
		store:  st,
		alerts: alerts.New(st, logger, alerts.Config{
			WebhookURL:   cfg.AlertWebhookURL,
			Secret:       cfg.AlertWebhookSecret,
			MaxAttempts:  cfg.AlertMaxAttempts,
			RetryBase:    cfg.AlertRetryBase,
			PollInterval: cfg.AlertPollInterval,
		}),
		broadcast: NewBroadcastHub(logger),
		logs:      NewLogRelay(logger),
	}
	server.storeMetrics = &storeMetrics{}
	server.policies = policy.NewEngine(st, 3*server.heartbeatInterval())

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /readyz", server.handleReadyz)
	mux.HandleFunc("GET /metrics", server.handleMetrics)
	mux.HandleFunc("GET /ws/clusters", server.requireAuth(server.handleWSClusters))

	// registerAPIRoutes declares every /api/v1 route. It runs twice: once
	// against the public mux and once against the fallback's shadow mux, which
	// answers unmatched /api requests with JSON 404/405 errors. New API routes
	// must be added here (or in a register*Routes method referenced below) so
	// the fallback can compute correct 405 Allow headers.
	registerAPIRoutes := func(m *http.ServeMux) {
		m.HandleFunc("GET /api/v1/meta", func(w http.ResponseWriter, _ *http.Request) {
			if err := api.WriteJSON(w, http.StatusOK, api.RuntimeInfo{
				Version:       version.String(),
				DemoMode:      cfg.DemoMode,
				ReadOnly:      cfg.DemoMode,
				SyntheticData: cfg.DemoMode,
				DataPolicy:    runtimeDataPolicy(cfg.DemoMode),
			}); err != nil {
				server.logger.Error("failed to write runtime metadata", "error", err)
			}
		})
		server.registerAuthRoutes(m)
		server.registerUserRoutes(m)
		server.registerAuditRoutes(m)
		server.registerAdminRoutes(m)
		server.registerAgentRoutes(m)
		server.registerClusterRoutes(m)
		server.registerAlertRoutes(m)
		server.registerEventRoutes(m)
		server.registerPolicyRoutes(m)
	}
	registerAPIRoutes(mux)

	fallback := apiJSONFallback(registerAPIRoutes)
	mux.Handle("/api/", fallback)
	mux.Handle("/api", fallback)
	mux.Handle("/", hubweb.Handler())

	handler := http.Handler(mux)
	if cfg.DemoMode {
		handler = server.withDemoReadOnly(handler)
	}
	handler = server.withSecurityHeaders(handler)
	handler = server.withLogging(handler)
	handler = server.withRequestID(handler)
	server.httpServer = &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
	return server
}

// readinessTimeout bounds the store probe so a wedged database cannot hold the
// readiness handler open indefinitely.
const readinessTimeout = 2 * time.Second

// handleReadyz reports readiness only when the backing store answers a query.
// /healthz stays a pure liveness check.
func (s *Server) handleReadyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), readinessTimeout)
	defer cancel()
	if err := s.store.Ping(ctx); err != nil {
		s.logger.Error("readiness probe failed", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"status":"unavailable","reason":"store unavailable"}`))
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ready"))
}

// Start serves HTTP requests until the context is cancelled or the server
// returns an error. Cancellation triggers a graceful shutdown.
func (s *Server) Start(ctx context.Context) error {
	hubCtx, stopHub := context.WithCancel(ctx)
	defer stopHub()
	go s.broadcast.Run(hubCtx)
	if !s.cfg.DemoMode {
		go s.monitorStaleClusters(hubCtx)
		go s.alerts.Run(hubCtx)
		go s.monitorEventRetention(hubCtx)
		go s.pruneExpiredSessions(hubCtx)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.httpServer.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		s.logger.Info("shutting down hub server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			return err
		}

		err := <-errCh
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func runtimeDataPolicy(demoMode bool) string {
	if demoMode {
		return "Synthetic sample data only. No live cluster identities, telemetry, or credentials."
	}
	return "Live hub data."
}

func (s *Server) withDemoReadOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead:
			next.ServeHTTP(w, r)
		case http.MethodOptions:
			w.Header().Set("Allow", "GET, HEAD, OPTIONS")
			w.WriteHeader(http.StatusNoContent)
		default:
			w.Header().Set("Allow", "GET, HEAD, OPTIONS")
			w.Header().Set("Cache-Control", "no-store")
			api.WriteError(w, http.StatusMethodNotAllowed, "public demo is read-only")
		}
	})
}

func (s *Server) withSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers := w.Header()
		headers.Set("Content-Security-Policy", "default-src 'self'; base-uri 'none'; connect-src 'self'; font-src 'self'; form-action 'none'; frame-ancestors 'none'; img-src 'self' data:; object-src 'none'; script-src 'self'; style-src 'self' 'unsafe-inline'")
		headers.Set("Cross-Origin-Opener-Policy", "same-origin")
		headers.Set("Cross-Origin-Resource-Policy", "same-origin")
		headers.Set("Permissions-Policy", "camera=(), geolocation=(), microphone=(), payment=(), usb=()")
		headers.Set("Referrer-Policy", "no-referrer")
		headers.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		headers.Set("X-Content-Type-Options", "nosniff")
		headers.Set("X-Frame-Options", "DENY")
		if r.URL.Path == "/healthz" || r.URL.Path == "/readyz" || r.URL.Path == "/metrics" || r.URL.Path == "/api" || strings.HasPrefix(r.URL.Path, "/api/") {
			headers.Set("Cache-Control", "no-store")
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) pruneExpiredSessions(ctx context.Context) {
	if err := s.store.DeleteExpiredSessions(ctx, time.Now().UTC()); err != nil {
		s.logger.Error("failed to prune expired sessions", "error", err)
	}
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			if err := s.store.DeleteExpiredSessions(ctx, now.UTC()); err != nil {
				s.logger.Error("failed to prune expired sessions", "error", err)
			}
		}
	}
}

func (s *Server) heartbeatInterval() time.Duration {
	if s.cfg.HeartbeatInterval > 0 {
		return s.cfg.HeartbeatInterval
	}
	return 30 * time.Second
}

func (s *Server) monitorStaleClusters(ctx context.Context) {
	ticker := time.NewTicker(s.heartbeatInterval())
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			s.markStaleClusters(ctx, now.UTC())
		}
	}
}

func (s *Server) markStaleClusters(ctx context.Context, now time.Time) {
	clusters, err := s.store.ListClusters(ctx)
	if err != nil {
		s.logger.Error("failed to list clusters for staleness check", "error", err)
		return
	}
	cutoff := now.Add(-3 * s.heartbeatInterval())
	for _, cluster := range clusters {
		if cluster.LastHeartbeat.IsZero() || !cluster.LastHeartbeat.Before(cutoff) || cluster.Health == types.HealthUnreachable {
			continue
		}
		oldHealth := cluster.Health
		if err := s.store.UpdateHealth(ctx, cluster.ID, types.HealthUnreachable, cluster.LastHeartbeat); err != nil {
			s.logger.Error("failed to mark cluster unreachable", "cluster_id", cluster.ID, "error", err)
			continue
		}
		cluster.Health = types.HealthUnreachable
		s.alerts.Evaluate(ctx, cluster)
		s.broadcast.Broadcast(ClusterUpdate{Type: "health_changed", Cluster: cluster})
		s.recordAgentDisconnected(ctx, cluster, "heartbeat_timeout", now)
		s.recordHeartbeatTransition(ctx, cluster, oldHealth, types.HealthUnreachable, now)
	}
}

func (s *Server) monitorEventRetention(ctx context.Context) {
	s.pruneExpiredEvents(ctx, time.Now().UTC())
	ticker := time.NewTicker(eventPruneInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			s.pruneExpiredEvents(ctx, now.UTC())
		}
	}
}

func (s *Server) eventRetention() time.Duration {
	if s.cfg.EventRetention > 0 {
		return s.cfg.EventRetention
	}
	return defaultEventRetention
}

func (s *Server) pruneExpiredEvents(ctx context.Context, now time.Time) {
	cutoff := now.UTC().Add(-s.eventRetention())
	removed, err := s.store.PruneEventsBefore(ctx, cutoff)
	if err != nil {
		s.logger.Error("failed to prune operational events", "error", err)
		return
	}
	if removed > 0 {
		s.logger.Info("pruned expired operational events", "removed", removed, "cutoff", cutoff)
	}
}

// requestIDHeader carries the per-request correlation ID both ways.
const requestIDHeader = "X-Request-ID"

// maxRequestIDLength bounds honored incoming IDs so a caller cannot plant an
// unbounded value into response headers and logs.
const maxRequestIDLength = 128

type requestIDContextKey struct{}

// withRequestID gives every request a correlation ID: an incoming
// X-Request-ID header is honored when well-formed, otherwise a UUID is
// generated. The ID is echoed in the X-Request-ID response header, placed in
// the request context for handlers and withLogging, and can therefore appear
// in every log line for the request. It must wrap withLogging so the log
// middleware can read the ID from the context.
func (s *Server) withRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := sanitizeRequestID(r.Header.Get(requestIDHeader))
		w.Header().Set(requestIDHeader, id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDContextKey{}, id)))
	})
}

// requestIDFromContext returns the request ID assigned by withRequestID.
func requestIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(requestIDContextKey{}).(string)
	return id, ok && id != ""
}

// isValidRequestID accepts printable ASCII IDs without whitespace, the only
// shape safe to echo back into response headers and structured logs.
func isValidRequestID(id string) bool {
	if id == "" || len(id) > maxRequestIDLength {
		return false
	}
	for i := 0; i < len(id); i++ {
		if id[i] < 0x21 || id[i] > 0x7e {
			return false
		}
	}
	return true
}

// sanitizeRequestID honors a well-formed incoming ID, otherwise it generates
// a fresh UUID.
func sanitizeRequestID(incoming string) string {
	if isValidRequestID(incoming) {
		return incoming
	}
	return uuid.NewString()
}

// statusWriter captures the response status for request logging while
// delegating everything else, including the flush and hijack upgrades the
// WebSocket and SSE handlers rely on.
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	if w.status == 0 {
		w.status = status
	}
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(p)
}

func (w *statusWriter) Flush() {
	_ = http.NewResponseController(w.ResponseWriter).Flush()
}

func (w *statusWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	// A hijacked connection has already answered with 101 Switching
	// Protocols, written directly over the raw connection.
	w.status = http.StatusSwitchingProtocols
	return http.NewResponseController(w.ResponseWriter).Hijack()
}

// withLogging logs the method, path, status, duration, and request ID of
// every HTTP request. Liveness and readiness probes log at Debug instead of
// Info so routine probe traffic never floods the operational log.
func (s *Server) withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		recorder := &statusWriter{ResponseWriter: w}
		next.ServeHTTP(recorder, r)
		attrs := []any{
			"method", r.Method,
			"path", r.URL.Path,
			"status", recorder.status,
			"duration", time.Since(started),
		}
		if id, ok := requestIDFromContext(r.Context()); ok {
			attrs = append(attrs, "request_id", id)
		}
		if r.URL.Path == "/healthz" || r.URL.Path == "/readyz" {
			s.logger.Debug("http request", attrs...)
			return
		}
		s.logger.Info("http request", attrs...)
	})
}
