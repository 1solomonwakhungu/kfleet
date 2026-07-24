package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/1solomonwakhungu/kfleet/internal/auth"
	"github.com/1solomonwakhungu/kfleet/internal/config"
	"github.com/1solomonwakhungu/kfleet/internal/store"
	"github.com/1solomonwakhungu/kfleet/pkg/types"
	"github.com/google/uuid"
)

const (
	minBootstrapPasswordLength = 12
	maxPasswordLength          = 72
)

// BootstrapAdmin creates the first admin user from KFLEET_BOOTSTRAP_ADMIN_*
// configuration when the hub has no users at all. It is a no-op once any
// user exists, so it is safe to call on every startup. The bootstrap
// password is never logged.
func BootstrapAdmin(ctx context.Context, cfg *config.Config, logger *slog.Logger, st store.Store) error {
	users, err := st.ListUsers(ctx)
	if err != nil {
		return fmt.Errorf("list users: %w", err)
	}
	if len(users) > 0 {
		return nil
	}

	username := strings.TrimSpace(cfg.BootstrapAdminUsername)
	email := strings.TrimSpace(cfg.BootstrapAdminEmail)
	password := cfg.BootstrapAdminPassword

	if username == "" && email == "" && password == "" {
		logger.Warn("no users exist yet and no bootstrap admin is configured; " +
			"set KFLEET_BOOTSTRAP_ADMIN_USERNAME, KFLEET_BOOTSTRAP_ADMIN_EMAIL, and " +
			"KFLEET_BOOTSTRAP_ADMIN_PASSWORD to create the first admin account")
		return nil
	}
	if username == "" || email == "" || password == "" {
		return errors.New("KFLEET_BOOTSTRAP_ADMIN_USERNAME, KFLEET_BOOTSTRAP_ADMIN_EMAIL, and " +
			"KFLEET_BOOTSTRAP_ADMIN_PASSWORD must all be set together")
	}
	if len(password) < minBootstrapPasswordLength {
		return fmt.Errorf("KFLEET_BOOTSTRAP_ADMIN_PASSWORD must be at least %d characters", minBootstrapPasswordLength)
	}
	if len(password) > maxPasswordLength {
		return fmt.Errorf("KFLEET_BOOTSTRAP_ADMIN_PASSWORD must be no more than %d bytes", maxPasswordLength)
	}

	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		return fmt.Errorf("hash bootstrap admin password: %w", err)
	}

	now := time.Now().UTC()
	user := types.User{
		ID:           uuid.NewString(),
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         types.RoleAdmin,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := st.CreateUser(ctx, user); err != nil {
		return fmt.Errorf("create bootstrap admin: %w", err)
	}
	if err := st.RecordAuditEvent(ctx, types.AuditEvent{
		ID:            uuid.NewString(),
		OccurredAt:    now,
		ActorUsername: "system",
		Action:        "user.bootstrap",
		TargetType:    "user",
		TargetID:      user.ID,
		Outcome:       types.AuditSuccess,
		Details:       "role=admin",
	}); err != nil {
		return fmt.Errorf("audit bootstrap admin creation: %w", err)
	}
	logger.Info("created bootstrap admin user", "username", username)
	return nil
}
