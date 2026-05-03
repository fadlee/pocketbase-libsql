# PocketBase with Turso/libSQL (Embedded Replica)

This project integrates **PocketBase v0.37.5** with **Turso (libSQL)** using the **Embedded Replica** mode. It provides microsecond read latency by serving reads from a local SQLite file while automatically syncing with a remote Turso primary database.

## Features

- **PocketBase v0.37.5**: Latest stable version with modern Go API.
- **Embedded Replicas** (Linux/macOS):
  - **Reads**: Served from local SQLite file (ultra-low latency).
  - **Writes**: Automatically forwarded to the remote primary database.
  - **Read-Your-Writes**: Immediate visibility of own writes.
  - **Periodic Sync**: Automatic background synchronization.
- **Cross-Platform Support**:
  - ✅ **Linux/macOS arm64**: Full embedded replica support (requires CGO).
  - ✅ **macOS amd64**: Remote-only fallback via HTTP (no CGO).
  - ✅ **Windows**: Remote-only fallback via HTTP (no CGO).
- **Hybrid Strategy**: libSQL for main data, local-only SQLite for auxiliary data (logs/system).
- **Graceful Shutdown**: Ensures all pending syncs are flushed on termination.

## Setup

1.  **Clone the repository**
2.  **Configure environment variables**:
    Create a `.env` file or set the following in your environment:
    ```env
    # Main database (Turso remote URL)
    LIBSQL_DATABASE_URL=libsql://your-db-name.turso.io
    LIBSQL_AUTH_TOKEN=your-auth-token

    # Optional: Fail fast if remote config is missing (recommended in production)
    LIBSQL_REQUIRE_REMOTE=false

    # Optional: Sync interval in seconds (defaults to 60)
    # LIBSQL_SYNC_INTERVAL=60
    ```

### Environment variables

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `LIBSQL_DATABASE_URL` | Production: yes | empty | Turso/libSQL remote database URL, for example `libsql://your-db-name.turso.io`. If empty and `LIBSQL_REQUIRE_REMOTE=false`, main DB falls back to local SQLite. |
| `LIBSQL_AUTH_TOKEN` | Production: yes | empty | Turso auth token. Required when `LIBSQL_REQUIRE_REMOTE=true`. |
| `LIBSQL_REQUIRE_REMOTE` | no | `false` | Set to `true` in production to fail fast when remote URL or token is missing. Prevents accidental local SQLite usage. |
| `LIBSQL_SYNC_INTERVAL` | no | `60` | Embedded replica background sync interval in seconds. Invalid values use default `60`. |

### Production safety

Set this in production:

```env
LIBSQL_REQUIRE_REMOTE=true
```

With this enabled, startup fails if `LIBSQL_DATABASE_URL` or `LIBSQL_AUTH_TOKEN` is missing. This prevents accidental production boot using local SQLite.

Remote URLs printed in logs are masked when an `authToken` query parameter is present.

3.  **Install dependencies**:
    ```bash
    go mod tidy
    ```

## Docker

Build image locally:

```bash
docker build -t fadlee/pocketbase-libsql:local .
```

Run container:

```bash
docker run --rm -p 8090:8090 \
  -e LIBSQL_DATABASE_URL=libsql://your-db-name.turso.io \
  -e LIBSQL_AUTH_TOKEN=your-auth-token \
  -e LIBSQL_REQUIRE_REMOTE=true \
  -v $(pwd)/pb_data:/pb/pb_data \
  fadlee/pocketbase-libsql:local
```

Image stores PocketBase data in `/pb/pb_data` and exposes port `8090`.

## Usage

### Development (Linux/macOS arm64)
Requires CGO for embedded replica mode:
```bash
CGO_ENABLED=1 go run . serve
```

### Development (macOS amd64)
Runs in remote-only mode (connects directly to Turso via HTTP):
```bash
CGO_ENABLED=0 go run . serve
```

### Development (Windows)
Runs in remote-only mode (connects directly to Turso via HTTP):
```bash
go run . serve
```

### Build
```bash
# Linux/macOS arm64 (Embedded Replica)
CGO_ENABLED=1 go build -ldflags="-s -w" -o pocketbase-turso .

# macOS Intel (amd64, Remote Fallback, CGO disabled)
CGO_ENABLED=0 go build -ldflags="-s -w" -o pocketbase-turso .

# Windows (Remote Fallback)
go build -ldflags="-s -w" -o pocketbase-turso.exe .
```

## How it Works

The project uses Go **build tags** to select the best driver for your platform:

- **Linux**: Uses `db_embedded.go` which leverages the CGO-based `go-libsql` driver. It creates a local replica of your Turso database in `pb_data/data.db`.
- **macOS arm64**: Uses `db_embedded.go` with the CGO-based `go-libsql` driver for embedded replicas.
- **macOS amd64**: Uses `db_darwin_amd64.go` with the pure-Go `libsql-client-go` driver for remote-only HTTP access.
- **Windows**: Uses `db_windows.go` which leverages the pure-Go `libsql-client-go` driver. It connects directly to Turso over HTTP.

## Shutdown behavior

On Linux and macOS arm64 embedded replica builds, the app attempts a final `Sync()` before closing the libSQL connector during PocketBase termination. If final sync fails, shutdown continues and logs a warning because network failures can happen during process termination.

## Release checksums

Release assets include `checksums.txt` with SHA-256 hashes. Verify a downloaded binary from the release asset directory with:

```bash
sha256sum -c checksums.txt --ignore-missing
```

## Platform Support

- ✅ **Linux** (amd64, arm64) - Full Embedded Replica
- ✅ **macOS** (arm64) - Full Embedded Replica
- ✅ **macOS** (amd64) - Remote-only Fallback
- ✅ **Windows** (amd64) - Remote-only Fallback

## License

MIT
