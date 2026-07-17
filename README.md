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

## Releases

Pushing a tag matching `v*` starts the release workflow. GoReleaser builds and
publishes hub and agent archives, pushes versioned and `latest` container images
to GHCR, and creates the GitHub release. A dependent job packages the hub and
agent Helm charts and publishes them to `oci://ghcr.io/1solomonwakhungu/charts`.
The workflow can also be started manually from the GitHub Actions page.
