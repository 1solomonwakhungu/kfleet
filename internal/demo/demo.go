// Package demo provides the isolated synthetic dataset used by the public demo.
package demo

import (
	"context"
	"fmt"
	"time"

	"github.com/1solomonwakhungu/kfleet/internal/store"
	"github.com/1solomonwakhungu/kfleet/pkg/types"
)

const (
	ClusterAlphaID   = "demo-cluster-alpha"
	ClusterBravoID   = "demo-cluster-bravo"
	ClusterCharlieID = "demo-cluster-charlie"
)

// Open creates a fresh in-memory store. It never opens the configured hub
// database, so demo mode cannot reveal persisted cluster identities.
func Open(ctx context.Context) (store.Store, error) {
	st, err := store.Open(":memory:")
	if err != nil {
		return nil, err
	}
	if err := Seed(ctx, st, time.Now().UTC()); err != nil {
		_ = st.Close()
		return nil, err
	}
	return st, nil
}

// Seed writes a deterministic set of clearly synthetic clusters and snapshots.
func Seed(ctx context.Context, st store.Store, now time.Time) error {
	for _, item := range fixtures(now) {
		if err := st.CreateCluster(ctx, item.cluster); err != nil {
			return fmt.Errorf("create synthetic cluster %q: %w", item.cluster.ID, err)
		}
		if err := st.ReplaceSnapshot(
			ctx,
			item.cluster.ID,
			item.snapshot,
			item.cluster.Version,
			item.cluster.AgentVersion,
			item.cluster.Health,
			item.cluster.LastHeartbeat,
		); err != nil {
			return fmt.Errorf("seed synthetic snapshot %q: %w", item.cluster.ID, err)
		}
	}
	return nil
}

type fixture struct {
	cluster  types.Cluster
	snapshot types.ClusterSnapshot
}

