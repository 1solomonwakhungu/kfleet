package server

import (
	"context"
	"net/http"
	"testing"

	"github.com/1solomonwakhungu/kfleet/pkg/types"
)

// TestRBACUnauthenticatedRequestsAreRejected proves every protected route
// rejects a request carrying no session cookie at all.
func TestRBACUnauthenticatedRequestsAreRejected(t *testing.T) {
	server := newTestHTTPServer(t)

	routes := []struct {
		method, path string
	}{
		{http.MethodGet, "/api/v1/clusters"},
		{http.MethodPost, "/api/v1/clusters/register"},
		{http.MethodDelete, "/api/v1/clusters/does-not-matter"},
		{http.MethodGet, "/api/v1/agents/pending"},
		{http.MethodPost, "/api/v1/agents/does-not-matter/approve"},
		{http.MethodGet, "/api/v1/users"},
		{http.MethodPost, "/api/v1/users"},
		{http.MethodGet, "/api/v1/audit"},
		{http.MethodPost, "/api/v1/admin/registration-token/rotate"},
		{http.MethodGet, "/api/v1/auth/me"},
	}
	for _, route := range routes {
		resp := requestWithSession(t, server, route.method, route.path, "", "")
		resp.Body.Close()
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("%s %s unauthenticated status = %d, want %d", route.method, route.path, resp.StatusCode, http.StatusUnauthorized)
		}
	}
}

// TestRBACReadOnlyRoleCanReadButNotMutate proves a read_only session can
// reach every read route but is forbidden from every mutation and from
// admin-only routes.
func TestRBACReadOnlyRoleCanReadButNotMutate(t *testing.T) {
	server := newTestHTTPServer(t)
	st := serverStoreForTest(t, server)
	readOnly := sessionCookieFor(t, st, types.RoleReadOnly)

	allowed := []struct {
		method, path string
	}{
		{http.MethodGet, "/api/v1/clusters"},
		{http.MethodGet, "/api/v1/agents/pending"},
		{http.MethodGet, "/api/v1/auth/me"},
	}
	for _, route := range allowed {
		resp := requestWithSession(t, server, route.method, route.path, readOnly, "")
		resp.Body.Close()
		if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
			t.Errorf("%s %s as read_only status = %d, want a non-auth-error status", route.method, route.path, resp.StatusCode)
		}
	}

	denied := []struct {
		method, path, body string
	}{
		{http.MethodPost, "/api/v1/clusters/register", `{"name":"blocked"}`},
		{http.MethodDelete, "/api/v1/clusters/does-not-matter", ""},
		{http.MethodPost, "/api/v1/agents/does-not-matter/approve", ""},
		{http.MethodGet, "/api/v1/users", ""},
		{http.MethodPost, "/api/v1/users", `{"username":"x","email":"x@example.com","password":"hunter2-hunter2","role":"operator"}`},
		{http.MethodGet, "/api/v1/audit", ""},
		{http.MethodPost, "/api/v1/admin/registration-token/rotate", ""},
	}
	for _, route := range denied {
		resp := requestWithSession(t, server, route.method, route.path, readOnly, route.body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("%s %s as read_only status = %d, want %d", route.method, route.path, resp.StatusCode, http.StatusForbidden)
		}
	}
}

