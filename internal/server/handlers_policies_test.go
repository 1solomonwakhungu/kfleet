package server

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/1solomonwakhungu/kfleet/internal/config"
	"github.com/1solomonwakhungu/kfleet/internal/store"
	"github.com/1solomonwakhungu/kfleet/pkg/api"
	"github.com/1solomonwakhungu/kfleet/pkg/types"
)

func TestPolicyAPIIsReadOnlyAndTenantScoped(t *testing.T) {
	httpServer, st := newPolicyTestServer(t)
	now := time.Now().UTC()
	for _, cluster := range []types.Cluster{
		{ID: "a", TenantID: "tenant-a", Name: "alpha", Health: types.HealthHealthy, Version: "v1.31.1", LastHeartbeat: now, RegisteredAt: now, Labels: map[string]string{}},
		{ID: "b", TenantID: "tenant-b", Name: "bravo", Health: types.HealthHealthy, Version: "v1.30.1", LastHeartbeat: now, RegisteredAt: now, Labels: map[string]string{}},
	} {
		if err := st.CreateCluster(context.Background(), cluster); err != nil {
			t.Fatalf("CreateCluster(%s) error = %v", cluster.ID, err)
		}
	}

	catalogResponse := tenantRequest(t, httpServer, http.MethodGet, "/api/v1/policies", "tenant-a", "")
	if catalogResponse.StatusCode != http.StatusOK {
		t.Fatalf("catalog status = %d, want 200", catalogResponse.StatusCode)
	}
	var catalog api.PolicyListResponse
	decodeResponse(t, catalogResponse, &catalog)
	if len(catalog.Policies) != 7 {
		t.Fatalf("catalog policies = %d, want 7", len(catalog.Policies))
	}

	resultsResponse := tenantRequest(t, httpServer, http.MethodGet, "/api/v1/policies/results", "tenant-a", "")
	var results api.PolicyResultsResponse
	decodeResponse(t, resultsResponse, &results)
	if results.Summary.ClusterCount != 1 || len(results.Results) == 0 {
		t.Fatalf("tenant-a response = %#v, want one cluster with results", results)
	}
	for _, result := range results.Results {
		if result.Subject.ClusterID == "b" || result.Subject.ClusterName == "bravo" {
			t.Fatalf("tenant-a policy result leaked tenant-b: %#v", result)
		}
	}

	crossTenant := tenantRequest(t, httpServer, http.MethodGet, "/api/v1/clusters/b/policy-results", "tenant-a", "")
	if crossTenant.StatusCode != http.StatusNotFound {
		t.Fatalf("cross-tenant cluster status = %d, want 404", crossTenant.StatusCode)
	}
	crossTenant.Body.Close()

	filtered := tenantRequest(t, httpServer, http.MethodGet, "/api/v1/policies/results?status=fail", "tenant-a", "")
	var failureResults api.PolicyResultsResponse
	decodeResponse(t, filtered, &failureResults)
	for _, result := range failureResults.Results {
		if result.Status != types.PolicyFail {
			t.Fatalf("filtered result status = %q, want fail", result.Status)
		}
	}

	invalid := tenantRequest(t, httpServer, http.MethodGet, "/api/v1/policies/results?status=broken", "tenant-a", "")
	if invalid.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid filter status = %d, want 400", invalid.StatusCode)
	}
	invalid.Body.Close()

	invalidTenant := tenantRequest(t, httpServer, http.MethodGet, "/api/v1/policies/results", "../tenant", "")
	if invalidTenant.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid tenant status = %d, want 400", invalidTenant.StatusCode)
	}
	invalidTenant.Body.Close()

	mutation := tenantRequest(t, httpServer, http.MethodPost, "/api/v1/policies", "tenant-a", `{}`)
	if mutation.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("policy mutation status = %d, want 405", mutation.StatusCode)
	}
	mutation.Body.Close()
}

func TestClusterInventoryEndpointsAreTenantScoped(t *testing.T) {
	httpServer, st := newPolicyTestServer(t)
	now := time.Now().UTC()
	for _, cluster := range []types.Cluster{
		{ID: "a", TenantID: "tenant-a", Name: "alpha", Health: types.HealthHealthy, RegisteredAt: now, Labels: map[string]string{}},
		{ID: "b", TenantID: "tenant-b", Name: "bravo", Health: types.HealthHealthy, RegisteredAt: now, Labels: map[string]string{}},
	} {
		if err := st.CreateCluster(context.Background(), cluster); err != nil {
			t.Fatalf("CreateCluster(%s) error = %v", cluster.ID, err)
		}
	}

	response := tenantRequest(t, httpServer, http.MethodGet, "/api/v1/clusters", "tenant-a", "")
	var list api.ListClustersResponse
	decodeResponse(t, response, &list)
	if len(list.Clusters) != 1 || list.Clusters[0].ID != "a" {
		t.Fatalf("tenant-a clusters = %#v, want only a", list.Clusters)
	}

	hidden := tenantRequest(t, httpServer, http.MethodGet, "/api/v1/clusters/b", "tenant-a", "")
	if hidden.StatusCode != http.StatusNotFound {
		t.Fatalf("cross-tenant GET status = %d, want 404", hidden.StatusCode)
	}
	hidden.Body.Close()
}

func newPolicyTestServer(t *testing.T) (*httptest.Server, store.Store) {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "kfleet.db"))
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := st.Close(); err != nil {
			t.Errorf("store.Close() error = %v", err)
		}
	})
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := New(&config.Config{ListenAddr: ":0", HeartbeatInterval: time.Minute}, logger, st)
	httpServer := httptest.NewServer(srv.httpServer.Handler)
	registerDefaultSession(httpServer, st, sessionCookieFor(t, st, types.RoleReadOnly))
	t.Cleanup(httpServer.Close)
	return httpServer, st
}

func tenantRequest(t *testing.T, server *httptest.Server, method, path, tenantID, body string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(method, server.URL+path, bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	req.Header.Set(tenantHeader, tenantID)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: defaultSessionFor(server)})
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s error = %v", method, path, err)
	}
	return response
}
