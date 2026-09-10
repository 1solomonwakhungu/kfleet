# kfleet documentation

| Document | What it covers |
| --- | --- |
| [REST API reference](api.md) | Every hub HTTP route: method, auth requirement, request/response shape, purpose, error format, and pagination. |
| [Operations runbook](operations.md) | Operator day-2: full environment variable reference, health probes and alerting, upgrade and backup procedure, troubleshooting scenarios. |
| [Hub authentication and authorization](authentication.md) | Bootstrap admin, roles and RBAC, sessions and CSRF, password lifecycle, audit log, backups, agent registration and token rotation. |
| [Fleet health alerts](alerts.md) | Alert rules and lifecycle, acknowledgement, signed webhook delivery with retries and dead-lettering, local receiver and failure injection. |
| [Pod log streaming](pod-logs.md) | The agent reverse-channel design, SSE frames, disconnect behavior, limits, and required agent RBAC. |
| [Policy and configuration drift](policy-drift.md) | Built-in read-only policies, result states, API filters, tenant isolation, and snapshot compatibility. |
| [Public read-only demo](public-demo.md) | The synthetic-data demo deployment, its security boundary, and how it is shipped. |
| [Brand](brand/README.md) | Logo assets and usage. |
| [UI mockups (GitHub Primer)](mockups/primer/README.md) | Static design exploration of the control plane screens; not shipped UI. |
