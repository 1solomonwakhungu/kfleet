package config

import "testing"

func TestLoadReadsRegistrationToken(t *testing.T) {
	t.Setenv("KFLEET_REGISTRATION_TOKEN", "bootstrap-token")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.RegistrationToken != "bootstrap-token" {
		t.Fatalf("RegistrationToken = %q, want configured token", cfg.RegistrationToken)
	}
}

func TestLoadReadsDemoMode(t *testing.T) {
	t.Setenv("KFLEET_DEMO_MODE", "true")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !cfg.DemoMode {
		t.Fatal("DemoMode = false, want true")
	}
}

func TestLoadRejectsInvalidDemoMode(t *testing.T) {
	t.Setenv("KFLEET_DEMO_MODE", "sometimes")
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want invalid boolean error")
	}
}
