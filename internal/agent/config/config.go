// Package config loads runtime configuration for the kfleet agent.
package config

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

var tenantIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,62}$`)

const (
	defaultReportInterval = 30 * time.Second
	defaultHealthAddress  = ":8081"
)

// Config contains the agent runtime configuration.
type Config struct {
	HubURL         string
	ClusterName    string
	HubToken       string
	TenantID       string
	ReportInterval time.Duration
	Kubeconfig     string
	HealthAddress  string
}

// Load reads agent configuration from environment variables.
func Load() (*Config, error) {
	hubURL := strings.TrimRight(strings.TrimSpace(os.Getenv("KFLEET_HUB_URL")), "/")
	if hubURL == "" {
		return nil, errors.New("KFLEET_HUB_URL is required")
	}
	clusterName := strings.TrimSpace(os.Getenv("KFLEET_CLUSTER_NAME"))
	if clusterName == "" {
		return nil, errors.New("KFLEET_CLUSTER_NAME is required")
	}

	interval := defaultReportInterval
	if raw := os.Getenv("KFLEET_REPORT_INTERVAL"); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil {
			return nil, fmt.Errorf("parse KFLEET_REPORT_INTERVAL: %w", err)
		}
		if parsed <= 0 {
			return nil, errors.New("KFLEET_REPORT_INTERVAL must be greater than zero")
		}
		interval = parsed
	}

	tenantID := envOrDefault("KFLEET_TENANT_ID", "default")
	if !tenantIDPattern.MatchString(tenantID) {
		return nil, errors.New("KFLEET_TENANT_ID must be a lowercase tenant identifier")
	}

	return &Config{
		HubURL:         hubURL,
		ClusterName:    clusterName,
		HubToken:       os.Getenv("KFLEET_HUB_TOKEN"),
		TenantID:       tenantID,
		ReportInterval: interval,
		Kubeconfig:     os.Getenv("KUBECONFIG"),
		HealthAddress:  envOrDefault("KFLEET_HEALTH_ADDR", defaultHealthAddress),
	}, nil
}

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
