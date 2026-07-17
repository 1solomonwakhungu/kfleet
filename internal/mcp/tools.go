package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/1solomonwakhungu/kfleet/internal/mcp/hubclient"
	"github.com/1solomonwakhungu/kfleet/pkg/api"
	"github.com/1solomonwakhungu/kfleet/pkg/types"
	protocol "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

const highRestartCount = 5

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