// TestRBACOperatorRoleCanMutateFleetButNotUsers proves an operator session
// can perform fleet operations (register/delete clusters, approve agents)
// but is forbidden from user management, the audit log, and admin routes.
func TestRBACOperatorRoleCanMutateFleetButNotUsers(t *testing.T) {
	server := newTestHTTPServer(t)
	st := serverStoreForTest(t, server)
	operator := sessionCookieFor(t, st, types.RoleOperator)

	registerResp := requestWithSession(t, server, http.MethodPost, "/api/v1/clusters/register", operator, `{"name":"operator-cluster"}`)
	if registerResp.StatusCode != http.StatusCreated {
		t.Fatalf("register cluster as operator status = %d, want %d", registerResp.StatusCode, http.StatusCreated)
	}
	var registration struct {
		ClusterID string `json:"clusterId"`
	}
	decodeResponse(t, registerResp, &registration)

	deleteResp := requestWithSession(t, server, http.MethodDelete, "/api/v1/clusters/"+registration.ClusterID, operator, "")
	deleteResp.Body.Close()
	if deleteResp.StatusCode != http.StatusNoContent {
		t.Errorf("delete cluster as operator status = %d, want %d", deleteResp.StatusCode, http.StatusNoContent)
	}

	denied := []struct {
		method, path, body string
	}{
		{http.MethodGet, "/api/v1/users", ""},
		{http.MethodPost, "/api/v1/users", `{"username":"x","email":"x@example.com","password":"hunter2-hunter2","role":"operator"}`},
		{http.MethodGet, "/api/v1/audit", ""},
		{http.MethodPost, "/api/v1/admin/registration-token/rotate", ""},
	}
	for _, route := range denied {
		resp := requestWithSession(t, server, route.method, route.path, operator, route.body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("%s %s as operator status = %d, want %d", route.method, route.path, resp.StatusCode, http.StatusForbidden)
		}
	}
}

// TestRBACAdminRoleCanManageUsersAndAudit proves an admin session can reach
// user management, the audit log, and admin routes.
func TestRBACAdminRoleCanManageUsersAndAudit(t *testing.T) {
	server := newTestHTTPServer(t)
	admin := defaultSessionFor(server)

	listUsersResp := requestWithSession(t, server, http.MethodGet, "/api/v1/users", admin, "")
	listUsersResp.Body.Close()
	if listUsersResp.StatusCode != http.StatusOK {
		t.Errorf("list users as admin status = %d, want %d", listUsersResp.StatusCode, http.StatusOK)
	}

	createResp := requestWithSession(t, server, http.MethodPost, "/api/v1/users", admin,
		`{"username":"new-operator","email":"new-operator@example.com","password":"hunter2-hunter2","role":"operator"}`)
	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("create user as admin status = %d, want %d", createResp.StatusCode, http.StatusCreated)
	}
	createResp.Body.Close()

	auditResp := requestWithSession(t, server, http.MethodGet, "/api/v1/audit", admin, "")
	auditResp.Body.Close()
	if auditResp.StatusCode != http.StatusOK {
		t.Errorf("list audit as admin status = %d, want %d", auditResp.StatusCode, http.StatusOK)
	}

	rotateResp := requestWithSession(t, server, http.MethodPost, "/api/v1/admin/registration-token/rotate", admin, "")
	rotateResp.Body.Close()
	if rotateResp.StatusCode != http.StatusOK {
		t.Errorf("rotate registration token as admin status = %d, want %d", rotateResp.StatusCode, http.StatusOK)
	}
}

// TestRBACDisabledUserSessionIsRejected proves a session for a disabled
// account is rejected even though the session row itself has not expired.
func TestRBACDisabledUserSessionIsRejected(t *testing.T) {
	server := newTestHTTPServer(t)
	st := serverStoreForTest(t, server)
	user := createTestUser(t, st, types.RoleOperator)
	cookie := createTestSession(t, st, user)

	if err := st.UpdateUser(context.Background(), user.ID, types.RoleOperator, true); err != nil {
		t.Fatalf("UpdateUser() disabling account error = %v", err)
	}

	resp := requestWithSession(t, server, http.MethodGet, "/api/v1/clusters", cookie, "")
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("disabled user status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestRBACMutationRequiresCSRFHeader(t *testing.T) {
	server := newTestHTTPServer(t)
	session := defaultSessionFor(server)

	req, err := http.NewRequest(http.MethodPost, server.URL+"/api/v1/clusters/register", nil)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: session})
	resp, err := server.Client().Do(req)
	if err != nil {
		t.Fatalf("request error = %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("mutation without CSRF header status = %d, want %d", resp.StatusCode, http.StatusForbidden)
	}

	events, err := serverStoreForTest(t, server).ListAuditEvents(context.Background(), 10)
	if err != nil {
		t.Fatalf("ListAuditEvents() error = %v", err)
	}
	if len(events) == 0 || events[0].Action != "authorization.csrf_denied" {
		t.Fatalf("audit events = %+v, want authorization.csrf_denied event", events)
	}
}
