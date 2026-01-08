# PocketBase with Turso/libSQL

This project integrates **PocketBase v0.35** with **Turso (libSQL)** as the primary database driver. It uses Turso for the main application data (`data.db`) and local SQLite for the auxiliary data (`auxiliary.db`).

## Features

- **PocketBase v0.35**: Latest stable version with modern Go API.
- **Turso/libSQL**: Cloud-native SQLite distribution for distributed data.
- **Dual Database Strategy**: Turso for main data, local SQLite for logs/system metadata.
- **Optimized Logging**: Connection status is logged only once on startup.
- **Pinned Dependencies**: Ensures stability by pinning `modernc.org/sqlite` and `libc` to PocketBase's tested versions.

## Setup

1.  **Clone the repository**
2.  **Configure environment variables**:
    Create a `.env` file or set the following in your environment:
    ```env
    TURSO_DATABASE_URL=libsql://your-db-name.turso.io
    TURSO_AUTH_TOKEN=your-auth-token
    ```
3.  **Install dependencies**:
    ```bash
    go mod tidy
    ```

## Usage

### Development
Run the server in development mode:
```bash
go run . serve
```

### Build
Build a production binary:
```bash
go build -o myapp main.go
```

To reduce binary size by excluding the default SQLite driver (since we use libSQL):
```bash
go build -tags no_default_driver -ldflags="-s -w" -o myapp main.go
```

## Configuration

The database connection logic is defined in `main.go` using the `pocketbase.Config.DBConnect` hook. This allows for seamless switching between Turso and local SQLite depending on the environment configuration.

## License

MIT
