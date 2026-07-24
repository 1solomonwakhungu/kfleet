# kfleet

> Lightweight multi-cluster Kubernetes management with a built-in AI interface.

[![CI](https://github.com/1solomonwakhungu/kfleet/actions/workflows/ci.yml/badge.svg)](https://github.com/1solomonwakhungu/kfleet/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/1solomonwakhungu/kfleet)](https://github.com/1solomonwakhungu/kfleet/releases)
[![License](https://img.shields.io/github/license/1solomonwakhungu/kfleet)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/1solomonwakhungu/kfleet)](https://goreportcard.com/report/github.com/1solomonwakhungu/kfleet)
[![GHCR](https://img.shields.io/badge/GHCR-kfleet--hub-blue?logo=github)](https://github.com/1solomonwakhungu/kfleet/pkgs/container/kfleet-hub)

## What is kfleet?

kfleet is a lightweight control plane for viewing and operating multiple Kubernetes clusters from one place. A small agent reports cluster state to a single Go hub that serves the REST API, real-time web UI, and SQLite-backed inventory, while the built-in Model Context Protocol (MCP) server lets AI clients inspect and diagnose the fleet using natural language.

## Architecture

```text
                         MCP over stdio
                    +---------------------+
                    | AI client / copilot |
                    +----------+----------+
                               |
                        +------v------+
                        | MCP server  |
                        | kfleet mcp  |
                        +------+------+
                               |
                               | Hub REST API
                               v
  +----------------------------+-----------------------------+
  |                         kfleet Hub                       |
  |  +----------+  +-----------+  +----------------------+  |
  |  | REST API |  | WebSocket |  | Embedded React UI    |  |
  |  +-----+----+  +-----+-----+  +----------------------+  |
  |        +-------------+--------------+                  |
  |                                    |                  |
  |                              +-----v------+           |
  |                              |  SQLite    |           |
  |                              +------------+           |
  +--------------------------+--------------------------------+
                             |
                    registration + state reports
              +--------------+---------------+
              |              |               |
       +------v------+ +-----v-------+ +-----v-------+
       | Agent       | | Agent       | | Agent       |
       | cluster-a   | | cluster-b   | | cluster-c   |
       +-------------+ +-------------+ +-------------+
```

The hub is the only service users connect to. Agents use Kubernetes RBAC to collect a normalized cluster snapshot, register with the hub, and send heartbeats and state updates.

## How it compares

This table compares native, bundled capabilities; integrations and third-party extensions may add more.

| Capability | kfleet | Rancher | Lens | Headlamp | OCM |
| --- | --- | --- | --- | --- | --- |
| Multi-cluster | Yes | Yes | Yes | Yes, via contexts | Yes |
| Agent footprint | One small agent | Multiple in-cluster components | None | None | Klusterlet components |
| Web UI | Built in | Built in | Desktop app | Built in | Not bundled |
| Real-time updates | WebSocket | Yes | Yes | Yes | Controller-driven |
| AI / MCP interface | Built in | No native MCP | No native MCP | No native MCP | No native MCP |
| Install complexity | Hub chart + agent chart | Moderate | Desktop install | Single cluster deployment | Multiple operators/controllers |
| Open source | Apache 2.0 | Yes | Partial | Apache 2.0 | Apache 2.0 |
| Single-binary hub | Yes | No | N/A | Yes | No |

## 60-second kind quickstart

You need Docker, [kind](https://kind.sigs.k8s.io/), `kubectl`, and Helm 3. Then run:

```bash
git clone https://github.com/1solomonwakhungu/kfleet.git
cd kfleet
./hack/quickstart.sh
```

The script creates three kind clusters, builds and loads the local images, installs the hub and agents, and waits for all agents to register. When it finishes, open [http://localhost:8080](http://localhost:8080).

Environment overrides are available for a different cluster count or image source:

```bash
KFLEET_CLUSTERS=2 IMAGE_TAG=dev ./hack/quickstart.sh
USE_GHCR=true IMAGE_TAG=latest ./hack/quickstart.sh
```

Remove the demo with `./hack/cleanup.sh`.

## Install with Helm

Create a shared registration token and install the hub:

```bash
export KFLEET_REGISTRATION_TOKEN="$(openssl rand -hex 32)"
export KFLEET_ADMIN_PASSWORD="$(openssl rand -base64 24)"

helm upgrade --install kfleet-hub ./charts/kfleet-hub \
  --namespace kfleet-system \
  --create-namespace \
  --set registration.token="$KFLEET_REGISTRATION_TOKEN" \
  --set-string auth.bootstrapAdmin.username=admin \
  --set-string auth.bootstrapAdmin.email=admin@example.com \
  --set-string auth.bootstrapAdmin.password="$KFLEET_ADMIN_PASSWORD" \
  --set persistence.enabled=true \
  --set persistence.size=1Gi
```

Expose the hub with an Ingress, a `LoadBalancer` service, or a local port-forward. Install the agent into every cluster you want to manage:

```bash
helm upgrade --install kfleet-agent ./charts/kfleet-agent \
  --kube-context my-cluster \
  --namespace kfleet-system \
  --create-namespace \
  --set hub.url=https://kfleet.example.com \
  --set hub.token="$KFLEET_REGISTRATION_TOKEN" \
  --set cluster.name=my-cluster
```

Important values include `hub.url`, `hub.token`, `cluster.name`, `cluster.labels`, `reportInterval`, and the image repository/tag settings in each chart's `values.yaml`.

The hub requires a human user session for its web UI and fleet APIs. It
supports admin, operator, and read-only roles, and records security-relevant
actions in an append-only audit log. See
[Hub authentication and authorization](docs/authentication.md) for bootstrap,
RBAC, cookie security, migrations, audit retention, and agent token rotation.

## Use the MCP server

The MCP server communicates over stdio and calls the hub REST API. Create a
dedicated read-only user, then set `KFLEET_HUB_URL`, `KFLEET_HUB_USERNAME`, and
`KFLEET_HUB_PASSWORD`. For Claude Desktop, add an entry like this to its MCP
configuration:

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

You can also build and use the standalone server at `./cmd/mcp`. Available tools list clusters, retrieve status, pods, and events, locate crashlooping pods, and produce a composite cluster diagnosis.

Example prompts:

- “List my clusters and summarize anything that is not healthy.”
- “Find crashlooping or frequently restarting pods across the fleet.”
- “Diagnose the production cluster and explain its recent Warning events.”
- “Show pods in the payments namespace on staging.”

## Detect policy and configuration drift

The Policy dashboard evaluates seven built-in, read-only checks across the latest cluster snapshots. It reports Kubernetes version, cluster label, namespace, deployment configuration, workload availability, namespace Pod Security, and pod security posture results as `pass`, `fail`, `unknown`, or `stale`.

Open `/policies` in the hub UI or query the REST API:

```bash
curl http://localhost:8080/api/v1/policies
curl 'http://localhost:8080/api/v1/policies/results?status=fail'
```

The hub never remediates policy findings. Multi-tenant installations scope agent and API traffic with `KFLEET_TENANT_ID` and `X-Kfleet-Tenant-ID`. See [Policy and configuration drift](docs/policy-drift.md) for policy definitions, API filters, freshness behavior, tenant isolation, and compatibility details.

## Development

Prerequisites:

- Go 1.23 or newer
- Node.js 20 and npm
- Docker
- `kubectl`, Helm 3, and kind for end-to-end development
- `golangci-lint` for `make lint`

Common workflows:

```bash
# Build the React UI and the Go hub/agent binaries.
make build

# Build only the embedded web application.
make web-build

# Run all Go tests with the race detector.
make test

# Run the hub locally with a disposable database.
KFLEET_DB_PATH=/tmp/kfleet.db \
KFLEET_SESSION_COOKIE_INSECURE=true \
KFLEET_BOOTSTRAP_ADMIN_USERNAME=admin \
KFLEET_BOOTSTRAP_ADMIN_EMAIL=admin@localhost \
KFLEET_BOOTSTRAP_ADMIN_PASSWORD='local-development-password' \
go run ./cmd/hub
```

The hub listens on `:8080` by default. Run the MCP subcommand during development with:

```bash
KFLEET_HUB_URL=http://localhost:8080 go run ./cmd/hub mcp
```

## Contributing

Contributions are welcome. Read [CONTRIBUTING.md](CONTRIBUTING.md) for the pull request flow, Conventional Commit format, DCO sign-off requirement, and local test/lint commands.

## License

kfleet is licensed under the [Apache License 2.0](LICENSE).
