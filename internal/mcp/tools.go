package mcp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/1solomonwakhungu/kfleet/internal/mcp/hubclient"
	"github.com/1solomonwakhungu/kfleet/pkg/api"
	"github.com/1solomonwakhungu/kfleet/pkg/types"
	protocol "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

const (
	highRestartCount = 5

	// defaultPodLogTailLines and maxPodLogTailLines keep get_pod_logs output
	// small enough for a model's context window by default.
	defaultPodLogTailLines = 100
	maxPodLogTailLines     = 1000
	// maxPodLogBytes caps the log text returned by get_pod_logs. When the
	// lines exceed it, the oldest lines are dropped so the newest output
	// (usually the failure) survives.
	maxPodLogBytes = 256 << 10
)

type toolHandlers struct {
	hub *hubclient.Client
}

type crashloopPod struct {
	ClusterID   string `json:"clusterId"`
	ClusterName string `json:"clusterName"`
	types.Pod
}

type diagnosis struct {
	Cluster       types.Cluster             `json:"cluster"`
	Status        api.ClusterStatusResponse `json:"status"`
	Crashloops    []crashloopPod            `json:"crashloopPods"`
	WarningEvents []types.Event             `json:"warningEvents"`
}

func registerTools(server *mcpserver.MCPServer, hub *hubclient.Client) {
	handlers := &toolHandlers{hub: hub}
	server.AddTools(
		mcpserver.ServerTool{
			Tool: protocol.NewTool("list_clusters",
				protocol.WithDescription("List every kfleet-managed cluster with its health, Kubernetes version, node count, and pod count.")),
			Handler: handlers.listClusters,
		},
		mcpserver.ServerTool{
			Tool: protocol.NewTool("get_cluster_status",
				protocol.WithDescription("Get the latest complete status snapshot and node details for a cluster by name or ID."),
				protocol.WithString("cluster", protocol.Required(), protocol.Description("Cluster name or ID"))),
			Handler: handlers.getClusterStatus,
		},
		mcpserver.ServerTool{
			Tool: protocol.NewTool("find_crashloop_pods",
				protocol.WithDescription("Find pods in CrashLoopBackOff or with at least five container restarts, across one cluster or the entire fleet."),
				protocol.WithString("cluster", protocol.Description("Optional cluster name or ID; omit to scan all clusters"))),
			Handler: handlers.findCrashloopPods,
		},
		mcpserver.ServerTool{
			Tool: protocol.NewTool("get_events",
				protocol.WithDescription("Get Kubernetes events for a cluster, optionally filtered by namespace and event type."),
				protocol.WithString("cluster", protocol.Required(), protocol.Description("Cluster name or ID")),
				protocol.WithString("namespace", protocol.Description("Optional Kubernetes namespace")),
				protocol.WithString("type", protocol.Enum("Normal", "Warning"), protocol.Description("Optional event type"))),
			Handler: handlers.getEvents,
		},
		mcpserver.ServerTool{
			Tool: protocol.NewTool("get_pods",
				protocol.WithDescription("List pods in a cluster, optionally limited to a Kubernetes namespace."),
				protocol.WithString("cluster", protocol.Required(), protocol.Description("Cluster name or ID")),
				protocol.WithString("namespace", protocol.Description("Optional Kubernetes namespace"))),
			Handler: handlers.getPods,
		},
		mcpserver.ServerTool{
			Tool: protocol.NewTool("diagnose_cluster",
				protocol.WithDescription("Diagnose a cluster by combining its status, crashlooping or frequently restarting pods, and recent Warning events."),
				protocol.WithString("cluster", protocol.Required(), protocol.Description("Cluster name or ID"))),
			Handler: handlers.diagnoseCluster,
		},
		mcpserver.ServerTool{
			Tool: protocol.NewTool("get_recent_events",
				protocol.WithDescription("Get recent operational timeline events (registration, approval, heartbeat state changes, version changes, reconnects, policy findings) for one cluster or the entire fleet."),
				protocol.WithString("cluster", protocol.Description("Optional cluster name or ID; omit for a fleet-wide timeline")),
				protocol.WithString("since", protocol.Description("Optional RFC3339 timestamp; only events at or after this time")),
				protocol.WithString("until", protocol.Description("Optional RFC3339 timestamp; only events strictly before this time")),
				protocol.WithNumber("before", protocol.Description("Optional positive cursor from nextCursor for the next older page")),
				protocol.WithNumber("limit", protocol.Description("Maximum number of events to return (default 50, max 500)"))),
			Handler: handlers.getTimeline,
		},
		mcpserver.ServerTool{
			Tool: protocol.NewTool("list_namespaces",
				protocol.WithDescription("List the Kubernetes namespaces in a cluster's latest snapshot."),
				protocol.WithString("cluster", protocol.Required(), protocol.Description("Cluster name or ID"))),
			Handler: handlers.listNamespaces,
		},
		mcpserver.ServerTool{
			Tool: protocol.NewTool("get_pod_logs",
				protocol.WithDescription("Fetch recent logs for a pod container. Set previous=true to read the last terminated instance, which holds the crash reason for a CrashLoopBackOff pod."),
				protocol.WithString("cluster", protocol.Required(), protocol.Description("Cluster name or ID")),
				protocol.WithString("namespace", protocol.Required(), protocol.Description("Kubernetes namespace of the pod")),
				protocol.WithString("pod", protocol.Required(), protocol.Description("Pod name")),
				protocol.WithString("container", protocol.Description("Optional container name; required for pods with more than one container")),
				protocol.WithNumber("tailLines", protocol.Description("Number of most recent lines to return (default 100, max 1000)")),
				protocol.WithBoolean("previous", protocol.Description("Read logs from the previous terminated container instance (default false)")),
				protocol.WithNumber("sinceSeconds", protocol.Description("Optional; only return lines newer than this many seconds"))),
			Handler: handlers.getPodLogs,
		},
	)
}

