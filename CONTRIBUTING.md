# Contributing to kfleet

Thanks for helping improve kfleet. Bug reports, documentation fixes, design discussions, and code contributions are all welcome.

## Pull request flow

1. Open or choose an issue and describe the problem you want to solve. For substantial changes, agree on the approach before investing in an implementation.
2. Fork the repository and create a focused branch from the latest `main`.
3. Make the smallest coherent change, including tests and documentation where appropriate.
4. Run the required checks locally.
5. Commit with a Conventional Commit message and DCO sign-off.
6. Push your branch and open a pull request against `main`. Explain what changed, why it changed, how it was tested, and any operational or compatibility impact.
7. Address review feedback with additional signed-off commits. Keep the PR focused; unrelated changes should use a separate issue and PR.

Maintainers may squash or rebase a PR during merge. CI must pass and all review conversations must be resolved before merging.

## Commit messages

Use [Conventional Commits](https://www.conventionalcommits.org/) with a short imperative description:

```text
feat(agent): collect deployment status
fix(hub): preserve cluster labels on heartbeat
docs: clarify kind quickstart
test(mcp): cover warning event filtering
```

Use `feat` for new behavior, `fix` for bug fixes, and types such as `docs`, `test`, `refactor`, `build`, or `chore` where appropriate. Add a scope when it makes the affected area clearer.

## Developer Certificate of Origin

All commits must include a `Signed-off-by` trailer to certify the [Developer Certificate of Origin](https://developercertificate.org/). Create it automatically with Git's `-s` flag:

```bash
git commit -s -m "feat(hub): add cluster filter"
```

The name and email in the trailer must match the commit author. If a commit is missing the trailer, amend it and force-push your branch safely:

```bash
git commit --amend --signoff --no-edit
git push --force-with-lease
```

## Tests and linting

Before opening or updating a pull request, run:

```bash
make test
make lint
go build ./...
go vet ./...
```

Frontend changes should also build the embedded web UI:

```bash
make web-build
```

Add or update tests for behavior changes. Do not commit generated binaries, local databases, credentials, or environment-specific configuration.

## Releases

Releases are automated with [release-please](https://github.com/googleapis/release-please).

1. Every push to `main` updates an open release PR titled `chore: release X.Y.Z`. It bumps the version in `.release-please-manifest.json` and adds a `CHANGELOG.md` entry built from the Conventional Commit messages merged since the last tag. `feat` commits bump the minor version, `fix` commits bump the patch version, and breaking changes bump the minor version while kfleet is below 1.0.0.
2. Merging the release PR creates the `vX.Y.Z` tag and GitHub release, then runs `.github/workflows/release.yml` from the same workflow run. That publishes the hub and agent images to GHCR, the Helm charts to `oci://ghcr.io/1solomonwakhungu/charts`, and the GoReleaser binaries and Homebrew cask.

Squash-merge titles become changelog entries, so keep PR titles in Conventional Commit form. To adjust the notes, edit `CHANGELOG.md` on the release PR branch right before merging; release-please rewrites that branch on every push to `main`.

The repository setting **Settings → Actions → General → Allow GitHub Actions to create and approve pull requests** must be enabled for release-please to open the release PR with `GITHUB_TOKEN`. Pull requests opened by `GITHUB_TOKEN` do not trigger CI. To run CI on release PRs, add a fine-grained personal access token with contents and pull-request write access as the `RELEASE_PLEASE_TOKEN` secret. With that secret set, the tag push itself triggers `release.yml` and the release-please workflow skips its own publish job.

Manual releases still work: push a `v*` tag, or run the Release workflow from the Actions tab with an existing tag.
