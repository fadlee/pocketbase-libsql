# AGENTS.md

## Package Identity

- `.github/` contains GitHub Actions for binary releases, npm publish, and Docker publish.
- Workflow scripts must stay aligned with Go build targets and npm downloader asset names.

## Setup & Run

- Go gate: `go test ./... && go vet ./...`
- npm gate: `npm test && npm pack --dry-run`
- Docker build: `docker build -t pocketbase-libsql:local .`
- Review workflows: `rg -n "workflow_run|PB_VERSION|checksums|docker/build-push-action|npm publish" .github/workflows`

## Patterns & Conventions

- Keep responsibilities split: `release.yml` builds GitHub release assets, `npm-publish.yml` publishes npm, `docker-release.yml` publishes images.
- ✅ DO: preserve `workflow_run` dependency from successful `release` into npm and Docker publish workflows.
- ✅ DO: keep checksum generation and Linux smoke checks in `.github/workflows/release.yml`.
- ✅ DO: resolve PocketBase version from `go.mod` with `go list -m -f '{{.Version}}' github.com/pocketbase/pocketbase`.
- ✅ DO: keep release asset names compatible with `lib/downloader.js`.
- ❌ DON'T: rename workflows or remove tag triggers without updating downstream references.
- ❌ DON'T: change Docker/npm tag derivation independently from release tags.

## Touch Points / Key Files

- Binary release: `.github/workflows/release.yml`
- npm publish: `.github/workflows/npm-publish.yml`
- Docker publish: `.github/workflows/docker-release.yml`
- Docker image: `Dockerfile`
- npm asset consumer: `lib/downloader.js`

## JIT Index Hints

- Triggers: `rg -n "tags:|workflow_run|refs/tags" .github/workflows`
- Matrix/platforms: `rg -n "matrix:|goos:|goarch:|CGO_ENABLED" .github/workflows/release.yml`
- Checksums/assets: `rg -n "checksums|upload-artifact|action-gh-release" .github/workflows/release.yml`
- npm publish path: `rg -n "npm publish|npm pack|package.json|NPM_README.md" .github/workflows/npm-publish.yml`
- Docker publish path: `rg -n "build-push-action|metadata-action|DOCKERHUB|GHCR" .github/workflows/docker-release.yml`

## Common Gotchas

- `npm-publish.yml` swaps `README.md` content from `NPM_README.md` during publish prep.
- Docker/npm publish jobs require a successful release workflow and a tag on the checked-out commit.
- Downloader asset naming in `lib/downloader.js` must match `release.yml` outputs.

## Pre-PR Checks

`go test ./... && go vet ./... && npm test && npm pack --dry-run`
