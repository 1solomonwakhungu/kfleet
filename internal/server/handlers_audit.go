package server

import (
	"net/http"
	"strconv"

	"github.com/1solomonwakhungu/kfleet/pkg/api"
	"github.com/1solomonwakhungu/kfleet/pkg/types"
)

const defaultAuditEventLimit = 100

func (s *Server) registerAuditRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/audit", s.requireRole(types.RoleAdmin, s.handleListAuditEvents))
}

func (s *Server) handleListAuditEvents(w http.ResponseWriter, r *http.Request) {
	limit := defaultAuditEventLimit
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 || parsed > 1000 {
			api.WriteError(w, http.StatusBadRequest, "limit must be a positive integer no greater than 1000")
			return
		}
		limit = parsed
	}

	events, err := s.store.ListAuditEvents(r.Context(), limit)
	if err != nil {
		s.logger.Error("failed to list audit events", "error", err)
		api.WriteError(w, http.StatusInternalServerError, "failed to list audit events")
		return
	}
	if err := api.WriteJSON(w, http.StatusOK, api.ListAuditEventsResponse{Events: events}); err != nil {
		s.logger.Error("failed to write audit events", "error", err)
	}
}
