package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

func TestAgentRegistrationApprovalAndHeartbeat(t *testing.T) {
	httpServer, _, _ := newAgentTestServer(t, 30*time.Second)

	register := agentRequest(t, httpServer, http.MethodPost, "/api/v1/agents/register", testRegistrationToken, `{
		"name":"production","labels":{"region":"us-central1"},"agentVersion":"0.1.0","k8sVersion":"v1.32.3"
	}`)
	if register.StatusCode != http.StatusCreated {
		t.Fatalf("register status = %d, want %d", register.StatusCode, http.StatusCreated)
	}
	var registration api.RegisterClusterResponse
	decodeResponse(t, register, &registration)
	if registration.ClusterID == "" || len(registration.Token) != 64 {
		t.Fatalf("registration = %#v, want ID and 32-byte hex token", registration)
	}

	pending := agentRequest(t, httpServer, http.MethodGet, "/api/v1/agents/pending", "", "")
	if pending.StatusCode != http.StatusOK {
		t.Fatalf("pending status = %d, want %d", pending.StatusCode, http.StatusOK)
	}
	var pendingList api.ListClustersResponse
	decodeResponse(t, pending, &pendingList)
	if len(pendingList.Clusters) != 1 || pendingList.Clusters[0].ID != registration.ClusterID {
		t.Fatalf("pending agents = %#v, want registered agent", pendingList)
	}
	if pendingList.Clusters[0].AgentVersion != "0.1.0" || pendingList.Clusters[0].Version != "v1.32.3" {
		t.Fatalf("pending agent versions = (%q, %q), want reported versions", pendingList.Clusters[0].AgentVersion, pendingList.Clusters[0].Version)
	}

	heartbeatBody := `{"clusterId":"` + registration.ClusterID + `","nodeCount":3,"healthyNodes":3,"podCount":12,"version":"v1.31.1"}`
	beforeApproval := agentRequest(t, httpServer, http.MethodPost, "/api/v1/agents/heartbeat", registration.Token, heartbeatBody)
	if beforeApproval.StatusCode != http.StatusForbidden {
		t.Fatalf("heartbeat before approval status = %d, want %d", beforeApproval.StatusCode, http.StatusForbidden)
	}
	beforeApproval.Body.Close()

	approve := agentRequest(t, httpServer, http.MethodPost, "/api/v1/agents/"+registration.ClusterID+"/approve", "", "")
	if approve.StatusCode != http.StatusOK {
		t.Fatalf("approve status = %d, want %d", approve.StatusCode, http.StatusOK)
	}
	approve.Body.Close()

	heartbeat := agentRequest(t, httpServer, http.MethodPost, "/api/v1/agents/heartbeat", registration.Token, heartbeatBody)
	if heartbeat.StatusCode != http.StatusOK {
		t.Fatalf("heartbeat status = %d, want %d", heartbeat.StatusCode, http.StatusOK)
	}
	var cluster types.Cluster
	decodeResponse(t, heartbeat, &cluster)
	if cluster.Health != types.HealthHealthy || cluster.NodeCount != 3 || cluster.PodCount != 12 {
		t.Fatalf("heartbeat cluster = %#v, want healthy snapshot", cluster)
	}

	wrongToken := agentRequest(t, httpServer, http.MethodPost, "/api/v1/agents/heartbeat", "wrong", heartbeatBody)
	if wrongToken.StatusCode != http.StatusUnauthorized {
		t.Fatalf("heartbeat wrong token status = %d, want %d", wrongToken.StatusCode, http.StatusUnauthorized)
	}
	wrongToken.Body.Close()

	missingApprove := agentRequest(t, httpServer, http.MethodPost, "/api/v1/agents/missing/approve", "", "")
	if missingApprove.StatusCode != http.StatusNotFound {
		t.Fatalf("missing approve status = %d, want %d", missingApprove.StatusCode, http.StatusNotFound)
	}
	missingApprove.Body.Close()
}

