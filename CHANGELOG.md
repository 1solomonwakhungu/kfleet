# Changelog

All notable changes to kfleet are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

This changelog begins in September 2026; older releases (v0.1.0 through
v0.1.2) predate it. Their notes live on the
[GitHub releases page](https://github.com/1solomonwakhungu/kfleet/releases).

## [Unreleased]

### Added

- Users can change their own password (verifying the current one) and admins
  can reset any other user's password, from the API and the admin Users page
  (#36).
- The hub Helm chart prints post-install notes, supports injecting bootstrap
  admin credentials from an existing Kubernetes Secret
  (`auth.existingSecret`), and validates required values at install time
  instead of failing at runtime (#44).
- The web UI gives every route its own page title, shows a real 404 page, and
  makes cluster cards clickable links (#39).
- Alerts and policy findings in the web UI link to the affected cluster's
  detail page, and live pages poll for fresh data (#42).

### Changed

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

### Fixed

- Release binaries embed the built web UI again: GoReleaser now builds the UI
  before packaging, so `v*` binaries no longer ship without it (#33).
- Pending agents re-register immediately once the hub reports them approved,
  instead of waiting for the next heartbeat cycle (#34).
- Agent logs include the hub's error response body on failed requests, so
  misconfiguration (wrong token, disabled registration) is visible from the
  agent side (#38).

### Security

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
