// Command agent runs the kfleet agent daemon.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/1solomonwakhungu/kfleet/internal/agent/collector"
	"github.com/1solomonwakhungu/kfleet/internal/agent/config"
	"github.com/1solomonwakhungu/kfleet/internal/agent/reporter"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load agent configuration", "error", err)
		os.Exit(1)
	}
	clusterCollector, err := collector.New(cfg)
	if err != nil {
		logger.Error("failed to create Kubernetes collector", "error", err)
		os.Exit(1)
	}
	clusterReporter := reporter.New(cfg)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	run(ctx, cfg.ReportInterval, clusterCollector, clusterReporter, logger)
}

func run(
	ctx context.Context,
	interval time.Duration,
	clusterCollector *collector.Collector,
	clusterReporter *reporter.Reporter,
	logger *slog.Logger,
) {
	report := func() {
		state, err := clusterCollector.Collect(ctx)
		if err != nil {
			logger.Error("failed to collect cluster state", "error", err)
			return
		}
		if err := clusterReporter.Report(ctx, state); err != nil {
			logger.Error("failed to report cluster state", "error", err)
			return
		}
		logger.Debug("cluster state reported")
	}

	report()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			report()
		}
	}
}
