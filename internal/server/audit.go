package server

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/1solomonwakhungu/kfleet/pkg/types"
	"github.com/google/uuid"
)

// auditActor identifies who performed an audited action. Username is always
// populated, even for failed logins where no session yet exists.
type auditActor struct {
	UserID   string
	Username string
	Role     types.Role
}

func auditActorFromUser(user types.User) auditActor {
	return auditActor{UserID: user.ID, Username: user.Username, Role: user.Role}
}

// recordAudit appends an immutable audit event. Callers must never pass
// secrets, tokens, kubeconfigs, or raw credential request bodies in details;
// only non-sensitive, structured metadata (usernames, roles, resource
// identifiers) belongs in the audit log.
func (s *Server) recordAudit(
	ctx context.Context,
	r *http.Request,
	actor auditActor,
	action, targetType, targetID string,
	outcome types.AuditOutcome,
	details string,
) {
	event := types.AuditEvent{
		ID:            uuid.NewString(),
		OccurredAt:    time.Now().UTC(),
		ActorUserID:   actor.UserID,
		ActorUsername: actor.Username,
		ActorRole:     actor.Role,
		Action:        action,
		TargetType:    targetType,
		TargetID:      targetID,
		Outcome:       outcome,
		Details:       details,
		SourceIP:      clientAddr(r),
	}
	if err := s.store.RecordAuditEvent(ctx, event); err != nil {
		s.logger.Error("failed to record audit event", "action", action, "error", err)
	}
}

// clientAddr returns the immediate network peer address for r, with any
// port stripped. It intentionally ignores X-Forwarded-For and similar
// client-supplied headers, which are trivially spoofable unless a specific
// trusted-proxy configuration validates them; operators fronting the hub
// with an ingress that needs the original client IP should configure that
// proxy to record it in their own access logs.
func clientAddr(r *http.Request) string {
	if r == nil {
		return ""
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
