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
	defaultEventRetention    = 90 * 24 * time.Hour
	defaultSessionDuration   = 24 * time.Hour
)

// Config contains the hub server configuration.
type Config struct {
	ListenAddr         string
	DBPath             string
	LogLevel           string
	HeartbeatInterval  time.Duration
	RegistrationToken  string
	DemoMode           bool
	AlertWebhookURL    string
	AlertWebhookSecret string
	AlertMaxAttempts   int
	AlertRetryBase     time.Duration
	AlertPollInterval  time.Duration
	EventRetention     time.Duration

	// SessionDuration controls how long a login session remains valid.
	SessionDuration time.Duration
	// SessionCookieSecure controls the Secure flag on the session cookie.
	// It must stay true in production (HTTPS); disable only for local,
	// plain-HTTP development.
	SessionCookieSecure bool

	// Bootstrap admin credentials. When set and no users exist yet, the hub
	// creates this admin account on startup. The password is never logged.
	BootstrapAdminUsername string
	BootstrapAdminEmail    string
	BootstrapAdminPassword string
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
	eventRetention := defaultEventRetention
	if value := os.Getenv("KFLEET_EVENT_RETENTION"); value != "" {
		parsed, err := time.ParseDuration(value)
		if err != nil || parsed <= 0 {
			return nil, fmt.Errorf("KFLEET_EVENT_RETENTION must be a positive duration")
		}
		eventRetention = parsed
	}
	demoMode := false
	if value := os.Getenv("KFLEET_DEMO_MODE"); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return nil, fmt.Errorf("KFLEET_DEMO_MODE must be a boolean")
		}
		demoMode = parsed
	}

	sessionDuration := defaultSessionDuration
	if value := os.Getenv("KFLEET_SESSION_DURATION"); value != "" {
		parsed, err := time.ParseDuration(value)
		if err != nil || parsed <= 0 {
			return nil, fmt.Errorf("KFLEET_SESSION_DURATION must be a positive duration")
		}
		sessionDuration = parsed
	}

	return &Config{
		ListenAddr:             envOrDefault("KFLEET_LISTEN_ADDR", defaultListenAddr),
		DBPath:                 envOrDefault("KFLEET_DB_PATH", defaultDBPath),
		LogLevel:               envOrDefault("KFLEET_LOG_LEVEL", defaultLogLevel),
		HeartbeatInterval:      heartbeatInterval,
		RegistrationToken:      os.Getenv("KFLEET_REGISTRATION_TOKEN"),
		DemoMode:               demoMode,
		AlertWebhookURL:        webhookURL,
		AlertWebhookSecret:     webhookSecret,
		AlertMaxAttempts:       maxAttempts,
		AlertRetryBase:         retryBase,
		AlertPollInterval:      pollInterval,
		EventRetention:         eventRetention,
		SessionDuration:        sessionDuration,
		SessionCookieSecure:    os.Getenv("KFLEET_SESSION_COOKIE_INSECURE") != "true",
		BootstrapAdminUsername: os.Getenv("KFLEET_BOOTSTRAP_ADMIN_USERNAME"),
		BootstrapAdminEmail:    os.Getenv("KFLEET_BOOTSTRAP_ADMIN_EMAIL"),
		BootstrapAdminPassword: os.Getenv("KFLEET_BOOTSTRAP_ADMIN_PASSWORD"),
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
