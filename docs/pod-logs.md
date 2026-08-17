# Pod log streaming

kfleet streams live pod logs into the cluster detail **Logs** tab without
giving the hub any Kubernetes credentials.

## Why a reverse channel

Agents run inside each cluster and push snapshots to the hub over HTTP. The
hub has no kubeconfig and cannot reach a cluster's API server, so it cannot
read pod logs itself. Instead, each approved agent opens an outbound
WebSocket to the hub and keeps it open. The hub pushes log requests down
that channel and relays the answers to the browser as Server-Sent Events.

```
Browser ──SSE──> Hub <──WebSocket── Agent ──client-go GetLogs──> kube-apiserver
```

## Endpoints

| Method and path | Auth | Purpose |
| --- | --- | --- |
| `GET /api/v1/clusters/{id}/pods/{namespace}/{pod}/logs` | Hub user session | Stream pod logs to the browser as SSE |
| `GET /api/v1/agents/{id}/logs` | Approved agent bearer token | Agent reverse channel (WebSocket) |

The browser route is tenant-scoped exactly like the other cluster resource
routes: the cluster must resolve inside the validated `X-Kfleet-Tenant-ID`
context or the request returns `404`.

### Query parameters

| Parameter | Default | Notes |
| --- | --- | --- |
| `container` | first container | Passed through to the Kubernetes log request |
| `follow` | `true` | `false` replays the backlog and ends the stream |
| `tailLines` | `200` | Clamped to 1–5000 |

### SSE frames

| Event | Payload | Meaning |
| --- | --- | --- |
| `message` (default) | the raw log line | One log line |
| `end` | `{}` | The stream finished normally |
| `stream-error` | `{"message": "..."}` | The agent could not stream (for example RBAC denied `pods/log`) |

The event is named `stream-error` rather than `error` because browsers
dispatch server-sent `error` events through the same handler as transport
failures.

## Agent-disconnected behavior

If the cluster has no live agent channel, the hub answers `503` with a JSON
body rather than an event stream:

```json
{ "error": "no agent is connected for this cluster", "code": 503 }
```

Because the response is not `text/event-stream`, the browser fails the
`EventSource` permanently instead of reconnecting in a loop. The Logs tab
reads the JSON error, shows an **Unavailable** state with the hub's message,
and offers a **Try again** button.

## Lifecycle and limits

- Closing the browser tab or switching pods cancels the request context. The
  hub sends a `stop` frame, and the agent cancels the Kubernetes stream, so
  no goroutine or API server watch is leaked.
- Each browser stream buffers up to 512 lines in the hub. A consumer that
  falls further behind has its stream cut rather than stalling the shared
  agent channel.
- The agent truncates individual lines at 16 KiB.
- A reconnecting agent replaces its previous channel; streams on the stale
  channel are terminated so the browser sees an explicit error.

## RBAC

The agent needs read access to the pod log subresource. The
`kfleet-agent` chart grants it:

```yaml
- apiGroups: [""]
  resources: ["pods/log"]
  verbs: ["get", "list"]
```

Upgrade the agent chart when moving to a version with log streaming;
otherwise the stream ends with a `stream-error` frame carrying the
Kubernetes `forbidden` message.

## Demo mode

The public demo has no real clusters and no agents, so the hub synthesizes
clearly generic log lines for demo clusters instead of failing. See
[Public demo](public-demo.md).
