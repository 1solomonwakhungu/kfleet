# Operations runbook

Day-2 operations for the kfleet hub and agents: configuration, health
probes, upgrades, and troubleshooting. API details live in
[api.md](api.md); security architecture lives in
[authentication.md](authentication.md).

## Environment variables

### Hub (`cmd/hub`)

| Variable | Default | Meaning |
| --- | --- | --- |
| `KFLEET_LISTEN_ADDR` | `:8080` | HTTP listen address. |
| `KFLEET_DB_PATH` | `./kfleet.db` | SQLite database path. Ignored in demo mode, which uses an in-memory database. |
| `KFLEET_LOG_LEVEL` | `info` | Hub log level: `debug`, `info`, `warn`, or `error`. |
| `KFLEET_HEARTBEAT_INTERVAL` | `30s` | Expected agent heartbeat cadence. Three missed intervals mark a cluster stale. Also paces the staleness monitor. |
| `KFLEET_REGISTRATION_TOKEN` | empty | Static agent registration token. When unset and no token has ever been rotated, agent registration fails closed with `403`. |
| `KFLEET_DEMO_MODE` | `false` | Read-only synthetic demo: in-memory fixture database, mutations rejected with `405`, background workers disabled. See [public-demo.md](public-demo.md). |
| `KFLEET_METRICS_ENABLED` | `true` | Set to `false` to disable the unauthenticated `/metrics` endpoint (it then answers `404`). Must be a boolean. The Helm chart exposes this as `metrics.enabled`. |
| `KFLEET_EVENT_RETENTION` | `2160h` (90 days) | How long operational timeline events are kept. Must be a positive Go duration. |
| `KFLEET_SESSION_DURATION` | `24h` | Login session lifetime. |
| `KFLEET_SESSION_COOKIE_INSECURE` | unset | Set to `true` to drop the `Secure` flag on the session cookie. Local plain-HTTP development only. |
| `KFLEET_BOOTSTRAP_ADMIN_USERNAME` | empty | Username for the first admin account. |
| `KFLEET_BOOTSTRAP_ADMIN_EMAIL` | empty | Email for the first admin account. |
| `KFLEET_BOOTSTRAP_ADMIN_PASSWORD` | empty | Password for the first admin account (12–72 bytes). All three bootstrap variables must be set together, and only when no user exists. |
| `KFLEET_ALERT_WEBHOOK_URL` | empty | Generic webhook endpoint for alert delivery. Empty disables delivery. Must be http(s) without URL user information. |
| `KFLEET_ALERT_WEBHOOK_SECRET` | empty | HMAC secret for webhook signing. Required when the URL is set; the hub refuses to start otherwise. |
| `KFLEET_ALERT_MAX_ATTEMPTS` | `5` | Total delivery attempts before an alert moves to `dead_letter`. |
| `KFLEET_ALERT_RETRY_BASE` | `5s` | Initial retry delay; later delays double up to one hour. |
| `KFLEET_ALERT_POLL_INTERVAL` | `1s` | Poll interval for durable pending deliveries. |

### Agent (`cmd/agent`)

`KFLEET_HUB_URL`, `KFLEET_HUB_TOKEN`, and `KFLEET_CLUSTER_NAME` are required;
the agent exits non-zero naming the missing variable instead of retrying
registration forever.

| Variable | Default | Meaning |
| --- | --- | --- |
| `KFLEET_HUB_URL` | required | Absolute http(s) URL of the hub. |
| `KFLEET_HUB_TOKEN` | required | Registration token used at `/api/v1/agents/register`. After registration the agent keeps its per-agent token in memory for the process lifetime. |
| `KFLEET_CLUSTER_NAME` | required | Cluster name registered with the hub. |
| `KFLEET_REPORT_INTERVAL` | `30s` | How often a full snapshot is reported. Liveness heartbeats are fixed at 15s. |
| `KFLEET_TENANT_ID` | `default` | Tenant identifier sent with agent traffic. Lowercase letters, numbers, dots, underscores, and hyphens. |
| `KFLEET_CLUSTER_LABELS` | empty | JSON object of labels stored on the cluster record, e.g. `{"environment":"prod","owner":"platform"}`. |
| `KFLEET_HEALTH_ADDR` | `:8081` | Listen address for the agent's `/healthz` and `/readyz` probes. |
| `KFLEET_LOG_LEVEL` | `info` | Agent log level: `debug`, `info`, `warn`, or `error`. |
| `KFLEET_HUB_CA` | unset | PEM-encoded CA bundle (one or more certificates) used to verify the hub's TLS certificate. Unset means system roots, which is correct for publicly trusted certificates. Must contain at least one valid PEM certificate. |
| `KUBECONFIG` | inherited | Kubernetes credentials for the in-cluster or kubeconfig-based collector. |