func TestAgentRegistrationRequiresConfiguredBootstrapToken(t *testing.T) {
	httpServer, _, _ := newAgentTestServerWithConfig(t, &config.Config{
		ListenAddr:        ":0",
		RegistrationToken: "bootstrap-token",
	})

	for _, token := range []string{"", "wrong-token"} {
		response := agentRequest(t, httpServer, http.MethodPost, "/api/v1/agents/register", token, `{"name":"secured"}`)
		response.Body.Close()
		if response.StatusCode != http.StatusUnauthorized {
			t.Fatalf("registration with token %q status = %d, want %d", token, response.StatusCode, http.StatusUnauthorized)
		}
	}

	response := agentRequest(t, httpServer, http.MethodPost, "/api/v1/agents/register", "bootstrap-token", `{"name":"secured"}`)
	response.Body.Close()
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("registration with bootstrap token status = %d, want %d", response.StatusCode, http.StatusCreated)
	}
}

func TestAgentRegistrationFailsClosedWithoutConfiguredToken(t *testing.T) {
	httpServer, _, _ := newAgentTestServerWithConfig(t, &config.Config{ListenAddr: ":0"})

	// With no rotated registration token and no static KFLEET_REGISTRATION_TOKEN
	// the hub must reject registration attempts instead of admitting them.
	for _, token := range []string{"", "any-token"} {
		response := agentRequest(t, httpServer, http.MethodPost, "/api/v1/agents/register", token, `{"name":"unauthenticated"}`)
		if response.StatusCode != http.StatusForbidden {
			response.Body.Close()
			t.Fatalf("registration without configured hub token status = %d, want %d", response.StatusCode, http.StatusForbidden)
		}
		var errResponse api.ErrorResponse
		decodeResponse(t, response, &errResponse)
		if errResponse.Error != "agent registration is disabled on this hub" {
			t.Fatalf("error = %q, want agent registration is disabled on this hub", errResponse.Error)
		}
	}

	// Other endpoints keep working while registration is disabled.
	pending := agentRequest(t, httpServer, http.MethodGet, "/api/v1/agents/pending", "", "")
	defer pending.Body.Close()
	if pending.StatusCode != http.StatusOK {
		t.Fatalf("pending agents status = %d, want %d", pending.StatusCode, http.StatusOK)
	}
}

func TestAgentReRegistrationRequiresCurrentAgentToken(t *testing.T) {
	httpServer, _, _ := newAgentTestServerWithConfig(t, &config.Config{
		ListenAddr:        ":0",
		RegistrationToken: testRegistrationToken,
	})

	register := agentRequest(t, httpServer, http.MethodPost, "/api/v1/agents/register", testRegistrationToken, `{"name":"production"}`)
	if register.StatusCode != http.StatusCreated {
		register.Body.Close()
		t.Fatalf("register status = %d, want %d", register.StatusCode, http.StatusCreated)
	}
	var registration api.RegisterClusterResponse
	decodeResponse(t, register, &registration)

	reRegister := func(token, existingAgentToken string) *http.Response {
		t.Helper()
		body := `{"name":"production"`
		if existingAgentToken != "" {
			body += `,"existingAgentToken":"` + existingAgentToken + `"`
		}
		body += `}`
		return agentRequest(t, httpServer, http.MethodPost, "/api/v1/agents/register", token, body)
	}

	// Re-registration without proof of the current agent token is rejected
	// instead of rotating the stored credential.
	for _, existing := range []string{"", "wrong-agent-token"} {
		response := reRegister(testRegistrationToken, existing)
		response.Body.Close()
		if response.StatusCode != http.StatusConflict {
			t.Fatalf("re-registration with existingAgentToken %q status = %d, want %d", existing, response.StatusCode, http.StatusConflict)
		}
	}

	// The legitimate re-register path presents the current token and rotates it.
	pending := reRegister(testRegistrationToken, registration.Token)
	if pending.StatusCode != http.StatusCreated {
		pending.Body.Close()
		t.Fatalf("pending re-registration status = %d, want %d", pending.StatusCode, http.StatusCreated)
	}
	var rotated api.RegisterClusterResponse
	decodeResponse(t, pending, &rotated)
	if rotated.ClusterID != registration.ClusterID {
		t.Fatalf("re-registration cluster ID = %q, want %q", rotated.ClusterID, registration.ClusterID)
	}
	if rotated.Token == "" || rotated.Token == registration.Token {
		t.Fatalf("re-registration token = %q, want a fresh token", rotated.Token)
	}

	approveAgent(t, httpServer, registration.ClusterID)

	// Once approved, re-registration keeps the approval and rotates again.
	approved := reRegister(testRegistrationToken, rotated.Token)
	if approved.StatusCode != http.StatusOK {
		approved.Body.Close()
		t.Fatalf("approved re-registration status = %d, want %d", approved.StatusCode, http.StatusOK)
	}
	var final api.RegisterClusterResponse
	decodeResponse(t, approved, &final)
	if final.Token == "" || final.Token == rotated.Token {
		t.Fatalf("approved re-registration token = %q, want a fresh token", final.Token)
	}

	// The rotation replaced the previous credential.
	heartbeatBody := `{"clusterId":"` + registration.ClusterID + `","nodeCount":1,"healthyNodes":1}`
	current := agentRequest(t, httpServer, http.MethodPost, "/api/v1/agents/heartbeat", final.Token, heartbeatBody)
	current.Body.Close()
	if current.StatusCode != http.StatusOK {
		t.Fatalf("heartbeat with rotated token status = %d, want %d", current.StatusCode, http.StatusOK)
	}
	stale := agentRequest(t, httpServer, http.MethodPost, "/api/v1/agents/heartbeat", rotated.Token, heartbeatBody)
	stale.Body.Close()
	if stale.StatusCode != http.StatusUnauthorized {
		t.Fatalf("heartbeat with previous token status = %d, want %d", stale.StatusCode, http.StatusUnauthorized)
	}
}

