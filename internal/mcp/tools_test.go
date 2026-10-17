package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/1solomonwakhungu/kfleet/internal/mcp/hubclient"
	"github.com/1solomonwakhungu/kfleet/pkg/api"
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

// newTestClient starts an in-process MCP client against a server backed by
// hub, which may be nil for tests that never call a hub-backed tool.
func newTestClient(t *testing.T, hub *hubclient.Client) *mcpclient.Client {
	t.Helper()
	server := NewServer(hub)
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
	return client
}

func TestRegistersTools(t *testing.T) {
	client := newTestClient(t, nil)
	result, err := client.ListTools(context.Background(), protocol.ListToolsRequest{})
	if err != nil {
		t.Fatalf("ListTools() error = %v", err)
	}
	registered := make(map[string]bool, len(result.Tools))
	for _, tool := range result.Tools {
		registered[tool.Name] = true
	}
	for _, name := range []string{"get_recent_events", "get_pod_logs", "list_namespaces"} {
		if !registered[name] {
			t.Errorf("%s not registered: %#v", name, result.Tools)
		}
	}
}

// fakeHub serves the hub endpoints the log and namespace tools call. Each
// log request is recorded so tests can assert the forwarded options.
type fakeHub struct {
	mu        sync.Mutex
	logQuery  url.Values
	logPath   string
	logStream string
	logStatus int
}

func newFakeHub(t *testing.T, hub *fakeHub) *hubclient.Client {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v1/clusters":
			_ = json.NewEncoder(w).Encode(api.ListClustersResponse{Clusters: []types.Cluster{{ID: "cluster-1", Name: "production"}}})
		case r.URL.Path == "/api/v1/clusters/cluster-1/namespaces":
			_ = json.NewEncoder(w).Encode([]string{"apps", "kube-system"})
		case strings.HasSuffix(r.URL.Path, "/logs"):
			hub.mu.Lock()
			hub.logQuery, hub.logPath = r.URL.Query(), r.URL.Path
			stream, status := hub.logStream, hub.logStatus
			hub.mu.Unlock()
			if status != 0 {
				w.WriteHeader(status)
				_ = json.NewEncoder(w).Encode(api.ErrorResponse{Error: "no agent is connected for this cluster"})
				return
			}
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = fmt.Fprint(w, stream)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	client, err := hubclient.New(server.URL, "token")
	if err != nil {
		t.Fatalf("hubclient.New() error = %v", err)
	}
	return client
}

func callTool(t *testing.T, client *mcpclient.Client, name string, arguments map[string]any) *protocol.CallToolResult {
	t.Helper()
	request := protocol.CallToolRequest{}
	request.Params.Name = name
	request.Params.Arguments = arguments
	result, err := client.CallTool(context.Background(), request)
	if err != nil {
		t.Fatalf("CallTool(%s) error = %v", name, err)
	}
	return result
}

func resultText(t *testing.T, result *protocol.CallToolResult) string {
	t.Helper()
	for _, content := range result.Content {
		if text, ok := protocol.AsTextContent(content); ok {
			return text.Text
		}
	}
	t.Fatalf("result has no text content: %#v", result.Content)
	return ""
}

func decodeResult[T any](t *testing.T, result *protocol.CallToolResult) T {
	t.Helper()
	if result.IsError {
		t.Fatalf("tool returned error: %s", resultText(t, result))
	}
	var value T
	if err := json.Unmarshal([]byte(resultText(t, result)), &value); err != nil {
		t.Fatalf("decode tool result error = %v", err)
	}
	return value
}

func TestListNamespacesTool(t *testing.T) {
	client := newTestClient(t, newFakeHub(t, &fakeHub{}))
	got := decodeResult[struct {
		Namespaces []string `json:"namespaces"`
		Count      int      `json:"count"`
	}](t, callTool(t, client, "list_namespaces", map[string]any{"cluster": "production"}))
	if got.Count != 2 || strings.Join(got.Namespaces, ",") != "apps,kube-system" {
		t.Fatalf("list_namespaces = %#v, want apps and kube-system", got)
	}
}

func TestGetPodLogsToolForwardsOptions(t *testing.T) {
	hub := &fakeHub{logStream: "data: starting\n\ndata: panic: boom\n\nevent: end\ndata: {}\n\n"}
	client := newTestClient(t, newFakeHub(t, hub))
	got := decodeResult[podLogs](t, callTool(t, client, "get_pod_logs", map[string]any{
		"cluster": "production", "namespace": "apps", "pod": "api-7f9", "container": "app",
		"tailLines": 40, "previous": true, "sinceSeconds": 300,
	}))
	if got.Logs != "starting\npanic: boom" || got.LineCount != 2 || got.Truncated {
		t.Fatalf("get_pod_logs = %#v, want two lines", got)
	}
	if got.Cluster.ID != "cluster-1" || got.Namespace != "apps" || got.Pod != "api-7f9" || got.Container != "app" || !got.Previous {
		t.Fatalf("get_pod_logs identity = %#v", got)
	}

	hub.mu.Lock()
	defer hub.mu.Unlock()
	if hub.logPath != "/api/v1/clusters/cluster-1/pods/apps/api-7f9/logs" {
		t.Fatalf("log path = %q", hub.logPath)
	}
	want := map[string]string{"follow": "false", "container": "app", "tailLines": "40", "previous": "true", "sinceSeconds": "300"}
	for key, value := range want {
		if got := hub.logQuery.Get(key); got != value {
			t.Errorf("query %s = %q, want %q", key, got, value)
		}
	}
}

