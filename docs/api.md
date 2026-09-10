# REST API reference

This reference enumerates every HTTP route the hub serves. It is generated
from the route registrations in `internal/server` (the `register*Routes`
methods and `server.New`), so it matches the code on `main`.

- Base path: `/api/v1` for application routes; `/healthz`, `/readyz`, and
  `/ws/clusters` sit outside it.
- All request and response bodies are JSON (`application/json`) except where
  noted: the two streaming endpoints are WebSocket and Server-Sent Events.
- Every `/api/v1` route validates the `X-Kfleet-Tenant-ID` header for tenant
  scoping (see [Policy and configuration drift](policy-drift.md#tenant-isolation)).
  A missing header means the `default` tenant. Invalid values return `400`.
- Agent-facing routes authenticate with bearer tokens, never user sessions.
- There is no `/metrics` endpoint. Scraping targets are limited to the
  health probes described in [Operations](operations.md#health-and-monitoring).

## Authentication classes

| Class | How requests authenticate |
| --- | --- |
| Public | No credential. Served to anyone. |
| Session | `kfleet_session` cookie from `POST /api/v1/auth/login`. Browser mutations must also send `X-Kfleet-CSRF: 1`. |
| Agent bearer | `Authorization: Bearer <per-agent token>` issued at registration. |
| Registration token | `Authorization: Bearer <registration token>` (rotated token or static `KFLEET_REGISTRATION_TOKEN`). |
| Demo read-only | In `KFLEET_DEMO_MODE=true`, every read route runs as a synthetic `read_only` user without a session; `POST`/`PUT`/`PATCH`/`DELETE` are rejected with `405` before routing. |

Role requirements use the hierarchy `read_only` < `operator` < `admin`;
"operator or higher" includes admins. A request that fails a role check
returns `403` with a message that names the required role.

## Error format

Every API error uses the same JSON contract:

```json
{ "error": "cluster not found", "code": 404 }
```

Unknown `/api` paths return this shape with `404`, and known paths requested
with the wrong method return `405` (with the mux-computed `Allow` header).
Role denials return `403` ("this action requires the admin role"), and
browser mutations without the CSRF header return `403`
("CSRF validation failed"). `405` also covers demo mode, where every mutation
is rejected with "public demo is read-only".

## Health and readiness

| Method | Path | Auth | Response | Purpose |
| --- | --- | --- | --- | --- |
| `GET` | `/healthz` | Public | `200` `ok` | Liveness probe; pure process check. |
| `GET` | `/readyz` | Public | `200` `ready`, or `503` `{"status":"unavailable","reason":"store unavailable"}` | Readiness probe; pings the SQLite store with a 2s timeout. |

| Method | Path | Auth | Response | Purpose |
| --- | --- | --- | --- | --- |
| `GET` | `/api/v1/meta` | Public | `RuntimeInfo` | Reports the runtime safety posture. |

`RuntimeInfo` returns `demoMode`, `readOnly`, `syntheticData`, and
`dataPolicy`; all three booleans are `true` only in demo mode.

## Real-time streams

| Method | Path | Auth | Purpose |
| --- | --- | --- | --- |
| `GET` | `/ws/clusters` | Session | WebSocket that sends one `snapshot` frame per cluster on connect, then `registered`, `health_changed`, `snapshot`, and `deleted` updates as they happen. Frames are `{"type":"...","cluster":{...}}`. |
| `GET` | `/api/v1/clusters/{id}/pods/{namespace}/{pod}/logs` | Session | SSE pod-log stream for the browser. Query parameters and frames are documented in [Pod log streaming](pod-logs.md). A cluster with no connected agent answers `503` with the standard JSON error. |
| `GET` | `/api/v1/agents/{id}/logs` | Agent bearer | The agent's outbound WebSocket reverse channel. Frames are `LogStreamMessage` JSON: hub→agent `start` and `stop`, agent→hub `data` and `end`. See [Pod log streaming](pod-logs.md). |

## Auth and users

| Method | Path | Auth | Request | Response | Purpose |
| --- | --- | --- | --- | --- | --- |
| `POST` | `/api/v1/auth/login` | Public | `{username, password}` | `200` user object; sets the `kfleet_session` cookie | Start a session. Wrong credentials return `401`. |
| `POST` | `/api/v1/auth/logout` | Session | — | `204` | Delete the current session. |
| `GET` | `/api/v1/auth/me` | Session | — | `200` user object | Current user's identity and role. |
| `GET` | `/api/v1/users` | Admin | — | `200` `{users:[...]}` | List accounts. |
| `POST` | `/api/v1/users` | Admin | `{username, email, password, role}` | `201` user object | Create a user. Password must be 12–72 bytes; role is `admin`, `operator`, or `read_only`; `409` when username or email exists. |
| `PATCH` | `/api/v1/users/{id}` | Admin | `{role, disabled}` | `200` user object | Change role or disable. `409` when the change would remove the last enabled admin. |
| `DELETE` | `/api/v1/users/{id}` | Admin | — | `204` | Delete a user. `409` for self-deletion or last-admin removal. |
| `PATCH` | `/api/v1/users/{id}/password` | Self (any role), or Admin for others | `{currentPassword?, newPassword}` | `204` | Self-changes must verify `currentPassword` (`401` when wrong); admins omit it. Passwords must be 12–72 bytes. Invalidates all sessions for the target account. |
| `GET` | `/api/v1/audit?limit=100` | Admin | — | `200` `{events:[...]}` | Newest audit entries. `limit` is 1–1000, default 100. |
| `POST` | `/api/v1/admin/registration-token/rotate` | Admin | — | `200` `{token}` | Mint a new registration token; shown exactly once, only the hash is stored. |

The user object is `{id, username, email, role, disabled, createdAt,
updatedAt}` and never includes password hashes. See
[authentication](authentication.md) for the full role matrix and password
lifecycle.

## Clusters and inventory

| Method | Path | Auth | Request | Response | Purpose |
| --- | --- | --- | --- | --- | --- |
| `GET` | `/api/v1/clusters` | Session | — | `200` `{clusters:[...]}` | List clusters in the tenant. |
| `GET` | `/api/v1/clusters/{id}` | Session | — | `200` cluster | One cluster by ID; `404` when unknown or outside the tenant. |
| `POST` | `/api/v1/clusters/register` | Operator or higher | `{name, labels}` | `201` `{clusterId, token}` | Operator-driven registration; the agent token is approved immediately and shown once. |
| `DELETE` | `/api/v1/clusters/{id}` | Operator or higher | — | `204` | Remove a cluster and its inventory. |
| `GET` | `/api/v1/clusters/{id}/status` | Session | — | `200` `{cluster, nodes}` | Cluster summary with node list. |
| `GET` | `/api/v1/clusters/{id}/pods` | Session | — | `200` pod array | Latest snapshot pods; optional `?namespace=`. |
| `GET` | `/api/v1/clusters/{id}/services` | Session | — | `200` service array | Latest snapshot services; optional `?namespace=`. |
| `GET` | `/api/v1/clusters/{id}/deployments` | Session | — | `200` deployment array | Latest snapshot deployments; optional `?namespace=`. |
| `GET` | `/api/v1/clusters/{id}/namespaces` | Session | — | `200` namespace array | Latest snapshot namespaces. |
| `GET` | `/api/v1/clusters/{id}/events` | Session | — | `200` event array | Kubernetes events from the latest snapshot; optional `?namespace=`. |

## Agent endpoints

These routes are for agents and the humans who approve them. Agents never use
user sessions.

| Method | Path | Auth | Request | Response | Purpose |
| --- | --- | --- | --- | --- | --- |
| `POST` | `/api/v1/agents/register` | Registration token | `{name, labels, agentVersion?, k8sVersion?, existingAgentToken?}` | `201` `{clusterId, token}` while pending, `200` when already approved | Self-service registration. Fails closed with `403` when no registration credential is configured; `401` for a wrong token; `409` when the cluster name exists and `existingAgentToken` does not prove the current credential. |
| `GET` | `/api/v1/agents/pending` | Session | — | `200` `{clusters:[...]}` | Clusters awaiting approval. |
| `POST` | `/api/v1/agents/{id}/approve` | Operator or higher | — | `200` cluster | Approve a pending agent. `404` when unknown. |
| `POST` | `/api/v1/agents/heartbeat` | Agent bearer | `{clusterId, nodeCount, podCount, healthyNodes, version}` | `200` cluster | Legacy count-based heartbeat; updates health and snapshot counters. `403` while pending approval. |
| `POST` | `/api/v1/agents/{id}/heartbeat` | Agent bearer | — | `200` `{clusterId, approved}` | Liveness heartbeat; also how a pending agent learns it was approved. |
| `POST` | `/api/v1/agents/{id}/deregister` | Agent bearer | — | `200` | Mark the cluster unreachable on agent shutdown. |
| `POST` | `/api/v1/clusters/{id}/status` | Agent bearer | Full snapshot (nodes, pods, services, deployments, namespaces, events, versions) | `200` cluster | Snapshot ingestion. Bodies over 4 MiB return `413`; malformed JSON returns `400`. |

## Timeline and policy findings

| Method | Path | Auth | Purpose |
| --- | --- | --- | --- |
| `GET` | `/api/v1/timeline` | Session | Fleet-wide operational events. |
| `GET` | `/api/v1/clusters/{id}/timeline` | Session | Operational events for one cluster ID or name. |
| `POST` | `/api/v1/clusters/{id}/policy-findings` | Agent bearer | Append an idempotent policy finding to the timeline. |

Timeline reads return `{ "events": [...], "nextCursor": 123 }` newest first
(`nextCursor` is omitted on the last page). Query parameters:

- `limit`: page size, 1–500, default 50 (`400` outside that range)
- `before`: positive `nextCursor` from the previous page
- `since`: inclusive RFC3339 timestamp
- `until`: exclusive RFC3339 timestamp (`400` when not after `since`)

Policy-finding ingestion requires `ruleId`, `resource`, and `message`;
`severity`, `details`, and `occurredAt` are optional. Bodies over 64 KiB
return `413`. A retry with identical content returns `200` instead of
duplicating; a first-time finding returns `201`.

## Alerts

| Method | Path | Auth | Request | Response | Purpose |
| --- | --- | --- | --- | --- | --- |
| `GET` | `/api/v1/alerts?status=&limit=` | Session | — | `200` `{alerts:[...]}` | Alert history. `status` is `firing`, `acknowledged`, or `resolved`; `limit` is 1–500, default 100. |
| `POST` | `/api/v1/alerts/{id}/acknowledge` | Operator or higher | `{acknowledgedBy}` (optional) | `200` alert | Acknowledge a firing alert; the acting username is recorded as the acknowledger. `409` when it is not firing; `404` when unknown. |
| `GET` | `/api/v1/alert-rules` | Session | — | `200` `{rules:[...]}` | List health alert rules. |
| `PUT` | `/api/v1/alert-rules/{id}` | Admin | `{name, health, severity, cooldownSeconds, enabled}` | `200` saved rule | Create or replace a rule. `health` is `degraded` or `unreachable`; `severity` is `warning` or `critical`; `cooldownSeconds` is 0–2592000. Unknown fields return `400`. |

Lifecycle, deduplication, and signed webhook delivery are documented in
[Fleet health alerts](alerts.md).

## Policies and drift

All routes are GET-only; the hub exposes no policy mutation or remediation.

| Method | Path | Auth | Purpose |
| --- | --- | --- | --- |
| `GET` | `/api/v1/policies` | Session | The immutable policy catalog. |
| `GET` | `/api/v1/policies/results` | Session | Fleet-wide evaluation: `{results, summary}`. |
| `GET` | `/api/v1/clusters/{id}/policy-results` | Session | Evaluation scoped to one cluster (`404` outside the tenant). |

Optional exact-match filters: `status` (`pass`, `fail`, `unknown`, `stale`),
`severity` (`low`, `medium`, `high`, `critical`), `scope` (`fleet`,
`cluster`, `namespace`, `workload`), `policyId`, `clusterId`. Invalid filter
values return `400`. See [Policy and configuration drift](policy-drift.md)
for the policy catalog and evaluation model.
