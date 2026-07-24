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

func TestLoadAlertConfiguration(t *testing.T) {
	t.Setenv("KFLEET_ALERT_WEBHOOK_URL", "http://127.0.0.1:9099/hooks/kfleet")
	t.Setenv("KFLEET_ALERT_WEBHOOK_SECRET", "local-secret")
	t.Setenv("KFLEET_ALERT_MAX_ATTEMPTS", "3")
	t.Setenv("KFLEET_ALERT_RETRY_BASE", "25ms")
	t.Setenv("KFLEET_ALERT_POLL_INTERVAL", "10ms")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.AlertWebhookURL != "http://127.0.0.1:9099/hooks/kfleet" ||
		cfg.AlertWebhookSecret != "local-secret" ||
		cfg.AlertMaxAttempts != 3 ||
		cfg.AlertRetryBase != 25_000_000 ||
		cfg.AlertPollInterval != 10_000_000 {
		t.Fatalf("alert configuration = %#v", cfg)
	}
}

func TestLoadRejectsUnsafeOrIncompleteWebhookConfiguration(t *testing.T) {
	tests := []struct {
		name   string
		url    string
		secret string
	}{
		{name: "missing secret", url: "https://example.test/hook"},
		{name: "unsupported scheme", url: "file:///tmp/hook", secret: "secret"},
		{name: "userinfo", url: "https://user:password@example.test/hook", secret: "secret"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("KFLEET_ALERT_WEBHOOK_URL", tt.url)
			t.Setenv("KFLEET_ALERT_WEBHOOK_SECRET", tt.secret)
			if _, err := Load(); err == nil {
				t.Fatal("Load() error = nil, want validation error")
			}
		})
	}
}
