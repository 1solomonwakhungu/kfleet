package policy

import (
	"context"
	"testing"
	"time"

	"github.com/1solomonwakhungu/kfleet/pkg/types"
)

type fakeReader struct {
	clusters    map[string][]types.Cluster
	pods        map[string][]types.Pod
	deployments map[string][]types.Deployment
	namespaces  map[string][]string
	configs     map[string][]types.Namespace
}

func (f *fakeReader) ListClustersForTenant(_ context.Context, tenantID string) ([]types.Cluster, error) {
	return append([]types.Cluster(nil), f.clusters[tenantID]...), nil
}

func (f *fakeReader) ListPods(_ context.Context, clusterID, _ string) ([]types.Pod, error) {
	return append([]types.Pod(nil), f.pods[clusterID]...), nil
}

func (f *fakeReader) ListDeployments(_ context.Context, clusterID, _ string) ([]types.Deployment, error) {
	return append([]types.Deployment(nil), f.deployments[clusterID]...), nil
}

func (f *fakeReader) ListNamespaces(_ context.Context, clusterID string) ([]string, error) {
	return append([]string(nil), f.namespaces[clusterID]...), nil
}

func (f *fakeReader) ListNamespaceConfigs(_ context.Context, clusterID string) ([]types.Namespace, error) {
	return append([]types.Namespace(nil), f.configs[clusterID]...), nil
}

func TestEngineEvaluatesPassFailUnknownAndStale(t *testing.T) {
	now := time.Date(2026, time.July, 23, 12, 0, 0, 0, time.UTC)
	fresh := now.Add(-time.Minute)
	reader := &fakeReader{
		clusters: map[string][]types.Cluster{
			"tenant-a": {
				{ID: "a", Name: "alpha", Version: "v1.31.4", LastHeartbeat: fresh, Labels: requiredLabels()},
				{ID: "b", Name: "bravo", Version: "v1.31.7", LastHeartbeat: fresh, Labels: requiredLabels()},
				{ID: "c", Name: "charlie", LastHeartbeat: fresh, Labels: map[string]string{}},
				{ID: "d", Name: "delta", Version: "v1.30.9", LastHeartbeat: now.Add(-time.Hour), Labels: requiredLabels()},
			},
		},
		namespaces: map[string][]string{
			"a": {"apps", "default"},
			"b": {"apps", "default"},
		},
		deployments: map[string][]types.Deployment{
			"a": {{
				Name: "api", Namespace: "apps", ConfigHash: "same", Images: []string{"api:v1"},
				DesiredReplicas: 2, UpdatedReplicas: 2, AvailableReplicas: 2,
			}},
			"b": {{
				Name: "api", Namespace: "apps", ConfigHash: "different", Images: []string{"api:v2"},
				DesiredReplicas: 2, UpdatedReplicas: 1, AvailableReplicas: 1,
			}},
			"c": {{
				Name: "api", Namespace: "apps", ConfigHash: "same", Images: []string{"api:v1"},
				DesiredReplicas: 1, UpdatedReplicas: 1, AvailableReplicas: 1,
			}},
		},
		pods: map[string][]types.Pod{
			"a": {{
				Name: "api", Namespace: "apps", SecurityContextKnown: true,
				RunAsNonRoot: true, ReadOnlyRootFilesystem: true, CapabilitiesDroppedAll: true,
			}},
			"b": {{
				Name: "api", Namespace: "apps", SecurityContextKnown: true,
				Privileged: true, AllowsPrivilegeEscalation: true,
			}},
			"c": {{Name: "legacy", Namespace: "apps"}},
		},
		configs: map[string][]types.Namespace{
			"a": {{Name: "apps", Labels: map[string]string{"pod-security.kubernetes.io/enforce": "restricted"}}},
			"b": {{Name: "apps", Labels: map[string]string{}}},
		},
	}

	engine := NewEngine(reader, 5*time.Minute)
	engine.now = func() time.Time { return now }
	results, err := engine.Evaluate(context.Background(), "tenant-a")
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}

	for _, status := range []types.PolicyStatus{types.PolicyPass, types.PolicyFail, types.PolicyUnknown, types.PolicyStale} {
		if countStatus(results, status) == 0 {
			t.Errorf("results have no %q outcome: %#v", status, results)
		}
	}
	assertResult(t, results, KubernetesVersionPolicyID, "a", "", types.PolicyPass)
	assertResult(t, results, KubernetesVersionPolicyID, "d", "", types.PolicyStale)
	assertResult(t, results, RequiredLabelsPolicyID, "c", "", types.PolicyFail)
	assertResult(t, results, WorkloadDriftPolicyID, "b", "api", types.PolicyFail)
	assertResult(t, results, WorkloadHealthPolicyID, "b", "api", types.PolicyFail)
	assertResult(t, results, PodSecurityPolicyID, "a", "api", types.PolicyPass)
	assertResult(t, results, PodSecurityPolicyID, "b", "api", types.PolicyFail)
	assertResult(t, results, PodSecurityPolicyID, "c", "legacy", types.PolicyUnknown)
	assertResult(t, results, NamespaceSecurityPolicyID, "a", "apps", types.PolicyPass)
	assertResult(t, results, NamespaceSecurityPolicyID, "b", "apps", types.PolicyFail)

	summary := Summarize(results, 4, now)
	if summary.Total != len(results) || summary.ClusterCount != 4 {
		t.Fatalf("summary = %#v, want total %d and 4 clusters", summary, len(results))
	}
	for _, status := range []types.PolicyStatus{types.PolicyPass, types.PolicyFail, types.PolicyUnknown, types.PolicyStale} {
		if summary.ByStatus[status] != countStatus(results, status) {
			t.Errorf("summary status %q = %d, want %d", status, summary.ByStatus[status], countStatus(results, status))
		}
	}
}

