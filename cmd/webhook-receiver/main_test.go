package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/1solomonwakhungu/kfleet/internal/alerts"
)

func TestReceiverVerifiesSignaturesAndInjectsFailures(t *testing.T) {
	const secret = "local-secret"
	state := &receiver{
		secret:        secret,
		failRemaining: 1,
		failStatus:    http.StatusServiceUnavailable,
		received:      make([]receivedWebhook, 0),
	}
	body := []byte(`{"event":"fleet.health.alert"}`)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	first := signedReceiverRequest(body, timestamp, alerts.Signature(secret, timestamp, body))
	firstResponse := httptest.NewRecorder()
	state.handleWebhook(firstResponse, first)
	if firstResponse.Code != http.StatusServiceUnavailable {
		t.Fatalf("first response = %d, want %d", firstResponse.Code, http.StatusServiceUnavailable)
	}

	second := signedReceiverRequest(body, timestamp, alerts.Signature(secret, timestamp, body))
	secondResponse := httptest.NewRecorder()
	state.handleWebhook(secondResponse, second)
	if secondResponse.Code != http.StatusNoContent {
		t.Fatalf("second response = %d, want %d", secondResponse.Code, http.StatusNoContent)
	}

	invalid := signedReceiverRequest(body, timestamp, "v1=invalid")
	invalidResponse := httptest.NewRecorder()
	state.handleWebhook(invalidResponse, invalid)
	if invalidResponse.Code != http.StatusUnauthorized {
		t.Fatalf("invalid signature response = %d, want %d", invalidResponse.Code, http.StatusUnauthorized)
	}

	state.mu.Lock()
	defer state.mu.Unlock()
	if len(state.received) != 2 || state.failRemaining != 0 {
		t.Fatalf("receiver state = received %d, failures %d", len(state.received), state.failRemaining)
	}
}

func TestReceiverFailureControl(t *testing.T) {
	state := &receiver{failStatus: http.StatusServiceUnavailable}
	request := httptest.NewRequest(
		http.MethodPost,
		"/control/failures",
		bytes.NewBufferString(`{"count":3,"status":502}`),
	)
	response := httptest.NewRecorder()
	state.handleFailures(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("control response = %d, want %d", response.Code, http.StatusOK)
	}
	if state.failRemaining != 3 || state.failStatus != http.StatusBadGateway {
		t.Fatalf("failure control state = (%d, %d)", state.failRemaining, state.failStatus)
	}
}

func TestLoopbackAddressValidation(t *testing.T) {
	for _, address := range []string{"127.0.0.1:9099", "[::1]:9099", "localhost:9099"} {
		if !isLoopbackAddress(address) {
			t.Errorf("isLoopbackAddress(%q) = false, want true", address)
		}
	}
	for _, address := range []string{":9099", "0.0.0.0:9099", "192.0.2.1:9099", "invalid"} {
		if isLoopbackAddress(address) {
			t.Errorf("isLoopbackAddress(%q) = true, want false", address)
		}
	}
}

func signedReceiverRequest(body []byte, timestamp, signature string) *http.Request {
	request := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	request.Header.Set("X-Kfleet-Timestamp", timestamp)
	request.Header.Set("X-Kfleet-Signature", signature)
	request.Header.Set("X-Kfleet-Delivery", "alert-1")
	request.Header.Set("X-Kfleet-Event", alerts.EventType)
	return request
}
