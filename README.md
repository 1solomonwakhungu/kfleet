<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="docs/brand/png/kfleet-logo-dark-128.png">
    <img src="docs/brand/png/kfleet-logo-128.png" alt="kfleet" width="128" height="128">
  </picture>
</p>

# kfleet

> See every Kubernetes cluster you run from one place, and let an AI client ask questions about them, without handing anyone a kubeconfig.

[![CI](https://github.com/1solomonwakhungu/kfleet/actions/workflows/ci.yml/badge.svg)](https://github.com/1solomonwakhungu/kfleet/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/1solomonwakhungu/kfleet)](https://github.com/1solomonwakhungu/kfleet/releases)
[![License](https://img.shields.io/github/license/1solomonwakhungu/kfleet)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/1solomonwakhungu/kfleet)](https://goreportcard.com/report/github.com/1solomonwakhungu/kfleet)
[![GHCR](https://img.shields.io/badge/GHCR-kfleet--hub-blue?logo=github)](https://github.com/1solomonwakhungu/kfleet/pkgs/container/kfleet-hub)

kfleet is a read-only control plane for a fleet of Kubernetes clusters. One small agent per cluster reports state to a single Go hub. The hub serves a web UI, a REST API, and a Model Context Protocol (MCP) server from one binary, and keeps everything in one SQLite file.

<p align="center">
  <a href="https://kfleet.sullylab.de">
    <img src="docs/assets/demo.gif" alt="A tour of the kfleet web UI: the fleet dashboard with three clusters, a degraded cluster's detail page, live pod logs, policy and drift results, and alert history" width="900">
  </a>
</p>

<p align="center">
  <a href="https://kfleet.sullylab.de"><strong>Open the live demo</strong></a> (synthetic data, read-only)
  ·
  <a href="#quickstart">Run it locally</a>
  ·
  <a href="docs/README.md">Docs</a>
</p>

The recording shows the hub running in [demo mode](docs/public-demo.md) with synthetic clusters. Nothing in it comes from a real environment.

## Contents

- [Why kfleet](#why-kfleet)
- [Architecture](#architecture)
- [Quickstart](#quickstart)
- [Install with Helm](#install-with-helm)
- [Use it from an AI client](#use-it-from-an-ai-client)
- [What you get](#what-you-get)
- [Development](#development)
- [Documentation](#documentation)

## Why kfleet

Rancher, Open Cluster Management (OCM), Lens, and Headlamp are good tools that do more than kfleet does. kfleet is for the case where you want to *see* a fleet and do not want to run a platform, or widen who can change your clusters, to get that view.

### The hub is one process

The hub is a 14 MB static Go binary with the web UI embedded and SQLite as its only store. In demo mode with three clusters loaded it idles at about 20 MB of memory. It needs no management cluster, cert-manager, CRDs, or operators, so it runs as a Deployment, a container, or a plain process on a VM.

Rancher's server runs on its own Kubernetes cluster, usually with cert-manager alongside it. OCM installs a cluster manager operator and its controllers on a hub cluster.

### The hub cannot touch your clusters

The hub stores no kubeconfigs, service account tokens, or cluster certificates. Each agent opens its own outbound connections, so no cluster API server has to be reachable from the hub or from anyone's laptop.

The agent's entire ClusterRole is three rules:

```yaml
rules:
  - apiGroups: [""]
    resources: ["pods", "services", "nodes", "events", "namespaces"]
    verbs: ["list"]
  - apiGroups: ["apps"]
    resources: ["deployments"]
    verbs: ["list"]
  - apiGroups: [""]
    resources: ["pods/log"]
    verbs: ["get"]
```

It cannot read Secrets or ConfigMaps, and it has no write verbs. If the hub is compromised, the attacker gets inventory and pod logs. They do not get a way into the clusters.

Lens and Headlamp work from each user's kubeconfig, so every user needs credentials for, and a network path to, every cluster. Rancher's downstream agent is bound to `cluster-admin`, and the Rancher server proxies the Kubernetes API with it.

### The AI interface is read-only and comes in the box

`kfleet-hub mcp` is an MCP server compiled into the same binary. It exposes nine tools, all of them reads, and signs in to the hub as an ordinary read-only user, so the model can do what that user can do and nothing else. It works with any MCP client and any model you choose.

The other tools have AI features now, with different trade-offs. Rancher's Liz assistant and its MCP server ship with SUSE Rancher Prime. Lens has the Prism assistant and a built-in MCP server inside its proprietary desktop app. Headlamp has an AI Assistant plugin that can also write to clusters after confirmation. OCM has an open proposal for an upstream MCP server and community-built ones today.

### It remembers

Lens and Headlamp read the Kubernetes API live, so a cluster you cannot reach shows you nothing. kfleet's hub keeps each cluster's last snapshot and a 90-day operational timeline, so an unreachable cluster still shows its last known state, when it dropped, and what changed before it did. The same store backs seven cross-cluster drift checks and health alerts with signed webhooks.

### At a glance

| | kfleet | Rancher | Lens | Headlamp | OCM |
| --- | --- | --- | --- | --- | --- |
| Runs centrally | One binary + SQLite | Server on a dedicated cluster | Nothing | Nothing, or one in-cluster app | Operator + controllers on a hub cluster |
| Runs in each cluster | One agent, three read rules | Agent with `cluster-admin` | Nothing | Nothing | Klusterlet agents |
| Cluster credentials held centrally | None | Yes | No, in each kubeconfig | No, in each kubeconfig | No, agents pull work |
| Can change clusters | No | Yes | Yes | Yes | Yes |
| State of an unreachable cluster | Last snapshot + timeline | Last known status | None | None | Last known status |
| AI interface | Read-only MCP server, built in | Liz + MCP (Rancher Prime) | Prism + MCP | AI Assistant plugin | Community MCP servers |
| License | Apache 2.0 | Apache 2.0 | Proprietary, free tier | Apache 2.0 | Apache 2.0 |

Competitor details were checked in November 2026 against each project's documentation and release notes: [Rancher AI](https://documentation.suse.com/cloudnative/rancher-ai/latest/en/introduction.html), [Lens MCP server](https://markets.financialcontent.com/spoke/article/bizwire-2026-3-18-lens-launches-built-in-mcp-server-connecting-ai-coding-assistants-to-kubernetes), [Headlamp AI Assistant](https://artifacthub.io/packages/headlamp/headlamp-plugins/headlamp_ai_assistant), [OCM MCP proposal](https://github.com/open-cluster-management-io/ocm/issues/1790). These tools move quickly. If a cell is out of date, please open an issue.

### When to pick something else

- You need to provision clusters, deploy workloads, or manage RBAC across the fleet. Use Rancher or OCM. kfleet never writes to a cluster.
- You want to edit resources, exec into pods, or port-forward. Use Lens, Headlamp, or `kubectl`.
- You need a highly available control plane. The hub is a single replica on SQLite. A hub outage costs you the view, not your clusters, but it is still an outage.
- You need more than pods, deployments, services, nodes, namespaces, and events. That list is what the agent collects today.

## Architecture

<p align="center">
  <img src="docs/assets/architecture.svg" alt="kfleet architecture: browsers, AI clients through the MCP server, and Prometheus connect to the hub. The hub contains the REST API, live updates, the log relay, the policy engine and alerts, and SQLite. One agent in each cluster connects out to the hub over HTTPS and WebSocket and reads from its own Kubernetes API server." width="900">
</p>

**Hub.** The only service people connect to. It serves the REST API, the embedded React UI, a WebSocket for live dashboard updates, and Prometheus metrics. It authenticates users with sessions, enforces admin, operator, and read-only roles, and writes an audit log. State lives in SQLite: the latest snapshot per cluster, the operational timeline, alerts, users, and audit records.

**Agent.** One Deployment per cluster. It lists resources with client-go, registers with the hub using a shared registration token, waits for an operator or admin to approve it, then sends a heartbeat and a normalized snapshot every 30 seconds. The hub marks a cluster stale after three missed intervals.

**Log channel.** The hub has no credentials, so it cannot fetch logs itself. Each approved agent holds an outbound WebSocket open to the hub. When someone opens a pod's logs, the hub sends the request down that socket, the agent streams lines back, and the hub relays them to the browser as Server-Sent Events. See [Pod log streaming](docs/pod-logs.md).

**MCP server.** A subcommand of the hub binary that speaks MCP over stdio and calls the same REST API as the UI.

**Alerts out.** When a cluster turns degraded or unreachable, the hub records an alert and, if configured, delivers an HMAC-SHA256-signed webhook with durable retries.

## Quickstart

### Look around without Kubernetes

Demo mode seeds an in-memory database with three synthetic clusters and rejects every write. It needs Go and Node.js, and no cluster.

```bash
git clone https://github.com/1solomonwakhungu/kfleet.git
cd kfleet
make web-build
KFLEET_DEMO_MODE=true go run ./cmd/hub
```

Open <http://localhost:8080>.

### Three real clusters on your laptop

The Homebrew cask installs the `kfleet` CLI along with kind, `kubectl`, and Helm. Docker must be running.

```bash
brew install --cask 1solomonwakhungu/tap/kfleet
kfleet quickstart
```

This creates three kind clusters, installs the released hub and agents, and starts a port-forward. Open the URL it prints, or run `kfleet open`.

```bash
kfleet quickstart --clusters 2 --port 9090   # different size or port
kfleet status                                # what is running
kfleet cleanup                               # remove everything
```

From a source checkout, `./hack/quickstart.sh` does the same with images built from your working tree.

## Install with Helm

Install the hub once:

```bash
export KFLEET_REGISTRATION_TOKEN="$(openssl rand -hex 32)"
export KFLEET_ADMIN_PASSWORD="$(openssl rand -base64 24)"

helm upgrade --install kfleet-hub \
  oci://ghcr.io/1solomonwakhungu/charts/kfleet-hub \
  --namespace kfleet-system \
  --create-namespace \
  --set registration.token="$KFLEET_REGISTRATION_TOKEN" \
  --set-string auth.bootstrapAdmin.username=admin \
  --set-string auth.bootstrapAdmin.email=admin@example.com \
  --set-string auth.bootstrapAdmin.password="$KFLEET_ADMIN_PASSWORD" \
  --set persistence.enabled=true \
  --set persistence.size=1Gi
```

Expose it with an Ingress, a `LoadBalancer` service, or a port-forward. Then install the agent in each cluster you want to see:

```bash
helm upgrade --install kfleet-agent \
  oci://ghcr.io/1solomonwakhungu/charts/kfleet-agent \
  --kube-context my-cluster \
  --namespace kfleet-system \
  --create-namespace \
  --set hub.url=https://kfleet.example.com \
  --set hub.token="$KFLEET_REGISTRATION_TOKEN" \
  --set cluster.name=my-cluster
```

Sign in to the hub as the bootstrap admin and approve the agent on the **Agents** page (operators can approve too). The cluster appears on the dashboard after its next report.

Things to know before production:

- `hub.url`, `hub.token`, and `cluster.name` are required. The agent exits with the missing variable named instead of retrying forever.
- For GitOps, keep tokens out of values with `hub.existingSecret` on the agent and `registration.existingSecret` and `auth.existingSecret` on the hub. Set `hub.caBundle` when the hub uses a private CA.
- Each release publishes matching charts. Pin with `--version 0.2.1`.
- The default requests are 100m CPU and 128Mi for the hub, and 50m CPU and 64Mi for each agent.

The [operations runbook](docs/operations.md) covers every environment variable, health probes, upgrades, backups, and troubleshooting. [Authentication and authorization](docs/authentication.md) covers roles, sessions, audit retention, and registration token rotation.

## Use it from an AI client

Create a dedicated read-only user in the hub at `/admin/users`, then point your MCP client at the hub binary. For Claude Desktop:

```json
{
  "mcpServers": {
    "kfleet": {
      "command": "/usr/local/bin/kfleet-hub",
      "args": ["mcp"],
      "env": {
        "KFLEET_HUB_URL": "http://localhost:8080",
        "KFLEET_HUB_USERNAME": "mcp-reader",
        "KFLEET_HUB_PASSWORD": "replace-with-the-read-only-user-password"
      }
    }
  }
}
```

| Tool | What it returns |
| --- | --- |
| `list_clusters` | Every cluster with health, version, and counts |
| `get_cluster_status` | One cluster's current status |
| `list_namespaces` | Namespaces in a cluster |
| `get_pods` | Pods, optionally filtered by namespace |
| `get_events` | Kubernetes events from the latest snapshot |
| `get_recent_events` | The hub's operational timeline, with time filters and paging |
| `find_crashloop_pods` | Crashlooping and frequently restarting pods across the fleet |
| `get_pod_logs` | Recent log lines for a pod, including the previous container of a crashlooping pod |
| `diagnose_cluster` | A composite report on one cluster |

Prompts to try:

- "List my clusters and summarize anything that is not healthy."
- "Find crashlooping pods across the fleet and show me the last logs from the worst one."
- "Diagnose the production cluster and explain its recent Warning events."
- "Show recent reconnects and policy findings across the fleet."

`get_pod_logs` returns the newest 100 lines by default, up to 1000, capped at 256 KiB. A standalone build of the server is also available at `./cmd/mcp`.

## What you get

| Feature | What it does | Details |
| --- | --- | --- |
| Fleet dashboard | Health, version, node and pod counts per cluster, with search by name or label and live updates over WebSocket | |
| Cluster detail | Pods, services, deployments, events, node capacity, and labels, with deep links to a tab, namespace, or pod | |
| Pod logs | Live log streaming through the agent channel, with follow, tail, and search | [pod-logs.md](docs/pod-logs.md) |
| Operational timeline | Append-only history of registrations, approvals, health changes, disconnects, reconnects, version changes, and policy findings. Kept for 90 days by default and survives cluster deletion | [api.md](docs/api.md#timeline-and-policy-findings) |
| Policy and drift | Seven built-in checks across the fleet: Kubernetes version, required labels, namespace and workload consistency, workload availability, and pod and namespace security. Results are `pass`, `fail`, `unknown`, or `stale`. kfleet reports and never remediates | [policy-drift.md](docs/policy-drift.md) |
| Health alerts | Alerts when a cluster turns degraded or unreachable, with acknowledgement and HMAC-signed webhooks retried from durable state | [alerts.md](docs/alerts.md) |
| Users and audit | Admin, operator, and read-only roles, with an append-only audit log of security-relevant actions | [authentication.md](docs/authentication.md) |
| Multi-tenancy | Agent and API traffic scoped by tenant ID | [policy-drift.md](docs/policy-drift.md) |
| Metrics | Prometheus endpoint at `/metrics` with non-sensitive aggregates only | [operations.md](docs/operations.md) |
| REST API | Everything the UI does | [api.md](docs/api.md) |

## Development

Prerequisites:

- Go 1.26 or newer
- Node.js 22 and npm
- Docker
- `kubectl`, Helm 3, and kind for end-to-end work
- `golangci-lint` for `make lint`

```bash
make build       # React UI plus the kfleet, hub, and agent binaries
make web-build   # only the embedded web application
make test        # Go tests with the race detector
make lint
```

Run the hub locally with a disposable database:

```bash
KFLEET_DB_PATH=/tmp/kfleet.db \
KFLEET_SESSION_COOKIE_INSECURE=true \
KFLEET_BOOTSTRAP_ADMIN_USERNAME=admin \
KFLEET_BOOTSTRAP_ADMIN_EMAIL=admin@localhost \
KFLEET_BOOTSTRAP_ADMIN_PASSWORD='local-development-password' \
KFLEET_REGISTRATION_TOKEN=local-dev-registration-token \
go run ./cmd/hub
```

The hub listens on `:8080`. Run the MCP server against it with:

```bash
KFLEET_HUB_URL=http://localhost:8080 go run ./cmd/hub mcp
```

Repository layout:

| Path | Contents |
| --- | --- |
| `cmd/hub`, `cmd/agent`, `cmd/kfleet`, `cmd/mcp` | Entry points for the hub, the agent, the quickstart CLI, and the standalone MCP server |
| `internal/server`, `internal/store` | Hub HTTP handlers and the SQLite store |
| `internal/agent` | Collector, registrar, reporter, and log channel client |
| `internal/policy`, `internal/alerts`, `internal/mcp` | Drift checks, alert delivery, and MCP tools |
| `web/` | React UI, embedded into the hub at build time |
| `charts/` | Helm charts for the hub and the agent |

The web UI is built with [Primer](https://primer.style), GitHub's design system. Components come from `@primer/react`, and colors, spacing, and radii resolve to `@primer/primitives` tokens, so the interface supports light and dark themes. Layout lives in CSS modules next to each component. Use the primitives custom properties (`--bgColor-*`, `--fgColor-*`, `--base-size-*`) there instead of hardcoded values.

To re-record the demo GIF after a UI change, see the header of [`hack/record-demo.mjs`](hack/record-demo.mjs).

## Documentation

The index is at [docs/README.md](docs/README.md).

- [REST API reference](docs/api.md)
- [Operations runbook](docs/operations.md)
- [Authentication and authorization](docs/authentication.md)
- [Fleet health alerts](docs/alerts.md)
- [Pod log streaming](docs/pod-logs.md)
- [Policy and configuration drift](docs/policy-drift.md)
- [Public read-only demo](docs/public-demo.md)

## Contributing

Contributions are welcome. [CONTRIBUTING.md](CONTRIBUTING.md) covers the pull request flow, Conventional Commit format, DCO sign-off, and local test and lint commands.

## License

kfleet is licensed under the [Apache License 2.0](LICENSE).
