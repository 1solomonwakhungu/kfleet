// Package config loads and validates hub runtime configuration.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	defaultListenAddr        = ":8080"
	defaultDBPath            = "./kfleet.db"
	defaultLogLevel          = "info"
	defaultHeartbeatInterval = 30 * time.Second
)

// Config contains the hub server configuration.
type Config struct {
	ListenAddr        string
	DBPath            string
	LogLevel          string
	HeartbeatInterval time.Duration
	RegistrationToken string
	DemoMode          bool
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
	demoMode := false
	if value := os.Getenv("KFLEET_DEMO_MODE"); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return nil, fmt.Errorf("KFLEET_DEMO_MODE must be a boolean")
		}
		demoMode = parsed
	}
	return &Config{
		ListenAddr:        envOrDefault("KFLEET_LISTEN_ADDR", defaultListenAddr),
		DBPath:            envOrDefault("KFLEET_DB_PATH", defaultDBPath),
		LogLevel:          envOrDefault("KFLEET_LOG_LEVEL", defaultLogLevel),
		HeartbeatInterval: heartbeatInterval,
		RegistrationToken: os.Getenv("KFLEET_REGISTRATION_TOKEN"),
		DemoMode:          demoMode,
	}, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
