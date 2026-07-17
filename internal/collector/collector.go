package collector

import (
	"context"
	"log/slog"
)

// ClusterSnapshot summarizes the current state of a managed cluster.
type ClusterSnapshot struct {
	NodeCount    int `json:"nodeCount"`
	PodCount     int `json:"podCount"`
	HealthyNodes int `json:"healthyNodes"`
}

// Collector gathers cluster state for agent heartbeats.
type Collector struct {
	logger *slog.Logger
}

// New constructs a cluster state collector.
func New(logger *slog.Logger) *Collector {
	return &Collector{logger: logger}
}

// CollectClusterInfo returns a placeholder snapshot until Kubernetes client
// integration is added.
func (c *Collector) CollectClusterInfo(ctx context.Context) (ClusterSnapshot, error) {
	if err := ctx.Err(); err != nil {
		return ClusterSnapshot{}, err
	}
	c.logger.Debug("collecting cluster information")
	return ClusterSnapshot{}, nil
}