func (h *toolHandlers) listClusters(ctx context.Context, _ protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	clusters, err := h.hub.ListClusters(ctx)
	if err != nil {
		return toolError("list clusters", err), nil
	}
	return jsonResult(map[string]any{"clusters": clusters, "count": len(clusters)})
}

func (h *toolHandlers) getClusterStatus(ctx context.Context, request protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	cluster, result := h.requiredCluster(ctx, request)
	if result != nil {
		return result, nil
	}
	status, err := h.hub.GetClusterStatus(ctx, cluster.ID)
	if err != nil {
		return toolError("get cluster status", err), nil
	}
	return jsonResult(status)
}

func (h *toolHandlers) findCrashloopPods(ctx context.Context, request protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	clusterName := strings.TrimSpace(request.GetString("cluster", ""))
	clusters, err := h.hub.ListClusters(ctx)
	if err != nil {
		return toolError("list clusters", err), nil
	}
	if clusterName != "" {
		cluster, ok := findCluster(clusters, clusterName)
		if !ok {
			return protocol.NewToolResultError(fmt.Sprintf("cluster %q was not found", clusterName)), nil
		}
		clusters = []types.Cluster{cluster}
	}

	found := make([]crashloopPod, 0)
	for _, cluster := range clusters {
		pods, err := h.hub.GetPods(ctx, cluster.ID, "")
		if err != nil {
			return toolError("get pods for cluster "+cluster.Name, err), nil
		}
		found = append(found, crashloopPods(cluster, pods)...)
	}
	return jsonResult(map[string]any{"pods": found, "count": len(found)})
}

func (h *toolHandlers) getEvents(ctx context.Context, request protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	cluster, result := h.requiredCluster(ctx, request)
	if result != nil {
		return result, nil
	}
	events, err := h.hub.GetEvents(ctx, cluster.ID, request.GetString("namespace", ""))
	if err != nil {
		return toolError("get events", err), nil
	}
	eventType := request.GetString("type", "")
	if eventType != "" {
		filtered := make([]types.Event, 0, len(events))
		for _, event := range events {
			if strings.EqualFold(event.Type, eventType) {
				filtered = append(filtered, event)
			}
		}
		events = filtered
	}
	return jsonResult(map[string]any{"cluster": cluster, "events": events, "count": len(events)})
}