func fixtures(now time.Time) []fixture {
	now = now.Truncate(time.Second)
	return []fixture{
		{
			cluster: types.Cluster{
				ID: ClusterAlphaID, Name: "demo-us-central", Health: types.HealthHealthy,
				Version: "v1.31.4", AgentVersion: "demo-1.0.0",
				RegisteredAt: now.Add(-45 * 24 * time.Hour), LastHeartbeat: now.Add(-18 * time.Second),
				Labels: map[string]string{"data": "synthetic", "environment": "demo", "region": "sample-central"},
			},
			snapshot: types.ClusterSnapshot{
				Nodes: []types.Node{
					{Name: "demo-alpha-control-1", Status: "Ready", Roles: []string{"control-plane"}, Version: "v1.31.4", CPUCapacity: "4", MemoryCapacity: "16Gi", Ready: true},
					{Name: "demo-alpha-worker-1", Status: "Ready", Roles: []string{"worker"}, Version: "v1.31.4", CPUCapacity: "8", MemoryCapacity: "32Gi", Ready: true},
					{Name: "demo-alpha-worker-2", Status: "Ready", Roles: []string{"worker"}, Version: "v1.31.4", CPUCapacity: "8", MemoryCapacity: "32Gi", Ready: true},
				},
				Pods: []types.Pod{
					demoPod("storefront", "catalog-api-7b9d-demo", "demo-alpha-worker-1", now.Add(-9*time.Hour), true, 0),
					demoPod("storefront", "checkout-api-5c84-demo", "demo-alpha-worker-2", now.Add(-7*time.Hour), true, 0),
					demoPod("observability", "metrics-agent-demo", "demo-alpha-worker-1", now.Add(-28*time.Hour), true, 1),
					demoPod("kube-system", "dns-demo-a", "demo-alpha-control-1", now.Add(-38*time.Hour), true, 0),
				},
				Services: []types.Service{
					demoService("catalog-api", "storefront", "10.240.0.20", 8080),
					demoService("checkout-api", "storefront", "10.240.0.21", 8080),
				},
				Deployments: []types.Deployment{
					demoDeployment("catalog-api", "storefront", 3, 3, "9h"),
					demoDeployment("checkout-api", "storefront", 2, 2, "7h"),
				},
				Events: []types.Event{
					demoEvent(ClusterAlphaID, "storefront", "ScalingReplicaSet", "Synthetic rollout completed for the catalog sample.", "Normal", now.Add(-34*time.Minute)),
				},
			},
		},
		{
			cluster: types.Cluster{
				ID: ClusterBravoID, Name: "demo-eu-west", Health: types.HealthDegraded,
				Version: "v1.30.8", AgentVersion: "demo-1.0.0",
				RegisteredAt: now.Add(-30 * 24 * time.Hour), LastHeartbeat: now.Add(-42 * time.Second),
				Labels: map[string]string{"data": "synthetic", "environment": "demo", "region": "sample-west"},
			},
			snapshot: types.ClusterSnapshot{
				Nodes: []types.Node{
					{Name: "demo-bravo-control-1", Status: "Ready", Roles: []string{"control-plane"}, Version: "v1.30.8", CPUCapacity: "4", MemoryCapacity: "16Gi", Ready: true},
					{Name: "demo-bravo-worker-1", Status: "NotReady", Roles: []string{"worker"}, Version: "v1.30.8", CPUCapacity: "8", MemoryCapacity: "32Gi", Ready: false},
				},
				Pods: []types.Pod{
					demoPod("payments", "ledger-api-6f75-demo", "demo-bravo-worker-1", now.Add(-3*time.Hour), false, 4),
					demoPod("payments", "queue-worker-798c-demo", "demo-bravo-control-1", now.Add(-16*time.Hour), true, 0),
					demoPod("kube-system", "dns-demo-b", "demo-bravo-control-1", now.Add(-24*time.Hour), true, 0),
				},
				Services: []types.Service{
					demoService("ledger-api", "payments", "10.241.0.30", 8080),
				},
				Deployments: []types.Deployment{
					demoDeployment("ledger-api", "payments", 3, 2, "3h"),
					demoDeployment("queue-worker", "payments", 2, 2, "16h"),
				},
				Events: []types.Event{
					demoEvent(ClusterBravoID, "payments", "Unhealthy", "Synthetic readiness check failure for demonstration only.", "Warning", now.Add(-6*time.Minute)),
				},
			},
		},
		{
			cluster: types.Cluster{
				ID: ClusterCharlieID, Name: "demo-edge-lab", Health: types.HealthUnreachable,
				Version: "v1.29.12", AgentVersion: "demo-0.9.4",
				RegisteredAt: now.Add(-14 * 24 * time.Hour), LastHeartbeat: now.Add(-22 * time.Minute),
				Labels: map[string]string{"data": "synthetic", "environment": "demo", "region": "sample-edge"},
			},
			snapshot: types.ClusterSnapshot{
				Nodes: []types.Node{
					{Name: "demo-charlie-node-1", Status: "Unknown", Roles: []string{"control-plane", "worker"}, Version: "v1.29.12", CPUCapacity: "4", MemoryCapacity: "8Gi", Ready: false},
				},
				Pods: []types.Pod{
					demoPod("edge-sample", "telemetry-demo", "demo-charlie-node-1", now.Add(-13*time.Hour), false, 2),
				},
				Services: []types.Service{
					demoService("telemetry", "edge-sample", "10.242.0.40", 4318),
				},
				Deployments: []types.Deployment{
					demoDeployment("telemetry", "edge-sample", 1, 0, "13h"),
				},
				Events: []types.Event{
					demoEvent(ClusterCharlieID, "edge-sample", "NodeStatusUnknown", "Synthetic heartbeat gap illustrates an unreachable cluster.", "Warning", now.Add(-22*time.Minute)),
				},
			},
		},
	}
}

func demoPod(namespace, name, node string, started time.Time, ready bool, restarts int32) types.Pod {
	phase := "Running"
	if !ready {
		phase = "Pending"
	}
	return types.Pod{Name: name, Namespace: namespace, Phase: phase, NodeName: node, RestartCount: restarts, Ready: ready, StartTime: started}
}

func demoService(name, namespace, clusterIP string, port int32) types.Service {
	return types.Service{
		Name: name, Namespace: namespace, Type: "ClusterIP", ClusterIP: clusterIP,
		ExternalIPs: []string{}, Ports: []types.ServicePort{{Name: "http", Port: port, TargetPort: port, Protocol: "TCP"}}, Age: "demo",
	}
}

func demoDeployment(name, namespace string, desired, ready int32, age string) types.Deployment {
	return types.Deployment{
		Name: name, Namespace: namespace, DesiredReplicas: desired, ReadyReplicas: ready,
		UpdatedReplicas: ready, AvailableReplicas: ready, Age: age,
	}
}

func demoEvent(clusterID, namespace, reason, message, eventType string, timestamp time.Time) types.Event {
	return types.Event{
		ClusterID: clusterID, Namespace: namespace, Reason: reason, Message: message,
		Type: eventType, Count: 1, LastTimestamp: timestamp,
	}
}
