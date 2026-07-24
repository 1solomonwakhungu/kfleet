// Package policy evaluates built-in, read-only fleet policies.
package policy

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/1solomonwakhungu/kfleet/pkg/types"
)

const (
	KubernetesVersionPolicyID = "kubernetes-version-consistency"
	RequiredLabelsPolicyID    = "required-cluster-labels"
	NamespaceDriftPolicyID    = "namespace-consistency"
	WorkloadDriftPolicyID     = "workload-configuration-consistency"
	WorkloadHealthPolicyID    = "workload-availability"
	PodSecurityPolicyID       = "pod-security-baseline"
	NamespaceSecurityPolicyID = "namespace-pod-security"
)

var catalog = []types.Policy{
	{
		ID: KubernetesVersionPolicyID, Name: "Kubernetes version consistency",
		Description: "Clusters should run the fleet's most common Kubernetes major and minor version.",
		Category:    "Kubernetes", Severity: types.SeverityHigh, Scope: types.ScopeFleet,
	},
	{
		ID: RequiredLabelsPolicyID, Name: "Required cluster labels",
		Description: "Clusters should declare environment, owner, and region labels.",
		Category:    "Governance", Severity: types.SeverityMedium, Scope: types.ScopeCluster,
	},
	{
		ID: NamespaceDriftPolicyID, Name: "Namespace consistency",
		Description: "Each cluster should have the tenant fleet's common namespace set.",
		Category:    "Namespaces", Severity: types.SeverityMedium, Scope: types.ScopeNamespace,
	},
	{
		ID: WorkloadDriftPolicyID, Name: "Workload configuration consistency",
		Description: "Matching deployments should have the same desired configuration across clusters.",
		Category:    "Workloads", Severity: types.SeverityHigh, Scope: types.ScopeWorkload,
	},
	{
		ID: WorkloadHealthPolicyID, Name: "Workload availability",
		Description: "Deployments should have all desired replicas updated and available.",
		Category:    "Workloads", Severity: types.SeverityHigh, Scope: types.ScopeWorkload,
	},
	{
		ID: PodSecurityPolicyID, Name: "Pod security baseline",
		Description: "Pods should run as non-root with restricted privileges, a read-only root filesystem, and all capabilities dropped.",
		Category:    "Security", Severity: types.SeverityCritical, Scope: types.ScopeWorkload,
	},
	{
		ID: NamespaceSecurityPolicyID, Name: "Namespace pod security",
		Description: "Application namespaces should enforce the Kubernetes baseline or restricted Pod Security Standard.",
		Category:    "Security", Severity: types.SeverityHigh, Scope: types.ScopeNamespace,
	},
}

var requiredClusterLabels = []string{"environment", "owner", "region"}

// Reader is the tenant-scoped, read-only snapshot surface required by the engine.
type Reader interface {
	ListClustersForTenant(ctx context.Context, tenantID string) ([]types.Cluster, error)
	ListPods(ctx context.Context, clusterID, namespace string) ([]types.Pod, error)
	ListDeployments(ctx context.Context, clusterID, namespace string) ([]types.Deployment, error)
	ListNamespaces(ctx context.Context, clusterID string) ([]string, error)
	ListNamespaceConfigs(ctx context.Context, clusterID string) ([]types.Namespace, error)
}

// Engine evaluates policies without changing cluster or Kubernetes state.
type Engine struct {
	reader     Reader
	staleAfter time.Duration
	now        func() time.Time
}

// NewEngine creates an evaluator. staleAfter defaults to five minutes.
func NewEngine(reader Reader, staleAfter time.Duration) *Engine {
	if staleAfter <= 0 {
		staleAfter = 5 * time.Minute
	}
	return &Engine{reader: reader, staleAfter: staleAfter, now: time.Now}
}

// Catalog returns a defensive copy of the built-in policy definitions.
func Catalog() []types.Policy {
	return append([]types.Policy(nil), catalog...)
}

