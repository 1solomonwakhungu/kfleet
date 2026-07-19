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
