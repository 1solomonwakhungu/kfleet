package server

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/1solomonwakhungu/kfleet/internal/config"
	"github.com/1solomonwakhungu/kfleet/internal/store"
	"github.com/1solomonwakhungu/kfleet/pkg/api"
	"github.com/1solomonwakhungu/kfleet/pkg/types"
)

func TestClusterHandlersLifecycle(t *testing.T) {
	server := newTestHTTPServer(t)

	registerResponse := request(t, server, http.MethodPost, "/api/v1/clusters/register", `{
		"name": "production",
		"labels": {"region": "us-central1"}
	}`)
	if registerResponse.StatusCode != http.StatusCreated {
		t.Fatalf("register status = %d, want %d", registerResponse.StatusCode, http.StatusCreated)
	}
	var registration api.RegisterClusterResponse
	decodeResponse(t, registerResponse, &registration)
	if registration.ClusterID == "" || registration.Token == "" {
		t.Fatalf("register response = %#v, want ID and token", registration)
	}

	listResponse := request(t, server, http.MethodGet, "/api/v1/clusters", "")
	if listResponse.StatusCode != http.StatusOK {
		t.Fatalf("list status = %d, want %d", listResponse.StatusCode, http.StatusOK)
	}
	var list api.ListClustersResponse
	decodeResponse(t, listResponse, &list)
	if len(list.Clusters) != 1 || list.Clusters[0].ID != registration.ClusterID {
		t.Fatalf("list response = %#v, want registered cluster", list)
	}

	getResponse := request(t, server, http.MethodGet, "/api/v1/clusters/"+registration.ClusterID, "")
	if getResponse.StatusCode != http.StatusOK {
		t.Fatalf("get status = %d, want %d", getResponse.StatusCode, http.StatusOK)
	}
	var cluster types.Cluster
	decodeResponse(t, getResponse, &cluster)
	if cluster.ID != registration.ClusterID || cluster.Name != "production" {
		t.Fatalf("get response = %#v, want registered cluster", cluster)
	}

	unknownResponse := request(t, server, http.MethodGet, "/api/v1/clusters/unknown", "")
	if unknownResponse.StatusCode != http.StatusNotFound {
		t.Fatalf("unknown get status = %d, want %d", unknownResponse.StatusCode, http.StatusNotFound)
	}
	unknownResponse.Body.Close()

	deleteResponse := request(t, server, http.MethodDelete, "/api/v1/clusters/"+registration.ClusterID, "")
	if deleteResponse.StatusCode != http.StatusNoContent {
		t.Fatalf("delete status = %d, want %d", deleteResponse.StatusCode, http.StatusNoContent)
	}
	deleteResponse.Body.Close()

	deletedResponse := request(t, server, http.MethodGet, "/api/v1/clusters/"+registration.ClusterID, "")
	if deletedResponse.StatusCode != http.StatusNotFound {
		t.Fatalf("deleted get status = %d, want %d", deletedResponse.StatusCode, http.StatusNotFound)
	}
	deletedResponse.Body.Close()
}

func TestRegisterClusterRejectsEmptyName(t *testing.T) {
	server := newTestHTTPServer(t)

	response := request(t, server, http.MethodPost, "/api/v1/clusters/register", `{"name":"  "}`)
	defer response.Body.Close()
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusBadRequest)
	}
}

func newTestHTTPServer(t *testing.T) *httptest.Server {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "kfleet.db"))
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := New(&config.Config{ListenAddr: ":0"}, logger, st)
	httpServer := httptest.NewServer(srv.httpServer.Handler)
	t.Cleanup(func() {
		httpServer.Close()
		if err := st.Close(); err != nil {
			t.Errorf("store.Close() error = %v", err)
		}
	})
	return httpServer
}

func request(t *testing.T, server *httptest.Server, method, path, body string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(method, server.URL+path, bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	response, err := server.Client().Do(req)
	if err != nil {
		t.Fatalf("request error = %v", err)
	}
	return response
}

func decodeResponse(t *testing.T, response *http.Response, target any) {
	t.Helper()
	defer response.Body.Close()
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		t.Fatalf("decode response error = %v", err)
	}
}
