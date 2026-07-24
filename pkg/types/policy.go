package types

import "time"

// PolicySeverity indicates the operational impact of a policy failure.
type PolicySeverity string

const (
	SeverityLow      PolicySeverity = "low"
	SeverityMedium   PolicySeverity = "medium"
	SeverityHigh     PolicySeverity = "high"
	SeverityCritical PolicySeverity = "critical"
)

// PolicyScope identifies the smallest object evaluated by a policy.
type PolicyScope string

const (
	ScopeFleet     PolicyScope = "fleet"
	ScopeCluster   PolicyScope = "cluster"
	ScopeNamespace PolicyScope = "namespace"
	ScopeWorkload  PolicyScope = "workload"
)

// PolicyStatus is the outcome of a read-only policy evaluation.
type PolicyStatus string

const (
	PolicyPass    PolicyStatus = "pass"
	PolicyFail    PolicyStatus = "fail"
	PolicyUnknown PolicyStatus = "unknown"
	PolicyStale   PolicyStatus = "stale"
)

// Policy is an immutable check provided by the hub.
type Policy struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Category    string         `json:"category"`
	Severity    PolicySeverity `json:"severity"`
	Scope       PolicyScope    `json:"scope"`
}

// PolicySubject identifies the tenant-owned object that was evaluated.
type PolicySubject struct {
	ClusterID   string `json:"clusterId,omitempty"`
	ClusterName string `json:"clusterName,omitempty"`
	Namespace   string `json:"namespace,omitempty"`
	Kind        string `json:"kind,omitempty"`
	Name        string `json:"name,omitempty"`
}

// PolicyResult records evidence from one point-in-time evaluation.
type PolicyResult struct {
	PolicyID    string            `json:"policyId"`
	PolicyName  string            `json:"policyName"`
	Category    string            `json:"category"`
	Severity    PolicySeverity    `json:"severity"`
	Scope       PolicyScope       `json:"scope"`
	Status      PolicyStatus      `json:"status"`
	Subject     PolicySubject     `json:"subject"`
	Message     string            `json:"message"`
	Expected    map[string]string `json:"expected,omitempty"`
	Actual      map[string]string `json:"actual,omitempty"`
	ObservedAt  time.Time         `json:"observedAt,omitempty"`
	EvaluatedAt time.Time         `json:"evaluatedAt"`
}

// PolicySummary contains fleet-wide counts for a tenant-safe result set.
type PolicySummary struct {
	Total        int                    `json:"total"`
	ByStatus     map[PolicyStatus]int   `json:"byStatus"`
	BySeverity   map[PolicySeverity]int `json:"bySeverity"`
	ClusterCount int                    `json:"clusterCount"`
	EvaluatedAt  time.Time              `json:"evaluatedAt"`
}