### MCP server (`cmd/mcp`, or `cmd/hub mcp`)

| Variable | Default | Meaning |
| --- | --- | --- |
| `KFLEET_HUB_URL` | required | Hub base URL the MCP server calls. |
| `KFLEET_HUB_USERNAME` | required | Hub user for the MCP server. Use a dedicated `read_only` account. |
| `KFLEET_HUB_PASSWORD` | required | Password for that user. |

### `kfleet` CLI (`cmd/kfleet`)

| Variable | Default | Meaning |
| --- | --- | --- |
| `KFLEET_CLUSTERS` | `3` | Default cluster count for `kfleet quickstart`. For `kfleet cleanup`, `0` means all clusters in the saved state. |
| `KFLEET_HUB_PORT` | `8080` | Local hub port for `kfleet quickstart`. |
| `KFLEET_ADMIN_USERNAME` | `admin` | Admin username used by `kfleet quickstart`. |
| `KFLEET_ADMIN_PASSWORD` | generated | Admin password used by `kfleet quickstart`; a random one is generated and saved to local state when unset. |
| `KFLEET_VERSION` | detected | Overrides the release version used by `kfleet version` and quickstart installs. |

### Webhook receiver (`cmd/webhook-receiver`)

A loopback-only development receiver for signed alert webhooks.

| Variable | Default | Meaning |
| --- | --- | --- |
| `KFLEET_RECEIVER_LISTEN_ADDR` | `127.0.0.1:9099` | Listen address. Must be a loopback host; the receiver exits otherwise. |
| `KFLEET_RECEIVER_SECRET` | required | HMAC secret used to verify `X-Kfleet-Signature`. |
| `KFLEET_RECEIVER_FAIL_FIRST` | `0` | Number of initial deliveries to fail before succeeding (failure injection). |
| `KFLEET_RECEIVER_FAIL_STATUS` | `503` | HTTP status returned by injected failures. Must be 400–599. |

The Helm charts set most of these through values (`registration.token`,
`auth.bootstrapAdmin`, `auth.sessionCookieInsecure`, `alerts.*`,
`timeline.retention`, `tenant.id`); see each chart's `values.yaml`. Secrets
can come from your own Kubernetes Secrets instead of chart-managed ones:
the hub chart supports `registration.existingSecret` (key
`registration-token`) and `auth.existingSecret` for bootstrap credentials,
and the agent chart supports `hub.existingSecret` with `hub.existingSecretKey`
(default `hub-token`). For a hub behind a private CA, set the agent chart's
`hub.caBundle` to the PEM bundle, which becomes the agent's `KFLEET_HUB_CA`.

## Health and monitoring

### Hub

- `GET /healthz` — liveness. Always `200 ok` while the process serves HTTP.
  Use it for a Kubernetes liveness probe.
- `GET /readyz` — readiness. Answers `200 ready` only when the SQLite store
  answers a query within 2 seconds; otherwise `503` with
  `{"status":"unavailable","reason":"store unavailable"}`. Use it for the
  readiness probe so traffic drains while the database is wedged.
