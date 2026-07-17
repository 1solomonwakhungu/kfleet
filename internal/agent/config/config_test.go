package config

import (
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	t.Setenv("KFLEET_HUB_URL", "https://hub.example.test/")
	t.Setenv("KFLEET_CLUSTER_NAME", "production")
	t.Setenv("KFLEET_HUB_TOKEN", "secret")
	t.Setenv("KFLEET_REPORT_INTERVAL", "45s")
	t.Setenv("KUBECONFIG", "/tmp/kubeconfig")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.HubURL != "https://hub.example.test" || cfg.ClusterName != "production" || cfg.HubToken != "secret" {
		t.Fatalf("Load() identity config = %#v", cfg)
	}
	if cfg.ReportInterval != 45*time.Second || cfg.Kubeconfig != "/tmp/kubeconfig" {
		t.Fatalf("Load() runtime config = %#v", cfg)
	}
}

func TestLoadDefaultReportInterval(t *testing.T) {
	t.Setenv("KFLEET_HUB_URL", "http://hub")
	t.Setenv("KFLEET_CLUSTER_NAME", "development")
	t.Setenv("KFLEET_REPORT_INTERVAL", "")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.ReportInterval != 30*time.Second {
		t.Fatalf("ReportInterval = %v, want 30s", cfg.ReportInterval)
	}
}

func TestLoadRejectsInvalidConfiguration(t *testing.T) {
	t.Setenv("KFLEET_HUB_URL", "")
	t.Setenv("KFLEET_CLUSTER_NAME", "production")
	if _, err := Load(); err == nil {
		t.Fatal("Load() with missing hub URL returned nil error")
	}

	t.Setenv("KFLEET_HUB_URL", "http://hub")
	t.Setenv("KFLEET_REPORT_INTERVAL", "not-a-duration")
	if _, err := Load(); err == nil {
		t.Fatal("Load() with invalid interval returned nil error")
	}
}
