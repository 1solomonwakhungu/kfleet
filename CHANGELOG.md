# Changelog

All notable changes to kfleet are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

This changelog begins in September 2026; older releases (v0.1.0 through
v0.1.2) predate it. Their notes live on the
[GitHub releases page](https://github.com/1solomonwakhungu/kfleet/releases).

## [0.2.1](https://github.com/1solomonwakhungu/kfleet/compare/v0.2.0...v0.2.1) (2026-10-17)


### Bug fixes

* **release:** restore web dist placeholders before GoReleaser ([#59](https://github.com/1solomonwakhungu/kfleet/issues/59)) ([3a31f1e](https://github.com/1solomonwakhungu/kfleet/commit/3a31f1ef06d41a6ef742c54ec3073249bfa89d1c))

## [0.2.0](https://github.com/1solomonwakhungu/kfleet/compare/v0.1.2...v0.2.0) (2026-10-17)

### Added

- The MCP server adds a `get_pod_logs` tool that returns recent pod
  container logs as text, including the previous terminated instance
  (`previous: true`) for crashlooping pods, and a `list_namespaces` tool.
  The hub log endpoint and agents accept matching `previous` and
  `sinceSeconds` options (#56).
- The hub sends `Cache-Control: no-cache` for `index.html` (including SPA
  fallback routes) and non-hashed static files, and
  `Cache-Control: public, max-age=31536000, immutable` for the content-hashed
  files under `assets/`. Deployed UIs now pick up new bundles on the next
  navigation instead of requiring a hard refresh.
- The hub exposes `GET /metrics` in the Prometheus text format with
  aggregate fleet gauges (registered agents, dead-lettered alerts, live log
  relays, dashboard clients, active log streams, database size). It is
  unauthenticated by design and can be disabled with
  `KFLEET_METRICS_ENABLED=false` (#48).
- Every hub response carries an `X-Request-ID` header — honoring a
  well-formed incoming value or minting a UUID — and the same ID appears in
  the request's log lines for end-to-end correlation (#48).
- `/api/v1/meta` now reports the hub build `version`, and hub startup logs
  include it (#48).
- Agents can verify a hub that uses a private certificate authority:
  `KFLEET_HUB_CA` (or the agent chart's `hub.caBundle`) supplies a PEM CA
  bundle, and the agent chart accepts the registration token from an
  existing Secret via `hub.existingSecret` (#40).
- Users can change their own password (verifying the current one) and admins
  can reset any other user's password, from the API and the admin Users page
  (#36).
- The agent honors `KFLEET_LOG_LEVEL` (`debug`, `info`, `warn`, `error`),
  matching the hub's log-level configuration (#46).
- The web UI makes cluster detail tabs, namespaces, and pods deep-linkable,
  gives every route its own page title, shows a real 404 page, and makes
  cluster cards clickable links (#47, #39).
- The hub Helm chart prints post-install notes, supports injecting
  registration and bootstrap-admin credentials from existing Kubernetes
  Secrets (`registration.existingSecret`, `auth.existingSecret`), and
  validates required values at install time instead of failing at runtime
  (#44).
- Alerts and policy findings in the web UI link to the affected cluster's
  detail page, and live pages poll for fresh data (#42).

### Changed

- Releases are automated with release-please: merging the release PR tags
  the version and publishes images, Helm charts, binaries, and the Homebrew
  cask from the same workflow run (#55).
- The agent's `/readyz` is honest about hub reachability: it reports ready
  only after a successful registration and while the last successful hub
  contact is within 45 seconds, answering `503` otherwise. Deployments that
  gate traffic on readiness will now hold back an agent that cannot reach
  the hub instead of routing to a disconnected one (#46).
- Agent environment parsing was consolidated with clearer startup errors
  (#46).
- Cleaned non-project debris out of the repository (#31).
- CI runs the web test suite, adds `govulncheck` and Dependabot, pins GitHub
  Action versions, and caches Docker builds (#35).
- Unknown or wrongly-methoded `/api` requests now return the standard JSON
  error format (with correct `Allow` headers on 405) instead of the Go mux's
  plain-text responses; unused legacy policy route aliases were removed
  (#37).
- The hub enables SQLite WAL journaling and `synchronous=NORMAL` pragmas and
  batches timeline/session pruning, improving write throughput and probe
  responsiveness (#43).
- Hub request logging now includes the HTTP status and request ID, and
  routine liveness/readiness probe traffic logs at debug level instead of
  flooding the operational log (#48).

### Fixed

- Release binaries embed the built web UI again: GoReleaser now builds the UI
  before packaging, so `v*` binaries no longer ship without it (#33).
- Pending agents re-register immediately once the hub reports them approved,
  instead of waiting for the next heartbeat cycle (#34).
- Agent logs include the hub's error response body on failed requests, so
  misconfiguration (wrong token, disabled registration) is visible from the
  agent side (#38).

### Security

- The agent chart ships least-privilege RBAC: the agent's Kubernetes role is
  narrowed to the collection and log-streaming calls it actually makes
  (#40).
- Agent registration fails closed: when no registration token is configured
  (and none has ever been rotated), the hub rejects registration attempts
  with `403` instead of admitting unauthenticated agents (#41).
- Re-registering an existing cluster name now requires proof of the current
  agent credential (`existingAgentToken`); without it the hub returns `409`
  instead of rotating the token for an unverified caller, preventing cluster
  identity takeover via the registration token (#41).
- Password changes and admin resets invalidate all sessions for the target
  account, verify the current password on self-changes, and record
  `user.password_update` audit events (#36).
- Role-denial responses name the required role, and heartbeat failures no
  longer leak internal store error strings to clients (#32).
- Go toolchain bumped to clear standard-library vulnerabilities flagged by
  `govulncheck` (#45).
- Go 1.26.9 and `golang.org/x/net` v0.60.0 clear GO-2026-6611, GO-2026-6612,
  GO-2026-6613, and GO-2026-6617 flagged by `govulncheck` (#57).

### Commits

#### Features

* add PostHog analytics to landing page ([#29](https://github.com/1solomonwakhungu/kfleet/issues/29)) ([93ae2ca](https://github.com/1solomonwakhungu/kfleet/commit/93ae2ca13df1ec0c37d70c384b40827506764ac8))
* **agent:** honest readiness probe, log-level config, consolidated env parsing ([#46](https://github.com/1solomonwakhungu/kfleet/issues/46)) ([28f2582](https://github.com/1solomonwakhungu/kfleet/commit/28f258294f1b14ba1b44c335d9c871915efb30c3))
* **agent:** least-privilege RBAC, optional hub CA bundle, existingSecret token ([#40](https://github.com/1solomonwakhungu/kfleet/issues/40)) ([6eac8de](https://github.com/1solomonwakhungu/kfleet/commit/6eac8def72a840bb48c4a8c6c8dacb378d347e26))
* **cli:** add Homebrew quickstart ([#19](https://github.com/1solomonwakhungu/kfleet/issues/19)) ([05646d2](https://github.com/1solomonwakhungu/kfleet/commit/05646d290edf245e9211ac53b60cf36a67e30845))
* **helm:** hub chart NOTES, existingSecret support, install-time guards ([#44](https://github.com/1solomonwakhungu/kfleet/issues/44)) ([6409464](https://github.com/1solomonwakhungu/kfleet/commit/6409464a100928c3a55c2c4ff9bfd68b000367e7))
* **hub,agent:** stream pod logs end to end ([#24](https://github.com/1solomonwakhungu/kfleet/issues/24)) ([963278c](https://github.com/1solomonwakhungu/kfleet/commit/963278c5e172e1cef14d393d270b94ddce66cce2))
* **hub:** /metrics endpoint, hub version in meta and logs, request IDs, quieter probes ([#48](https://github.com/1solomonwakhungu/kfleet/issues/48)) ([eaf51fe](https://github.com/1solomonwakhungu/kfleet/commit/eaf51fecc7781fffb1f7bdb6efeddd67f8d8eaf9))
* **mcp:** add get_pod_logs and list_namespaces tools ([#56](https://github.com/1solomonwakhungu/kfleet/issues/56)) ([dd950a2](https://github.com/1solomonwakhungu/kfleet/commit/dd950a2ee9d80bc3e5853c02c69a4df39dd1b323))
* user password change and reset ([#36](https://github.com/1solomonwakhungu/kfleet/issues/36)) ([0fa08d3](https://github.com/1solomonwakhungu/kfleet/commit/0fa08d39d2d4a83be6d82ef8266b26fb2a6a10c9))
* **web:** add admin UI for users, audit log, token rotation, and cluster removal ([#23](https://github.com/1solomonwakhungu/kfleet/issues/23)) ([a66df2c](https://github.com/1solomonwakhungu/kfleet/commit/a66df2cf465fcaa91891e3b02fadd15e2f1d61b4))
* **web:** cascading searchable Namespace + Pod dropdowns in Logs tab ([#51](https://github.com/1solomonwakhungu/kfleet/issues/51)) ([a3447fc](https://github.com/1solomonwakhungu/kfleet/commit/a3447fcfb2483826f57c7550f170032bb3f2a59f))
* **web:** link alerts and policy findings to cluster detail; poll live data ([#42](https://github.com/1solomonwakhungu/kfleet/issues/42)) ([1091194](https://github.com/1solomonwakhungu/kfleet/commit/1091194d20589504181e18e722550e7f6f1f1057))
* **web:** make cluster detail tab, namespace, and pod deep-linkable ([#47](https://github.com/1solomonwakhungu/kfleet/issues/47)) ([93477b9](https://github.com/1solomonwakhungu/kfleet/commit/93477b93c35c4c5323b9e87c2a611ccd0e846fc6))
* **web:** per-route titles, 404 page, cluster card as link ([#39](https://github.com/1solomonwakhungu/kfleet/issues/39)) ([9ed446c](https://github.com/1solomonwakhungu/kfleet/commit/9ed446c12f7ba9a8fc36687b1523df12e113c943))


#### Bug fixes

* **agent,hub:** validate agent config, stamp real version, probe readiness ([#22](https://github.com/1solomonwakhungu/kfleet/issues/22)) ([c1745a9](https://github.com/1solomonwakhungu/kfleet/commit/c1745a9d1dba1429519e0a39d330e7c3e678379e))
* **agent:** re-register as soon as the hub reports the agent approved ([#34](https://github.com/1solomonwakhungu/kfleet/issues/34)) ([790e95b](https://github.com/1solomonwakhungu/kfleet/commit/790e95b1e6892ee1db260d28046dd8aa83821be7))
* **agent:** surface hub error bodies in agent logs ([#38](https://github.com/1solomonwakhungu/kfleet/issues/38)) ([607eca4](https://github.com/1solomonwakhungu/kfleet/commit/607eca442ec2fa3a9ced1b3fb7281a82df7cf500))
* **deps:** bump golang.org/x/net and Go toolchain for govulncheck ([#57](https://github.com/1solomonwakhungu/kfleet/issues/57)) ([208ce33](https://github.com/1solomonwakhungu/kfleet/commit/208ce33afb3057262d0a3112533af021b1ea520d))
* **hub:** cache-control headers for embedded web bundle ([#54](https://github.com/1solomonwakhungu/kfleet/issues/54)) ([b412f0f](https://github.com/1solomonwakhungu/kfleet/commit/b412f0f65fe38fceb2789b612e8da14a04f32d98))
* **hub:** enable WAL pragmas and batch event/session pruning ([#43](https://github.com/1solomonwakhungu/kfleet/issues/43)) ([e522b9a](https://github.com/1solomonwakhungu/kfleet/commit/e522b9a84d4ecaf7e33cac515b362e156a7b9a5f))
* **hub:** fail closed on missing registration token and guard cluster re-registration ([#41](https://github.com/1solomonwakhungu/kfleet/issues/41)) ([76421b6](https://github.com/1solomonwakhungu/kfleet/commit/76421b60cc5e8b2c1a75f4e774a06910a5eb69a4))
* **hub:** issue real agent tokens on cluster registration ([#21](https://github.com/1solomonwakhungu/kfleet/issues/21)) ([76f3415](https://github.com/1solomonwakhungu/kfleet/commit/76f34157567689be4149d361a8a4d01f4fb3f496))
* **hub:** name the required role in 403s; stop leaking heartbeat store errors ([#32](https://github.com/1solomonwakhungu/kfleet/issues/32)) ([6e1a99f](https://github.com/1solomonwakhungu/kfleet/commit/6e1a99f887dde26b8fd6e5c7dc8a9a07efa63ef6))
* **hub:** return JSON errors for unknown API routes; drop policy aliases ([#37](https://github.com/1solomonwakhungu/kfleet/issues/37)) ([1326f24](https://github.com/1solomonwakhungu/kfleet/commit/1326f24e7cc6bceea077571c944b4f0de9768a7f))
* **release:** build the web UI before GoReleaser so release binaries embed it ([#33](https://github.com/1solomonwakhungu/kfleet/issues/33)) ([887a8bb](https://github.com/1solomonwakhungu/kfleet/commit/887a8bbb00586caedc9cf94e36305152d7b1381f))
* **web:** remove health-colored left border on cluster cards ([#52](https://github.com/1solomonwakhungu/kfleet/issues/52)) ([06937d3](https://github.com/1solomonwakhungu/kfleet/commit/06937d3f33b80eb3f289fc3696faa9f63c75e0c7))
* **web:** remove provisioned-account hint from login page ([#50](https://github.com/1solomonwakhungu/kfleet/issues/50)) ([b0b3b7d](https://github.com/1solomonwakhungu/kfleet/commit/b0b3b7d210fb234250072dd0f60491929833856e))
* **web:** tighten Logs tab controls spacing to match button gap ([#53](https://github.com/1solomonwakhungu/kfleet/issues/53)) ([c2a2606](https://github.com/1solomonwakhungu/kfleet/commit/c2a260613306fd16be810a3223a7def69637f579))


#### Documentation

* add API reference, operations runbook, changelog, docs index ([#49](https://github.com/1solomonwakhungu/kfleet/issues/49)) ([e69c6f2](https://github.com/1solomonwakhungu/kfleet/commit/e69c6f230f9478efbe368101f391d893ae17608b))
* record v0.1.2 publication ([#20](https://github.com/1solomonwakhungu/kfleet/issues/20)) ([095f95d](https://github.com/1solomonwakhungu/kfleet/commit/095f95db72a3da5c4a9a5fb5aa8efa70835d052b))
