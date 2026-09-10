package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	t.Setenv("KFLEET_HUB_URL", "https://hub.example.test/")
	t.Setenv("KFLEET_CLUSTER_NAME", "production")
	t.Setenv("KFLEET_HUB_TOKEN", "secret")
	t.Setenv("KFLEET_REPORT_INTERVAL", "45s")
	t.Setenv("KUBECONFIG", "/tmp/kubeconfig")
	t.Setenv("KFLEET_HEALTH_ADDR", "127.0.0.1:19090")
	t.Setenv("KFLEET_TENANT_ID", "platform-a")
	t.Setenv("KFLEET_LOG_LEVEL", "debug")
	t.Setenv("KFLEET_CLUSTER_LABELS", `{"env":"prod","region":"eu-west"}`)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.HubURL != "https://hub.example.test" || cfg.ClusterName != "production" || cfg.HubToken != "secret" {
		t.Fatalf("Load() identity config = %#v", cfg)
	}
	if cfg.ReportInterval != 45*time.Second || cfg.Kubeconfig != "/tmp/kubeconfig" || cfg.HealthAddress != "127.0.0.1:19090" {
		t.Fatalf("Load() runtime config = %#v", cfg)
	}
	if cfg.TenantID != "platform-a" {
		t.Fatalf("TenantID = %q, want platform-a", cfg.TenantID)
	}
	if cfg.LogLevel != "debug" {
		t.Fatalf("LogLevel = %q, want debug", cfg.LogLevel)
	}
	wantLabels := map[string]string{"env": "prod", "region": "eu-west"}
	if len(cfg.ClusterLabels) != len(wantLabels) || cfg.ClusterLabels["env"] != "prod" || cfg.ClusterLabels["region"] != "eu-west" {
		t.Fatalf("ClusterLabels = %#v, want %#v", cfg.ClusterLabels, wantLabels)
	}
}

func TestLoadDefaultReportInterval(t *testing.T) {
	t.Setenv("KFLEET_HUB_URL", "http://hub")
	t.Setenv("KFLEET_CLUSTER_NAME", "development")
	t.Setenv("KFLEET_HUB_TOKEN", "secret")
	t.Setenv("KFLEET_REPORT_INTERVAL", "")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.ReportInterval != 30*time.Second {
		t.Fatalf("ReportInterval = %v, want 30s", cfg.ReportInterval)
	}
	if cfg.HealthAddress != ":8081" {
		t.Fatalf("HealthAddress = %q, want :8081", cfg.HealthAddress)
	}
	if cfg.TenantID != "default" {
		t.Fatalf("TenantID = %q, want default", cfg.TenantID)
	}
	if cfg.LogLevel != "info" {
		t.Fatalf("LogLevel = %q, want info", cfg.LogLevel)
	}
	if cfg.ClusterLabels == nil || len(cfg.ClusterLabels) != 0 {
		t.Fatalf("ClusterLabels = %#v, want an empty map", cfg.ClusterLabels)
	}
}

func TestLoadRejectsInvalidConfiguration(t *testing.T) {
	t.Setenv("KFLEET_HUB_URL", "")
	t.Setenv("KFLEET_CLUSTER_NAME", "production")
	t.Setenv("KFLEET_HUB_TOKEN", "secret")
	if _, err := Load(); err == nil {
		t.Fatal("Load() with missing hub URL returned nil error")
	}

	t.Setenv("KFLEET_HUB_URL", "http://hub")
	t.Setenv("KFLEET_REPORT_INTERVAL", "not-a-duration")
	if _, err := Load(); err == nil {
		t.Fatal("Load() with invalid interval returned nil error")
	}

	t.Setenv("KFLEET_REPORT_INTERVAL", "30s")
	t.Setenv("KFLEET_TENANT_ID", "../invalid")
	if _, err := Load(); err == nil {
		t.Fatal("Load() with invalid tenant ID returned nil error")
	}
}

