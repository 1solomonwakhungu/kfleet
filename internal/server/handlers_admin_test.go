package server

import (
	"net/http"
	"testing"

	"github.com/1solomonwakhungu/kfleet/internal/config"
	"github.com/1solomonwakhungu/kfleet/pkg/api"
	"github.com/1solomonwakhungu/kfleet/pkg/types"
)

// TestHandleRotateRegistrationTokenReplacesStaticToken proves that after an
// admin rotates the agent registration token, the previously valid static
// KFLEET_REGISTRATION_TOKEN configured at startup is rejected and only the
// newly issued token is accepted, while the agent registration endpoint
// itself remains reachable without a session cookie (preserving the
// existing agent onboarding flow).
func TestHandleRotateRegistrationTokenReplacesStaticToken(t *testing.T) {
	httpServer, _, st := newAgentTestServerWithConfig(t, &config.Config{
		ListenAddr:        ":0",
		RegistrationToken: "bootstrap-token",
	})
	registerDefaultSession(httpServer, st, sessionCookieFor(t, st, types.RoleAdmin))

	beforeRotation := agentRequest(t, httpServer, http.MethodPost, "/api/v1/agents/register", "bootstrap-token", `{"name":"pre-rotation"}`)
	beforeRotation.Body.Close()
	if beforeRotation.StatusCode != http.StatusCreated {
		t.Fatalf("register with static token before rotation status = %d, want %d", beforeRotation.StatusCode, http.StatusCreated)
	}

	rotateResp := requestWithSession(t, httpServer, http.MethodPost, "/api/v1/admin/registration-token/rotate", defaultSessionFor(httpServer), "")
	if rotateResp.StatusCode != http.StatusOK {
		t.Fatalf("rotate status = %d, want %d", rotateResp.StatusCode, http.StatusOK)
	}
	var rotated api.RotateRegistrationTokenResponse
	decodeResponse(t, rotateResp, &rotated)
	if rotated.Token == "" {
		t.Fatal("rotate response has no token")
	}

	staleToken := agentRequest(t, httpServer, http.MethodPost, "/api/v1/agents/register", "bootstrap-token", `{"name":"post-rotation-stale"}`)
	staleToken.Body.Close()
	if staleToken.StatusCode != http.StatusUnauthorized {
		t.Errorf("register with stale static token after rotation status = %d, want %d", staleToken.StatusCode, http.StatusUnauthorized)
	}

	newToken := agentRequest(t, httpServer, http.MethodPost, "/api/v1/agents/register", rotated.Token, `{"name":"post-rotation-fresh"}`)
	newToken.Body.Close()
	if newToken.StatusCode != http.StatusCreated {
		t.Errorf("register with rotated token status = %d, want %d", newToken.StatusCode, http.StatusCreated)
	}
}
