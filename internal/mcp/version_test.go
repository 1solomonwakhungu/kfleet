package mcp

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/1solomonwakhungu/kfleet/internal/version"
)

// TestServerReportsStampedVersion proves the MCP handshake advertises the build
// version rather than a hard-coded constant.
func TestServerReportsStampedVersion(t *testing.T) {
	original := version.Version
	t.Cleanup(func() { version.Version = original })
	version.Version = "v9.9.9"

	server := NewServer(nil)
	response := server.server.HandleMessage(context.Background(), []byte(`{
		"jsonrpc":"2.0",
		"id":1,
		"method":"initialize",
		"params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1"}}
	}`))
	encoded, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal initialize response: %v", err)
	}
	var decoded struct {
		Result struct {
			ServerInfo struct {
				Name    string `json:"name"`
				Version string `json:"version"`
			} `json:"serverInfo"`
		} `json:"result"`
	}
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("decode initialize response: %v", err)
	}
	if decoded.Result.ServerInfo.Version != "v9.9.9" {
		t.Fatalf("serverInfo = %#v, want the stamped build version", decoded.Result.ServerInfo)
	}
}