func TestEngineNeverReadsAnotherTenant(t *testing.T) {
	now := time.Now().UTC()
	reader := &fakeReader{
		clusters: map[string][]types.Cluster{
			"tenant-a": {{ID: "a", Name: "alpha", LastHeartbeat: now, Labels: requiredLabels()}},
			"tenant-b": {{ID: "b", Name: "bravo", LastHeartbeat: now, Labels: requiredLabels()}},
		},
		namespaces: map[string][]string{"a": {"apps"}, "b": {"secret"}},
	}
	engine := NewEngine(reader, time.Hour)
	engine.now = func() time.Time { return now }

	results, err := engine.Evaluate(context.Background(), "tenant-a")
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	for _, result := range results {
		if result.Subject.ClusterID == "b" || result.Subject.ClusterName == "bravo" {
			t.Fatalf("tenant-a result leaked tenant-b subject: %#v", result)
		}
	}
}

func TestCatalogReturnsIndependentCopy(t *testing.T) {
	first := Catalog()
	if len(first) != 7 {
		t.Fatalf("Catalog() length = %d, want 7", len(first))
	}
	first[0].Name = "changed"
	if Catalog()[0].Name == "changed" {
		t.Fatal("Catalog() returned mutable package state")
	}
}

func requiredLabels() map[string]string {
	return map[string]string{"environment": "test", "owner": "platform", "region": "us-central1"}
}

func countStatus(results []types.PolicyResult, status types.PolicyStatus) int {
	count := 0
	for _, result := range results {
		if result.Status == status {
			count++
		}
	}
	return count
}

func assertResult(
	t *testing.T,
	results []types.PolicyResult,
	policyID, clusterID, subjectName string,
	status types.PolicyStatus,
) {
	t.Helper()
	for _, result := range results {
		if result.PolicyID == policyID && result.Subject.ClusterID == clusterID &&
			(subjectName == "" || result.Subject.Name == subjectName) {
			if result.Status != status {
				t.Fatalf("%s/%s/%s status = %q, want %q: %#v", policyID, clusterID, subjectName, result.Status, status, result)
			}
			return
		}
	}
	t.Fatalf("missing result %s/%s/%s", policyID, clusterID, subjectName)
}