func TestStalenessMarksClusterUnreachable(t *testing.T) {
	_, srv, st := newAgentTestServer(t, time.Second)
	ctx := context.Background()
	cluster := types.Cluster{
		ID:            "stale-cluster",
		Name:          "stale",
		Health:        types.HealthHealthy,
		RegisteredAt:  time.Now().UTC().Add(-time.Hour),
		LastHeartbeat: time.Now().UTC().Add(-4 * time.Second),
		Labels:        map[string]string{},
	}
	if err := st.CreateCluster(ctx, cluster); err != nil {
		t.Fatalf("CreateCluster() error = %v", err)
	}

	srv.markStaleClusters(ctx, time.Now().UTC())
	got, err := st.GetCluster(ctx, cluster.ID)
	if err != nil {
		t.Fatalf("GetCluster() error = %v", err)
	}
	if got.Health != types.HealthUnreachable {
		t.Fatalf("cluster health = %q, want %q", got.Health, types.HealthUnreachable)
	}
}

func newAgentTestServer(t *testing.T, interval time.Duration) (*httptest.Server, *Server, store.Store) {
	return newAgentTestServerWithConfig(t, &config.Config{ListenAddr: ":0", HeartbeatInterval: interval, RegistrationToken: testRegistrationToken})
}

func newAgentTestServerWithConfig(t *testing.T, cfg *config.Config) (*httptest.Server, *Server, store.Store) {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "kfleet.db"))
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := New(cfg, logger, st)
	httpServer := httptest.NewServer(srv.httpServer.Handler)
	t.Cleanup(func() {
		httpServer.Close()
		if err := st.Close(); err != nil {
			t.Errorf("store.Close() error = %v", err)
		}
	})
	registerDefaultSession(httpServer, st, sessionCookieFor(t, st, types.RoleAdmin))
	return httpServer, srv, st
}

// agentRequest issues a request carrying an agent bearer token (when token
// is non-empty) and the server's default admin session cookie, so tests
// exercising the operator-facing pending/approve endpoints alongside the
// agent-facing register/heartbeat endpoints keep working from one helper.
// Tests exercising a specific human role should use agentRequestWithSession.
func agentRequest(t *testing.T, server *httptest.Server, method, path, token, body string) *http.Response {
	t.Helper()
	return agentRequestWithSession(t, server, method, path, token, defaultSessionFor(server), body)
}

func agentRequestWithSession(t *testing.T, server *httptest.Server, method, path, token, sessionCookie, body string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(method, server.URL+path, bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if sessionCookie != "" {
		req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionCookie})
		if mutationRequest(req) {
			req.Header.Set(csrfHeaderName, "1")
		}
	}
	response, err := server.Client().Do(req)
	if err != nil {
		t.Fatalf("request error = %v", err)
	}
	return response
}

func TestGenerateToken(t *testing.T) {
	raw, hash := generateToken()
	if len(raw) != 64 || len(hash) != 64 || hash != hashToken(raw) {
		t.Fatalf("generateToken() returned invalid raw/hash pair")
	}
	if _, err := json.Marshal(raw); err != nil {
		t.Fatalf("generated token is not JSON-safe: %v", err)
	}
}

