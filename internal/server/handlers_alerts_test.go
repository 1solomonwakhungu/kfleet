package server

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/1solomonwakhungu/kfleet/pkg/types"
)

func TestAlertHistoryAndAcknowledgementAreTenantScoped(t *testing.T) {
	server := newTestHTTPServer(t)
	st := serverStoreForTest(t, server)
	now := time.Now().UTC()
	for _, alert := range []types.Alert{
		{
			ID: "alert-a", TenantID: "tenant-a", RuleID: "fleet-health-degraded",
			RuleName: "Cluster health degraded", ClusterID: "cluster-a", ClusterName: "alpha",
			DedupeKey: "degraded:cluster-a", Health: types.HealthDegraded,
			Severity: types.AlertSeverityWarning, Summary: "alpha is degraded",
			Status: types.AlertStatusFiring, TriggeredAt: now, UpdatedAt: now,
			DeliveryStatus: types.AlertDeliveryDisabled,
		},
		{
			ID: "alert-b", TenantID: "tenant-b", RuleID: "fleet-health-degraded",
			RuleName: "Cluster health degraded", ClusterID: "cluster-b", ClusterName: "bravo",
			DedupeKey: "degraded:cluster-b", Health: types.HealthDegraded,
			Severity: types.AlertSeverityWarning, Summary: "bravo is degraded",
			Status: types.AlertStatusFiring, TriggeredAt: now, UpdatedAt: now,
			DeliveryStatus: types.AlertDeliveryDisabled,
		},
	} {
		created, err := st.CreateAlertIfDue(context.Background(), alert, 0)
		if err != nil || !created {
			t.Fatalf("CreateAlertIfDue(%s) = %v, %v, want true, nil", alert.ID, created, err)
		}
	}

	list := tenantRequest(t, server, http.MethodGet, "/api/v1/alerts", "tenant-a", "")
	var history alertListResponse
	decodeResponse(t, list, &history)
	if len(history.Alerts) != 1 || history.Alerts[0].ID != "alert-a" {
		t.Fatalf("tenant-a alert history = %#v, want only alert-a", history.Alerts)
	}

	crossTenant := tenantRequest(
		t,
		server,
		http.MethodPost,
		"/api/v1/alerts/alert-b/acknowledge",
		"tenant-a",
		`{"acknowledgedBy":"spoofed"}`,
	)
	crossTenant.Body.Close()
	if crossTenant.StatusCode != http.StatusNotFound {
		t.Fatalf("cross-tenant acknowledge status = %d, want %d", crossTenant.StatusCode, http.StatusNotFound)
	}
}

func TestAlertHistoryAndAcknowledgementLifecycle(t *testing.T) {
	server, registration := registeredAgent(t)
	approveAgent(t, server, registration.ClusterID)

	snapshot := agentRequest(
		t,
		server,
		http.MethodPost,
		"/api/v1/clusters/"+registration.ClusterID+"/status",
		registration.Token,
		snapshotPayload(),
	)
	if snapshot.StatusCode != http.StatusOK {
		t.Fatalf("snapshot status = %d, want %d", snapshot.StatusCode, http.StatusOK)
	}
	snapshot.Body.Close()

	list := request(t, server, http.MethodGet, "/api/v1/alerts", "")
	if list.StatusCode != http.StatusOK {
		t.Fatalf("alert list status = %d, want %d", list.StatusCode, http.StatusOK)
	}
	var history alertListResponse
	decodeResponse(t, list, &history)
	if len(history.Alerts) != 1 {
		t.Fatalf("alert history length = %d, want 1", len(history.Alerts))
	}
	alert := history.Alerts[0]
	if alert.RuleID != "fleet-health-degraded" ||
		alert.Status != types.AlertStatusFiring ||
		alert.DeliveryStatus != types.AlertDeliveryDisabled {
		t.Fatalf("created alert = %#v", alert)
	}

	acknowledged := request(
		t,
		server,
		http.MethodPost,
		"/api/v1/alerts/"+alert.ID+"/acknowledge",
		`{"acknowledgedBy":"on-call"}`,
	)
	if acknowledged.StatusCode != http.StatusOK {
		t.Fatalf("acknowledge status = %d, want %d", acknowledged.StatusCode, http.StatusOK)
	}
	var updated types.Alert
	decodeResponse(t, acknowledged, &updated)
	if updated.Status != types.AlertStatusAcknowledged ||
		updated.AcknowledgedBy == "" ||
		updated.AcknowledgedBy == "on-call" ||
		updated.AcknowledgedAt == nil {
		t.Fatalf("acknowledged alert = %#v", updated)
	}

	repeated := request(
		t,
		server,
		http.MethodPost,
		"/api/v1/alerts/"+alert.ID+"/acknowledge",
		`{"acknowledgedBy":"on-call"}`,
	)
	defer repeated.Body.Close()
	if repeated.StatusCode != http.StatusConflict {
		t.Fatalf("repeated acknowledge status = %d, want %d", repeated.StatusCode, http.StatusConflict)
	}
}

func TestAlertRuleEndpoints(t *testing.T) {
	server := newTestHTTPServer(t)
	list := request(t, server, http.MethodGet, "/api/v1/alert-rules", "")
	var rules alertRuleListResponse
	decodeResponse(t, list, &rules)
	if len(rules.Rules) != 2 {
		t.Fatalf("default alert rules = %d, want 2", len(rules.Rules))
	}

	update := request(t, server, http.MethodPut, "/api/v1/alert-rules/fleet-health-degraded", `{
		"name":"Degraded cluster",
		"health":"degraded",
		"severity":"critical",
		"cooldownSeconds":30,
		"enabled":true
	}`)
	if update.StatusCode != http.StatusOK {
		t.Fatalf("update rule status = %d, want %d", update.StatusCode, http.StatusOK)
	}
	var saved types.AlertRule
	decodeResponse(t, update, &saved)
	if saved.ID != "fleet-health-degraded" ||
		saved.Name != "Degraded cluster" ||
		saved.Severity != types.AlertSeverityCritical ||
		saved.CooldownSeconds != 30 {
		t.Fatalf("saved rule = %#v", saved)
	}

	invalid := request(t, server, http.MethodPut, "/api/v1/alert-rules/invalid", `{
		"name":"Healthy is not alertable",
		"health":"healthy",
		"severity":"warning",
		"cooldownSeconds":30,
		"enabled":true
	}`)
	defer invalid.Body.Close()
	if invalid.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid rule status = %d, want %d", invalid.StatusCode, http.StatusBadRequest)
	}
}
