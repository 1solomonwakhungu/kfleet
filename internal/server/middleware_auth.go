package server

import (
	"errors"
	"net/http"
	"time"

	"github.com/1solomonwakhungu/kfleet/internal/auth"
	"github.com/1solomonwakhungu/kfleet/internal/store"
	"github.com/1solomonwakhungu/kfleet/pkg/api"
	"github.com/1solomonwakhungu/kfleet/pkg/types"
)

const sessionCookieName = "kfleet_session"
const csrfHeaderName = "X-Kfleet-CSRF"

func (s *Server) setSessionCookie(w http.ResponseWriter, rawToken string, expiresAt time.Time) {
	maxAge := int(time.Until(expiresAt).Seconds())
	if maxAge < 1 {
		maxAge = 1
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    rawToken,
		Path:     "/",
		Expires:  expiresAt,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   s.cfg.SessionCookieSecure,
		SameSite: http.SameSiteStrictMode,
	})
}

func (s *Server) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   s.cfg.SessionCookieSecure,
		SameSite: http.SameSiteStrictMode,
	})
}

// requireAuth wraps next so it only runs for requests carrying a valid,
// unexpired session cookie for an enabled user. The authenticated user is
// attached to the request context for downstream handlers.
func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil || cookie.Value == "" {
			api.WriteError(w, http.StatusUnauthorized, "authentication required")
			return
		}

		user, err := s.store.GetSessionUser(r.Context(), auth.HashToken(cookie.Value), time.Now().UTC())
		if err != nil {
			if !errors.Is(err, store.ErrNotFound) {
				s.logger.Error("failed to validate session", "error", err)
			}
			s.clearSessionCookie(w)
			api.WriteError(w, http.StatusUnauthorized, "session is invalid or has expired")
			return
		}
		if user.Disabled {
			s.clearSessionCookie(w)
			api.WriteError(w, http.StatusUnauthorized, "account is disabled")
			return
		}
		if mutationRequest(r) && r.Header.Get(csrfHeaderName) != "1" {
			s.recordAudit(
				r.Context(),
				r,
				auditActorFromUser(user),
				"authorization.csrf_denied",
				"http_route",
				r.Method+" "+r.URL.Path,
				types.AuditFailure,
				"missing or invalid CSRF header",
			)
			api.WriteError(w, http.StatusForbidden, "CSRF validation failed")
			return
		}

		next(w, r.WithContext(withAuthenticatedUser(r.Context(), user)))
	}
}

// requireRole wraps next with requireAuth and additionally rejects
// authenticated users whose role does not meet the minimum privilege level.
func (s *Server) requireRole(minimum types.Role, next http.HandlerFunc) http.HandlerFunc {
	return s.requireAuth(func(w http.ResponseWriter, r *http.Request) {
		user, ok := authenticatedUser(r.Context())
		if !ok || !auth.HasAtLeast(user.Role, minimum) {
			if ok {
				s.recordAudit(
					r.Context(),
					r,
					auditActorFromUser(user),
					"authorization.role_denied",
					"http_route",
					r.Method+" "+r.URL.Path,
					types.AuditFailure,
					"required_role="+string(minimum),
				)
			}
			api.WriteError(w, http.StatusForbidden, "this action requires a higher role")
			return
		}
		next(w, r)
	})
}

func mutationRequest(r *http.Request) bool {
	switch r.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return false
	default:
		return true
	}
}