func (h *toolHandlers) getPods(ctx context.Context, request protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	cluster, result := h.requiredCluster(ctx, request)
	if result != nil {
		return result, nil
	}
	pods, err := h.hub.GetPods(ctx, cluster.ID, request.GetString("namespace", ""))
	if err != nil {
		return toolError("get pods", err), nil
	}
	return jsonResult(map[string]any{"cluster": cluster, "pods": pods, "count": len(pods)})
}

func (h *toolHandlers) diagnoseCluster(ctx context.Context, request protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	cluster, result := h.requiredCluster(ctx, request)
	if result != nil {
		return result, nil
	}
	status, err := h.hub.GetClusterStatus(ctx, cluster.ID)
	if err != nil {
		return toolError("get cluster status", err), nil
	}
	pods, err := h.hub.GetPods(ctx, cluster.ID, "")
	if err != nil {
		return toolError("get pods", err), nil
	}
	events, err := h.hub.GetEvents(ctx, cluster.ID, "")
	if err != nil {
		return toolError("get events", err), nil
	}
	warnings := make([]types.Event, 0)
	for _, event := range events {
		if strings.EqualFold(event.Type, "Warning") {
			warnings = append(warnings, event)
		}
	}
	return jsonResult(diagnosis{
		Cluster: cluster, Status: status,
		Crashloops: crashloopPods(cluster, pods), WarningEvents: warnings,
	})
}

func (h *toolHandlers) getTimeline(ctx context.Context, request protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	clusterID := ""
	if name := strings.TrimSpace(request.GetString("cluster", "")); name != "" {
		clusters, err := h.hub.ListClusters(ctx)
		if err != nil {
			return toolError("list clusters", err), nil
		}
		cluster, ok := findCluster(clusters, name)
		if !ok {
			return protocol.NewToolResultError(fmt.Sprintf("cluster %q was not found", name)), nil
		}
		clusterID = cluster.ID
	}

	query := hubclient.TimelineQuery{
		ClusterID: clusterID,
		Before:    int64(request.GetInt("before", 0)),
		Limit:     request.GetInt("limit", 0),
	}
	if query.Before < 0 {
		return protocol.NewToolResultError("before must be a positive cursor"), nil
	}
	if query.Limit < 0 || query.Limit > 500 {
		return protocol.NewToolResultError("limit must be between 1 and 500"), nil
	}
	if since := strings.TrimSpace(request.GetString("since", "")); since != "" {
		parsed, err := time.Parse(time.RFC3339, since)
		if err != nil {
			return protocol.NewToolResultError("since must be an RFC3339 timestamp"), nil
		}
		query.Since = &parsed
	}
	if until := strings.TrimSpace(request.GetString("until", "")); until != "" {
		parsed, err := time.Parse(time.RFC3339, until)
		if err != nil {
			return protocol.NewToolResultError("until must be an RFC3339 timestamp"), nil
		}
		query.Until = &parsed
	}

	page, err := h.hub.GetTimeline(ctx, query)
	if err != nil {
		return toolError("get timeline", err), nil
	}
	return jsonResult(map[string]any{"events": page.Events, "count": len(page.Events), "nextCursor": page.NextCursor})
}

func (h *toolHandlers) listNamespaces(ctx context.Context, request protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	cluster, result := h.requiredCluster(ctx, request)
	if result != nil {
		return result, nil
	}
	namespaces, err := h.hub.ListNamespaces(ctx, cluster.ID)
	if err != nil {
		return toolError("list namespaces", err), nil
	}
	return jsonResult(map[string]any{"cluster": cluster, "namespaces": namespaces, "count": len(namespaces)})
}

type podLogs struct {
	Cluster   types.Cluster `json:"cluster"`
	Namespace string        `json:"namespace"`
	Pod       string        `json:"pod"`
	Container string        `json:"container,omitempty"`
	Previous  bool          `json:"previous"`
	LineCount int           `json:"lineCount"`
	Truncated bool          `json:"truncated,omitempty"`
	Note      string        `json:"note,omitempty"`
	Logs      string        `json:"logs"`
}