func TestAgentReRegistrationRejectsCrossTenantToken(t *testing.T) {
	httpServer, _, st := newAgentTestServerWithConfig(t, &config.Config{
		ListenAddr:        ":0",
		RegistrationToken: testRegistrationToken,
	})

	registerInTenant := func(tenant string) api.RegisterClusterResponse {
		t.Helper()
		req, err := http.NewRequest(http.MethodPost, httpServer.URL+"/api/v1/agents/register", bytes.NewBufferString(`{"name":"shared"}`))
		if err != nil {
			t.Fatalf("http.NewRequest() error = %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+testRegistrationToken)
		req.Header.Set(tenantHeader, tenant)
		response, err := httpServer.Client().Do(req)
		if err != nil {
			t.Fatalf("register request error = %v", err)
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusCreated {
			t.Fatalf("register in %s status = %d, want %d", tenant, response.StatusCode, http.StatusCreated)
		}
		var registration api.RegisterClusterResponse
		decodeResponse(t, response, &registration)
		return registration
	}

	tenantA := registerInTenant("tenant-a")
	tenantB := registerInTenant("tenant-b")

	// Tenant A's credential must not re-register tenant B's same-named cluster.
	req, err := http.NewRequest(http.MethodPost, httpServer.URL+"/api/v1/agents/register",
		bytes.NewBufferString(`{"name":"shared","existingAgentToken":"`+tenantA.Token+`"}`))
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+testRegistrationToken)
	req.Header.Set(tenantHeader, "tenant-b")
	response, err := httpServer.Client().Do(req)
	if err != nil {
		t.Fatalf("cross-tenant re-registration request error = %v", err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusConflict {
		t.Fatalf("cross-tenant re-registration status = %d, want %d", response.StatusCode, http.StatusConflict)
	}

	// Tenant B's own credential still holds; tenant A's never did.
	if err := st.ApproveAgent(context.Background(), tenantB.ClusterID); err != nil {
		t.Fatalf("ApproveAgent() error = %v", err)
	}
	heartbeatBody := `{"clusterId":"` + tenantB.ClusterID + `","nodeCount":1,"healthyNodes":1}`
	own := agentRequest(t, httpServer, http.MethodPost, "/api/v1/agents/heartbeat", tenantB.Token, heartbeatBody)
	own.Body.Close()
	if own.StatusCode != http.StatusOK {
		t.Fatalf("heartbeat with tenant B token status = %d, want %d", own.StatusCode, http.StatusOK)
	}
	foreign := agentRequest(t, httpServer, http.MethodPost, "/api/v1/agents/heartbeat", tenantA.Token, heartbeatBody)
	foreign.Body.Close()
	if foreign.StatusCode != http.StatusUnauthorized {
		t.Fatalf("heartbeat with tenant A token status = %d, want %d", foreign.StatusCode, http.StatusUnauthorized)
	}
}

// erroringSettingsStore fails setting reads while delegating everything else,
// so tests can exercise the registration path's fail-closed store handling.
type erroringSettingsStore struct {
	store.Store
	getSettingErr error
}

func (s erroringSettingsStore) GetSetting(context.Context, string) (string, bool, error) {
	return "", false, s.getSettingErr
}

func TestAgentRegistrationFailsClosedOnStoreError(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "kfleet.db"))
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := New(&config.Config{ListenAddr: ":0"}, logger, erroringSettingsStore{Store: st, getSettingErr: errors.New("settings unavailable")})
	httpServer := httptest.NewServer(srv.httpServer.Handler)
	t.Cleanup(httpServer.Close)

	// A settings store failure is a server error, not "registration disabled".
	response := agentRequest(t, httpServer, http.MethodPost, "/api/v1/agents/register", "some-token", `{"name":"boom"}`)
	if response.StatusCode != http.StatusInternalServerError {
		response.Body.Close()
		t.Fatalf("registration on store error status = %d, want %d", response.StatusCode, http.StatusInternalServerError)
	}
	var errResponse api.ErrorResponse
	decodeResponse(t, response, &errResponse)
	if errResponse.Error != "failed to register agent" {
		t.Fatalf("error = %q, want a generic failure message", errResponse.Error)
	}
}
