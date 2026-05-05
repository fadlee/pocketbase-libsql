# AGENTS.md

## Package Identity

- `lib/` contains npm-side download, cache, platform, and checksum support for release binaries.
- It is CommonJS Node.js using built-in modules only.

## Setup & Run

- Trigger through wrapper: `node bin/runner.js --help`
- Smoke test: `npm test`
- Preview package: `npm pack --dry-run`
- Inspect target info: `node -e "const d=require('./lib/downloader'); d.getBinaryInfo().then(console.log)"`

## Patterns & Conventions

- Keep `lib/downloader.js` as the source of truth for platform mapping, asset naming, cache layout, and checksums.
- ✅ DO: update `getPlatformInfo()` and `getAssetName()` together when platform support changes.
- ✅ DO: keep cache paths versioned under `~/.cache/pocketbase-libsql-bin` via `getCacheDir()` and `getBinaryInfo()`.
- ✅ DO: preserve checksum verification through `tryVerifyChecksum()`; best-effort warnings are okay, mismatches are not.
- ✅ DO: keep download URLs aligned with `REPO = 'fadlee/pocketbase-libsql'` and `PACKAGE_VERSION`.
- ❌ DON'T: add HTTP/download dependencies unless built-ins are insufficient.
- ❌ DON'T: duplicate asset naming rules in `bin/runner.js` or workflow scripts without updating this file.

## Touch Points / Key Files

- Downloader implementation: `lib/downloader.js`
- Wrapper caller: `bin/runner.js`
- Version source: `package.json`
- Release asset creator: `.github/workflows/release.yml`
- npm publisher: `.github/workflows/npm-publish.yml`

## JIT Index Hints

- Platform mapping: `rg -n "getPlatformInfo|platformName|archName" lib`
- Cache/version logic: `rg -n "getCacheDir|getBinaryInfo|getRequestedVersion" lib`
- Checksums: `rg -n "checksums|sha256|tryVerifyChecksum" lib`
- Download flow: `rg -n "downloadFile|downloadBinary|ensureBinary" lib`
- Version coupling: `rg -n "PACKAGE_VERSION|REPO|releases/download" lib package.json .github/workflows`

## Common Gotchas

- Asset names must exactly match GitHub release filenames.
- Windows needs `.exe`; Unix targets need executable permissions.
- `PACKAGE_VERSION` comes from `package.json`, which release automation rewrites before publish.

## Pre-PR Checks

`npm test && node -e "const d=require('./lib/downloader'); d.getBinaryInfo().then(console.log)"`
