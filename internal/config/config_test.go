package config

import (
	"strings"
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
	if cfg.EventRetention != defaultEventRetention {
		t.Fatalf("EventRetention = %s, want default %s", cfg.EventRetention, defaultEventRetention)
	}
}

func TestLoadReadsEventRetention(t *testing.T) {
	t.Setenv("KFLEET_EVENT_RETENTION", "720h")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.EventRetention != 30*24*time.Hour {
		t.Fatalf("EventRetention = %s, want 720h", cfg.EventRetention)
	}
}

func TestLoadRejectsInvalidEventRetention(t *testing.T) {
	for _, value := range []string{"never", "0s", "-1h"} {
		t.Run(strings.ReplaceAll(value, "-", "negative"), func(t *testing.T) {
			t.Setenv("KFLEET_EVENT_RETENTION", value)
			if _, err := Load(); err == nil {
				t.Fatalf("Load() with KFLEET_EVENT_RETENTION=%q error = nil, want error", value)
			}
		})
	}
}
