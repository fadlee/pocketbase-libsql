# AGENTS.md

## Package Identity

- `bin/` is the npm CLI entrypoint for `npx pocketbase-libsql-bin`.
- It is CommonJS Node.js that prepares args and spawns the downloaded Go binary.

## Setup & Run

- Wrapper help: `node bin/runner.js --help`
- Smoke test: `npm test`
- Run wrapper: `node bin/runner.js serve`
- Run with explicit data dir: `node bin/runner.js --dir=./pb_data serve`
- Preview package: `npm pack --dry-run`

## Patterns & Conventions

- Keep `bin/runner.js` focused on args, help text, child process lifecycle, and signal forwarding.
- ✅ DO: preserve `withDefaultDir()` in `bin/runner.js` so `--dir` defaults to `path.join(process.cwd(), 'pb_data')` only when absent.
- ✅ DO: keep binary resolution delegated to `lib/downloader.js` via `ensureBinary()`.
- ✅ DO: keep `spawn(..., { stdio: 'inherit', shell: false })` unless there is a tested cross-platform reason to change it.
- ❌ DON'T: move download, cache, asset naming, or checksum logic into `bin/runner.js`; that belongs in `lib/downloader.js`.
- ❌ DON'T: hardcode absolute data paths or swallow child process signals/errors.

## Touch Points / Key Files

- Wrapper entrypoint: `bin/runner.js`
- Downloader/cache implementation: `lib/downloader.js`
- npm metadata: `package.json`
- npm publish workflow: `.github/workflows/npm-publish.yml`

## JIT Index Hints

- Arg handling: `rg -n "hasDirArg|withDefaultDir|rawArgs" bin`
- Process lifecycle: `rg -n "spawn\\(|child\\.on|process\\.on" bin`
- Downloader calls: `rg -n "ensureBinary" bin lib`
- Package scripts/bin: `rg -n "\"scripts\"|\"bin\"" package.json`

## Common Gotchas

- `--help` prints wrapper help and still resolves the binary.
- Default `--dir=./pb_data` affects local UX and docs.
- Binary version is tied to `package.json` and GitHub release asset names.

## Pre-PR Checks

`npm test && npm pack --dry-run`
