// Package server provides the kfleet hub HTTP server.
package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/1solomonwakhungu/kfleet/internal/alerts"
	"github.com/1solomonwakhungu/kfleet/internal/config"
	hubweb "github.com/1solomonwakhungu/kfleet/internal/hub/web"
	"github.com/1solomonwakhungu/kfleet/internal/policy"
	"github.com/1solomonwakhungu/kfleet/internal/store"
	"github.com/1solomonwakhungu/kfleet/pkg/api"
	"github.com/1solomonwakhungu/kfleet/pkg/types"
)

const shutdownTimeout = 5 * time.Second

// eventPruneInterval controls how often the retention sweep runs. Hourly is
// frequent enough to keep the table bounded without adding meaningful load.
const eventPruneInterval = time.Hour
const defaultEventRetention = 90 * 24 * time.Hour

// Server is the kfleet hub HTTP server.
type Server struct {
	cfg        *config.Config
	logger     *slog.Logger
	store      store.Store
	alerts     *alerts.Manager
	policies   *policy.Engine
	broadcast  *BroadcastHub
	httpServer *http.Server
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
	}
	server.policies = policy.NewEngine(st, 3*server.heartbeatInterval())

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	})
	mux.HandleFunc("GET /api/v1/meta", func(w http.ResponseWriter, _ *http.Request) {
		if err := api.WriteJSON(w, http.StatusOK, api.RuntimeInfo{
			DemoMode:      cfg.DemoMode,
			ReadOnly:      cfg.DemoMode,
			SyntheticData: cfg.DemoMode,
			DataPolicy:    runtimeDataPolicy(cfg.DemoMode),
		}); err != nil {
			server.logger.Error("failed to write runtime metadata", "error", err)
		}
	})
	server.registerAuthRoutes(mux)
	server.registerUserRoutes(mux)
	server.registerAuditRoutes(mux)
	server.registerAdminRoutes(mux)
	server.registerAgentRoutes(mux)
	server.registerClusterRoutes(mux)
	server.registerAlertRoutes(mux)
	server.registerEventRoutes(mux)
	server.registerPolicyRoutes(mux)
	mux.HandleFunc("GET /ws/clusters", server.requireAuth(server.handleWSClusters))
	mux.Handle("/", hubweb.Handler())

	handler := http.Handler(mux)
	if cfg.DemoMode {
		handler = server.withDemoReadOnly(handler)
	}
	handler = server.withSecurityHeaders(handler)
	handler = server.withLogging(handler)
	server.httpServer = &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
	return server
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
		if r.URL.Path == "/healthz" || r.URL.Path == "/readyz" || strings.HasPrefix(r.URL.Path, "/api/") {
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

// withLogging logs the method, path, and duration of every HTTP request.
func (s *Server) withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		next.ServeHTTP(w, r)
		s.logger.Info("http request",
			"method", r.Method,
			"path", r.URL.Path,
			"duration", time.Since(started),
		)
	})
}
