package server

import (
	"context"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/1solomonwakhungu/kfleet/internal/auth"
	"github.com/1solomonwakhungu/kfleet/internal/store"
	"github.com/1solomonwakhungu/kfleet/pkg/types"
	"github.com/google/uuid"
)

// testUserPassword is used only for users created directly by tests; it
// never needs to be memorable or environment-specific.
const testUserPassword = "Sup3rSecretPassw0rd!"

// createTestUser persists a user with the given role and returns it,
// including its bcrypt hash. Tests use this to exercise RBAC without going
// through the HTTP login flow.
func createTestUser(t *testing.T, st store.Store, role types.Role) types.User {
	t.Helper()
	passwordHash, err := auth.HashPassword(testUserPassword)
	if err != nil {
		t.Fatalf("auth.HashPassword() error = %v", err)
	}
	now := time.Now().UTC()
	user := types.User{
		ID:           uuid.NewString(),
		Username:     "test-" + string(role) + "-" + uuid.NewString(),
		Email:        "test-" + uuid.NewString() + "@example.com",
		PasswordHash: passwordHash,
		Role:         role,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := st.CreateUser(context.Background(), user); err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	return user
}

// createTestSession creates a valid session for user and returns the raw
// session token to send as the kfleet_session cookie value.
func createTestSession(t *testing.T, st store.Store, user types.User) string {
	t.Helper()
	raw, hash, err := auth.NewSessionToken()
	if err != nil {
		t.Fatalf("auth.NewSessionToken() error = %v", err)
	}
	if err := st.CreateSession(context.Background(), hash, user.ID, time.Now().UTC().Add(time.Hour)); err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	return raw
}

// sessionCookieFor is a convenience wrapper that creates a user with role
// and an active session, returning the raw session cookie value.
func sessionCookieFor(t *testing.T, st store.Store, role types.Role) string {
	t.Helper()
	return createTestSession(t, st, createTestUser(t, st, role))
}

// testServerSessions associates each test httptest.Server with a default
// admin session cookie, so existing request()/agentRequest() call sites
// across the test suite keep working against RBAC-protected routes without
// every call site individually threading a cookie through. Tests that need
// to exercise a specific role use requestWithSession/agentRequestWithSession
// with an explicit cookie instead. testServerStores lets those same tests
// recover the underlying store.Store to create additional non-admin
// sessions (see serverStoreForTest).
var (
	testServerSessionsMu sync.Mutex
	testServerSessions   = map[*httptest.Server]string{}
	testServerStores     = map[*httptest.Server]store.Store{}
)

func registerDefaultSession(server *httptest.Server, st store.Store, cookie string) {
	testServerSessionsMu.Lock()
	defer testServerSessionsMu.Unlock()
	testServerSessions[server] = cookie
	testServerStores[server] = st
}

func defaultSessionFor(server *httptest.Server) string {
	testServerSessionsMu.Lock()
	defer testServerSessionsMu.Unlock()
	return testServerSessions[server]
}

func serverStoreForTest(t *testing.T, server *httptest.Server) store.Store {
	t.Helper()
	testServerSessionsMu.Lock()
	defer testServerSessionsMu.Unlock()
	st, ok := testServerStores[server]
	if !ok {
		t.Fatalf("no store registered for test server")
	}
	return st
}
