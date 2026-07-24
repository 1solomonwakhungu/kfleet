package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/1solomonwakhungu/kfleet/pkg/types"
)

const createAlertRulesTable = `
CREATE TABLE IF NOT EXISTS alert_rules (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	health TEXT NOT NULL,
	severity TEXT NOT NULL,
	cooldown_seconds INTEGER NOT NULL CHECK (cooldown_seconds >= 0),
	enabled INTEGER NOT NULL CHECK (enabled IN (0, 1)),
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL
)`

const createAlertsTable = `
CREATE TABLE IF NOT EXISTS alerts (
	id TEXT PRIMARY KEY,
	rule_id TEXT NOT NULL,
	rule_name TEXT NOT NULL,
	cluster_id TEXT NOT NULL,
	cluster_name TEXT NOT NULL,
	dedupe_key TEXT NOT NULL,
	health TEXT NOT NULL,
	severity TEXT NOT NULL,
	summary TEXT NOT NULL,
	status TEXT NOT NULL,
	triggered_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL,
	acknowledged_at TIMESTAMP,
	acknowledged_by TEXT NOT NULL DEFAULT '',
	resolved_at TIMESTAMP,
	delivery_status TEXT NOT NULL,
	delivery_attempts INTEGER NOT NULL DEFAULT 0,
	next_delivery_at TIMESTAMP,
	last_delivery_error TEXT NOT NULL DEFAULT '',
	delivered_at TIMESTAMP,
	dead_lettered_at TIMESTAMP,
	FOREIGN KEY (rule_id) REFERENCES alert_rules(id)
)`

const createAlertHistoryIndex = `
CREATE INDEX IF NOT EXISTS idx_alerts_history
ON alerts(triggered_at DESC, id DESC)`

const createAlertDeliveryIndex = `
CREATE INDEX IF NOT EXISTS idx_alerts_delivery
ON alerts(delivery_status, next_delivery_at)`

