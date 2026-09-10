package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/1solomonwakhungu/kfleet/internal/auth"
	"github.com/1solomonwakhungu/kfleet/internal/store"
	"github.com/1solomonwakhungu/kfleet/pkg/api"
	"github.com/1solomonwakhungu/kfleet/pkg/types"
)

func TestHandleCreateUserValidatesInput(t *testing.T) {
	server := newTestHTTPServer(t)
	admin := defaultSessionFor(server)

	cases := []struct {
		name string
		body string
	}{
		{"missing username", `{"username":"","email":"a@example.com","password":"hunter2-hunter2","role":"operator"}`},
		{"missing email", `{"username":"a","email":"","password":"hunter2-hunter2","role":"operator"}`},
		{"missing password", `{"username":"a","email":"a@example.com","password":"","role":"operator"}`},
		{"short password", `{"username":"a","email":"a@example.com","password":"short","role":"operator"}`},
		{"invalid role", `{"username":"a","email":"a@example.com","password":"hunter2-hunter2","role":"superadmin"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := requestWithSession(t, server, http.MethodPost, "/api/v1/users", admin, tc.body)
			resp.Body.Close()
			if resp.StatusCode != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
			}
		})
	}
}

func TestHandleCreateUserRejectsDuplicateUsername(t *testing.T) {
	server := newTestHTTPServer(t)
	admin := defaultSessionFor(server)

	body := `{"username":"dupe","email":"dupe@example.com","password":"hunter2-hunter2","role":"read_only"}`
	first := requestWithSession(t, server, http.MethodPost, "/api/v1/users", admin, body)
	first.Body.Close()
	if first.StatusCode != http.StatusCreated {
		t.Fatalf("first create status = %d, want %d", first.StatusCode, http.StatusCreated)
	}

	second := requestWithSession(t, server, http.MethodPost, "/api/v1/users", admin,
		`{"username":"dupe","email":"other@example.com","password":"hunter2-hunter2","role":"read_only"}`)
	second.Body.Close()
	if second.StatusCode != http.StatusConflict {
		t.Errorf("duplicate username status = %d, want %d", second.StatusCode, http.StatusConflict)
	}
}

func TestHandleCreateUserNeverReturnsPasswordHash(t *testing.T) {
	server := newTestHTTPServer(t)
	admin := defaultSessionFor(server)

	resp := requestWithSession(t, server, http.MethodPost, "/api/v1/users", admin,
		`{"username":"nohash","email":"nohash@example.com","password":"hunter2-hunter2","role":"read_only"}`)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create status = %d, want %d", resp.StatusCode, http.StatusCreated)
	}
	var got api.UserResponse
	decodeResponse(t, resp, &got)
	if got.Username != "nohash" || got.Role != types.RoleReadOnly {
		t.Fatalf("created user = %+v, want matching username/role", got)
	}
}

func TestHandleUpdateUserChangesRoleAndDisabled(t *testing.T) {
	server := newTestHTTPServer(t)
	st := serverStoreForTest(t, server)
	admin := defaultSessionFor(server)
	target := createTestUser(t, st, types.RoleReadOnly)

	resp := requestWithSession(t, server, http.MethodPatch, "/api/v1/users/"+target.ID, admin,
		`{"role":"operator","disabled":true}`)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("update status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var got api.UserResponse
	decodeResponse(t, resp, &got)
	if got.Role != types.RoleOperator || !got.Disabled {
		t.Fatalf("updated user = %+v, want operator role and disabled", got)
	}
}

func TestHandleUpdateUserRejectsUnknownUserAndInvalidRole(t *testing.T) {
	server := newTestHTTPServer(t)
	admin := defaultSessionFor(server)

	notFound := requestWithSession(t, server, http.MethodPatch, "/api/v1/users/does-not-exist", admin, `{"role":"operator","disabled":false}`)
	notFound.Body.Close()
	if notFound.StatusCode != http.StatusNotFound {
		t.Errorf("update unknown user status = %d, want %d", notFound.StatusCode, http.StatusNotFound)
	}

	invalidRole := requestWithSession(t, server, http.MethodPatch, "/api/v1/users/does-not-exist", admin, `{"role":"bogus","disabled":false}`)
	invalidRole.Body.Close()
	if invalidRole.StatusCode != http.StatusBadRequest {
		t.Errorf("update with invalid role status = %d, want %d", invalidRole.StatusCode, http.StatusBadRequest)
	}
}

func TestHandleUpdateUserRejectsRemovingLastAdmin(t *testing.T) {
	server := newTestHTTPServer(t)
	admin := defaultSessionFor(server)
	soleAdminID := defaultAdminIDForTest(t, server)

	resp := requestWithSession(t, server, http.MethodPatch, "/api/v1/users/"+soleAdminID, admin, `{"role":"operator","disabled":false}`)
	resp.Body.Close()
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("demote sole admin status = %d, want %d", resp.StatusCode, http.StatusConflict)
	}
}

func TestHandleDeleteUserRemovesAccount(t *testing.T) {
	server := newTestHTTPServer(t)
	st := serverStoreForTest(t, server)
	admin := defaultSessionFor(server)
	target := createTestUser(t, st, types.RoleOperator)

	resp := requestWithSession(t, server, http.MethodDelete, "/api/v1/users/"+target.ID, admin, "")
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete status = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}

	getResp := requestWithSession(t, server, http.MethodPatch, "/api/v1/users/"+target.ID, admin, `{"role":"operator","disabled":false}`)
	getResp.Body.Close()
	if getResp.StatusCode != http.StatusNotFound {
		t.Errorf("update deleted user status = %d, want %d", getResp.StatusCode, http.StatusNotFound)
	}
}

func TestHandleDeleteUserRejectsSelfDeletion(t *testing.T) {
	server := newTestHTTPServer(t)
	admin := defaultSessionFor(server)
	adminID := defaultAdminIDForTest(t, server)

	resp := requestWithSession(t, server, http.MethodDelete, "/api/v1/users/"+adminID, admin, "")
	resp.Body.Close()
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("self-delete status = %d, want %d", resp.StatusCode, http.StatusConflict)
	}
}

// TestHandleDeleteUserConcurrentMutualDeletionLeavesOneAdmin has two admins
// (the default admin and a second one) each try, at the same time, to
// delete the other. The self-delete guard rules out either deleting their
// own account, so both requests target real admin accounts; the store's
// last-admin guard must ensure exactly one of the two deletions succeeds
// and the loser gets either a 409 from the last-admin guard or a 401 when
// its session is deleted before authentication completes. Run with -race.
func TestHandleDeleteUserConcurrentMutualDeletionLeavesOneAdmin(t *testing.T) {
	server := newTestHTTPServer(t)
	st := serverStoreForTest(t, server)
	adminA := defaultSessionFor(server)
	adminAID := defaultAdminIDForTest(t, server)
	adminBUser := createTestUser(t, st, types.RoleAdmin)
	adminB := createTestSession(t, st, adminBUser)

	results := make(chan int, 2)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		resp := requestWithSession(t, server, http.MethodDelete, "/api/v1/users/"+adminBUser.ID, adminA, "")
		resp.Body.Close()
		results <- resp.StatusCode
	}()
	go func() {
		defer wg.Done()
		resp := requestWithSession(t, server, http.MethodDelete, "/api/v1/users/"+adminAID, adminB, "")
		resp.Body.Close()
		results <- resp.StatusCode
	}()
	wg.Wait()
	close(results)

	var successes, rejected int
	for status := range results {
		switch status {
		case http.StatusNoContent:
			successes++
		case http.StatusConflict, http.StatusUnauthorized:
			rejected++
		default:
			t.Errorf("concurrent mutual admin deletion returned unexpected status %d", status)
		}
	}
	if successes != 1 || rejected != 1 {
		t.Fatalf("successes=%d rejected=%d, want exactly one of each", successes, rejected)
	}

	users, err := st.ListUsers(context.Background())
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}
	enabledAdmins := 0
	for _, u := range users {
		if u.Role == types.RoleAdmin && !u.Disabled {
			enabledAdmins++
		}
	}
	if enabledAdmins != 1 {
		t.Fatalf("enabled admins after concurrent mutual deletion = %d, want 1", enabledAdmins)
	}
}

// defaultAdminIDForTest returns the user ID behind server's default admin
// session cookie by calling GET /api/v1/auth/me.
func defaultAdminIDForTest(t *testing.T, server *httptest.Server) string {
	t.Helper()
	resp := requestWithSession(t, server, http.MethodGet, "/api/v1/auth/me", defaultSessionFor(server), "")
	var got api.UserResponse
	decodeResponse(t, resp, &got)
	if got.ID == "" {
		t.Fatal("defaultAdminIDForTest: /api/v1/auth/me returned no user ID")
	}
	return got.ID
}

func TestHandleUpdateUserPasswordSelfChange(t *testing.T) {
	server := newTestHTTPServer(t)
	st := serverStoreForTest(t, server)
	adminID := defaultAdminIDForTest(t, server)

	resp := requestWithSession(t, server, http.MethodPatch, "/api/v1/users/"+adminID+"/password", defaultSessionFor(server),
		`{"currentPassword":"`+testUserPassword+`","newPassword":"brand-new-password-42"}`)
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("self-change status = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}

	target, err := st.GetUserByID(context.Background(), adminID)
	if err != nil {
		t.Fatalf("GetUserByID() error = %v", err)
	}
	if err := auth.VerifyPassword(target.PasswordHash, "brand-new-password-42"); err != nil {
		t.Fatalf("new password does not verify: %v", err)
	}
	if err := auth.VerifyPassword(target.PasswordHash, testUserPassword); !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("old password still verifies after change: %v", err)
	}

	// Every session for the account is invalidated, including the one used
	// for the change itself.
	after := requestWithSession(t, server, http.MethodGet, "/api/v1/auth/me", defaultSessionFor(server), "")
	after.Body.Close()
	if after.StatusCode != http.StatusUnauthorized {
		t.Errorf("request with pre-change session status = %d, want %d", after.StatusCode, http.StatusUnauthorized)
	}
}

func TestHandleUpdateUserPasswordSelfChangeRejectsWrongCurrentPassword(t *testing.T) {
	server := newTestHTTPServer(t)
	adminID := defaultAdminIDForTest(t, server)

	resp := requestWithSession(t, server, http.MethodPatch, "/api/v1/users/"+adminID+"/password", defaultSessionFor(server),
		`{"currentPassword":"not-my-password","newPassword":"brand-new-password-42"}`)
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("wrong current password status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}

	missing := requestWithSession(t, server, http.MethodPatch, "/api/v1/users/"+adminID+"/password", defaultSessionFor(server),
		`{"newPassword":"brand-new-password-42"}`)
	missing.Body.Close()
	if missing.StatusCode != http.StatusBadRequest {
		t.Errorf("missing current password status = %d, want %d", missing.StatusCode, http.StatusBadRequest)
	}
}

func TestHandleUpdateUserPasswordAdminResetsAnotherUser(t *testing.T) {
	server := newTestHTTPServer(t)
	st := serverStoreForTest(t, server)
	admin := defaultSessionFor(server)
	target := createTestUser(t, st, types.RoleReadOnly)
	targetSession := createTestSession(t, st, target)

	resp := requestWithSession(t, server, http.MethodPatch, "/api/v1/users/"+target.ID+"/password", admin,
		`{"newPassword":"reset-password-99"}`)
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("admin reset status = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}

	updated, err := st.GetUserByID(context.Background(), target.ID)
	if err != nil {
		t.Fatalf("GetUserByID() error = %v", err)
	}
	if err := auth.VerifyPassword(updated.PasswordHash, "reset-password-99"); err != nil {
		t.Fatalf("new password does not verify: %v", err)
	}
	if err := auth.VerifyPassword(updated.PasswordHash, testUserPassword); !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("old password still verifies after reset: %v", err)
	}
	if _, err := st.GetSessionUser(context.Background(), auth.HashToken(targetSession), time.Now().UTC()); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("target session survived password reset: err = %v, want %v", err, store.ErrNotFound)
	}
}

func TestHandleUpdateUserPasswordNonAdminCannotChangeOthers(t *testing.T) {
	server := newTestHTTPServer(t)
	st := serverStoreForTest(t, server)
	target := createTestUser(t, st, types.RoleOperator)
	operator := createTestSession(t, st, createTestUser(t, st, types.RoleReadOnly))

	resp := requestWithSession(t, server, http.MethodPatch, "/api/v1/users/"+target.ID+"/password", operator,
		`{"newPassword":"hijacked-password-1"}`)
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("non-admin changing another user status = %d, want %d", resp.StatusCode, http.StatusForbidden)
	}

	// The route itself stays open to non-admins for self-service changes.
	me := requestWithSession(t, server, http.MethodGet, "/api/v1/auth/me", operator, "")
	var meResp api.UserResponse
	decodeResponse(t, me, &meResp)
	self := requestWithSession(t, server, http.MethodPatch, "/api/v1/users/"+meResp.ID+"/password", operator,
		`{"currentPassword":"`+testUserPassword+`","newPassword":"own-new-password-77"}`)
	self.Body.Close()
	if self.StatusCode != http.StatusNoContent {
		t.Errorf("non-admin self-change status = %d, want %d", self.StatusCode, http.StatusNoContent)
	}
}

func TestHandleUpdateUserPasswordValidatesInput(t *testing.T) {
	server := newTestHTTPServer(t)
	st := serverStoreForTest(t, server)
	admin := defaultSessionFor(server)
	target := createTestUser(t, st, types.RoleReadOnly)

	cases := []struct {
		name string
		body string
		want int
	}{
		{"short password", `{"newPassword":"short"}`, http.StatusBadRequest},
		{"missing password", `{"currentPassword":"x"}`, http.StatusBadRequest},
		{"unknown user", `{"newPassword":"reset-password-99"}`, http.StatusNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := "/api/v1/users/" + target.ID + "/password"
			if tc.want == http.StatusNotFound {
				path = "/api/v1/users/does-not-exist/password"
			}
			resp := requestWithSession(t, server, http.MethodPatch, path, admin, tc.body)
			resp.Body.Close()
			if resp.StatusCode != tc.want {
				t.Errorf("status = %d, want %d", resp.StatusCode, tc.want)
			}
		})
	}
}
