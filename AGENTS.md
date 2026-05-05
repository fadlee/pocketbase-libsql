# AGENTS.md

## Project Snapshot

- Single-project repo: Go PocketBase/libSQL binary plus a small npm `npx` wrapper.
- Runtime code is at root (`main.go`, `db_*.go`, `libsql_conn.go`); wrapper code is in `bin/` and `lib/`.
- Release automation is in `.github/workflows/`; nearest subdirectory `AGENTS.md` files add local rules.

## Root Setup Commands

- Install Go deps: `go mod download`
- Install npm deps when touching package metadata: `npm install`
- Test/vet Go: `go test ./... && go vet ./...`
- Test npm wrapper: `npm test`
- Run embedded-replica dev target: `CGO_ENABLED=1 go run . serve`
- Run remote-only fallback target: `CGO_ENABLED=0 go run . serve`
- Build binary: `go build -trimpath -ldflags="-s -w" -o pocketbase-libsql .`
- Build Docker image: `docker build -t pocketbase-libsql:local .`

## Universal Conventions

- Keep startup/config in `main.go`; keep platform behavior in build-tagged `db_*.go` files.
- Use `libsql_conn.go` helpers for libSQL URL construction/masking; do not duplicate token/query logic.
- Keep wrapper split: CLI/process handling in `bin/`, download/cache logic in `lib/`.
- Preserve version and asset naming alignment across `go.mod`, `package.json`, docs, and workflows.
- Prefer focused root-package tests in `*_test.go`.

## Security & Secrets

- Never commit real `LIBSQL_AUTH_TOKEN` values or populated `.env` files.
- Runtime secrets belong in env vars (`LIBSQL_DATABASE_URL`, `LIBSQL_AUTH_TOKEN`).
- Mask token-bearing URLs in logs; preserve the `maskLibSQLURL` pattern.
- Treat `pb_data/` as runtime state, not source-controlled fixture data.

## JIT Index

- Core Go runtime: root (`main.go`, `db_*.go`, `libsql_conn.go`, `*_test.go`)
- npm wrapper entrypoint: `bin/` -> `bin/AGENTS.md`
- npm downloader/cache logic: `lib/` -> `lib/AGENTS.md`
- CI/release automation: `.github/` -> `.github/AGENTS.md`
- Planning/spec docs: `docs/superpowers/` -> `docs/superpowers/AGENTS.md`

## Quick Find Commands

- Startup/config: `rg -n "loadAppConfigFromEnv|validateAppConfig|shouldSkipDBInit" .`
- DB connectors: `rg -n "func dbConnect|closeConnectors" .`
- Build tags: `rg -n "^//go:build" .`
- Tests: `rg -n "^func Test" .`
- Wrapper behavior: `rg -n "ensureBinary|withDefaultDir|printWrapperHelp" bin lib`
- Release/version wiring: `rg -n "PB_VERSION|checksums|npm publish|workflow_run" .github package.json README.md NPM_README.md`

## Definition of Done

- Run `go test ./... && go vet ./...` for Go changes.
- Run `npm test` for wrapper/packaging changes.
- Update docs/workflows when behavior, release assets, env vars, or commands change.