// TestLoadRequiresHubToken proves a misconfigured agent fails fast instead of
// starting up and retrying registration forever with an empty bearer token.
func TestLoadRequiresHubToken(t *testing.T) {
	t.Setenv("KFLEET_HUB_URL", "https://hub.example.test")
	t.Setenv("KFLEET_CLUSTER_NAME", "production")
	t.Setenv("KFLEET_HUB_TOKEN", "   ")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() with blank hub token returned nil error")
	}
	if !strings.Contains(err.Error(), "KFLEET_HUB_TOKEN") {
		t.Fatalf("Load() error = %v, want the missing variable named", err)
	}
}

func TestLoadRejectsMalformedHubURL(t *testing.T) {
	t.Setenv("KFLEET_CLUSTER_NAME", "production")
	t.Setenv("KFLEET_HUB_TOKEN", "secret")
	for _, raw := range []string{"hub.example.test", "ftp://hub.example.test", "://nope"} {
		t.Setenv("KFLEET_HUB_URL", raw)
		if _, err := Load(); err == nil {
			t.Fatalf("Load() with hub URL %q returned nil error", raw)
		}
	}
}

// TestLoadLogLevel proves KFLEET_LOG_LEVEL accepts the hub's levels, defaults
// to info, and fails fast on anything else instead of silently falling back.
func TestLoadLogLevel(t *testing.T) {
	t.Setenv("KFLEET_HUB_URL", "https://hub.example.test")
	t.Setenv("KFLEET_CLUSTER_NAME", "production")
	t.Setenv("KFLEET_HUB_TOKEN", "secret")

	for _, level := range []string{"debug", "info", "warn", "error"} {
		t.Setenv("KFLEET_LOG_LEVEL", level)
		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() with log level %q error = %v", level, err)
		}
		if cfg.LogLevel != level {
			t.Fatalf("LogLevel = %q, want %q", cfg.LogLevel, level)
		}
	}

	for _, level := range []string{"verbose", "DEBUG", "trace", "warning"} {
		t.Setenv("KFLEET_LOG_LEVEL", level)
		_, err := Load()
		if err == nil {
			t.Fatalf("Load() with invalid log level %q returned nil error", level)
		}
		if !strings.Contains(err.Error(), "KFLEET_LOG_LEVEL") {
			t.Fatalf("Load() error = %v, want the invalid variable named", err)
		}
	}
}

// TestLoadClusterLabels proves KFLEET_CLUSTER_LABELS is parsed in Load with
// the same fail-fast behavior as the other environment variables.
func TestLoadClusterLabels(t *testing.T) {
	t.Setenv("KFLEET_HUB_URL", "https://hub.example.test")
	t.Setenv("KFLEET_CLUSTER_NAME", "production")
	t.Setenv("KFLEET_HUB_TOKEN", "secret")

	t.Setenv("KFLEET_CLUSTER_LABELS", "not-json")
	_, err := Load()
	if err == nil {
		t.Fatal("Load() with malformed cluster labels returned nil error")
	}
	if !strings.Contains(err.Error(), "KFLEET_CLUSTER_LABELS") {
		t.Fatalf("Load() error = %v, want the failing variable named", err)
	}

	t.Setenv("KFLEET_CLUSTER_LABELS", `{"team":"platform"}`)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() with valid cluster labels error = %v", err)
	}
	if cfg.ClusterLabels["team"] != "platform" {
		t.Fatalf("ClusterLabels = %#v, want team=platform", cfg.ClusterLabels)
	}

	// A blank value means no labels, not an error.
	t.Setenv("KFLEET_CLUSTER_LABELS", "")
	cfg, err = Load()
	if err != nil {
		t.Fatalf("Load() with unset cluster labels error = %v", err)
	}
	if len(cfg.ClusterLabels) != 0 {
		t.Fatalf("ClusterLabels = %#v, want empty", cfg.ClusterLabels)
	}
}
