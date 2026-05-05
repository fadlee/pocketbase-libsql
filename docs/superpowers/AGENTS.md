# AGENTS.md

## Package Identity

- `docs/superpowers/` stores Markdown plans and specs, not runtime code.
- Use it for decision history, requirements, and acceptance checks.

## Setup & Run

- List docs: `find docs/superpowers -maxdepth 2 -type f | sort`
- Search decisions: `rg -n "Problem|Goals|Requirements|Acceptance Checks" docs/superpowers`
- Cross-check behavior: `rg -n "LIBSQL_REQUIRE_REMOTE|checksums|final sync|mask" README.md main.go libsql_conn.go .github/workflows`

## Patterns & Conventions

- Keep docs decision-oriented: problem, goals, non-goals, requirements, acceptance checks.
- ✅ DO: follow `docs/superpowers/specs/2026-05-03-production-hardening.md` for spec structure.
- ✅ DO: follow `docs/superpowers/plans/2026-05-03-production-hardening.md` for implementation-plan style.
- ✅ DO: include copy-pasteable validation commands when behavior changes.
- ❌ DON'T: paste full code listings; point to files like `main.go`, `libsql_conn.go`, or `.github/workflows/release.yml`.
- ❌ DON'T: let env vars, commands, or workflow names drift from the repo.

## Touch Points / Key Files

- Example plan: `docs/superpowers/plans/2026-05-03-production-hardening.md`
- Example spec: `docs/superpowers/specs/2026-05-03-production-hardening.md`
- Runtime config: `main.go`
- URL handling: `libsql_conn.go`
- Release workflow: `.github/workflows/release.yml`

## JIT Index Hints

- Acceptance criteria: `rg -n "Acceptance Checks|Requirements|Non-Goals" docs/superpowers`
- Env vars: `rg -n "LIBSQL_" docs/superpowers README.md main.go`
- Release hardening: `rg -n "checksums|smoke checks|go vet|go test" docs/superpowers .github/workflows`
- Code/doc comparison: `rg -n "final sync|maskLibSQLURL|shouldSkipDBInit" .`

## Common Gotchas

- These docs describe intent; verify against code before relying on them.
- Acceptance commands should match actual repository commands.

## Pre-PR Checks

`rg -n "LIBSQL_|checksums|final sync|workflow_run" docs/superpowers README.md main.go libsql_conn.go .github/workflows`