func TestGetPodLogsToolDefaults(t *testing.T) {
	hub := &fakeHub{logStream: "event: end\ndata: {}\n\n"}
	client := newTestClient(t, newFakeHub(t, hub))
	got := decodeResult[podLogs](t, callTool(t, client, "get_pod_logs", map[string]any{
		"cluster": "cluster-1", "namespace": "apps", "pod": "api",
	}))
	if got.LineCount != 0 || got.Previous || !strings.Contains(got.Note, "previous=true") {
		t.Fatalf("get_pod_logs = %#v, want empty output with a previous=true hint", got)
	}
	hub.mu.Lock()
	defer hub.mu.Unlock()
	if got := hub.logQuery.Get("tailLines"); got != strconv.Itoa(defaultPodLogTailLines) {
		t.Fatalf("tailLines = %q, want default %d", got, defaultPodLogTailLines)
	}
	if hub.logQuery.Has("previous") || hub.logQuery.Has("sinceSeconds") || hub.logQuery.Has("container") {
		t.Fatalf("query = %v, want unset options omitted", hub.logQuery)
	}
}

func TestGetPodLogsToolRejectsInvalidArguments(t *testing.T) {
	client := newTestClient(t, newFakeHub(t, &fakeHub{}))
	tests := map[string]struct {
		arguments map[string]any
		want      string
	}{
		"missing pod":     {map[string]any{"cluster": "production", "namespace": "apps"}, "namespace and pod are required"},
		"missing cluster": {map[string]any{"namespace": "apps", "pod": "api"}, "cluster is required"},
		"unknown cluster": {map[string]any{"cluster": "staging", "namespace": "apps", "pod": "api"}, `cluster "staging" was not found`},
		"tail too large":  {map[string]any{"cluster": "production", "namespace": "apps", "pod": "api", "tailLines": maxPodLogTailLines + 1}, "tailLines must be between 1 and 1000"},
		"tail zero":       {map[string]any{"cluster": "production", "namespace": "apps", "pod": "api", "tailLines": 0}, "tailLines must be between 1 and 1000"},
		"negative since":  {map[string]any{"cluster": "production", "namespace": "apps", "pod": "api", "sinceSeconds": -1}, "sinceSeconds must be a positive"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := callTool(t, client, "get_pod_logs", tt.arguments)
			if !result.IsError || !strings.Contains(resultText(t, result), tt.want) {
				t.Fatalf("get_pod_logs result = %#v, want error containing %q", result, tt.want)
			}
		})
	}
}

func TestGetPodLogsToolReportsHubErrors(t *testing.T) {
	client := newTestClient(t, newFakeHub(t, &fakeHub{logStatus: http.StatusServiceUnavailable}))
	result := callTool(t, client, "get_pod_logs", map[string]any{"cluster": "production", "namespace": "apps", "pod": "api"})
	if !result.IsError || !strings.Contains(resultText(t, result), "no agent is connected") {
		t.Fatalf("get_pod_logs result = %#v, want hub error", result)
	}
}

func TestNewestLinesWithin(t *testing.T) {
	lines := []string{"aaaa", "bbbb", "cccc"}
	if kept, truncated := newestLinesWithin(lines, 14); truncated || len(kept) != 3 {
		t.Fatalf("newestLinesWithin(14) = %#v, %t, want all lines", kept, truncated)
	}
	kept, truncated := newestLinesWithin(lines, 13)
	if !truncated || strings.Join(kept, ",") != "bbbb,cccc" {
		t.Fatalf("newestLinesWithin(13) = %#v, %t, want newest two lines", kept, truncated)
	}
}

func TestGetPodLogsToolTruncatesLargeOutput(t *testing.T) {
	line := strings.Repeat("x", 1023)
	var stream strings.Builder
	total := maxPodLogBytes/1024 + 10
	for i := 0; i < total; i++ {
		stream.WriteString("data: " + line + "\n\n")
	}
	stream.WriteString("event: end\ndata: {}\n\n")
	client := newTestClient(t, newFakeHub(t, &fakeHub{logStream: stream.String()}))
	got := decodeResult[podLogs](t, callTool(t, client, "get_pod_logs", map[string]any{
		"cluster": "production", "namespace": "apps", "pod": "api", "tailLines": maxPodLogTailLines,
	}))
	if !got.Truncated || len(got.Logs) > maxPodLogBytes || got.LineCount != maxPodLogBytes/1024 {
		t.Fatalf("truncated=%t bytes=%d lines=%d, want output capped at %d bytes", got.Truncated, len(got.Logs), got.LineCount, maxPodLogBytes)
	}
	if !strings.Contains(got.Note, "dropped the 10 oldest lines") {
		t.Fatalf("note = %q, want dropped-line count", got.Note)
	}
}
