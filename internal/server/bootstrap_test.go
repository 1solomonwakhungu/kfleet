package server

import (
	"context"
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"testing"

	"github.com/1solomonwakhungu/kfleet/internal/config"
	"github.com/1solomonwakhungu/kfleet/internal/store"
	"github.com/1solomonwakhungu/kfleet/pkg/types"
)

func newBootstrapTestStore(t *testing.T) store.Store {
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
	return st
}

func silentLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestBootstrapAdminCreatesAdminWhenConfigured(t *testing.T) {
	st := newBootstrapTestStore(t)
	cfg := &config.Config{
		BootstrapAdminUsername: "admin",
		BootstrapAdminEmail:    "admin@example.com",
		BootstrapAdminPassword: "hunter2-hunter2-hunter2",
	}

	if err := BootstrapAdmin(context.Background(), cfg, silentLogger(), st); err != nil {
		t.Fatalf("BootstrapAdmin() error = %v", err)
	}

	user, err := st.GetUserByUsername(context.Background(), "admin")
	if err != nil {
		t.Fatalf("GetUserByUsername() error = %v", err)
	}
	if user.Role != types.RoleAdmin || user.Disabled {
		t.Fatalf("bootstrap user = %+v, want enabled admin", user)
	}
	if user.PasswordHash == cfg.BootstrapAdminPassword {
		t.Fatal("bootstrap admin password was stored in plaintext")
	}
	events, err := st.ListAuditEvents(context.Background(), 10)
	if err != nil {
		t.Fatalf("ListAuditEvents() error = %v", err)
	}
	if len(events) != 1 || events[0].Action != "user.bootstrap" || events[0].TargetID != user.ID {
		t.Fatalf("bootstrap audit events = %+v, want one user.bootstrap event", events)
	}
}

func TestBootstrapAdminIsNoOpWhenUsersAlreadyExist(t *testing.T) {
	st := newBootstrapTestStore(t)
	existing := newTestUserRecordForBootstrap(types.RoleReadOnly)
	if err := st.CreateUser(context.Background(), existing); err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	cfg := &config.Config{
		BootstrapAdminUsername: "admin",
		BootstrapAdminEmail:    "admin@example.com",
		BootstrapAdminPassword: "hunter2-hunter2-hunter2",
	}
	if err := BootstrapAdmin(context.Background(), cfg, silentLogger(), st); err != nil {
		t.Fatalf("BootstrapAdmin() error = %v", err)
	}

	if _, err := st.GetUserByUsername(context.Background(), "admin"); err == nil {
		t.Fatal("BootstrapAdmin() created an admin even though users already existed")
	}
}

func TestBootstrapAdminIsNoOpWhenUnconfigured(t *testing.T) {
	st := newBootstrapTestStore(t)
	cfg := &config.Config{}

	if err := BootstrapAdmin(context.Background(), cfg, silentLogger(), st); err != nil {
		t.Fatalf("BootstrapAdmin() error = %v", err)
	}

	users, err := st.ListUsers(context.Background())
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}
	if len(users) != 0 {
		t.Fatalf("ListUsers() = %d users, want 0 when bootstrap admin is unconfigured", len(users))
	}
}

func TestBootstrapAdminRejectsPartialConfiguration(t *testing.T) {
	st := newBootstrapTestStore(t)
	cfg := &config.Config{BootstrapAdminUsername: "admin"}

	if err := BootstrapAdmin(context.Background(), cfg, silentLogger(), st); err == nil {
		t.Fatal("BootstrapAdmin() error = nil, want error for partial configuration")
	}
}

func TestBootstrapAdminRejectsShortPassword(t *testing.T) {
	st := newBootstrapTestStore(t)
	cfg := &config.Config{
		BootstrapAdminUsername: "admin",
		BootstrapAdminEmail:    "admin@example.com",
		BootstrapAdminPassword: "short",
	}

	if err := BootstrapAdmin(context.Background(), cfg, silentLogger(), st); err == nil {
		t.Fatal("BootstrapAdmin() error = nil, want error for short password")
	}
}

func TestBootstrapAdminRejectsPasswordBeyondBcryptLimit(t *testing.T) {
	st := newBootstrapTestStore(t)
	cfg := &config.Config{
		BootstrapAdminUsername: "admin",
		BootstrapAdminEmail:    "admin@example.com",
		BootstrapAdminPassword: strings.Repeat("x", maxPasswordLength+1),
	}

	if err := BootstrapAdmin(context.Background(), cfg, silentLogger(), st); err == nil {
		t.Fatal("BootstrapAdmin() error = nil, want error for password beyond bcrypt limit")
	}
}

func newTestUserRecordForBootstrap(role types.Role) types.User {
	return types.User{
		ID:           "seed-user",
		Username:     "seed",
		Email:        "seed@example.com",
		PasswordHash: "bcrypt-hash-placeholder",
		Role:         role,
	}
}
