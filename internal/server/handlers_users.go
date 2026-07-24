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
	"github.com/google/uuid"
)

func (s *Server) registerUserRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/users", s.requireRole(types.RoleAdmin, s.handleListUsers))
	mux.HandleFunc("POST /api/v1/users", s.requireRole(types.RoleAdmin, s.handleCreateUser))
	mux.HandleFunc("PATCH /api/v1/users/{id}", s.requireRole(types.RoleAdmin, s.handleUpdateUser))
	mux.HandleFunc("DELETE /api/v1/users/{id}", s.requireRole(types.RoleAdmin, s.handleDeleteUser))
}

func (s *Server) handleListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := s.store.ListUsers(r.Context())
	if err != nil {
		s.logger.Error("failed to list users", "error", err)
		api.WriteError(w, http.StatusInternalServerError, "failed to list users")
		return
	}
	responses := make([]api.UserResponse, 0, len(users))
	for _, user := range users {
		responses = append(responses, toUserResponse(user))
	}
	if err := api.WriteJSON(w, http.StatusOK, api.ListUsersResponse{Users: responses}); err != nil {
		s.logger.Error("failed to write user list", "error", err)
	}
}

func (s *Server) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	actor, _ := authenticatedUser(r.Context())

	var request api.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	username := strings.TrimSpace(request.Username)
	email := strings.TrimSpace(request.Email)
	if username == "" || email == "" || request.Password == "" {
		api.WriteError(w, http.StatusBadRequest, "username, email, and password are required")
		return
	}
	if len(request.Password) < minBootstrapPasswordLength || len(request.Password) > maxPasswordLength {
		api.WriteError(w, http.StatusBadRequest, "password must be between 12 and 72 bytes")
		return
	}
	if !types.ValidRole(request.Role) {
		api.WriteError(w, http.StatusBadRequest, "role must be admin, operator, or read_only")
		return
	}

	passwordHash, err := auth.HashPassword(request.Password)
	if err != nil {
		s.logger.Error("failed to hash password for new user", "error", err)
		api.WriteError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	now := time.Now().UTC()
	user := types.User{
		ID:           uuid.NewString(),
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         request.Role,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.store.CreateUser(r.Context(), user); err != nil {
		if errors.Is(err, store.ErrConflict) {
			api.WriteError(w, http.StatusConflict, "username or email is already in use")
			return
		}
		s.logger.Error("failed to create user", "error", err)
		api.WriteError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	s.recordAudit(r.Context(), r, auditActorFromUser(actor), "user.create", "user", user.ID, types.AuditSuccess, "role="+string(user.Role))
	if err := api.WriteJSON(w, http.StatusCreated, toUserResponse(user)); err != nil {
		s.logger.Error("failed to write created user", "error", err)
	}
}

func (s *Server) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	actor, _ := authenticatedUser(r.Context())
	targetID := r.PathValue("id")

	var request api.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if !types.ValidRole(request.Role) {
		api.WriteError(w, http.StatusBadRequest, "role must be admin, operator, or read_only")
		return
	}

	if err := s.store.UpdateUser(r.Context(), targetID, request.Role, request.Disabled); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			api.WriteError(w, http.StatusNotFound, "user not found")
			return
		}
		if errors.Is(err, store.ErrLastAdmin) {
			api.WriteError(w, http.StatusConflict, "at least one enabled admin is required")
			return
		}
		s.logger.Error("failed to update user", "error", err)
		api.WriteError(w, http.StatusInternalServerError, "failed to update user")
		return
	}

	updated, err := s.store.GetUserByID(r.Context(), targetID)
	if err != nil {
		s.logger.Error("failed to read updated user", "error", err)
		api.WriteError(w, http.StatusInternalServerError, "failed to update user")
		return
	}

	s.recordAudit(r.Context(), r, auditActorFromUser(actor), "user.update", "user", targetID, types.AuditSuccess, "role="+string(request.Role))
	if err := api.WriteJSON(w, http.StatusOK, toUserResponse(updated)); err != nil {
		s.logger.Error("failed to write updated user", "error", err)
	}
}

func (s *Server) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	actor, _ := authenticatedUser(r.Context())
	targetID := r.PathValue("id")

	if targetID == actor.ID {
		api.WriteError(w, http.StatusConflict, "you cannot delete your own account")
		return
	}
	if err := s.store.DeleteUser(r.Context(), targetID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			api.WriteError(w, http.StatusNotFound, "user not found")
			return
		}
		if errors.Is(err, store.ErrLastAdmin) {
			api.WriteError(w, http.StatusConflict, "at least one enabled admin is required")
			return
		}
		s.logger.Error("failed to delete user", "error", err)
		api.WriteError(w, http.StatusInternalServerError, "failed to delete user")
		return
	}

	s.recordAudit(r.Context(), r, auditActorFromUser(actor), "user.delete", "user", targetID, types.AuditSuccess, "")
	w.WriteHeader(http.StatusNoContent)
}
