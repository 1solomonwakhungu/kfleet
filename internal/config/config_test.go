package config

import (
	"testing"
	"time"
)

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

func TestLoadDefaultsSessionSettings(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.SessionDuration != 24*time.Hour {
		t.Fatalf("SessionDuration = %v, want 24h default", cfg.SessionDuration)
	}
	if !cfg.SessionCookieSecure {
		t.Fatalf("SessionCookieSecure = false, want true by default")
	}
}

func TestLoadReadsBootstrapAdminAndSessionOverrides(t *testing.T) {
	t.Setenv("KFLEET_BOOTSTRAP_ADMIN_USERNAME", "admin")
	t.Setenv("KFLEET_BOOTSTRAP_ADMIN_EMAIL", "admin@example.com")
	t.Setenv("KFLEET_BOOTSTRAP_ADMIN_PASSWORD", "hunter2-hunter2")
	t.Setenv("KFLEET_SESSION_DURATION", "1h")
	t.Setenv("KFLEET_SESSION_COOKIE_INSECURE", "true")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.BootstrapAdminUsername != "admin" || cfg.BootstrapAdminEmail != "admin@example.com" || cfg.BootstrapAdminPassword != "hunter2-hunter2" {
		t.Fatalf("bootstrap admin config = %+v, want configured values", cfg)
	}
	if cfg.SessionDuration != time.Hour {
		t.Fatalf("SessionDuration = %v, want 1h", cfg.SessionDuration)
	}
	if cfg.SessionCookieSecure {
		t.Fatalf("SessionCookieSecure = true, want false when KFLEET_SESSION_COOKIE_INSECURE=true")
	}
}

func TestLoadRejectsInvalidSessionDuration(t *testing.T) {
	t.Setenv("KFLEET_SESSION_DURATION", "not-a-duration")
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want error for invalid KFLEET_SESSION_DURATION")
	}
}
