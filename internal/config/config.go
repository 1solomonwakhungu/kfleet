// Package config loads and validates hub runtime configuration.
package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"
)

const (
	defaultListenAddr        = ":8080"
	defaultDBPath            = "./kfleet.db"
	defaultLogLevel          = "info"
	defaultHeartbeatInterval = 30 * time.Second
	defaultAlertMaxAttempts  = 5
	defaultAlertRetryBase    = 5 * time.Second
	defaultAlertPollInterval = time.Second
)

// Config contains the hub server configuration.
type Config struct {
	ListenAddr         string
	DBPath             string
	LogLevel           string
	HeartbeatInterval  time.Duration
	RegistrationToken  string
	AlertWebhookURL    string
	AlertWebhookSecret string
	AlertMaxAttempts   int
	AlertRetryBase     time.Duration
	AlertPollInterval  time.Duration
}

// Load reads hub configuration from the environment, applying defaults where
// values are not set.

func Load() (*Config, error) {
	heartbeatInterval, err := positiveDurationEnv("KFLEET_HEARTBEAT_INTERVAL", defaultHeartbeatInterval)
	if err != nil {
		return nil, err
	}
	retryBase, err := positiveDurationEnv("KFLEET_ALERT_RETRY_BASE", defaultAlertRetryBase)
	if err != nil {
		return nil, err
	}
	pollInterval, err := positiveDurationEnv("KFLEET_ALERT_POLL_INTERVAL", defaultAlertPollInterval)
	if err != nil {
		return nil, err
	}
	maxAttempts := defaultAlertMaxAttempts
	if value := os.Getenv("KFLEET_ALERT_MAX_ATTEMPTS"); value != "" {
		parsed, parseErr := strconv.Atoi(value)
		if parseErr != nil || parsed <= 0 {
			return nil, fmt.Errorf("KFLEET_ALERT_MAX_ATTEMPTS must be a positive integer")
		}
		maxAttempts = parsed
	}
	webhookURL := os.Getenv("KFLEET_ALERT_WEBHOOK_URL")
	webhookSecret := os.Getenv("KFLEET_ALERT_WEBHOOK_SECRET")
	if webhookURL != "" {
		parsed, parseErr := url.Parse(webhookURL)
		if parseErr != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" || parsed.User != nil {
			return nil, fmt.Errorf("KFLEET_ALERT_WEBHOOK_URL must be an http or https URL without user information")
		}
		if webhookSecret == "" {
			return nil, fmt.Errorf("KFLEET_ALERT_WEBHOOK_SECRET is required when KFLEET_ALERT_WEBHOOK_URL is set")
		}
	}
	return &Config{
		ListenAddr:         envOrDefault("KFLEET_LISTEN_ADDR", defaultListenAddr),
		DBPath:             envOrDefault("KFLEET_DB_PATH", defaultDBPath),
		LogLevel:           envOrDefault("KFLEET_LOG_LEVEL", defaultLogLevel),
		HeartbeatInterval:  heartbeatInterval,
		RegistrationToken:  os.Getenv("KFLEET_REGISTRATION_TOKEN"),
		AlertWebhookURL:    webhookURL,
		AlertWebhookSecret: webhookSecret,
		AlertMaxAttempts:   maxAttempts,
		AlertRetryBase:     retryBase,
		AlertPollInterval:  pollInterval,
	}, nil
}

func positiveDurationEnv(key string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s must be a positive duration", key)
	}
	return parsed, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
