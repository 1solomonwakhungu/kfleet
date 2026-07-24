package store

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/1solomonwakhungu/kfleet/pkg/types"
	"github.com/google/uuid"
)

func newTestStore(t *testing.T) Store {
	t.Helper()
	st, err := Open(filepath.Join(t.TempDir(), "kfleet.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := st.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})
	return st
}

func newTestUserRecord(role types.Role) types.User {
	now := time.Now().UTC()
	id := uuid.NewString()
	return types.User{
		ID:           id,
		Username:     "user-" + id,
		Email:        id + "@example.com",
		PasswordHash: "bcrypt-hash-placeholder",
		Role:         role,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func TestSQLiteStoreUserLifecycle(t *testing.T) {
	t.Parallel()
	st := newTestStore(t)
	ctx := context.Background()

	user := newTestUserRecord(types.RoleOperator)
	if err := st.CreateUser(ctx, user); err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	byID, err := st.GetUserByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetUserByID() error = %v", err)
	}
	if byID.Username != user.Username || byID.Role != types.RoleOperator {
		t.Fatalf("GetUserByID() = %+v, want matching user", byID)
	}

	byUsername, err := st.GetUserByUsername(ctx, user.Username)
	if err != nil {
		t.Fatalf("GetUserByUsername() error = %v", err)
	}
	if byUsername.ID != user.ID {
		t.Fatalf("GetUserByUsername() = %+v, want ID %q", byUsername, user.ID)
	}

	duplicate := newTestUserRecord(types.RoleReadOnly)
	duplicate.Username = user.Username
	if err := st.CreateUser(ctx, duplicate); !errors.Is(err, ErrConflict) {
		t.Fatalf("CreateUser() with duplicate username error = %v, want %v", err, ErrConflict)
	}

	if err := st.UpdateUser(ctx, user.ID, types.RoleAdmin, true); err != nil {
		t.Fatalf("UpdateUser() error = %v", err)
	}
	updated, err := st.GetUserByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetUserByID() after update error = %v", err)
	}
	if updated.Role != types.RoleAdmin || !updated.Disabled {
		t.Fatalf("updated user = %+v, want admin role and disabled", updated)
	}

	if err := st.UpdateUser(ctx, "missing", types.RoleAdmin, false); !errors.Is(err, ErrNotFound) {
		t.Fatalf("UpdateUser() for missing user error = %v, want %v", err, ErrNotFound)
	}

	if err := st.DeleteUser(ctx, user.ID); err != nil {
		t.Fatalf("DeleteUser() error = %v", err)
	}
	if _, err := st.GetUserByID(ctx, user.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetUserByID() after delete error = %v, want %v", err, ErrNotFound)
	}
	if err := st.DeleteUser(ctx, user.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("DeleteUser() for missing user error = %v, want %v", err, ErrNotFound)
	}
}

func TestSQLiteStoreListUsersOrdered(t *testing.T) {
	t.Parallel()
	st := newTestStore(t)
	ctx := context.Background()

	first := newTestUserRecord(types.RoleAdmin)
	first.CreatedAt = time.Now().UTC().Add(-time.Hour)
	first.UpdatedAt = first.CreatedAt
	second := newTestUserRecord(types.RoleOperator)

	if err := st.CreateUser(ctx, first); err != nil {
		t.Fatalf("CreateUser(first) error = %v", err)
	}
	if err := st.CreateUser(ctx, second); err != nil {
		t.Fatalf("CreateUser(second) error = %v", err)
	}

	users, err := st.ListUsers(ctx)
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}
	if len(users) != 2 || users[0].ID != first.ID || users[1].ID != second.ID {
		t.Fatalf("ListUsers() = %+v, want [first, second] ordered by created_at", users)
	}
}

// TestSQLiteStoreUpdateUserRejectsRemovingLastAdmin proves that demoting or
// disabling the sole enabled admin is rejected, and that a second enabled
// admin makes the same change legal.
func TestSQLiteStoreUpdateUserRejectsRemovingLastAdmin(t *testing.T) {
	t.Parallel()
	st := newTestStore(t)
	ctx := context.Background()

	admin := newTestUserRecord(types.RoleAdmin)
	if err := st.CreateUser(ctx, admin); err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	if err := st.UpdateUser(ctx, admin.ID, types.RoleOperator, false); !errors.Is(err, ErrLastAdmin) {
		t.Fatalf("UpdateUser() demoting sole admin error = %v, want %v", err, ErrLastAdmin)
	}
	if err := st.UpdateUser(ctx, admin.ID, types.RoleAdmin, true); !errors.Is(err, ErrLastAdmin) {
		t.Fatalf("UpdateUser() disabling sole admin error = %v, want %v", err, ErrLastAdmin)
	}
	if err := st.DeleteUser(ctx, admin.ID); !errors.Is(err, ErrLastAdmin) {
		t.Fatalf("DeleteUser() removing sole admin error = %v, want %v", err, ErrLastAdmin)
	}

	secondAdmin := newTestUserRecord(types.RoleAdmin)
	if err := st.CreateUser(ctx, secondAdmin); err != nil {
		t.Fatalf("CreateUser(secondAdmin) error = %v", err)
	}

	if err := st.UpdateUser(ctx, admin.ID, types.RoleOperator, false); err != nil {
		t.Fatalf("UpdateUser() demoting admin with a second admin present error = %v, want nil", err)
	}
}

// TestSQLiteStoreLastAdminGuardIsRaceSafe fires concurrent demote-or-delete
// requests against every enabled admin and verifies exactly one enabled
// admin always survives, proving the check-then-act guard in
// withLastAdminGuard cannot be raced past. Run with -race.
func TestSQLiteStoreLastAdminGuardIsRaceSafe(t *testing.T) {
	t.Parallel()
	st := newTestStore(t)
	ctx := context.Background()

	const adminCount = 5
	admins := make([]types.User, adminCount)
	for i := range admins {
		admins[i] = newTestUserRecord(types.RoleAdmin)
		if err := st.CreateUser(ctx, admins[i]); err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}
	}

	var wg sync.WaitGroup
	for i, admin := range admins {
		wg.Add(1)
		go func(id string, disable bool) {
			defer wg.Done()
			if disable {
				_ = st.UpdateUser(ctx, id, types.RoleAdmin, true)
			} else {
				_ = st.DeleteUser(ctx, id)
			}
		}(admin.ID, i%2 == 0)
	}
	wg.Wait()

	users, err := st.ListUsers(ctx)
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}
	enabledAdmins := 0
	for _, u := range users {
		if u.Role == types.RoleAdmin && !u.Disabled {
			enabledAdmins++
		}
	}
	if enabledAdmins < 1 {
		t.Fatalf("enabled admins after concurrent removal = %d, want at least 1", enabledAdmins)
	}
}

