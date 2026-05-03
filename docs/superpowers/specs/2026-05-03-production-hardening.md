# Production Hardening Spec

## Problem

PocketBase libSQL/Turso wrapper can accidentally run production without remote database configuration. Current remote fallback token handling risks malformed connection strings or token leakage in logs. Embedded replica shutdown does not explicitly run final sync. Release workflow lacks test/vet gate, smoke checks, and checksums.

## Goals

- Fail fast in production when remote libSQL URL or auth token is missing.
- Preserve local development fallback to SQLite unless strict production mode is enabled.
- Encode `LIBSQL_AUTH_TOKEN` safely when building remote libSQL connection strings.
- Mask `authToken` query values in logged remote URLs.
- Run final embedded replica sync before connector close during shutdown.
- Add unit tests for env parsing, strict validation, command DB-init skip behavior, URL construction, and URL masking.
- Add CI test/vet gate, Linux smoke checks, and release checksums.
- Document production env behavior, shutdown behavior, token handling, and checksum verification.
- Rename Go module to `github.com/fadlee/pocketbase-libsql`.

## Non-Goals

- No database schema changes.
- No PocketBase API behavior changes.
- No change to local developer defaults except better validation helpers.
- No new runtime dependency beyond Go standard library.
- No automatic secret rotation or Turso token management.

## Requirements

### Configuration

- Add `LIBSQL_REQUIRE_REMOTE` boolean env flag.
- Accepted truthy values: `1`, `true`, `yes`, `on`, case-insensitive, whitespace-trimmed.
- Default `LIBSQL_REQUIRE_REMOTE=false`.
- When `LIBSQL_REQUIRE_REMOTE=true`, app must fail before startup if either:
  - `LIBSQL_DATABASE_URL` empty
  - `LIBSQL_AUTH_TOKEN` empty
- When `LIBSQL_REQUIRE_REMOTE=false`, missing remote config must keep current local fallback behavior.
- Invalid `LIBSQL_SYNC_INTERVAL` must keep default `60s` behavior.

### DB init skip behavior

- DB init must be skipped for non-DB commands/flags:
  - `--help`, `-h`, `--version`, `-v`, `help`, `update`
- No-arg invocation must not skip DB init because default command may serve/start app.

### Connection string safety

- Build remote connection strings via helper, not manual string concatenation.
- Token must be URL query encoded through `net/url`.
- Existing query params must be preserved.
- Empty token must return validated raw URL unchanged.
- Invalid URL must return error.
- Logged URL masking must replace `authToken` value with `***` without re-encoding unrelated existing query values.
- Invalid masked URL input must render `<invalid-url>`.

### Embedded shutdown

- On embedded replica builds, logged remote URLs must be masked when an `authToken` query parameter is present.
- On embedded replica builds, termination hook must attempt final `Sync()` before connector `Close()`.
- If final sync fails, log warning and continue shutdown.
- Connector close warnings must remain logged.

### CI and release

- Release workflow must run `go test ./...` and `go vet ./...` before builds.
- Build job must depend on test job.
- Linux amd64 release binaries must pass `--help` and `--version` smoke checks.
- Release assets must include `checksums.txt` with SHA-256 hashes for asset files, excluding `checksums.txt` itself.
- Workflow YAML must have one top-level `jobs:` key.

### Docs

- `.env.example` must document:
  - `LIBSQL_DATABASE_URL`
  - `LIBSQL_AUTH_TOKEN`
  - `LIBSQL_REQUIRE_REMOTE`
  - `LIBSQL_SYNC_INTERVAL`
- README must include:
  - env reference table
  - production safety section
  - token masking note
  - shutdown/final sync behavior
  - checksum usage

## Acceptance Checks

```bash
go test ./...
go vet ./...
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go test ./...
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go test ./...
CGO_ENABLED=1 go build -trimpath -ldflags="-s -w" -o /tmp/pocketbase-libsql .
/tmp/pocketbase-libsql --help >/dev/null
/tmp/pocketbase-libsql --version >/dev/null
```

All commands must exit 0.
