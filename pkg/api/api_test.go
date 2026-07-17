package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	payload := RegisterClusterResponse{ClusterID: "cluster-1", Token: "secret"}

	if err := WriteJSON(recorder, http.StatusCreated, payload); err != nil {
		t.Fatalf("WriteJSON() error = %v", err)
	}
	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusCreated)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want %q", got, "application/json")
	}

	var got RegisterClusterResponse
	if err := json.NewDecoder(recorder.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got != payload {
		t.Fatalf("body = %#v, want %#v", got, payload)
	}
}

func TestWriteError(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	WriteError(recorder, http.StatusBadRequest, "invalid cluster")

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}

	var got ErrorResponse
	if err := json.NewDecoder(recorder.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	want := ErrorResponse{Error: "invalid cluster", Code: http.StatusBadRequest}
	if got != want {
		t.Fatalf("body = %#v, want %#v", got, want)
	}
}
