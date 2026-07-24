// Package server provides the kfleet hub HTTP server.
package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/1solomonwakhungu/kfleet/internal/config"
	hubweb "github.com/1solomonwakhungu/kfleet/internal/hub/web"
	"github.com/1solomonwakhungu/kfleet/internal/policy"
	"github.com/1solomonwakhungu/kfleet/internal/store"
	"github.com/1solomonwakhungu/kfleet/pkg/types"
)

const shutdownTimeout = 5 * time.Second

// Server is the kfleet hub HTTP server.
type Server struct {
	cfg        *config.Config
	logger     *slog.Logger
	store      store.Store
	policies   *policy.Engine
	broadcast  *BroadcastHub
	httpServer *http.Server
}

// New constructs a hub server with its routes configured.
func New(cfg *config.Config, logger *slog.Logger, st store.Store) *Server {
	server := &Server{
		cfg:       cfg,
		logger:    logger,
		store:     st,
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
	server.registerAgentRoutes(mux)
	server.registerClusterRoutes(mux)
	server.registerPolicyRoutes(mux)
	mux.HandleFunc("GET /ws/clusters", server.handleWSClusters)
	mux.Handle("/", hubweb.Handler())

	server.httpServer = &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           server.withLogging(mux),
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
	go s.monitorStaleClusters(hubCtx)

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
		if err := s.store.UpdateHealth(ctx, cluster.ID, types.HealthUnreachable, cluster.LastHeartbeat); err != nil {
			s.logger.Error("failed to mark cluster unreachable", "cluster_id", cluster.ID, "error", err)
			continue
		}
		cluster.Health = types.HealthUnreachable
		s.broadcast.Broadcast(ClusterUpdate{Type: "health_changed", Cluster: cluster})
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
