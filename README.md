# PocketBase with Turso/libSQL (Embedded Replica)

This project integrates **PocketBase v0.35.1** with **Turso (libSQL)** using the **Embedded Replica** mode. It provides microsecond read latency by serving reads from a local SQLite file while automatically syncing with a remote Turso primary database.

## Features

- **PocketBase v0.35.1**: Latest stable version with modern Go API.
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

    # Optional: Sync interval in seconds (defaults to 60)
    # LIBSQL_SYNC_INTERVAL=60
    ```
3.  **Install dependencies**:
    ```bash
    go mod tidy
    ```

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

## Platform Support

- ✅ **Linux** (amd64, arm64) - Full Embedded Replica
- ✅ **macOS** (arm64) - Full Embedded Replica
- ✅ **macOS** (amd64) - Remote-only Fallback
- ✅ **Windows** (amd64) - Remote-only Fallback

## License

MIT