// Evaluate evaluates all policies against clusters belonging to exactly one tenant.
func (e *Engine) Evaluate(ctx context.Context, tenantID string) ([]types.PolicyResult, error) {
	clusters, err := e.reader.ListClustersForTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list tenant clusters: %w", err)
	}
	now := e.now().UTC()
	results := make([]types.PolicyResult, 0)

	results = append(results, e.evaluateVersions(clusters, now)...)
	results = append(results, e.evaluateLabels(clusters, now)...)

	namespaceResults, err := e.evaluateNamespaces(ctx, clusters, now)
	if err != nil {
		return nil, err
	}
	results = append(results, namespaceResults...)

	workloadResults, err := e.evaluateWorkloads(ctx, clusters, now)
	if err != nil {
		return nil, err
	}
	results = append(results, workloadResults...)

	securityResults, err := e.evaluateSecurity(ctx, clusters, now)
	if err != nil {
		return nil, err
	}
	results = append(results, securityResults...)

	sort.SliceStable(results, func(i, j int) bool {
		left, right := results[i], results[j]
		if left.PolicyID != right.PolicyID {
			return left.PolicyID < right.PolicyID
		}
		if left.Subject.ClusterName != right.Subject.ClusterName {
			return left.Subject.ClusterName < right.Subject.ClusterName
		}
		if left.Subject.Namespace != right.Subject.Namespace {
			return left.Subject.Namespace < right.Subject.Namespace
		}
		if left.Subject.Kind != right.Subject.Kind {
			return left.Subject.Kind < right.Subject.Kind
		}
		return left.Subject.Name < right.Subject.Name
	})
	return results, nil
}

// Summarize produces stable counters for a result set.
func Summarize(results []types.PolicyResult, clusterCount int, evaluatedAt time.Time) types.PolicySummary {
	summary := types.PolicySummary{
		Total:        len(results),
		ByStatus:     map[types.PolicyStatus]int{},
		BySeverity:   map[types.PolicySeverity]int{},
		ClusterCount: clusterCount,
		EvaluatedAt:  evaluatedAt.UTC(),
	}
	for _, status := range []types.PolicyStatus{types.PolicyPass, types.PolicyFail, types.PolicyUnknown, types.PolicyStale} {
		summary.ByStatus[status] = 0
	}
	for _, severity := range []types.PolicySeverity{types.SeverityLow, types.SeverityMedium, types.SeverityHigh, types.SeverityCritical} {
		summary.BySeverity[severity] = 0
	}
	for _, result := range results {
		summary.ByStatus[result.Status]++
		summary.BySeverity[result.Severity]++
	}
	return summary
}

func (e *Engine) evaluateVersions(clusters []types.Cluster, now time.Time) []types.PolicyResult {
	policy := policyByID(KubernetesVersionPolicyID)
	freshVersions := make([]string, 0, len(clusters))
	for _, cluster := range clusters {
		if !e.isStale(cluster, now) && versionLine(cluster.Version) != "" {
			freshVersions = append(freshVersions, versionLine(cluster.Version))
		}
	}
	expected := mode(freshVersions)
	results := make([]types.PolicyResult, 0, len(clusters))
	for _, cluster := range clusters {
		result := newResult(policy, cluster, now)
		actual := versionLine(cluster.Version)
		switch {
		case e.isStale(cluster, now):
			setStale(&result, e.staleAfter)
		case actual == "":
			result.Status = types.PolicyUnknown
			result.Message = "Kubernetes version was not reported"
		case expected == "":
			result.Status = types.PolicyUnknown
			result.Message = "No fresh fleet version baseline is available"
		case actual == expected:
			result.Status = types.PolicyPass
			result.Message = "Kubernetes version matches the fleet baseline"
			result.Expected = map[string]string{"majorMinor": expected}
			result.Actual = map[string]string{"majorMinor": actual, "version": cluster.Version}
		default:
			result.Status = types.PolicyFail
			result.Message = "Kubernetes version differs from the fleet baseline"
			result.Expected = map[string]string{"majorMinor": expected}
			result.Actual = map[string]string{"majorMinor": actual, "version": cluster.Version}
		}
		results = append(results, result)
	}
	return results
}

func (e *Engine) evaluateLabels(clusters []types.Cluster, now time.Time) []types.PolicyResult {
	policy := policyByID(RequiredLabelsPolicyID)
	results := make([]types.PolicyResult, 0, len(clusters))
	for _, cluster := range clusters {
		result := newResult(policy, cluster, now)
		if e.isStale(cluster, now) {
			setStale(&result, e.staleAfter)
			results = append(results, result)
			continue
		}
		missing := make([]string, 0)
		for _, label := range requiredClusterLabels {
			if strings.TrimSpace(cluster.Labels[label]) == "" {
				missing = append(missing, label)
			}
		}
		result.Expected = map[string]string{"labels": strings.Join(requiredClusterLabels, ",")}
		if len(missing) == 0 {
			result.Status = types.PolicyPass
			result.Message = "All required cluster labels are present"
		} else {
			result.Status = types.PolicyFail
			result.Message = "Required cluster labels are missing"
			result.Actual = map[string]string{"missing": strings.Join(missing, ",")}
		}
		results = append(results, result)
	}
	return results
}

