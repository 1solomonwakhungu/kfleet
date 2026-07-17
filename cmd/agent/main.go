package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/1solomonwakhungu/kfleet/internal/agentconfig"
	"github.com/1solomonwakhungu/kfleet/internal/collector"
	"github.com/1solomonwakhungu/kfleet/internal/registration"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := agentconfig.Load()
	if err != nil {
		logger.Error("failed to load agent configuration", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	clusterCollector := collector.New(logger)
	registrationClient := registration.NewClient(cfg.HubURL, logger)

	token, err := registrationClient.Register(ctx, cfg.ClusterName)
	if err != nil {
		logger.Error("failed to register agent", "error", err)
		os.Exit(1)
	}
	logger.Info("agent registered", "cluster", cfg.ClusterName)

	if err := run(ctx, cfg.HeartbeatInterval, token, clusterCollector, registrationClient, logger); err != nil {
		logger.Error("agent stopped with an error", "error", err)
		os.Exit(1)
	}
	logger.Info("agent stopped")
}

func run(
	ctx context.Context,
	interval time.Duration,
	token string,
	clusterCollector *collector.Collector,
	registrationClient *registration.Client,
	logger *slog.Logger,
) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	logger.Info("starting heartbeat loop", "interval", interval)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			snapshot, err := clusterCollector.CollectClusterInfo(ctx)
			if err != nil {
				logger.Error("failed to collect cluster information", "error", err)
				continue
			}
			if err := registrationClient.Heartbeat(ctx, token, snapshot); err != nil {
				logger.Error("failed to send heartbeat", "error", err)
				continue
			}
			logger.Debug("heartbeat sent")
		}
	}
}
