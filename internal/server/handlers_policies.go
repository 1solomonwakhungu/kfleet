package server

import (
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/1solomonwakhungu/kfleet/internal/policy"
	"github.com/1solomonwakhungu/kfleet/internal/store"
	"github.com/1solomonwakhungu/kfleet/pkg/api"
	"github.com/1solomonwakhungu/kfleet/pkg/types"
)

const tenantHeader = "X-Kfleet-Tenant-ID"

var tenantIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,62}$`)

func (s *Server) registerPolicyRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/policies", s.requireAuth(s.handleListPolicies))
	mux.HandleFunc("GET /api/v1/policies/results", s.requireAuth(s.handlePolicyResults))
	mux.HandleFunc("GET /api/v1/policies/summary", s.requireAuth(s.handlePolicyResults))
	mux.HandleFunc("GET /api/v1/policy-results", s.requireAuth(s.handlePolicyResults))
	mux.HandleFunc("GET /api/v1/drift", s.requireAuth(s.handlePolicyResults))
	mux.HandleFunc("GET /api/v1/clusters/{id}/policy-results", s.requireAuth(s.handleClusterPolicyResults))
}

func (s *Server) handleListPolicies(w http.ResponseWriter, r *http.Request) {
	if _, ok := tenantIDFromRequest(w, r); !ok {
		return
	}
	if err := api.WriteJSON(w, http.StatusOK, api.PolicyListResponse{Policies: policy.Catalog()}); err != nil {
		s.logger.Error("failed to write policy catalog", "error", err)
	}
}

func (s *Server) handlePolicyResults(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := tenantIDFromRequest(w, r)
	if !ok {
		return
	}
	results, err := s.policies.Evaluate(r.Context(), tenantID)
	if err != nil {
		s.logger.Error("failed to evaluate policies", "tenant_id", tenantID, "error", err)
		api.WriteError(w, http.StatusInternalServerError, "failed to evaluate policies")
		return
	}
	results, ok = filterPolicyResults(w, r, results)
	if !ok {
		return
	}
	s.writePolicyResults(w, results)
}

func (s *Server) handleClusterPolicyResults(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := tenantIDFromRequest(w, r)
	if !ok {
		return
	}
	cluster, err := s.store.GetClusterForTenant(r.Context(), tenantID, r.PathValue("id"))
	if errors.Is(err, store.ErrNotFound) {
		api.WriteError(w, http.StatusNotFound, "cluster not found")
		return
	}
	if err != nil {
		s.logger.Error("failed to load policy cluster", "tenant_id", tenantID, "error", err)
		api.WriteError(w, http.StatusInternalServerError, "failed to evaluate policies")
		return
	}
	results, err := s.policies.Evaluate(r.Context(), tenantID)
	if err != nil {
		s.logger.Error("failed to evaluate cluster policies", "tenant_id", tenantID, "cluster_id", cluster.ID, "error", err)
		api.WriteError(w, http.StatusInternalServerError, "failed to evaluate policies")
		return
	}
	filtered := make([]types.PolicyResult, 0)
	for _, result := range results {
		if result.Subject.ClusterID == cluster.ID {
			filtered = append(filtered, result)
		}
	}
	filtered, ok = filterPolicyResults(w, r, filtered)
	if !ok {
		return
	}
	s.writePolicyResults(w, filtered)
}

func (s *Server) writePolicyResults(w http.ResponseWriter, results []types.PolicyResult) {
	clusters := make(map[string]struct{})
	for _, result := range results {
		if result.Subject.ClusterID != "" {
			clusters[result.Subject.ClusterID] = struct{}{}
		}
	}
	now := time.Now().UTC()
	if len(results) > 0 {
		now = results[0].EvaluatedAt
	}
	response := api.PolicyResultsResponse{
		Results: results,
		Summary: policy.Summarize(results, len(clusters), now),
	}
	if err := api.WriteJSON(w, http.StatusOK, response); err != nil {
		s.logger.Error("failed to write policy results", "error", err)
	}
}

func filterPolicyResults(w http.ResponseWriter, r *http.Request, results []types.PolicyResult) ([]types.PolicyResult, bool) {
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	severity := strings.TrimSpace(r.URL.Query().Get("severity"))
	scope := strings.TrimSpace(r.URL.Query().Get("scope"))
	policyID := strings.TrimSpace(r.URL.Query().Get("policyId"))
	clusterID := strings.TrimSpace(r.URL.Query().Get("clusterId"))

	if status != "" && !validPolicyStatus(status) {
		api.WriteError(w, http.StatusBadRequest, "invalid policy status")
		return nil, false
	}
	if severity != "" && !validPolicySeverity(severity) {
		api.WriteError(w, http.StatusBadRequest, "invalid policy severity")
		return nil, false
	}
	if scope != "" && !validPolicyScope(scope) {
		api.WriteError(w, http.StatusBadRequest, "invalid policy scope")
		return nil, false
	}

	filtered := make([]types.PolicyResult, 0, len(results))
	for _, result := range results {
		if status != "" && string(result.Status) != status {
			continue
		}
		if severity != "" && string(result.Severity) != severity {
			continue
		}
		if scope != "" && string(result.Scope) != scope {
			continue
		}
		if policyID != "" && result.PolicyID != policyID {
			continue
		}
		if clusterID != "" && result.Subject.ClusterID != clusterID {
			continue
		}
		filtered = append(filtered, result)
	}
	return filtered, true
}

func tenantIDFromRequest(w http.ResponseWriter, r *http.Request) (string, bool) {
	w.Header().Add("Vary", tenantHeader)
	values := r.Header.Values(tenantHeader)
	if len(values) == 0 || strings.TrimSpace(values[0]) == "" {
		return store.DefaultTenantID, true
	}
	if len(values) != 1 {
		api.WriteError(w, http.StatusBadRequest, "invalid tenant ID")
		return "", false
	}
	tenantID := strings.TrimSpace(values[0])
	if strings.Contains(tenantID, ",") || !tenantIDPattern.MatchString(tenantID) {
		api.WriteError(w, http.StatusBadRequest, "invalid tenant ID")
		return "", false
	}
	return tenantID, true
}

func validPolicyStatus(value string) bool {
	switch types.PolicyStatus(value) {
	case types.PolicyPass, types.PolicyFail, types.PolicyUnknown, types.PolicyStale:
		return true
	default:
		return false
	}
}

func validPolicySeverity(value string) bool {
	switch types.PolicySeverity(value) {
	case types.SeverityLow, types.SeverityMedium, types.SeverityHigh, types.SeverityCritical:
		return true
	default:
		return false
	}
}

func validPolicyScope(value string) bool {
	switch types.PolicyScope(value) {
	case types.ScopeFleet, types.ScopeCluster, types.ScopeNamespace, types.ScopeWorkload:
		return true
	default:
		return false
	}
}
