# Fleet health alerts

kfleet evaluates persisted fleet health rules whenever the hub observes a cluster health update. Two rules are installed by default:

| Rule | Health | Severity | Cooldown |
| --- | --- | --- | --- |
| `fleet-health-degraded` | `degraded` | warning | 15 minutes |
| `fleet-health-unreachable` | `unreachable` | critical | 15 minutes |

Rules are persisted in SQLite. They can be listed with `GET /api/v1/alert-rules` and replaced by ID with `PUT /api/v1/alert-rules/{id}`. Rule updates accept `name`, `health`, `severity`, `cooldownSeconds`, and `enabled`. Health must be `degraded` or `unreachable`.

## Lifecycle and deduplication

An enabled matching rule creates one durable alert with a dedupe key composed from the rule, cluster, and health state. Another observation with the same key inside the rule cooldown does not create a duplicate. The cooldown check and insert run in one SQLite transaction.

When the cluster moves to a different health state, previously firing or acknowledged alerts for that cluster are marked resolved. Acknowledgement does not stop webhook delivery. Operator state and delivery state are independent so the history shows both what operators did and what the notification transport did.

Alert operator states:

- `firing`
- `acknowledged`
- `resolved`

Webhook delivery states:

- `pending`
- `retrying`
- `delivered`
- `dead_letter`
- `disabled`

Delivery is `disabled` when no webhook URL is configured. Alert creation, history, resolution, and acknowledgement continue to work in this mode.

## API

```text
GET  /api/v1/alerts?status=firing&limit=100
POST /api/v1/alerts/{id}/acknowledge
GET  /api/v1/alert-rules
PUT  /api/v1/alert-rules/{id}
```

Acknowledgement body:

```json
{
  "acknowledgedBy": "on-call"
}
```

Only a firing alert can be acknowledged. A repeated acknowledgement returns HTTP 409.

## Signed generic webhooks

Set both values to enable delivery:

```bash
export KFLEET_ALERT_WEBHOOK_URL="http://127.0.0.1:9099/webhook"
export KFLEET_ALERT_WEBHOOK_SECRET="replace-with-a-long-random-local-secret"
```

The hub refuses to start if the URL is not HTTP or HTTPS, contains URL user information, or is configured without a secret.

Each request uses `Content-Type: application/json` and includes:

```text
X-Kfleet-Event: fleet.health.alert
X-Kfleet-Delivery: <alert UUID>
X-Kfleet-Timestamp: <Unix seconds>
X-Kfleet-Signature: v1=<hex HMAC-SHA256>
```

The signed bytes are:

```text
<X-Kfleet-Timestamp>.<raw request body>
```

Receivers should calculate HMAC-SHA256 with the shared secret, compare the result in constant time, reject old timestamps, and treat `X-Kfleet-Delivery` as an idempotency key.

Example payload:

```json
{
  "id": "18e56ff2-b390-4d8c-8602-4b686c2a1ff7",
  "event": "fleet.health.alert",
  "occurredAt": "2026-07-23T17:00:00Z",
  "rule": {
    "id": "fleet-health-unreachable",
    "name": "Cluster unreachable"
  },
  "cluster": {
    "id": "cluster-a",
    "name": "production",
    "health": "unreachable"
  },
  "severity": "critical",
  "summary": "production is unreachable"
}
```

Any HTTP 2xx response is successful. Network errors and other status codes are retried with exponential backoff. After the configured maximum number of attempts, the alert moves to `dead_letter`. The last error and every attempt count remain visible in alert history.

## Configuration

| Environment variable | Default | Purpose |
| --- | --- | --- |
| `KFLEET_ALERT_WEBHOOK_URL` | empty | Generic webhook endpoint. Empty disables delivery. |
| `KFLEET_ALERT_WEBHOOK_SECRET` | empty | HMAC secret. Required when the URL is set. |
| `KFLEET_ALERT_MAX_ATTEMPTS` | `5` | Total attempts before dead letter. |
| `KFLEET_ALERT_RETRY_BASE` | `5s` | Initial retry delay. Later delays double up to one hour. |
| `KFLEET_ALERT_POLL_INTERVAL` | `1s` | Interval for polling durable pending deliveries. |

The Helm chart exposes the same settings under `alerts`. The webhook secret is stored in the chart-managed Kubernetes Secret.

## Safe local receiver and failure injection

The bundled development receiver refuses non-loopback listen addresses. It verifies signatures, stores received payloads in memory, and provides failure injection.

Terminal 1:

```bash
KFLEET_RECEIVER_SECRET="local-test-secret" \
KFLEET_RECEIVER_FAIL_FIRST=2 \
make receiver
```

Terminal 2:

```bash
alert_db="$(mktemp -t kfleet-alerts)"
KFLEET_DB_PATH="$alert_db" \
KFLEET_ALERT_WEBHOOK_URL="http://127.0.0.1:9099/webhook" \
KFLEET_ALERT_WEBHOOK_SECRET="local-test-secret" \
KFLEET_ALERT_RETRY_BASE=500ms \
go run ./cmd/hub
```

Inspect accepted attempts:

```bash
curl --fail http://127.0.0.1:9099/received
```

Change failure behavior without restarting the receiver:

```bash
curl --fail \
  -H 'Content-Type: application/json' \
  -d '{"count":3,"status":503}' \
  http://127.0.0.1:9099/control/failures
```

Keep local verification pointed at `127.0.0.1`. Do not use production notification endpoints for failure tests.