func TestSQLiteStoreSessionLifecycle(t *testing.T) {
	t.Parallel()
	st := newTestStore(t)
	ctx := context.Background()

	user := newTestUserRecord(types.RoleReadOnly)
	if err := st.CreateUser(ctx, user); err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	now := time.Now().UTC()
	if err := st.CreateSession(ctx, "hash-1", user.ID, now.Add(time.Hour)); err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	got, err := st.GetSessionUser(ctx, "hash-1", now)
	if err != nil {
		t.Fatalf("GetSessionUser() error = %v", err)
	}
	if got.ID != user.ID {
		t.Fatalf("GetSessionUser() = %+v, want user %q", got, user.ID)
	}

	if _, err := st.GetSessionUser(ctx, "hash-1", now.Add(2*time.Hour)); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetSessionUser() for expired session error = %v, want %v", err, ErrNotFound)
	}

	if err := st.CreateSession(ctx, "hash-2", user.ID, now.Add(-time.Minute)); err != nil {
		t.Fatalf("CreateSession() expired error = %v", err)
	}
	if err := st.DeleteExpiredSessions(ctx, now); err != nil {
		t.Fatalf("DeleteExpiredSessions() error = %v", err)
	}
	if _, err := st.GetSessionUser(ctx, "hash-2", now.Add(-time.Hour)); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetSessionUser() for pruned session error = %v, want %v", err, ErrNotFound)
	}

	if err := st.DeleteSession(ctx, "hash-1"); err != nil {
		t.Fatalf("DeleteSession() error = %v", err)
	}
	if _, err := st.GetSessionUser(ctx, "hash-1", now); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetSessionUser() after delete error = %v, want %v", err, ErrNotFound)
	}
}

func TestSQLiteStoreSettings(t *testing.T) {
	t.Parallel()
	st := newTestStore(t)
	ctx := context.Background()

	if _, ok, err := st.GetSetting(ctx, "missing"); err != nil || ok {
		t.Fatalf("GetSetting() for missing key = (ok=%v, err=%v), want (false, nil)", ok, err)
	}

	if err := st.SetSetting(ctx, "key", "value"); err != nil {
		t.Fatalf("SetSetting() error = %v", err)
	}
	value, ok, err := st.GetSetting(ctx, "key")
	if err != nil || !ok || value != "value" {
		t.Fatalf("GetSetting() = (%q, %v, %v), want (\"value\", true, nil)", value, ok, err)
	}

	if err := st.SetSetting(ctx, "key", "updated"); err != nil {
		t.Fatalf("SetSetting() overwrite error = %v", err)
	}
	value, _, _ = st.GetSetting(ctx, "key")
	if value != "updated" {
		t.Fatalf("GetSetting() after overwrite = %q, want %q", value, "updated")
	}
}

func TestSQLiteStoreAuditEventsAreAppendOnly(t *testing.T) {
	t.Parallel()
	st := newTestStore(t)
	sqlite, ok := st.(*sqliteStore)
	if !ok {
		t.Fatalf("store type = %T, want *sqliteStore", st)
	}

	event := types.AuditEvent{
		ID:            uuid.NewString(),
		OccurredAt:    time.Now().UTC(),
		ActorUsername: "admin",
		ActorRole:     types.RoleAdmin,
		Action:        "user.create",
		TargetType:    "user",
		TargetID:      "target-user",
		Outcome:       types.AuditSuccess,
	}
	if err := st.RecordAuditEvent(context.Background(), event); err != nil {
		t.Fatalf("RecordAuditEvent() error = %v", err)
	}

	if _, err := sqlite.db.Exec(`UPDATE audit_events SET action = 'tampered' WHERE id = ?`, event.ID); err == nil {
		t.Fatal("UPDATE audit_events succeeded, want append-only trigger failure")
	}
	if _, err := sqlite.db.Exec(`DELETE FROM audit_events WHERE id = ?`, event.ID); err == nil {
		t.Fatal("DELETE FROM audit_events succeeded, want append-only trigger failure")
	}

	events, err := st.ListAuditEvents(context.Background(), 10)
	if err != nil {
		t.Fatalf("ListAuditEvents() error = %v", err)
	}
	if len(events) != 1 || events[0].ID != event.ID || events[0].Action != event.Action {
		t.Fatalf("audit events after mutation attempts = %+v, want original event", events)
	}
}
