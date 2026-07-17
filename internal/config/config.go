package config

import "os"

const (
	defaultListenAddr = ":8080"
	defaultDBPath     = "./kfleet.db"
	defaultLogLevel   = "info"
)

// Config contains the hub server configuration.
type Config struct {
	ListenAddr string
	DBPath     string
	LogLevel   string
}

// Load reads hub configuration from the environment, applying defaults where
// values are not set.
func Load() (*Config, error) {
	return &Config{
		ListenAddr: envOrDefault("KFLEET_LISTEN_ADDR", defaultListenAddr),
		DBPath:     envOrDefault("KFLEET_DB_PATH", defaultDBPath),
		LogLevel:   envOrDefault("KFLEET_LOG_LEVEL", defaultLogLevel),
	}, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
