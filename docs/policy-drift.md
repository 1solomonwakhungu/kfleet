# Policy and configuration drift

kfleet evaluates a fixed catalog of read-only policies against the latest durable agent snapshots. Evaluation never calls a Kubernetes mutation API and the hub exposes no policy create, update, remediate, or delete endpoint.

## Result model

Each policy declares a category, severity, and scope. Severities are `low`, `medium`, `high`, and `critical`. Scopes are `fleet`, `cluster`, `namespace`, and `workload`.

Every evaluation produces one of four states:

| State | Meaning |
| --- | --- |
| `pass` | The reported evidence satisfies the policy. |
| `fail` | The reported evidence explicitly differs from the expected policy or fleet baseline. |
| `unknown` | Required evidence is missing, or a workload has no peer for a drift comparison. |
| `stale` | The cluster snapshot is older than three configured heartbeat intervals. |

`unknown` and `stale` are intentionally distinct. Missing evidence cannot be treated as compliance, and old evidence cannot represent current state.

## Built-in policies

| ID | Severity | Scope | Check |
| --- | --- | --- | --- |
| `kubernetes-version-consistency` | high | fleet | Compares every fresh cluster's Kubernetes major and minor version with the tenant fleet mode. |
| `required-cluster-labels` | medium | cluster | Requires non-empty `environment`, `owner`, and `region` cluster labels. |
| `namespace-consistency` | medium | namespace | Compares each cluster's namespace set with the tenant fleet mode. |
| `workload-configuration-consistency` | high | workload | Compares deterministic deployment desired-configuration hashes for matching namespace and deployment names. |
| `workload-availability` | high | workload | Requires desired, updated, and available deployment replica counts to agree. |
| `pod-security-baseline` | critical | workload | Requires application pods to run as non-root, disable privilege escalation, use a read-only root filesystem, drop all capabilities, avoid privileged mode, and avoid host namespaces. |
| `namespace-pod-security` | high | namespace | Requires application namespaces to enforce the Kubernetes `baseline` or `restricted` Pod Security Standard. |

The security checks exclude Kubernetes and kfleet system namespaces. Workload hashes include the complete desired Deployment spec and therefore include images, replicas, selectors, strategy, and pod-template configuration.

The baseline for a drift policy is the most common fresh value inside one tenant. Ties use lexical order so repeated evaluations are deterministic. A matching workload that exists in only one cluster is `unknown`, because there is no peer evidence.

## REST API

All endpoints are GET-only:

```text
GET /api/v1/policies
GET /api/v1/policies/results
GET /api/v1/policies/summary
GET /api/v1/clusters/{id}/policy-results
```

`GET /api/v1/policy-results` and `GET /api/v1/drift` are compatibility aliases for the results endpoint.

Results accept optional exact-match filters:

```text
?status=fail
?severity=critical
?scope=workload
?policyId=pod-security-baseline
?clusterId=<cluster-id>
```

The response contains `results` and a `summary` with total, status, severity, cluster, and evaluation-time fields. Invalid status, severity, scope, or tenant values return `400`. A cluster outside the current tenant returns `404`, not a cross-tenant authorization hint.

## Tenant isolation

The hub stores a `tenant_id` on every cluster and scopes inventory, pending-agent, policy, drift-baseline, and cluster policy reads before resource data is loaded. Tenant A's baselines and results cannot include Tenant B's clusters.

Single-tenant installations use `default`. Multi-tenant installations set `X-Kfleet-Tenant-ID` to a lowercase identifier containing letters, numbers, dots, underscores, or hyphens. The agent sends the configured `KFLEET_TENANT_ID`; the Helm chart exposes it as `tenant.id`.

The header is an isolation context, not authentication by itself. At a multi-tenant ingress, an authenticated trusted proxy must remove any client-supplied value and set the verified tenant ID. This preserves a clean boundary for the authentication layer while preventing accidental cross-tenant queries inside kfleet.

Example:

```bash
curl -H 'X-Kfleet-Tenant-ID: platform-a' \
  'http://localhost:8080/api/v1/policies/results?status=fail'
```

## Snapshot compatibility

New agents report namespace labels, deployment hashes and images, and normalized pod security posture. Existing agents remain accepted. Policies that require fields unavailable from an older agent return `unknown` rather than `pass` or `fail`.