func seedDefaultAlertRules(db *sql.DB, now time.Time) error {
	defaults := []types.AlertRule{
		{
			ID:              "fleet-health-degraded",
			Name:            "Cluster health degraded",
			Health:          types.HealthDegraded,
			Severity:        types.AlertSeverityWarning,
			CooldownSeconds: int64((15 * time.Minute) / time.Second),
			Enabled:         true,
		},
		{
			ID:              "fleet-health-unreachable",
			Name:            "Cluster unreachable",
			Health:          types.HealthUnreachable,
			Severity:        types.AlertSeverityCritical,
			CooldownSeconds: int64((15 * time.Minute) / time.Second),
			Enabled:         true,
		},
	}
	for _, rule := range defaults {
		if _, err := db.Exec(`
			INSERT OR IGNORE INTO alert_rules (
				id, name, health, severity, cooldown_seconds, enabled, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			rule.ID, rule.Name, rule.Health, rule.Severity, rule.CooldownSeconds, rule.Enabled, now, now,
		); err != nil {
			return fmt.Errorf("insert %s: %w", rule.ID, err)
		}
	}
	return nil
}

func (s *sqliteStore) ListAlertRules(ctx context.Context) ([]types.AlertRule, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, health, severity, cooldown_seconds, enabled, created_at, updated_at
		FROM alert_rules
		ORDER BY severity DESC, name`)
	if err != nil {
		return nil, fmt.Errorf("list alert rules: %w", err)
	}
	defer rows.Close()

	rules := make([]types.AlertRule, 0)
	for rows.Next() {
		var rule types.AlertRule
		if err := rows.Scan(
			&rule.ID, &rule.Name, &rule.Health, &rule.Severity, &rule.CooldownSeconds,
			&rule.Enabled, &rule.CreatedAt, &rule.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan alert rule: %w", err)
		}
		rules = append(rules, rule)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list alert rules: %w", err)
	}
	return rules, nil
}

func (s *sqliteStore) UpsertAlertRule(ctx context.Context, rule types.AlertRule) error {
	now := rule.UpdatedAt
	if now.IsZero() {
		now = time.Now().UTC()
	}
	createdAt := rule.CreatedAt
	if createdAt.IsZero() {
		createdAt = now
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO alert_rules (
			id, name, health, severity, cooldown_seconds, enabled, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			health = excluded.health,
			severity = excluded.severity,
			cooldown_seconds = excluded.cooldown_seconds,
			enabled = excluded.enabled,
			updated_at = excluded.updated_at`,
		rule.ID, rule.Name, rule.Health, rule.Severity, rule.CooldownSeconds,
		rule.Enabled, createdAt, now,
	)
	if err != nil {
		return fmt.Errorf("upsert alert rule: %w", err)
	}
	return nil
}

func (s *sqliteStore) CreateAlertIfDue(
	ctx context.Context,
	alert types.Alert,
	cooldown time.Duration,
) (created bool, err error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin create alert transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var lastTriggered time.Time
	err = tx.QueryRowContext(ctx, `
		SELECT triggered_at
		FROM alerts
		WHERE dedupe_key = ?
		ORDER BY triggered_at DESC
		LIMIT 1`, alert.DedupeKey).Scan(&lastTriggered)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, fmt.Errorf("read alert cooldown: %w", err)
	}
	if err == nil && alert.TriggeredAt.Before(lastTriggered.Add(cooldown)) {
		if commitErr := tx.Commit(); commitErr != nil {
			return false, fmt.Errorf("commit alert cooldown check: %w", commitErr)
		}
		return false, nil
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO alerts (
			id, rule_id, rule_name, cluster_id, cluster_name, dedupe_key,
			health, severity, summary, status, triggered_at, updated_at,
			delivery_status, delivery_attempts, next_delivery_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		alert.ID, alert.RuleID, alert.RuleName, alert.ClusterID, alert.ClusterName, alert.DedupeKey,
		alert.Health, alert.Severity, alert.Summary, alert.Status, alert.TriggeredAt, alert.UpdatedAt,
		alert.DeliveryStatus, alert.DeliveryAttempts, alert.NextDeliveryAt,
	)
	if err != nil {
		return false, fmt.Errorf("create alert: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return false, fmt.Errorf("commit alert: %w", err)
	}
	return true, nil
}

func (s *sqliteStore) ResolveClusterAlerts(
	ctx context.Context,
	clusterID string,
	health types.ClusterHealth,
	resolvedAt time.Time,
) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE alerts
		SET status = ?, resolved_at = ?, updated_at = ?
		WHERE cluster_id = ?
		  AND health <> ?
		  AND status IN (?, ?)`,
		types.AlertStatusResolved, resolvedAt, resolvedAt, clusterID, health,
		types.AlertStatusFiring, types.AlertStatusAcknowledged,
	)
	if err != nil {
		return fmt.Errorf("resolve cluster alerts: %w", err)
	}
	return nil
}

func (s *sqliteStore) ListAlerts(
	ctx context.Context,
	status types.AlertStatus,
	limit int,
) ([]types.Alert, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	query := alertSelect + `
		FROM alerts`
	args := make([]any, 0, 2)
	if status != "" {
		query += ` WHERE status = ?`
		args = append(args, status)
	}
	query += ` ORDER BY triggered_at DESC, id DESC LIMIT ?`
	args = append(args, limit)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list alerts: %w", err)
	}
	defer rows.Close()

	alerts := make([]types.Alert, 0)
	for rows.Next() {
		alert, err := scanAlert(rows)
		if err != nil {
			return nil, fmt.Errorf("scan alert: %w", err)
		}
		alerts = append(alerts, alert)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list alerts: %w", err)
	}
	return alerts, nil
}

func (s *sqliteStore) GetAlert(ctx context.Context, id string) (types.Alert, error) {
	row := s.db.QueryRowContext(ctx, alertSelect+` FROM alerts WHERE id = ?`, id)
	alert, err := scanAlert(row)
	if errors.Is(err, sql.ErrNoRows) {
		return types.Alert{}, ErrNotFound
	}
	if err != nil {
		return types.Alert{}, fmt.Errorf("get alert: %w", err)
	}
	return alert, nil
}

func (s *sqliteStore) AcknowledgeAlert(
	ctx context.Context,
	id, acknowledgedBy string,
	acknowledgedAt time.Time,
) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE alerts
		SET status = ?, acknowledged_at = ?, acknowledged_by = ?, updated_at = ?
		WHERE id = ? AND status = ?`,
		types.AlertStatusAcknowledged, acknowledgedAt, acknowledgedBy, acknowledgedAt,
		id, types.AlertStatusFiring,
	)
	if err != nil {
		return fmt.Errorf("acknowledge alert: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read acknowledge result: %w", err)
	}
	if affected == 1 {
		return nil
	}
	if _, err := s.GetAlert(ctx, id); errors.Is(err, ErrNotFound) {
		return ErrNotFound
	} else if err != nil {
		return err
	}
	return ErrInvalidState
}

func (s *sqliteStore) ListDueAlertDeliveries(
	ctx context.Context,
	now time.Time,
	limit int,
) ([]types.Alert, error) {
	if limit <= 0 || limit > 100 {
		limit = 25
	}
	rows, err := s.db.QueryContext(ctx, alertSelect+`
		FROM alerts
		WHERE delivery_status IN (?, ?)
		  AND next_delivery_at IS NOT NULL
		  AND next_delivery_at <= ?
		ORDER BY next_delivery_at, triggered_at
		LIMIT ?`,
		types.AlertDeliveryPending, types.AlertDeliveryRetrying, now, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list due alert deliveries: %w", err)
	}
	defer rows.Close()

	alerts := make([]types.Alert, 0)
	for rows.Next() {
		alert, err := scanAlert(rows)
		if err != nil {
			return nil, fmt.Errorf("scan due alert: %w", err)
		}
		alerts = append(alerts, alert)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list due alert deliveries: %w", err)
	}
	return alerts, nil
}

func (s *sqliteStore) RecordAlertDelivered(
	ctx context.Context,
	id string,
	attempts int,
	deliveredAt time.Time,
) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE alerts
		SET delivery_status = ?, delivery_attempts = ?, next_delivery_at = NULL,
		    last_delivery_error = '', delivered_at = ?, updated_at = ?
		WHERE id = ?`,
		types.AlertDeliveryDelivered, attempts, deliveredAt, deliveredAt, id,
	)
	if err != nil {
		return fmt.Errorf("record alert delivered: %w", err)
	}
	return requireAffectedRow(result)
}

