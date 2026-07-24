package server

import (
	"context"
	"net/http"
	"testing"

	"github.com/1solomonwakhungu/kfleet/internal/auth"
	"github.com/1solomonwakhungu/kfleet/pkg/api"
	"github.com/1solomonwakhungu/kfleet/pkg/types"
)

func TestHandleLoginSuccessSetsSessionCookie(t *testing.T) {
	server := newTestHTTPServer(t)
	st := serverStoreForTest(t, server)
	user := createTestUser(t, st, types.RoleOperator)

	resp := requestWithSession(t, server, http.MethodPost, "/api/v1/auth/login", "",
		`{"username":"`+user.Username+`","password":"`+testUserPassword+`"}`)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var got api.UserResponse
	decodeResponse(t, resp, &got)
	if got.Username != user.Username || got.Role != types.RoleOperator {
		t.Fatalf("login response = %+v, want matching user", got)
	}

	found := false
	for _, cookie := range resp.Cookies() {
		if cookie.Name == sessionCookieName {
			found = true
			if cookie.Value == "" {
				t.Fatal("session cookie value is empty")
			}
			if !cookie.HttpOnly {
				t.Fatal("session cookie is not HttpOnly")
			}
			if cookie.SameSite != http.SameSiteStrictMode {
				t.Fatalf("session cookie SameSite = %v, want Strict", cookie.SameSite)
			}
		}
	}
	if !found {
		t.Fatal("login response did not set a session cookie")
	}
}

func TestHandleLoginRejectsWrongPasswordAndUnknownUser(t *testing.T) {
	server := newTestHTTPServer(t)
	st := serverStoreForTest(t, server)
	user := createTestUser(t, st, types.RoleReadOnly)

	wrongPassword := requestWithSession(t, server, http.MethodPost, "/api/v1/auth/login", "",
		`{"username":"`+user.Username+`","password":"wrong-password-entirely"}`)
	wrongPassword.Body.Close()
	if wrongPassword.StatusCode != http.StatusUnauthorized {
		t.Errorf("login with wrong password status = %d, want %d", wrongPassword.StatusCode, http.StatusUnauthorized)
	}

	unknownUser := requestWithSession(t, server, http.MethodPost, "/api/v1/auth/login", "",
		`{"username":"nobody-registered","password":"whatever-password"}`)
	unknownUser.Body.Close()
	if unknownUser.StatusCode != http.StatusUnauthorized {
		t.Errorf("login with unknown user status = %d, want %d", unknownUser.StatusCode, http.StatusUnauthorized)
	}

	missingFields := requestWithSession(t, server, http.MethodPost, "/api/v1/auth/login", "", `{"username":"","password":""}`)
	missingFields.Body.Close()
	if missingFields.StatusCode != http.StatusBadRequest {
		t.Errorf("login with empty fields status = %d, want %d", missingFields.StatusCode, http.StatusBadRequest)
	}
}

func TestHandleLoginRejectsDisabledAccount(t *testing.T) {
	server := newTestHTTPServer(t)
	st := serverStoreForTest(t, server)
	user := createTestUser(t, st, types.RoleOperator)
	if err := st.UpdateUser(context.Background(), user.ID, types.RoleOperator, true); err != nil {
		t.Fatalf("UpdateUser() disabling account error = %v", err)
	}

	resp := requestWithSession(t, server, http.MethodPost, "/api/v1/auth/login", "",
		`{"username":"`+user.Username+`","password":"`+testUserPassword+`"}`)
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("login for disabled account status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestHandleLoginRecordsAuditEventsOnSuccessAndFailure(t *testing.T) {
	server := newTestHTTPServer(t)
	st := serverStoreForTest(t, server)
	user := createTestUser(t, st, types.RoleAdmin)

	ok := requestWithSession(t, server, http.MethodPost, "/api/v1/auth/login", "",
		`{"username":"`+user.Username+`","password":"`+testUserPassword+`"}`)
	ok.Body.Close()
	bad := requestWithSession(t, server, http.MethodPost, "/api/v1/auth/login", "",
		`{"username":"`+user.Username+`","password":"wrong"}`)
	bad.Body.Close()

	events, err := st.ListAuditEvents(context.Background(), 100)
	if err != nil {
		t.Fatalf("ListAuditEvents() error = %v", err)
	}
	var successCount, failureCount int
	for _, event := range events {
		if event.Action != "login" || event.ActorUsername != user.Username {
			continue
		}
		switch event.Outcome {
		case types.AuditSuccess:
			successCount++
		case types.AuditFailure:
			failureCount++
		}
	}
	if successCount == 0 {
		t.Error("no successful login audit event recorded")
	}
	if failureCount == 0 {
		t.Error("no failed login audit event recorded")
	}
}

func TestHandleLogoutClearsSessionCookie(t *testing.T) {
	server := newTestHTTPServer(t)
	st := serverStoreForTest(t, server)
	user := createTestUser(t, st, types.RoleReadOnly)
	cookie := createTestSession(t, st, user)

	logoutResp := requestWithSession(t, server, http.MethodPost, "/api/v1/auth/logout", cookie, "")
	if logoutResp.StatusCode != http.StatusNoContent {
		t.Fatalf("logout status = %d, want %d", logoutResp.StatusCode, http.StatusNoContent)
	}
	logoutResp.Body.Close()

	afterLogout := requestWithSession(t, server, http.MethodGet, "/api/v1/auth/me", cookie, "")
	afterLogout.Body.Close()
	if afterLogout.StatusCode != http.StatusUnauthorized {
		t.Errorf("me after logout status = %d, want %d", afterLogout.StatusCode, http.StatusUnauthorized)
	}
}

func TestHandleMeReturnsAuthenticatedUser(t *testing.T) {
	server := newTestHTTPServer(t)
	st := serverStoreForTest(t, server)
	user := createTestUser(t, st, types.RoleOperator)
	cookie := createTestSession(t, st, user)

	resp := requestWithSession(t, server, http.MethodGet, "/api/v1/auth/me", cookie, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("me status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var got api.UserResponse
	decodeResponse(t, resp, &got)
	if got.ID != user.ID || got.Role != types.RoleOperator {
		t.Fatalf("me response = %+v, want user %q", got, user.ID)
	}
}

func TestNewSessionTokenNeverAppearsInStoredHash(t *testing.T) {
	raw, hash, err := auth.NewSessionToken()
	if err != nil {
		t.Fatalf("NewSessionToken() error = %v", err)
	}
	if hash == raw {
		t.Fatal("session token hash equals raw token; raw token must never be persisted verbatim")
	}
}
