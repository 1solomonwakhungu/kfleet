// Package alerts evaluates fleet health rules and delivers signed webhooks.
package alerts

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/1solomonwakhungu/kfleet/internal/store"
	"github.com/1solomonwakhungu/kfleet/pkg/types"
	"github.com/google/uuid"
)

const (
	// EventType is sent in the webhook body and X-Kfleet-Event header.
	EventType            = "fleet.health.alert"
	maxResponseBodyBytes = 1024
	maxRetryDelay        = time.Hour
)

// Config controls webhook delivery. An empty WebhookURL keeps durable alert
// history enabled while marking delivery disabled.
type Config struct {
	WebhookURL   string
	Secret       string
	MaxAttempts  int
	RetryBase    time.Duration
	PollInterval time.Duration
	HTTPClient   *http.Client
}

// Manager evaluates rules and processes durable webhook deliveries.
type Manager struct {
	store        store.Store
	logger       *slog.Logger
	webhookURL   string
	secret       string
	maxAttempts  int
	retryBase    time.Duration
	pollInterval time.Duration
	client       *http.Client
	now          func() time.Time
}

// New constructs an alert manager.
func New(st store.Store, logger *slog.Logger, cfg Config) *Manager {
	maxAttempts := cfg.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 5
	}
	retryBase := cfg.RetryBase
	if retryBase <= 0 {
		retryBase = 5 * time.Second
	}
	pollInterval := cfg.PollInterval
	if pollInterval <= 0 {
		pollInterval = time.Second
	}
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &Manager{
		store:        st,
		logger:       logger,
		webhookURL:   cfg.WebhookURL,
		secret:       cfg.Secret,
		maxAttempts:  maxAttempts,
		retryBase:    retryBase,
		pollInterval: pollInterval,
		client:       client,
		now:          func() time.Time { return time.Now().UTC() },
	}
}

// Evaluate resolves stale alert states and creates alerts for every matching
// enabled rule. Creation is atomically suppressed inside each rule cooldown.
func (m *Manager) Evaluate(ctx context.Context, cluster types.Cluster) {
	now := m.now()
	if err := m.store.ResolveClusterAlerts(ctx, cluster.ID, cluster.Health, now); err != nil {
		m.logger.Error("failed to resolve cluster alerts", "cluster_id", cluster.ID, "error", err)
		return
	}
	rules, err := m.store.ListAlertRules(ctx)
	if err != nil {
		m.logger.Error("failed to list alert rules", "cluster_id", cluster.ID, "error", err)
		return
	}
	for _, rule := range rules {
		if !rule.Enabled || rule.Health != cluster.Health {
			continue
		}
		deliveryStatus := types.AlertDeliveryDisabled
		var nextDeliveryAt *time.Time
		if m.webhookURL != "" {
			deliveryStatus = types.AlertDeliveryPending
			nextDeliveryAt = &now
		}
		alert := types.Alert{
			ID:             uuid.NewString(),
			RuleID:         rule.ID,
			RuleName:       rule.Name,
			ClusterID:      cluster.ID,
			ClusterName:    cluster.Name,
			DedupeKey:      rule.ID + ":" + cluster.ID + ":" + string(cluster.Health),
			Health:         cluster.Health,
			Severity:       rule.Severity,
			Summary:        fmt.Sprintf("%s is %s", cluster.Name, cluster.Health),
			Status:         types.AlertStatusFiring,
			TriggeredAt:    now,
			UpdatedAt:      now,
			DeliveryStatus: deliveryStatus,
			NextDeliveryAt: nextDeliveryAt,
		}
		created, err := m.store.CreateAlertIfDue(ctx, alert, time.Duration(rule.CooldownSeconds)*time.Second)
		if err != nil {
			m.logger.Error("failed to create fleet health alert", "rule_id", rule.ID, "cluster_id", cluster.ID, "error", err)
			continue
		}
		if created {
			m.logger.Warn("fleet health alert created",
				"alert_id", alert.ID,
				"rule_id", rule.ID,
				"cluster_id", cluster.ID,
				"health", cluster.Health,
				"delivery_status", deliveryStatus,
			)
		}
	}
}

// Run processes due deliveries until the context is cancelled.
func (m *Manager) Run(ctx context.Context) {
	m.ProcessDue(ctx)
	ticker := time.NewTicker(m.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.ProcessDue(ctx)
		}
	}
}