func (e *Engine) evaluateNamespaces(ctx context.Context, clusters []types.Cluster, now time.Time) ([]types.PolicyResult, error) {
	policy := policyByID(NamespaceDriftPolicyID)
	sets := make(map[string][]string, len(clusters))
	signatures := make([]string, 0, len(clusters))
	for _, cluster := range clusters {
		if e.isStale(cluster, now) {
			continue
		}
		namespaces, err := e.reader.ListNamespaces(ctx, cluster.ID)
		if err != nil {
			return nil, fmt.Errorf("list namespaces for %s: %w", cluster.ID, err)
		}
		sort.Strings(namespaces)
		sets[cluster.ID] = namespaces
		if len(namespaces) > 0 {
			signatures = append(signatures, strings.Join(namespaces, "\x00"))
		}
	}
	baselineSignature := mode(signatures)
	baseline := splitSignature(baselineSignature)
	results := make([]types.PolicyResult, 0, len(clusters))
	for _, cluster := range clusters {
		result := newResult(policy, cluster, now)
		if e.isStale(cluster, now) {
			setStale(&result, e.staleAfter)
		} else if len(sets[cluster.ID]) == 0 || baselineSignature == "" {
			result.Status = types.PolicyUnknown
			result.Message = "No namespace inventory is available"
		} else {
			missing, extra := setDifference(baseline, sets[cluster.ID]), setDifference(sets[cluster.ID], baseline)
			result.Expected = map[string]string{"namespaces": strings.Join(baseline, ",")}
			result.Actual = map[string]string{"namespaces": strings.Join(sets[cluster.ID], ",")}
			if len(missing) == 0 && len(extra) == 0 {
				result.Status = types.PolicyPass
				result.Message = "Namespace inventory matches the fleet baseline"
			} else {
				result.Status = types.PolicyFail
				result.Message = "Namespace inventory differs from the fleet baseline"
				result.Actual["missing"] = strings.Join(missing, ",")
				result.Actual["extra"] = strings.Join(extra, ",")
			}
		}
		results = append(results, result)
	}
	return results, nil
}

type clusterDeployments struct {
	cluster     types.Cluster
	deployments []types.Deployment
}

func (e *Engine) evaluateWorkloads(ctx context.Context, clusters []types.Cluster, now time.Time) ([]types.PolicyResult, error) {
	driftPolicy := policyByID(WorkloadDriftPolicyID)
	healthPolicy := policyByID(WorkloadHealthPolicyID)
	all := make([]clusterDeployments, 0, len(clusters))
	hashes := make(map[string][]string)
	for _, cluster := range clusters {
		if e.isStale(cluster, now) {
			all = append(all, clusterDeployments{cluster: cluster})
			continue
		}
		deployments, err := e.reader.ListDeployments(ctx, cluster.ID, "")
		if err != nil {
			return nil, fmt.Errorf("list deployments for %s: %w", cluster.ID, err)
		}
		all = append(all, clusterDeployments{cluster: cluster, deployments: deployments})
		for _, deployment := range deployments {
			if deployment.ConfigHash != "" {
				key := deployment.Namespace + "\x00" + deployment.Name
				hashes[key] = append(hashes[key], deployment.ConfigHash)
			}
		}
	}

	results := make([]types.PolicyResult, 0)
	for _, item := range all {
		if e.isStale(item.cluster, now) {
			drift := newResult(driftPolicy, item.cluster, now)
			setStale(&drift, e.staleAfter)
			results = append(results, drift)
			health := newResult(healthPolicy, item.cluster, now)
			setStale(&health, e.staleAfter)
			results = append(results, health)
			continue
		}
		if len(item.deployments) == 0 {
			drift := newResult(driftPolicy, item.cluster, now)
			drift.Status = types.PolicyUnknown
			drift.Message = "No deployment configuration was reported"
			results = append(results, drift)
			health := newResult(healthPolicy, item.cluster, now)
			health.Status = types.PolicyUnknown
			health.Message = "No deployments were reported"
			results = append(results, health)
			continue
		}
		for _, deployment := range item.deployments {
			subject := workloadSubject(item.cluster, "Deployment", deployment.Namespace, deployment.Name)
			drift := newResultWithSubject(driftPolicy, subject, item.cluster.LastHeartbeat, now)
			key := deployment.Namespace + "\x00" + deployment.Name
			baseline := mode(hashes[key])
			switch {
			case deployment.ConfigHash == "":
				drift.Status = types.PolicyUnknown
				drift.Message = "Deployment configuration hash was not reported"
			case len(hashes[key]) < 2:
				drift.Status = types.PolicyUnknown
				drift.Message = "No peer deployment is available for drift comparison"
				drift.Actual = map[string]string{"configHash": deployment.ConfigHash}
			case deployment.ConfigHash == baseline:
				drift.Status = types.PolicyPass
				drift.Message = "Deployment configuration matches its fleet peers"
				drift.Expected = map[string]string{"configHash": baseline}
				drift.Actual = map[string]string{"configHash": deployment.ConfigHash, "images": strings.Join(deployment.Images, ",")}
			default:
				drift.Status = types.PolicyFail
				drift.Message = "Deployment configuration differs from its fleet peers"
				drift.Expected = map[string]string{"configHash": baseline}
				drift.Actual = map[string]string{"configHash": deployment.ConfigHash, "images": strings.Join(deployment.Images, ",")}
			}
			results = append(results, drift)

			health := newResultWithSubject(healthPolicy, subject, item.cluster.LastHeartbeat, now)
			health.Expected = map[string]string{
				"updatedReplicas":   fmt.Sprint(deployment.DesiredReplicas),
				"availableReplicas": fmt.Sprint(deployment.DesiredReplicas),
			}
			health.Actual = map[string]string{
				"desiredReplicas":   fmt.Sprint(deployment.DesiredReplicas),
				"updatedReplicas":   fmt.Sprint(deployment.UpdatedReplicas),
				"availableReplicas": fmt.Sprint(deployment.AvailableReplicas),
			}
			if deployment.DesiredReplicas == deployment.UpdatedReplicas && deployment.DesiredReplicas == deployment.AvailableReplicas {
				health.Status = types.PolicyPass
				health.Message = "All desired deployment replicas are updated and available"
			} else {
				health.Status = types.PolicyFail
				health.Message = "Deployment replicas are not fully updated and available"
			}
			results = append(results, health)
		}
	}
	return results, nil
}

