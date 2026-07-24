# Public read-only demo

The deployment at `https://kfleet.sullylab.de` is a product demonstration, not
a connection to a live Kubernetes environment.

## Safety guarantees

`KFLEET_DEMO_MODE=true` changes the hub's runtime boundary:

- The configured SQLite path is ignored. The process creates a fresh in-memory
  database and seeds only fixtures whose names and labels explicitly identify
  them as synthetic.
- No agent registration tokens, Kubernetes credentials, kubeconfigs, cluster
  certificates, real hostnames, or live cluster identifiers are loaded.
- `POST`, `PUT`, `PATCH`, and `DELETE` are rejected with HTTP 405 before they
  reach an application handler. `GET`, `HEAD`, and `OPTIONS` remain available.
- The stale-cluster monitor, alert delivery worker, timeline retention worker,
  and session-pruning worker are disabled so the demo process does not perform
  background mutations.
- Read routes run as a synthetic `read_only` viewer without creating a user or
  session. Admin-only reads remain forbidden, and every mutation is still
  rejected before routing.
- Synthetic alert history and timeline entries are seeded alongside the
  cluster snapshots so the demo covers the operational features without
  connecting to a live system.
- The UI identifies the environment as a read-only synthetic demo and removes
  the agent approval navigation.
- The server applies CSP, HSTS, anti-framing, MIME sniffing, referrer, browser
  permissions, cross-origin isolation, and no-store API headers.

The demo mode is deliberately opt-in. A normal kfleet installation behaves as
before when `KFLEET_DEMO_MODE` is unset or false.

## Run locally

```bash
make web-build
KFLEET_DEMO_MODE=true go run ./cmd/hub
```

Then visit `http://localhost:8080`.

Verify the boundary:

```bash
curl -i http://localhost:8080/api/v1/meta
curl -i http://localhost:8080/api/v1/clusters
curl -i -X POST http://localhost:8080/api/v1/clusters/register \
  -H 'Content-Type: application/json' \
  --data '{"name":"should-not-be-created"}'
```

The metadata response must report `demoMode`, `readOnly`, and `syntheticData` as
true. The mutation must return HTTP 405.

## GitOps deployment

The manifests in `deploy/homelab` create a restricted `kfleet-demo` namespace
with two non-root, read-only-filesystem replicas, no service account token,
default-deny networking, explicit Traefik-only ingress, and no application
secrets. The only registry credential is the cluster-managed `ghcr-pull`
image-pull secret. It is never mounted into the container.

Flux registration and the `kfleet.sullylab.de` Cloudflare tunnel and DNS records
are maintained in `VentureForgre/homelab`.

Deployment order:

1. Merge the kfleet PR so CI publishes `ghcr.io/1solomonwakhungu/kfleet-hub:main`.
2. Merge the homelab PR so Flux registers `deploy/homelab`.
3. Provision the cluster-managed `ghcr-pull` secret in `kfleet-demo` using the
   homelab registry-secret runbook. Do not commit or mount registry credentials.
4. Confirm Flux readiness, pod readiness, response headers, synthetic metadata,
   blocked mutations, and HTTP 200 from `https://kfleet.sullylab.de`.
