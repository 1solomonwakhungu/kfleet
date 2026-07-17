// Command agent runs the kfleet agent daemon.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/1solomonwakhungu/kfleet/internal/agent/collector"
	"github.com/1solomonwakhungu/kfleet/internal/agent/config"
	"github.com/1solomonwakhungu/kfleet/internal/agent/registrar"
	"github.com/1solomonwakhungu/kfleet/internal/agent/reporter"
)

const (
	heartbeatInterval = 15 * time.Second
	deregisterTimeout = 5 * time.Second
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load agent configuration", "error", err)
		os.Exit(1)
	}
	labels, err := clusterLabels(os.Getenv("KFLEET_CLUSTER_LABELS"))
	if err != nil {
		logger.Error("failed to parse cluster labels", "error", err)
		os.Exit(1)
	}
	clusterCollector, err := collector.New(cfg)
	if err != nil {
		logger.Error("failed to create Kubernetes collector", "error", err)
		os.Exit(1)
	}
	agentRegistrar := registrar.New(cfg, labels)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	run(ctx, cfg, clusterCollector, agentRegistrar, logger)
}

func run(
	ctx context.Context,
	cfg *config.Config,
	clusterCollector *collector.Collector,
	agentRegistrar *registrar.Registrar,
	logger *slog.Logger,
) {
	backoff := registrar.NewBackoff()
	k8sVersion := ""
	if state, err := clusterCollector.Collect(ctx); err != nil {
		logger.Warn("failed to collect Kubernetes version before registration", "error", err)
	} else {
		k8sVersion = state.K8sVersion
	}

	delayBeforeRegister := false
	for ctx.Err() == nil {
		if delayBeforeRegister && !waitForRetry(ctx, backoff.Next()) {
			break
		}
		registration, err := agentRegistrar.Register(ctx, k8sVersion)
		if err != nil {
			if ctx.Err() != nil {
				break
			}
			logger.Error("agent registration failed", "error", err)
			delayBeforeRegister = true
			continue
		}
		backoff.Reset()
		delayBeforeRegister = false
		if registration.Approved {
			logger.Info("agent registered", "cluster_id", registration.ClusterID)
		} else {
			logger.Info("pending approval", "cluster_id", registration.ClusterID)
		}

		reportCfg := *cfg
		reportCfg.HubToken = agentRegistrar.Token()
		clusterReporter := reporter.New(&reportCfg)
		reportState(ctx, clusterCollector, clusterReporter, logger)
		if runAgentLoop(ctx, cfg.ReportInterval, clusterCollector, clusterReporter, agentRegistrar, logger) {
			break
		}
		delayBeforeRegister = true
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), deregisterTimeout)
	defer cancel()
	if err := agentRegistrar.Deregister(shutdownCtx); err != nil {
		logger.Warn("failed to deregister agent", "error", err)
	}
}

// runAgentLoop returns true when the root context is cancelled and false when
// registration should be retried after a heartbeat failure.
func runAgentLoop(
	ctx context.Context,
	reportInterval time.Duration,
	clusterCollector *collector.Collector,
	clusterReporter *reporter.Reporter,
	agentRegistrar *registrar.Registrar,
	logger *slog.Logger,
) bool {
	heartbeats := time.NewTicker(heartbeatInterval)
	reports := time.NewTicker(reportInterval)
	defer heartbeats.Stop()
	defer reports.Stop()
	for {
		select {
		case <-ctx.Done():
			return true
		case <-heartbeats.C:
			if err := agentRegistrar.Heartbeat(ctx); err != nil {
				if ctx.Err() != nil {
					return true
				}
				logger.Error("agent heartbeat failed; re-registering", "error", err)
				return false
			}
		case <-reports.C:
			reportState(ctx, clusterCollector, clusterReporter, logger)
		}
	}
}

func reportState(
	ctx context.Context,
	clusterCollector *collector.Collector,
	clusterReporter *reporter.Reporter,
	logger *slog.Logger,
) {
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

func waitForRetry(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func clusterLabels(raw string) (map[string]string, error) {
	if raw == "" {
		return map[string]string{}, nil
	}
	var labels map[string]string
	if err := json.Unmarshal([]byte(raw), &labels); err != nil {
		return nil, fmt.Errorf("decode KFLEET_CLUSTER_LABELS: %w", err)
	}
	return labels, nil
}