func (h *toolHandlers) getPodLogs(ctx context.Context, request protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	namespace := strings.TrimSpace(request.GetString("namespace", ""))
	pod := strings.TrimSpace(request.GetString("pod", ""))
	if namespace == "" || pod == "" {
		return protocol.NewToolResultError("namespace and pod are required and must be non-empty strings"), nil
	}
	tailLines := request.GetInt("tailLines", defaultPodLogTailLines)
	if tailLines < 1 || tailLines > maxPodLogTailLines {
		return protocol.NewToolResultError(fmt.Sprintf("tailLines must be between 1 and %d", maxPodLogTailLines)), nil
	}
	sinceSeconds := request.GetInt("sinceSeconds", 0)
	if sinceSeconds < 0 {
		return protocol.NewToolResultError("sinceSeconds must be a positive number of seconds"), nil
	}
	cluster, result := h.requiredCluster(ctx, request)
	if result != nil {
		return result, nil
	}

	query := hubclient.PodLogsQuery{
		ClusterID:    cluster.ID,
		Namespace:    namespace,
		Pod:          pod,
		Container:    strings.TrimSpace(request.GetString("container", "")),
		TailLines:    int64(tailLines),
		Previous:     request.GetBool("previous", false),
		SinceSeconds: int64(sinceSeconds),
	}
	lines, err := h.hub.GetPodLogs(ctx, query)
	if err != nil {
		return toolError("get pod logs", err), nil
	}

	kept, truncated := newestLinesWithin(lines, maxPodLogBytes)
	output := podLogs{
		Cluster: cluster, Namespace: namespace, Pod: pod, Container: query.Container,
		Previous: query.Previous, LineCount: len(kept), Truncated: truncated,
		Logs: strings.Join(kept, "\n"),
	}
	switch {
	case truncated:
		output.Note = fmt.Sprintf("output exceeded %d KiB; dropped the %d oldest lines", maxPodLogBytes>>10, len(lines)-len(kept))
	case len(lines) == 0 && !query.Previous:
		output.Note = "no log output; for a crashlooping container, retry with previous=true"
	case len(lines) == 0:
		output.Note = "no log output"
	}
	return jsonResult(output)
}

// newestLinesWithin keeps the newest lines whose joined size fits in budget
// bytes and reports whether any older lines were dropped.
func newestLinesWithin(lines []string, budget int) ([]string, bool) {
	size := 0
	for i := len(lines) - 1; i >= 0; i-- {
		size += len(lines[i]) + 1
		if size > budget+1 {
			return lines[i+1:], true
		}
	}
	return lines, false
}

func (h *toolHandlers) requiredCluster(ctx context.Context, request protocol.CallToolRequest) (types.Cluster, *protocol.CallToolResult) {
	name, err := request.RequireString("cluster")
	if err != nil || strings.TrimSpace(name) == "" {
		return types.Cluster{}, protocol.NewToolResultError("cluster is required and must be a non-empty string")
	}
	clusters, err := h.hub.ListClusters(ctx)
	if err != nil {
		return types.Cluster{}, toolError("list clusters", err)
	}
	cluster, ok := findCluster(clusters, name)
	if !ok {
		return types.Cluster{}, protocol.NewToolResultError(fmt.Sprintf("cluster %q was not found", name))
	}
	return cluster, nil
}

func findCluster(clusters []types.Cluster, nameOrID string) (types.Cluster, bool) {
	for _, cluster := range clusters {
		if cluster.ID == nameOrID || strings.EqualFold(cluster.Name, nameOrID) {
			return cluster, true
		}
	}
	return types.Cluster{}, false
}

func crashloopPods(cluster types.Cluster, pods []types.Pod) []crashloopPod {
	found := make([]crashloopPod, 0)
	for _, pod := range pods {
		if strings.EqualFold(pod.Phase, "CrashLoopBackOff") || pod.RestartCount >= highRestartCount {
			found = append(found, crashloopPod{ClusterID: cluster.ID, ClusterName: cluster.Name, Pod: pod})
		}
	}
	return found
}

func jsonResult(value any) (*protocol.CallToolResult, error) {
	result, err := protocol.NewToolResultJSON(value)
	if err != nil {
		return toolError("encode tool result", err), nil
	}
	return result, nil
}

func toolError(action string, err error) *protocol.CallToolResult {
	return protocol.NewToolResultError(fmt.Sprintf("%s: %v", action, err))
}
