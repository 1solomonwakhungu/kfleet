// Package agentconfig loads and validates agent runtime configuration.
package agentconfig

import (
	"errors"
	"fmt"
	"os"
	"time"
)

const (
	defaultHubURL            = "http://localhost:8080"
	defaultHeartbeatInterval = 30 * time.Second
)

// Config contains the agent daemon configuration.
type Config struct {
	HubURL            string
	ClusterName       string
	Kubeconfig        string
	HeartbeatInterval time.Duration
}

// Load reads agent configuration from the environment.
func Load() (*Config, error) {
	clusterName := os.Getenv("KFLEET_CLUSTER_NAME")
	if clusterName == "" {
		return nil, errors.New("KFLEET_CLUSTER_NAME is required")
	}

	interval := defaultHeartbeatInterval
	if rawInterval := os.Getenv("KFLEET_HEARTBEAT_INTERVAL"); rawInterval != "" {
		parsedInterval, err := time.ParseDuration(rawInterval)
		if err != nil {
			return nil, fmt.Errorf("parse KFLEET_HEARTBEAT_INTERVAL: %w", err)
		}
		if parsedInterval <= 0 {
			return nil, errors.New("KFLEET_HEARTBEAT_INTERVAL must be greater than zero")
		}
		interval = parsedInterval
	}

	hubURL := os.Getenv("KFLEET_HUB_URL")
	if hubURL == "" {
		hubURL = defaultHubURL
	}

	return &Config{
		HubURL:            hubURL,
		ClusterName:       clusterName,
		Kubeconfig:        os.Getenv("KUBECONFIG"),
		HeartbeatInterval: interval,
	}, nil
}
