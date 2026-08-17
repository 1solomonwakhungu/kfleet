# Status

## Completed

- Redesigned the landing page with Hallmark's Map / Diagram macrostructure.
- Replaced faux dashboard and terminal chrome with an accessible SVG fleet topology and plain command ledger.
- Added a Cobalt-dark semantic token sheet, N13 command navigation, and Ft2 inline footer.
- Added the Better Design Linear-based React landing page under `landing-page/`.
- Added responsive desktop and mobile layouts, project metadata, social card, favicon, and Vercel configuration.
- Installed Better Design registry components and owned Tabler icon components.
- Verified the production build and linted the landing page.
- Passed Better Design accessibility/content review and comprehension review.
- Deployed the production site to https://kfleet-landing.vercel.app.
- Opened and merged GitHub PR #10 after all five CI checks passed.
- Added the production URL to the GitHub repository homepage and updated the repository description.
- Published deployment is now defined by the production-ready `kfleet-hub` and `kfleet-agent` Helm charts.
- Added tag-driven OCI chart packaging and publishing to the release workflow.
- Added Helm lint and render checks to CI and removed duplicate placeholder charts.
- Added a no-clone `kfleet` quickstart CLI with status, browser opening, and cleanup commands.
- Configured GoReleaser to publish the CLI through the existing Homebrew tap.
- Added release-attached Helm charts for version-aligned quickstarts.
- Added the admin web surface: user management, audit log, agent registration
  token rotation, and guarded cluster removal, each gated on the signed-in
  user's role.

## Results

- Landing page build: passed
- GitHub CI checks: 5/5 passed
- Better Design comprehension issues: 0 blocking
- Hallmark slop test: 58/58 passed
- Production deployment: ready
- Helm charts linted and rendered: 2/2 passed
- Version synchronization: chart `1.2.3` deploys image `v1.2.3`
- Go test suite: passed
- CLI tests with race detector: passed
- Full Go test suite with race detector: passed
- Web test suite: 61 tests passed across 16 files
- GoReleaser cross-platform snapshot: passed
- Helm chart lint: 2/2 passed
- Release `v0.1.2`: published successfully
- GitHub release assets: 9
- OCI Helm charts pulled and rendered: 2/2 passed
- Multi-architecture image manifests verified: linux/amd64 and linux/arm64

## Next

- Future semantic version tags publish matching binaries, images, and OCI Helm charts automatically.
