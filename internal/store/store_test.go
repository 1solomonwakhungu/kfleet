package store

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/1solomonwakhungu/kfleet/pkg/types"
)

func TestSQLiteStoreClusterLifecycle(t *testing.T) {
	t.Parallel()

	st, err := Open(filepath.Join(t.TempDir(), "kfleet.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := st.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})

	ctx := context.Background()
	registeredAt := time.Date(2025, time.January, 2, 3, 4, 5, 0, time.UTC)
	cluster := types.Cluster{
		ID:           "cluster-1",
		Name:         "production",
		Health:       types.HealthUnknown,
		RegisteredAt: registeredAt,
		Labels:       map[string]string{"region": "us-central1"},
	}

	if err := st.CreateCluster(ctx, cluster); err != nil {
		t.Fatalf("CreateCluster() error = %v", err)
	}

	got, err := st.GetCluster(ctx, cluster.ID)
	if err != nil {
		t.Fatalf("GetCluster() error = %v", err)
	}
	if !reflect.DeepEqual(got, cluster) {
		t.Fatalf("GetCluster() = %#v, want %#v", got, cluster)
	}

	clusters, err := st.ListClusters(ctx)
	if err != nil {
		t.Fatalf("ListClusters() error = %v", err)
	}
	if len(clusters) != 1 || !reflect.DeepEqual(clusters[0], cluster) {
		t.Fatalf("ListClusters() = %#v, want [%#v]", clusters, cluster)
	}

	heartbeat := registeredAt.Add(time.Minute)
	if err := st.UpdateHealth(ctx, cluster.ID, types.HealthHealthy, heartbeat); err != nil {
		t.Fatalf("UpdateHealth() error = %v", err)
	}
	got, err = st.GetCluster(ctx, cluster.ID)
	if err != nil {
		t.Fatalf("GetCluster() after UpdateHealth error = %v", err)
	}
	if got.Health != types.HealthHealthy || !got.LastHeartbeat.Equal(heartbeat) {
		t.Fatalf("health update = (%q, %v), want (%q, %v)", got.Health, got.LastHeartbeat, types.HealthHealthy, heartbeat)
	}

	if err := st.UpdateSnapshot(ctx, cluster.ID, 3, 42, "v1.31.1"); err != nil {
		t.Fatalf("UpdateSnapshot() error = %v", err)
	}
	got, err = st.GetCluster(ctx, cluster.ID)
	if err != nil {
		t.Fatalf("GetCluster() after UpdateSnapshot error = %v", err)
	}
	if got.NodeCount != 3 || got.PodCount != 42 || got.Version != "v1.31.1" {
		t.Fatalf("snapshot update = (%d, %d, %q), want (3, 42, %q)", got.NodeCount, got.PodCount, got.Version, "v1.31.1")
	}

	if err := st.DeleteCluster(ctx, cluster.ID); err != nil {
		t.Fatalf("DeleteCluster() error = %v", err)
	}
	_, err = st.GetCluster(ctx, cluster.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetCluster() after delete error = %v, want ErrNotFound", err)
	}
}

func TestSQLiteStoreAgentApprovalLifecycle(t *testing.T) {
	t.Parallel()

	st, err := Open(filepath.Join(t.TempDir(), "kfleet.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := st.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})

	ctx := context.Background()
	cluster := types.Cluster{
		ID:           "cluster-agent",
		Name:         "agent-cluster",
		Health:       types.HealthUnknown,
		RegisteredAt: time.Now().UTC(),
		Labels:       map[string]string{},
	}
	if err := st.CreateCluster(ctx, cluster); err != nil {
		t.Fatalf("CreateCluster() error = %v", err)
	}
	if err := st.IssueAgentToken(ctx, cluster.ID, "hash-one"); err != nil {
		t.Fatalf("IssueAgentToken() error = %v", err)
	}

	approved, err := st.ValidateAgentToken(ctx, cluster.ID, "hash-one")
	if err != nil || approved {
		t.Fatalf("ValidateAgentToken() = (%v, %v), want (false, nil)", approved, err)
	}
	if _, err := st.ValidateAgentToken(ctx, cluster.ID, "wrong-hash"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("ValidateAgentToken() wrong hash error = %v, want ErrNotFound", err)
	}

	pending, err := st.ListPendingAgents(ctx)
	if err != nil {
		t.Fatalf("ListPendingAgents() error = %v", err)
	}
	if len(pending) != 1 || pending[0].ID != cluster.ID {
		t.Fatalf("ListPendingAgents() = %#v, want cluster %q", pending, cluster.ID)
	}

	if err := st.ApproveAgent(ctx, cluster.ID); err != nil {
		t.Fatalf("ApproveAgent() error = %v", err)
	}
	approved, err = st.ValidateAgentToken(ctx, cluster.ID, "hash-one")
	if err != nil || !approved {
		t.Fatalf("ValidateAgentToken() = (%v, %v), want (true, nil)", approved, err)
	}
	pending, err = st.ListPendingAgents(ctx)
	if err != nil || len(pending) != 0 {
		t.Fatalf("ListPendingAgents() after approval = (%#v, %v), want empty", pending, err)
	}

	if err := st.IssueAgentToken(ctx, cluster.ID, "hash-two"); err != nil {
		t.Fatalf("IssueAgentToken() rotation error = %v", err)
	}
	if _, err := st.ValidateAgentToken(ctx, cluster.ID, "hash-one"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("ValidateAgentToken() old token error = %v, want ErrNotFound", err)
	}
	approved, err = st.ValidateAgentToken(ctx, cluster.ID, "hash-two")
	if err != nil || !approved {
		t.Fatalf("ValidateAgentToken() rotated = (%v, %v), want approval preserved", approved, err)
	}

	if err := st.ApproveAgent(ctx, "missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("ApproveAgent() missing error = %v, want ErrNotFound", err)
	}
	if err := st.IssueAgentToken(ctx, "missing", "hash"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("IssueAgentToken() missing error = %v, want ErrNotFound", err)
	}
}

func TestSQLiteStoreReplaceSnapshot(t *testing.T) {
	t.Parallel()

	st, err := Open(filepath.Join(t.TempDir(), "kfleet.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := st.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})

	ctx := context.Background()
	cluster := types.Cluster{
		ID:           "snapshot-cluster",
		Name:         "snapshot",
		Health:       types.HealthUnknown,
		RegisteredAt: time.Now().UTC(),
		Labels:       map[string]string{},
	}
	if err := st.CreateCluster(ctx, cluster); err != nil {
		t.Fatalf("CreateCluster() error = %v", err)
	}
	if err := st.IssueAgentToken(ctx, cluster.ID, "token-hash"); err != nil {
		t.Fatalf("IssueAgentToken() error = %v", err)
	}

	firstHeartbeat := time.Date(2026, time.July, 19, 12, 0, 0, 0, time.UTC)
	first := types.ClusterSnapshot{
		Nodes: []types.Node{
			{Name: "node-a", Status: "Ready", Ready: true, Roles: []string{"worker"}},
			{Name: "node-b", Status: "NotReady", Ready: false, Roles: []string{}},
		},
		Pods: []types.Pod{
			{Name: "api", Namespace: "apps", Phase: "Running"},
			{Name: "dns", Namespace: "kube-system", Phase: "Running"},
		},
		Services:    []types.Service{{Name: "api", Namespace: "apps", Type: "ClusterIP"}},
		Deployments: []types.Deployment{{Name: "api", Namespace: "apps", DesiredReplicas: 2}},
		Events:      []types.Event{{Namespace: "apps", Reason: "Started", Message: "started"}},
	}
	if err := st.ReplaceSnapshot(ctx, cluster.ID, first, "v1.31.1", "0.1.0", types.HealthDegraded, firstHeartbeat); err != nil {
		t.Fatalf("ReplaceSnapshot(first) error = %v", err)
	}
	invalid := first
	invalid.Nodes = []types.Node{
		{Name: "duplicate", Status: "Ready", Ready: true},
		{Name: "duplicate", Status: "Ready", Ready: true},
	}
	if err := st.ReplaceSnapshot(ctx, cluster.ID, invalid, "invalid", "invalid", types.HealthHealthy, firstHeartbeat.Add(30*time.Second)); err == nil {
		t.Fatal("ReplaceSnapshot(invalid) error = nil, want duplicate node error")
	}
	unchangedNodes, err := st.ListNodes(ctx, cluster.ID)
	if err != nil || len(unchangedNodes) != 2 || unchangedNodes[0].Name != "node-a" || unchangedNodes[1].Name != "node-b" {
		t.Fatalf("ListNodes() after rollback = (%#v, %v), want original nodes", unchangedNodes, err)
	}
	unchangedCluster, err := st.GetCluster(ctx, cluster.ID)
	if err != nil || unchangedCluster.Version != "v1.31.1" || !unchangedCluster.LastHeartbeat.Equal(firstHeartbeat) {
		t.Fatalf("GetCluster() after rollback = (%#v, %v), want original metadata", unchangedCluster, err)
	}

	secondHeartbeat := firstHeartbeat.Add(time.Minute)
	second := types.ClusterSnapshot{
		Nodes:       []types.Node{{Name: "node-c", Status: "Ready", Ready: true, Roles: []string{}}},
		Pods:        []types.Pod{{Name: "worker", Namespace: "batch", Phase: "Running"}},
		Services:    []types.Service{},
		Deployments: []types.Deployment{},
		Events:      []types.Event{},
	}
	if err := st.ReplaceSnapshot(ctx, cluster.ID, second, "v1.32.0", "0.2.0", types.HealthHealthy, secondHeartbeat); err != nil {
		t.Fatalf("ReplaceSnapshot(second) error = %v", err)
	}

	gotCluster, err := st.GetCluster(ctx, cluster.ID)
	if err != nil {
		t.Fatalf("GetCluster() error = %v", err)
	}
	if gotCluster.NodeCount != 1 || gotCluster.PodCount != 1 || gotCluster.Version != "v1.32.0" || gotCluster.AgentVersion != "0.2.0" ||
		gotCluster.Health != types.HealthHealthy || !gotCluster.LastHeartbeat.Equal(secondHeartbeat) {
		t.Fatalf("cluster metadata = %#v, want replacement metadata", gotCluster)
	}

	nodes, err := st.ListNodes(ctx, cluster.ID)
	if err != nil || len(nodes) != 1 || nodes[0].Name != "node-c" {
		t.Fatalf("ListNodes() = (%#v, %v), want only node-c", nodes, err)
	}
	pods, err := st.ListPods(ctx, cluster.ID, "")
	if err != nil || len(pods) != 1 || pods[0].Name != "worker" {
		t.Fatalf("ListPods() = (%#v, %v), want only worker", pods, err)
	}
	appsPods, err := st.ListPods(ctx, cluster.ID, "apps")
	if err != nil || len(appsPods) != 0 {
		t.Fatalf("ListPods(apps) = (%#v, %v), want empty", appsPods, err)
	}
	namespaces, err := st.ListNamespaces(ctx, cluster.ID)
	if err != nil || !reflect.DeepEqual(namespaces, []string{"batch"}) {
		t.Fatalf("ListNamespaces() = (%#v, %v), want [batch]", namespaces, err)
	}
	services, err := st.ListServices(ctx, cluster.ID, "")
	if err != nil || len(services) != 0 {
		t.Fatalf("ListServices() = (%#v, %v), want empty", services, err)
	}
	deployments, err := st.ListDeployments(ctx, cluster.ID, "")
	if err != nil || len(deployments) != 0 {
		t.Fatalf("ListDeployments() = (%#v, %v), want empty", deployments, err)
	}
	events, err := st.ListEvents(ctx, cluster.ID, "")
	if err != nil || len(events) != 0 {
		t.Fatalf("ListEvents() = (%#v, %v), want empty", events, err)
	}
}

func TestSQLiteStoreTenantIsolationAndPolicyEvidence(t *testing.T) {
	t.Parallel()

	st, err := Open(filepath.Join(t.TempDir(), "kfleet.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := st.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})

	ctx := context.Background()
	now := time.Now().UTC()
	for _, cluster := range []types.Cluster{
		{ID: "a", TenantID: "tenant-a", Name: "alpha", Health: types.HealthHealthy, RegisteredAt: now, Labels: map[string]string{}},
		{ID: "b", TenantID: "tenant-b", Name: "alpha", Health: types.HealthHealthy, RegisteredAt: now, Labels: map[string]string{}},
	} {
		if err := st.CreateCluster(ctx, cluster); err != nil {
			t.Fatalf("CreateCluster(%s) error = %v", cluster.ID, err)
		}
	}

	clusters, err := st.ListClustersForTenant(ctx, "tenant-a")
	if err != nil || len(clusters) != 1 || clusters[0].ID != "a" {
		t.Fatalf("ListClustersForTenant(tenant-a) = (%#v, %v), want only a", clusters, err)
	}
	if _, err := st.GetClusterForTenant(ctx, "tenant-a", "b"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetClusterForTenant(tenant-a, b) error = %v, want ErrNotFound", err)
	}

	snapshot := types.ClusterSnapshot{
		Pods: []types.Pod{{
			Name: "api", Namespace: "apps", SecurityContextKnown: true,
			RunAsNonRoot: true, ReadOnlyRootFilesystem: true, CapabilitiesDroppedAll: true,
		}},
		Deployments: []types.Deployment{{
			Name: "api", Namespace: "apps", ConfigHash: "abc123", Images: []string{"example/api:v1"},
		}},
		Namespaces: []types.Namespace{{
			Name: "apps", Labels: map[string]string{"pod-security.kubernetes.io/enforce": "restricted"},
		}},
	}
	if err := st.ReplaceSnapshot(ctx, "a", snapshot, "v1.31.1", "0.2.0", types.HealthHealthy, now); err != nil {
		t.Fatalf("ReplaceSnapshot() error = %v", err)
	}
	pods, err := st.ListPods(ctx, "a", "")
	if err != nil || len(pods) != 1 || !pods[0].SecurityContextKnown || !pods[0].RunAsNonRoot ||
		!pods[0].ReadOnlyRootFilesystem || !pods[0].CapabilitiesDroppedAll {
		t.Fatalf("ListPods() policy evidence = (%#v, %v)", pods, err)
	}
	deployments, err := st.ListDeployments(ctx, "a", "")
	if err != nil || len(deployments) != 1 || deployments[0].ConfigHash != "abc123" ||
		!reflect.DeepEqual(deployments[0].Images, []string{"example/api:v1"}) {
		t.Fatalf("ListDeployments() policy evidence = (%#v, %v)", deployments, err)
	}
	namespaces, err := st.ListNamespaceConfigs(ctx, "a")
	if err != nil || len(namespaces) != 1 ||
		namespaces[0].Labels["pod-security.kubernetes.io/enforce"] != "restricted" {
		t.Fatalf("ListNamespaceConfigs() = (%#v, %v)", namespaces, err)
	}
}