// ProcessDue attempts each delivery currently due. It is public to support
// deterministic integration tests and one-shot operational verification.
func (m *Manager) ProcessDue(ctx context.Context) {
	if m.webhookURL == "" {
		return
	}
	due, err := m.store.ListDueAlertDeliveries(ctx, m.now(), 25)
	if err != nil {
		m.logger.Error("failed to list due alert deliveries", "error", err)
		return
	}
	for _, alert := range due {
		if ctx.Err() != nil {
			return
		}
		m.deliver(ctx, alert)
	}
}

func (m *Manager) deliver(ctx context.Context, alert types.Alert) {
	attempts := alert.DeliveryAttempts + 1
	deliveredAt := m.now()
	err := m.send(ctx, alert, deliveredAt)
	if err == nil {
		if err := m.store.RecordAlertDelivered(ctx, alert.ID, attempts, deliveredAt); err != nil {
			m.logger.Error("failed to record delivered alert", "alert_id", alert.ID, "error", err)
		}
		return
	}

	var nextAttemptAt *time.Time
	if attempts < m.maxAttempts {
		next := deliveredAt.Add(m.retryDelay(attempts))
		nextAttemptAt = &next
	}
	deliveryError := truncateError(err.Error(), 512)
	if recordErr := m.store.RecordAlertDeliveryFailure(
		ctx, alert.ID, attempts, nextAttemptAt, deliveryError, deliveredAt,
	); recordErr != nil {
		m.logger.Error("failed to record alert delivery failure", "alert_id", alert.ID, "error", recordErr)
		return
	}
	if nextAttemptAt == nil {
		m.logger.Error("alert webhook moved to dead letter",
			"alert_id", alert.ID,
			"attempts", attempts,
			"error", deliveryError,
		)
		return
	}
	m.logger.Warn("alert webhook delivery failed; retry scheduled",
		"alert_id", alert.ID,
		"attempts", attempts,
		"next_attempt_at", nextAttemptAt,
		"error", deliveryError,
	)
}

func (m *Manager) send(ctx context.Context, alert types.Alert, sentAt time.Time) error {
	payload := struct {
		ID         string    `json:"id"`
		Event      string    `json:"event"`
		OccurredAt time.Time `json:"occurredAt"`
		Rule       struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"rule"`
		Cluster struct {
			ID     string              `json:"id"`
			Name   string              `json:"name"`
			Health types.ClusterHealth `json:"health"`
		} `json:"cluster"`
		Severity types.AlertSeverity `json:"severity"`
		Summary  string              `json:"summary"`
	}{
		ID:         alert.ID,
		Event:      EventType,
		OccurredAt: alert.TriggeredAt,
		Severity:   alert.Severity,
		Summary:    alert.Summary,
	}
	payload.Rule.ID = alert.RuleID
	payload.Rule.Name = alert.RuleName
	payload.Cluster.ID = alert.ClusterID
	payload.Cluster.Name = alert.ClusterName
	payload.Cluster.Health = alert.Health

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode webhook: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create webhook request: %w", err)
	}
	timestamp := strconv.FormatInt(sentAt.Unix(), 10)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "kfleet-alerts/1")
	req.Header.Set("X-Kfleet-Event", EventType)
	req.Header.Set("X-Kfleet-Delivery", alert.ID)
	req.Header.Set("X-Kfleet-Timestamp", timestamp)
	req.Header.Set("X-Kfleet-Signature", Signature(m.secret, timestamp, body))

	response, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("send webhook: %w", err)
	}
	defer response.Body.Close()
	responseBody, readErr := io.ReadAll(io.LimitReader(response.Body, maxResponseBodyBytes))
	if readErr != nil {
		return fmt.Errorf("read webhook response: %w", readErr)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		detail := strings.TrimSpace(string(responseBody))
		if detail == "" {
			detail = http.StatusText(response.StatusCode)
		}
		return fmt.Errorf("webhook returned %d: %s", response.StatusCode, detail)
	}
	return nil
}

// Signature calculates the versioned HMAC-SHA256 webhook signature.
func Signature(secret, timestamp string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(timestamp))
	_, _ = mac.Write([]byte("."))
	_, _ = mac.Write(body)
	return "v1=" + hex.EncodeToString(mac.Sum(nil))
}

func (m *Manager) retryDelay(attempts int) time.Duration {
	delay := m.retryBase
	for range attempts - 1 {
		if delay >= maxRetryDelay/2 {
			return maxRetryDelay
		}
		delay *= 2
	}
	if delay > maxRetryDelay {
		return maxRetryDelay
	}
	return delay
}

func truncateError(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[:limit]
}