- `GET /metrics` — unauthenticated Prometheus scrape target (text exposition
  format 0.0.4). Serves aggregate gauges only; see
  [api.md — /metrics](api.md#metrics). Disable with `KFLEET_METRICS_ENABLED=false`.
- `GET /api/v1/meta` — public runtime metadata; verifies the hub is serving
  and reports the build version and demo mode.

Alert on:

- `readyz` failing (store unavailable — check disk space and I/O on the PVC)
- `healthz` failing (process down — the rollout should restart it)
- `kfleet_alerts_dead_letter` above zero on `/metrics` (webhook receiver failing)
- cluster `unreachable` alerts, which the hub raises after three missed
  heartbeat intervals, via the [alerts](alerts.md) webhook
- repeated `401`s from agents in hub logs (rotated or wrong tokens)

### Agent

The agent serves `/healthz` and `/readyz` on `KFLEET_HEALTH_ADDR`
(`:8081` by default).

- `GET /healthz` — liveness. Always `200 ok` while the process can serve.
- `GET /readyz` — readiness, now honest about hub reachability. It answers
  `200 ok` only when the agent has completed a registration call with the
  hub **and** the last successful hub contact (registration, heartbeat, or
  status report) is within the staleness window: three times the agent's
  fixed 15-second heartbeat interval, so 45 seconds. Otherwise it answers
  `503` with `{"error":"no successful hub contact"}`. Registration counts as
  contact even while the agent is still pending approval; until that first
  registration succeeds, the agent reports not-ready.

This makes agent `/readyz` a usable fleet-connectivity signal: an agent cut
off from the hub goes unready within 45 seconds instead of reporting healthy
while it retries. Transient unready during registration backoff is expected;
persistent unready means the hub is unreachable, the token is wrong, or the
agent cannot reach the API server for its first collection. Watch agent logs
for `agent registration failed` or `agent heartbeat failed; re-registering`
to distinguish them.

## Upgrades

- **Pin chart versions.** Each `v*` release publishes matching OCI charts
  (`oci://ghcr.io/1solomonwakhungu/charts/kfleet-hub` and `.../kfleet-agent`)
  whose `appVersion` matches the release. Install with
  `helm upgrade --install ... --version X.Y.Z` so the chart and the images it
  deploys stay in lockstep.
- **Back up before upgrading.** The hub database holds users, sessions, the
  audit log, and the timeline. Take a consistent SQLite snapshot first; the
  procedure (including the WAL caveat and a one-off-pod recipe) is in
  [authentication.md — backup and restore](authentication.md#backup-and-restore).
- **Migrations run at boot.** On startup the hub applies additive SQLite
  migrations — creating tables such as `users`, `sessions`, `audit_events`,
  `alerts`, `operational_events`, and `settings`, plus indexes, triggers, and
  tenant/security-context columns — and seeds the two default alert rules.
  Existing cluster and agent records stay in place. There is no separate
  migration step to run; upgrade is a normal `helm upgrade` (or binary
  replacement) followed by a hub restart. Roll back by restoring the backup
  taken before the upgrade, because authentication data lives in the same
  database.
- **Upgrade hub and agent charts together** when the release changes the
  wire protocol (for example, log streaming), and re-check the agent's
  RBAC rules against the chart's current `values.yaml`.

## Troubleshooting

Each scenario: symptom → cause → fix.

### Agent registration returns 403 "agent registration is disabled on this hub"

The hub fails closed when no registration credential exists: no
`KFLEET_REGISTRATION_TOKEN` was set and no admin has ever rotated a token.
The hub logs a warning at startup naming the variable. **Fix:** set
`KFLEET_REGISTRATION_TOKEN` (or rotate a token in the UI at `/agents` /
`POST /api/v1/admin/registration-token/rotate`) and restart the hub, then
update the agent's `KFLEET_HUB_TOKEN`. See
[authentication.md — agent registration](authentication.md#agent-registration).

### Agent reports 401 "invalid registration token"

Registration is enabled but the presented token is wrong — usually the token
was rotated, so the previous static `KFLEET_REGISTRATION_TOKEN` is rejected,
or the agent's `KFLEET_HUB_TOKEN` does not match the hub. **Fix:** put the
current token in the agent's `KFLEET_HUB_TOKEN` (re-run the `helm upgrade`
with the new value) and restart the agent. Existing approved agents keep
their per-agent credentials and are unaffected.

### Agent gets 409 re-registering "cluster already registered; provide the current agent token or remove the cluster first"

Re-registration rotates the cluster's agent token, so the hub demands proof
the caller still holds the current credential (`existingAgentToken`). The
bundled agent carries its token in memory and re-registers fine, but a
restarted agent process has lost it and the hub never re-issues a stored
token. **Fix:** remove the cluster (operator or admin, from the cluster
detail page or `DELETE /api/v1/clusters/{id}`) and let the agent register
fresh. This is the supported recovery path today; there is no agent-token
reset endpoint.

### Pod logs return 503 "no agent is connected for this cluster"

The browser-side SSE route only works when the cluster's approved agent has
its outbound WebSocket connected. The agent may be down, pending approval,
running a version without log streaming, or cut off by network policy.
**Fix:** check the agent pod is running and approved (`GET
/api/v1/agents/pending`), confirm the agent chart version supports log
streaming and grants `pods/log` RBAC (see
[pod-logs.md — RBAC](pod-logs.md#rbac)), then retry. The JSON error body is
deliberately not an event stream so the UI shows an explicit unavailable
state instead of reconnecting in a loop.

### Login or API call returns 403 "this action requires the admin role" (or operator)

The hub names the minimum role in every role-denied 403. This is expected RBAC
behavior, not a login failure — the session is valid but insufficient.
**Fix:** have an admin grant the needed role on the Users page
(`PATCH /api/v1/users/{id}`), or perform the action as an admin. The role
matrix is in [authentication.md — roles](authentication.md#roles). Browser
mutations that fail with "CSRF validation failed" are a different problem:
the client must send `X-Kfleet-CSRF: 1` (the embedded UI does this
automatically; custom scripts must too).
