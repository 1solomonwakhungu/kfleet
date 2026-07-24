package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/1solomonwakhungu/kfleet/internal/auth"
	"github.com/1solomonwakhungu/kfleet/internal/store"
	"github.com/1solomonwakhungu/kfleet/pkg/api"
	"github.com/1solomonwakhungu/kfleet/pkg/types"
)

func (s *Server) registerAuthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/auth/login", s.handleLogin)
	mux.HandleFunc("POST /api/v1/auth/logout", s.requireAuth(s.handleLogout))
	mux.HandleFunc("GET /api/v1/auth/me", s.requireAuth(s.handleMe))
}

// handleLogin authenticates a username and password and, on success, issues
// a session cookie. The request body (which contains the plaintext
// password) is never logged; only the outcome and username are audited.
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var request api.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	username := strings.TrimSpace(request.Username)
	if username == "" || request.Password == "" {
		api.WriteError(w, http.StatusBadRequest, "username and password are required")
		return
	}

	user, err := s.store.GetUserByUsername(r.Context(), username)
	if err != nil {
		auth.ConsumePasswordVerification(request.Password)
		if !errors.Is(err, store.ErrNotFound) {
			s.logger.Error("failed to look up user for login", "error", err)
		}
		s.recordAudit(r.Context(), r, auditActor{Username: username}, "login", "user", username, types.AuditFailure, "user not found")
		api.WriteError(w, http.StatusUnauthorized, "invalid username or password")
		return
	}
	if user.Disabled {
		s.recordAudit(r.Context(), r, auditActorFromUser(user), "login", "user", user.ID, types.AuditFailure, "account disabled")
		api.WriteError(w, http.StatusUnauthorized, "invalid username or password")
		return
	}
	if err := auth.VerifyPassword(user.PasswordHash, request.Password); err != nil {
		s.recordAudit(r.Context(), r, auditActorFromUser(user), "login", "user", user.ID, types.AuditFailure, "incorrect password")
		api.WriteError(w, http.StatusUnauthorized, "invalid username or password")
		return
	}

	rawToken, tokenHash, err := auth.NewSessionToken()
	if err != nil {
		s.logger.Error("failed to generate session token", "error", err)
		api.WriteError(w, http.StatusInternalServerError, "failed to start session")
		return
	}
	expiresAt := time.Now().UTC().Add(s.sessionDuration())
	if err := s.store.CreateSession(r.Context(), tokenHash, user.ID, expiresAt); err != nil {
		s.logger.Error("failed to persist session", "error", err)
		api.WriteError(w, http.StatusInternalServerError, "failed to start session")
		return
	}

	s.setSessionCookie(w, rawToken, expiresAt)
	s.recordAudit(r.Context(), r, auditActorFromUser(user), "login", "user", user.ID, types.AuditSuccess, "")
	if err := api.WriteJSON(w, http.StatusOK, toUserResponse(user)); err != nil {
		s.logger.Error("failed to write login response", "error", err)
	}
}

func (s *Server) sessionDuration() time.Duration {
	if s.cfg.SessionDuration > 0 {
		return s.cfg.SessionDuration
	}
	return 24 * time.Hour
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	user, _ := authenticatedUser(r.Context())
	if cookie, err := r.Cookie(sessionCookieName); err == nil && cookie.Value != "" {
		if err := s.store.DeleteSession(r.Context(), auth.HashToken(cookie.Value)); err != nil {
			s.logger.Error("failed to delete session", "error", err)
		}
	}
	s.clearSessionCookie(w)
	s.recordAudit(r.Context(), r, auditActorFromUser(user), "logout", "user", user.ID, types.AuditSuccess, "")
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	user, ok := authenticatedUser(r.Context())
	if !ok {
		api.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	if err := api.WriteJSON(w, http.StatusOK, toUserResponse(user)); err != nil {
		s.logger.Error("failed to write current user response", "error", err)
	}
}

func toUserResponse(user types.User) api.UserResponse {
	return api.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		Disabled:  user.Disabled,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
