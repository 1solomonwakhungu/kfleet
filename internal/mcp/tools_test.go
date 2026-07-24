package mcp

import (
	"context"
	"testing"

	"github.com/1solomonwakhungu/kfleet/pkg/types"
	mcpclient "github.com/mark3labs/mcp-go/client"
	protocol "github.com/mark3labs/mcp-go/mcp"
)

func TestCrashloopPods(t *testing.T) {
	cluster := types.Cluster{ID: "cluster-1", Name: "production"}
	pods := []types.Pod{
		{Name: "crashing", Phase: "CrashLoopBackOff"},
		{Name: "restarting", Phase: "Running", RestartCount: highRestartCount},
		{Name: "healthy", Phase: "Running", RestartCount: highRestartCount - 1},
	}
	found := crashloopPods(cluster, pods)
	if len(found) != 2 {
		t.Fatalf("crashloopPods() = %#v, want two pods", found)
	}
	if found[0].ClusterID != cluster.ID || found[0].ClusterName != cluster.Name {
		t.Fatalf("cluster identity = %#v, want %#v", found[0], cluster)
	}
}

func TestFindClusterByNameOrID(t *testing.T) {
	clusters := []types.Cluster{{ID: "cluster-1", Name: "Production"}}
	for _, query := range []string{"cluster-1", "production"} {
		if _, ok := findCluster(clusters, query); !ok {
			t.Errorf("findCluster(%q) did not find cluster", query)
		}
	}
}

func TestRegistersRecentEventsTool(t *testing.T) {
	server := NewServer(nil)
	client, err := mcpclient.NewInProcessClient(server.server)
	if err != nil {
		t.Fatalf("NewInProcessClient() error = %v", err)
	}
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Errorf("client.Close() error = %v", err)
		}
	})
	ctx := context.Background()
	if err := client.Start(ctx); err != nil {
		t.Fatalf("client.Start() error = %v", err)
	}
	initialize := protocol.InitializeRequest{}
	initialize.Params.ProtocolVersion = protocol.LATEST_PROTOCOL_VERSION
	initialize.Params.ClientInfo = protocol.Implementation{Name: "test", Version: "1"}
	if _, err := client.Initialize(ctx, initialize); err != nil {
		t.Fatalf("client.Initialize() error = %v", err)
	}

	result, err := client.ListTools(ctx, protocol.ListToolsRequest{})
	if err != nil {
		t.Fatalf("ListTools() error = %v", err)
	}
	for _, tool := range result.Tools {
		if tool.Name == "get_recent_events" {
			return
		}
	}
	t.Fatalf("get_recent_events not registered: %#v", result.Tools)
}
