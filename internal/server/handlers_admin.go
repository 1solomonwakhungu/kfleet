package server

import (
	"net/http"

	"github.com/1solomonwakhungu/kfleet/internal/auth"
	"github.com/1solomonwakhungu/kfleet/pkg/api"
	"github.com/1solomonwakhungu/kfleet/pkg/types"
)

// settingRegistrationTokenHash stores a rotated agent registration token's
// hash. When unset, the hub falls back to the KFLEET_REGISTRATION_TOKEN
// environment variable, preserving the original agent registration flow.
const settingRegistrationTokenHash = "registration_token_hash"

func (s *Server) registerAdminRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/admin/registration-token/rotate", s.requireRole(types.RoleAdmin, s.handleRotateRegistrationToken))
}

// handleRotateRegistrationToken issues a new agent registration token and
// persists only its hash, invalidating any previously issued registration
// token (including the static KFLEET_REGISTRATION_TOKEN, once rotated at
// least once). The raw token is returned exactly once in the response body
// and is never logged.
func (s *Server) handleRotateRegistrationToken(w http.ResponseWriter, r *http.Request) {
	actor, _ := authenticatedUser(r.Context())

	rawToken, tokenHash, err := auth.NewSessionToken()
	if err != nil {
		s.logger.Error("failed to generate registration token", "error", err)
		api.WriteError(w, http.StatusInternalServerError, "failed to rotate registration token")
		return
	}
	if err := s.store.SetSetting(r.Context(), settingRegistrationTokenHash, tokenHash); err != nil {
		s.logger.Error("failed to persist rotated registration token", "error", err)
		api.WriteError(w, http.StatusInternalServerError, "failed to rotate registration token")
		return
	}

	s.recordAudit(r.Context(), r, auditActorFromUser(actor), "config.registration_token_rotate", "registration_token", "", types.AuditSuccess, "")
	if err := api.WriteJSON(w, http.StatusOK, api.RotateRegistrationTokenResponse{Token: rawToken}); err != nil {
		s.logger.Error("failed to write rotated registration token", "error", err)
	}
}
