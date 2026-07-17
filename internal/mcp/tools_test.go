package mcp

import (
	"testing"

	"github.com/1solomonwakhungu/kfleet/pkg/types"
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
