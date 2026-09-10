package server

import (
	"net/http"
	"testing"

	"github.com/1solomonwakhungu/kfleet/pkg/api"
)

func TestUnknownAPIPathReturnsJSONNotFound(t *testing.T) {
	httpServer, _ := newPolicyTestServer(t)

	response := tenantRequest(t, httpServer, http.MethodGet, "/api/v1/does-not-exist", "tenant-a", "")
	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("unknown API path status = %d, want 404", response.StatusCode)
	}
	assertJSONErrorBody(t, response, http.StatusNotFound)
}

func TestUnknownAPIPathWithMutationMethodReturnsNotFound(t *testing.T) {
	httpServer, _ := newPolicyTestServer(t)

	response := tenantRequest(t, httpServer, http.MethodPost, "/api/v1/does-not-exist", "tenant-a", `{}`)
	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("unknown API path status = %d, want 404", response.StatusCode)
	}
	assertJSONErrorBody(t, response, http.StatusNotFound)
}

func TestBareAPIPathReturnsJSONNotFound(t *testing.T) {
	httpServer, _ := newPolicyTestServer(t)

	response := tenantRequest(t, httpServer, http.MethodGet, "/api", "tenant-a", "")
	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("bare /api status = %d, want 404", response.StatusCode)
	}
	if cc := response.Header.Get("Cache-Control"); cc != "no-store" {
		t.Fatalf("bare /api cache control = %q, want no-store", cc)
	}
	assertJSONErrorBody(t, response, http.StatusNotFound)
}

func TestWrongMethodOnKnownAPIRouteReturnsJSONNotAllowed(t *testing.T) {
	httpServer, _ := newPolicyTestServer(t)

	cases := []struct {
		name   string
		method string
		path   string
		allow  string
	}{
		{"POST on GET-only catalog", http.MethodPost, "/api/v1/policies", "GET, HEAD"},
		{"DELETE on results", http.MethodDelete, "/api/v1/policies/results", "GET, HEAD"},
		{"GET on POST-only login", http.MethodGet, "/api/v1/auth/login", "POST"},
		{"GET on POST-only heartbeat", http.MethodGet, "/api/v1/agents/heartbeat", "POST"},
		{"DELETE on GET-only wildcard pods", http.MethodDelete, "/api/v1/clusters/some-cluster/pods", "GET, HEAD"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			response := tenantRequest(t, httpServer, tc.method, tc.path, "tenant-a", "")
			if response.StatusCode != http.StatusMethodNotAllowed {
				t.Fatalf("%s %s status = %d, want 405", tc.method, tc.path, response.StatusCode)
			}
			if allow := response.Header.Get("Allow"); allow != tc.allow {
				t.Fatalf("Allow header = %q, want %q", allow, tc.allow)
			}
			assertJSONErrorBody(t, response, http.StatusMethodNotAllowed)
		})
	}
}

func TestPolicyAliasRoutesRemoved(t *testing.T) {
	httpServer, _ := newPolicyTestServer(t)

	results := tenantRequest(t, httpServer, http.MethodGet, "/api/v1/policies/results", "tenant-a", "")
	if results.StatusCode != http.StatusOK {
		t.Fatalf("policy results status = %d, want 200", results.StatusCode)
	}
	results.Body.Close()

	for _, alias := range []string{"/api/v1/policies/summary", "/api/v1/policy-results", "/api/v1/drift"} {
		response := tenantRequest(t, httpServer, http.MethodGet, alias, "tenant-a", "")
		if response.StatusCode != http.StatusNotFound {
			t.Fatalf("removed alias %s status = %d, want 404", alias, response.StatusCode)
		}
		assertJSONErrorBody(t, response, http.StatusNotFound)
	}
}

func TestMetaRouteStillServed(t *testing.T) {
	httpServer, _ := newPolicyTestServer(t)

	response := tenantRequest(t, httpServer, http.MethodGet, "/api/v1/meta", "tenant-a", "")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("meta status = %d, want 200", response.StatusCode)
	}
	var meta api.RuntimeInfo
	decodeResponse(t, response, &meta)
	if meta.DataPolicy == "" {
		t.Fatalf("meta data policy = empty, want a description")
	}
}

func TestSPAFallbackUnchangedForNonAPIPaths(t *testing.T) {
	httpServer, _ := newPolicyTestServer(t)

	response, err := http.Get(httpServer.URL + "/not-an-api-route")
	if err != nil {
		t.Fatalf("SPA fallback request error = %v", err)
	}
	defer response.Body.Close()
	if ct := response.Header.Get("Content-Type"); ct == "application/json" {
		t.Fatalf("non-API path content type = application/json, want the SPA handler's response")
	}
}

func assertJSONErrorBody(t *testing.T, response *http.Response, wantCode int) {
	t.Helper()
	if ct := response.Header.Get("Content-Type"); ct != "application/json" {
		t.Fatalf("content type = %q, want application/json", ct)
	}
	var payload api.ErrorResponse
	decodeResponse(t, response, &payload)
	if payload.Error == "" {
		t.Fatalf("error body %q, want a non-empty message", payload.Error)
	}
	if payload.Code != wantCode {
		t.Fatalf("error code = %d, want %d", payload.Code, wantCode)
	}
}
