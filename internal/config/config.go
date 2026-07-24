// Package config loads and validates hub runtime configuration.
package config

import (
	"fmt"
	"os"
	"time"
)

const (
	defaultListenAddr        = ":8080"
	defaultDBPath            = "./kfleet.db"
	defaultLogLevel          = "info"
	defaultHeartbeatInterval = 30 * time.Second
	defaultSessionDuration   = 24 * time.Hour
)

// Config contains the hub server configuration.
type Config struct {
	ListenAddr        string
	DBPath            string
	LogLevel          string
	HeartbeatInterval time.Duration
	RegistrationToken string

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
	heartbeatInterval := defaultHeartbeatInterval
	if value := os.Getenv("KFLEET_HEARTBEAT_INTERVAL"); value != "" {
		parsed, err := time.ParseDuration(value)
		if err != nil || parsed <= 0 {
			return nil, fmt.Errorf("KFLEET_HEARTBEAT_INTERVAL must be a positive duration")
		}
		heartbeatInterval = parsed
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
		SessionDuration:        sessionDuration,
		SessionCookieSecure:    os.Getenv("KFLEET_SESSION_COOKIE_INSECURE") != "true",
		BootstrapAdminUsername: os.Getenv("KFLEET_BOOTSTRAP_ADMIN_USERNAME"),
		BootstrapAdminEmail:    os.Getenv("KFLEET_BOOTSTRAP_ADMIN_EMAIL"),
		BootstrapAdminPassword: os.Getenv("KFLEET_BOOTSTRAP_ADMIN_PASSWORD"),
	}, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