func (s *sqliteStore) RecordAlertDeliveryFailure(
	ctx context.Context,
	id string,
	attempts int,
	nextAttemptAt *time.Time,
	deliveryError string,
	failedAt time.Time,
) error {
	status := types.AlertDeliveryRetrying
	var deadLetteredAt *time.Time
	if nextAttemptAt == nil {
		status = types.AlertDeliveryDeadLetter
		deadLetteredAt = &failedAt
	}
	result, err := s.db.ExecContext(ctx, `
		UPDATE alerts
		SET delivery_status = ?, delivery_attempts = ?, next_delivery_at = ?,
		    last_delivery_error = ?, dead_lettered_at = ?, updated_at = ?
		WHERE id = ?`,
		status, attempts, nextAttemptAt, deliveryError, deadLetteredAt, failedAt, id,
	)
	if err != nil {
		return fmt.Errorf("record alert delivery failure: %w", err)
	}
	return requireAffectedRow(result)
}

const alertSelect = `
	SELECT id, rule_id, rule_name, cluster_id, cluster_name, dedupe_key,
	       health, severity, summary, status, triggered_at, updated_at,
	       acknowledged_at, acknowledged_by, resolved_at, delivery_status,
	       delivery_attempts, next_delivery_at, last_delivery_error,
	       delivered_at, dead_lettered_at`

type alertScanner interface {
	Scan(dest ...any) error
}

func scanAlert(scanner alertScanner) (types.Alert, error) {
	var alert types.Alert
	var acknowledgedAt, resolvedAt, nextDeliveryAt, deliveredAt, deadLetteredAt sql.NullTime
	err := scanner.Scan(
		&alert.ID, &alert.RuleID, &alert.RuleName, &alert.ClusterID, &alert.ClusterName, &alert.DedupeKey,
		&alert.Health, &alert.Severity, &alert.Summary, &alert.Status, &alert.TriggeredAt, &alert.UpdatedAt,
		&acknowledgedAt, &alert.AcknowledgedBy, &resolvedAt, &alert.DeliveryStatus,
		&alert.DeliveryAttempts, &nextDeliveryAt, &alert.LastDeliveryError,
		&deliveredAt, &deadLetteredAt,
	)
	if err != nil {
		return types.Alert{}, err
	}
	alert.AcknowledgedAt = nullTimePointer(acknowledgedAt)
	alert.ResolvedAt = nullTimePointer(resolvedAt)
	alert.NextDeliveryAt = nullTimePointer(nextDeliveryAt)
	alert.DeliveredAt = nullTimePointer(deliveredAt)
	alert.DeadLetteredAt = nullTimePointer(deadLetteredAt)
	return alert, nil
}

func nullTimePointer(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	return &value.Time
}
