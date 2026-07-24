package server

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/1solomonwakhungu/kfleet/pkg/api"
)

func TestHandleListAuditEventsReturnsRecordedEvents(t *testing.T) {
	server := newTestHTTPServer(t)
	admin := defaultSessionFor(server)

	createResp := requestWithSession(t, server, http.MethodPost, "/api/v1/users", admin,
		`{"username":"audited","email":"audited@example.com","password":"hunter2-hunter2","role":"read_only"}`)
	createResp.Body.Close()

	resp := requestWithSession(t, server, http.MethodGet, "/api/v1/audit", admin, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list audit status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var got api.ListAuditEventsResponse
	decodeResponse(t, resp, &got)
	if len(got.Events) == 0 {
		t.Fatal("audit log is empty after a user.create action")
	}

	found := false
	for _, event := range got.Events {
		if event.Action == "user.create" && event.TargetType == "user" {
			found = true
			if event.Details == "" {
				t.Error("user.create audit event has no details")
			}
		}
	}
	if !found {
		t.Error("no user.create audit event was recorded")
	}
}

func TestHandleListAuditEventsValidatesLimit(t *testing.T) {
	server := newTestHTTPServer(t)
	admin := defaultSessionFor(server)

	for _, limit := range []string{"0", "-1", "not-a-number", "1001"} {
		resp := requestWithSession(t, server, http.MethodGet, "/api/v1/audit?limit="+limit, admin, "")
		resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("limit=%q status = %d, want %d", limit, resp.StatusCode, http.StatusBadRequest)
		}
	}

	ok := requestWithSession(t, server, http.MethodGet, "/api/v1/audit?limit=5", admin, "")
	ok.Body.Close()
	if ok.StatusCode != http.StatusOK {
		t.Errorf("limit=5 status = %d, want %d", ok.StatusCode, http.StatusOK)
	}
}

func TestHandleListAuditEventsNeverIncludesSecrets(t *testing.T) {
	server := newTestHTTPServer(t)
	admin := defaultSessionFor(server)

	rotate := requestWithSession(t, server, http.MethodPost, "/api/v1/admin/registration-token/rotate", admin, "")
	var rotated api.RotateRegistrationTokenResponse
	decodeResponse(t, rotate, &rotated)
	if rotated.Token == "" {
		t.Fatal("rotate response has no token")
	}

	resp := requestWithSession(t, server, http.MethodGet, "/api/v1/audit", admin, "")
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read audit response body error = %v", err)
	}
	if strings.Contains(string(raw), rotated.Token) {
		t.Fatal("audit log body contains the raw rotated registration token; secrets must never be audited")
	}
}