func (e *Engine) evaluateSecurity(ctx context.Context, clusters []types.Cluster, now time.Time) ([]types.PolicyResult, error) {
	podPolicy := policyByID(PodSecurityPolicyID)
	namespacePolicy := policyByID(NamespaceSecurityPolicyID)
	results := make([]types.PolicyResult, 0)
	for _, cluster := range clusters {
		if e.isStale(cluster, now) {
			for _, policy := range []types.Policy{podPolicy, namespacePolicy} {
				result := newResult(policy, cluster, now)
				setStale(&result, e.staleAfter)
				results = append(results, result)
			}
			continue
		}

		pods, err := e.reader.ListPods(ctx, cluster.ID, "")
		if err != nil {
			return nil, fmt.Errorf("list pods for %s: %w", cluster.ID, err)
		}
		applicationPods := make([]types.Pod, 0, len(pods))
		for _, pod := range pods {
			if !isSystemNamespace(pod.Namespace) {
				applicationPods = append(applicationPods, pod)
			}
		}
		if len(applicationPods) == 0 {
			result := newResult(podPolicy, cluster, now)
			result.Status = types.PolicyUnknown
			result.Message = "No pod security evidence was reported"
			results = append(results, result)
		}
		for _, pod := range applicationPods {
			result := newResultWithSubject(
				podPolicy,
				workloadSubject(cluster, "Pod", pod.Namespace, pod.Name),
				cluster.LastHeartbeat,
				now,
			)
			if !pod.SecurityContextKnown {
				result.Status = types.PolicyUnknown
				result.Message = "Pod security context was not reported by the agent"
			} else {
				violations := podSecurityViolations(pod)
				result.Expected = map[string]string{"profile": "restricted"}
				if len(violations) == 0 {
					result.Status = types.PolicyPass
					result.Message = "Pod meets the built-in restricted security profile"
				} else {
					result.Status = types.PolicyFail
					result.Message = "Pod violates the built-in restricted security profile"
					result.Actual = map[string]string{"violations": strings.Join(violations, ",")}
				}
			}
			results = append(results, result)
		}

		namespaces, err := e.reader.ListNamespaceConfigs(ctx, cluster.ID)
		if err != nil {
			return nil, fmt.Errorf("list namespace configs for %s: %w", cluster.ID, err)
		}
		applicationNamespaces := make([]types.Namespace, 0, len(namespaces))
		for _, namespace := range namespaces {
			if !isSystemNamespace(namespace.Name) {
				applicationNamespaces = append(applicationNamespaces, namespace)
			}
		}
		if len(applicationNamespaces) == 0 {
			result := newResult(namespacePolicy, cluster, now)
			result.Status = types.PolicyUnknown
			result.Message = "No application namespace label evidence was reported"
			results = append(results, result)
		}
		for _, namespace := range applicationNamespaces {
			subject := types.PolicySubject{
				ClusterID: cluster.ID, ClusterName: cluster.Name,
				Namespace: namespace.Name, Kind: "Namespace", Name: namespace.Name,
			}
			result := newResultWithSubject(namespacePolicy, subject, cluster.LastHeartbeat, now)
			enforce := namespace.Labels["pod-security.kubernetes.io/enforce"]
			result.Expected = map[string]string{"pod-security.kubernetes.io/enforce": "baseline|restricted"}
			result.Actual = map[string]string{"pod-security.kubernetes.io/enforce": enforce}
			if enforce == "baseline" || enforce == "restricted" {
				result.Status = types.PolicyPass
				result.Message = "Namespace enforces a supported Pod Security Standard"
			} else {
				result.Status = types.PolicyFail
				result.Message = "Namespace does not enforce the baseline or restricted Pod Security Standard"
			}
			results = append(results, result)
		}
	}
	return results, nil
}

