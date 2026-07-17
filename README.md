# kfleet

Multi-cluster Kubernetes management platform

## Architecture

- Hub controller (Go)
- Agent daemon (Go)
- Web UI (React + Tailwind + shadcn/ui)
- MCP Server (Go)
- Helm charts
- GitHub Actions CI/CD

## Repository Layout

- `cmd/` contains executable entry points.
- `internal/` contains private application packages.
- `pkg/` contains reusable public packages.
- `web/` contains the React web application.
- `charts/` contains Helm charts.
- `.github/` contains GitHub Actions workflows and repository configuration.

## Development

Run `make help` to list the available Makefile targets for building, testing,
linting, tidying, and cleaning the project.
