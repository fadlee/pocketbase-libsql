# pocketbase-libsql-bin

Run `pocketbase-libsql` release binaries from npm via `npx`.

## Usage

```bash
npx pocketbase-libsql-bin serve
```

Use specific release:

```bash
npx pocketbase-libsql-bin --pbl-version 0.37.5 serve
```

Use environment variable:

```bash
POCKETBASE_LIBSQL_VERSION=0.37.5 npx pocketbase-libsql-bin serve
```

## How it works

- detects current OS and architecture
- downloads matching binary from GitHub Releases of `fadlee/pocketbase-libsql`
- caches binary in `~/.cache/pocketbase-libsql-bin`
- forwards all remaining arguments to binary

## Supported platforms

- Linux amd64
- Linux arm64
- macOS amd64
- macOS arm64
- Windows amd64

## Notes

Wrapper expects release assets from this repository with names like:

- `pocketbase-libsql-v0.37.5-linux-amd64`
- `pocketbase-libsql-v0.37.5-linux-arm64`
- `pocketbase-libsql-v0.37.5-darwin-amd64`
- `pocketbase-libsql-v0.37.5-darwin-arm64`
- `pocketbase-libsql-v0.37.5-windows-amd64.exe`