func (e *Engine) isStale(cluster types.Cluster, now time.Time) bool {
	return cluster.LastHeartbeat.IsZero() || now.Sub(cluster.LastHeartbeat) > e.staleAfter
}

func newResult(policy types.Policy, cluster types.Cluster, now time.Time) types.PolicyResult {
	return newResultWithSubject(policy, types.PolicySubject{
		ClusterID: cluster.ID, ClusterName: cluster.Name,
	}, cluster.LastHeartbeat, now)
}

func newResultWithSubject(policy types.Policy, subject types.PolicySubject, observedAt, now time.Time) types.PolicyResult {
	return types.PolicyResult{
		PolicyID: policy.ID, PolicyName: policy.Name, Category: policy.Category,
		Severity: policy.Severity, Scope: policy.Scope, Subject: subject,
		ObservedAt: observedAt, EvaluatedAt: now,
	}
}

func setStale(result *types.PolicyResult, staleAfter time.Duration) {
	result.Status = types.PolicyStale
	result.Message = "Cluster snapshot is older than the freshness threshold"
	result.Expected = map[string]string{"freshWithin": staleAfter.String()}
}

func workloadSubject(cluster types.Cluster, kind, namespace, name string) types.PolicySubject {
	return types.PolicySubject{
		ClusterID: cluster.ID, ClusterName: cluster.Name,
		Namespace: namespace, Kind: kind, Name: name,
	}
}

func policyByID(id string) types.Policy {
	for _, policy := range catalog {
		if policy.ID == id {
			return policy
		}
	}
	panic("unknown built-in policy: " + id)
}

func versionLine(version string) string {
	version = strings.TrimPrefix(strings.TrimSpace(version), "v")
	parts := strings.Split(version, ".")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return ""
	}
	return parts[0] + "." + parts[1]
}

func mode(values []string) string {
	counts := make(map[string]int)
	best, bestCount := "", 0
	for _, value := range values {
		if value == "" {
			continue
		}
		counts[value]++
		if counts[value] > bestCount || (counts[value] == bestCount && value < best) {
			best, bestCount = value, counts[value]
		}
	}
	return best
}

func splitSignature(signature string) []string {
	if signature == "" {
		return nil
	}
	return strings.Split(signature, "\x00")
}

func setDifference(left, right []string) []string {
	rightSet := make(map[string]struct{}, len(right))
	for _, value := range right {
		rightSet[value] = struct{}{}
	}
	result := make([]string, 0)
	for _, value := range left {
		if _, ok := rightSet[value]; !ok {
			result = append(result, value)
		}
	}
	return result
}

func podSecurityViolations(pod types.Pod) []string {
	violations := make([]string, 0)
	if pod.Privileged {
		violations = append(violations, "privileged")
	}
	if !pod.RunAsNonRoot {
		violations = append(violations, "runAsNonRoot")
	}
	if !pod.ReadOnlyRootFilesystem {
		violations = append(violations, "readOnlyRootFilesystem")
	}
	if pod.AllowsPrivilegeEscalation {
		violations = append(violations, "allowPrivilegeEscalation")
	}
	if !pod.CapabilitiesDroppedAll {
		violations = append(violations, "dropAllCapabilities")
	}
	if pod.HostNetwork {
		violations = append(violations, "hostNetwork")
	}
	if pod.HostPID {
		violations = append(violations, "hostPID")
	}
	if pod.HostIPC {
		violations = append(violations, "hostIPC")
	}
	return violations
}

func isSystemNamespace(namespace string) bool {
	return namespace == "kube-system" || namespace == "kube-public" ||
		namespace == "kube-node-lease" || namespace == "kfleet-system"
}
