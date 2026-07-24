# Hub authentication and authorization

kfleet authenticates human users with username and password credentials and
stores only bcrypt password hashes. A successful login creates a random,
server-side session whose raw token is held in an `HttpOnly`, `Secure`, and
`SameSite=Strict` cookie. Session tokens are stored in SQLite only as SHA-256
digests.

Health probes, static web assets, agent registration, agent heartbeats, and
agent snapshot reports do not use human sessions. Agent endpoints retain the
shared registration token and per-agent bearer token checks that existed before
human authentication was added.

## Bootstrap the first administrator

Set all three variables before the first hub start:

```bash
export KFLEET_BOOTSTRAP_ADMIN_USERNAME=admin
export KFLEET_BOOTSTRAP_ADMIN_EMAIL=admin@example.com
export KFLEET_BOOTSTRAP_ADMIN_PASSWORD='use-a-unique-password-of-12-or-more-characters'
```

The account is created only when the database contains no users. Later changes
to these variables do not overwrite accounts or passwords. Remove the password
from the runtime environment after the first successful start when your
deployment system supports that workflow.

The Helm chart exposes the same settings under `auth.bootstrapAdmin`. Protect
the values file with your normal secret-management controls. The chart defaults
to secure cookies and expects HTTPS:

```bash
helm upgrade --install kfleet-hub ./charts/kfleet-hub \
  --set registration.token="$KFLEET_REGISTRATION_TOKEN" \
  --set-string auth.bootstrapAdmin.username=admin \
  --set-string auth.bootstrapAdmin.email=admin@example.com \
  --set-string auth.bootstrapAdmin.password="$KFLEET_ADMIN_PASSWORD"
```

For local plain HTTP only, set `KFLEET_SESSION_COOKIE_INSECURE=true` or
`auth.sessionCookieInsecure=true`. Never use that setting on an internet-facing
hub.

## Roles

| Capability | read_only | operator | admin |
| --- | --- | --- | --- |
| View clusters, resources, timeline, policies, and live updates | Yes | Yes | Yes |
| Register or remove clusters | No | Yes | Yes |
| Approve pending agents | No | Yes | Yes |
| Create, update, disable, or delete users | No | No | Yes |
| Read the audit log | No | No | Yes |
| Rotate the agent registration token | No | No | Yes |

The store rejects changes that would leave the hub without an enabled admin.
Disabling or deleting a user invalidates access immediately because every
request reloads the current user state.

## Browser and API behavior

Login at `/login` or call `POST /api/v1/auth/login`. Browser mutations must
include `X-Kfleet-CSRF: 1`; the embedded UI sends it automatically. The custom
header and strict same-site cookie prevent cross-site form submissions from
performing authenticated mutations.

Authentication endpoints:

| Method | Path | Access |
| --- | --- | --- |
| `POST` | `/api/v1/auth/login` | Public |
| `POST` | `/api/v1/auth/logout` | Authenticated |
| `GET` | `/api/v1/auth/me` | Authenticated |
| `GET`, `POST` | `/api/v1/users` | Admin |
| `PATCH`, `DELETE` | `/api/v1/users/{id}` | Admin |
| `GET` | `/api/v1/audit?limit=100` | Admin |
| `POST` | `/api/v1/admin/registration-token/rotate` | Admin |

`KFLEET_SESSION_DURATION` controls session lifetime and defaults to `24h`.

## Audit log

Security-relevant events include first-admin bootstrap, successful and failed
logins, logout, user management, cluster registration and deletion, agent
approval, registration token rotation, CSRF rejection, and role denial. Audit
records never include passwords, raw session tokens, or registration tokens.

The `audit_events` table is append-only. SQLite triggers reject `UPDATE` and
`DELETE` statements, and the application store exposes only append and list
operations. Keep the database file on durable storage and include it in backup
and retention procedures.

## Database migration

The hub applies additive SQLite migrations at startup. Existing databases gain
the `users`, `sessions`, `audit_events`, and `settings` tables, append-only audit
triggers, and lookup indexes. Existing cluster and agent records remain in
place. Back up the database before upgrading, and roll back by restoring that
backup because authentication data is not written to a separate database.

## Agent registration token rotation

`POST /api/v1/admin/registration-token/rotate` returns a new raw registration
token once and persists only its hash. After the first rotation, the previous
static `KFLEET_REGISTRATION_TOKEN` is rejected. Existing approved agents keep
their per-agent credentials and continue reporting normally.
