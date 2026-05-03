# PocketBase libSQL

PocketBase image with libSQL/Turso support.

## Image

- Docker Hub: `fadlee/pocketbase-libsql`
- GHCR: `ghcr.io/fadlee/pocketbase-libsql`

## Supported tags

- `latest`
- `v0.37.5`

## Exposed port

- `8090`

## Volumes

Container stores PocketBase data in:

- `/pb/pb_data`

Optional directories you may also mount:

- `/pb/pb_hooks`
- `/pb/pb_migrations`

## Required environment variables

For production with Turso/libSQL remote database:

- `LIBSQL_DATABASE_URL`
- `LIBSQL_AUTH_TOKEN`
- `LIBSQL_REQUIRE_REMOTE=true`

Optional:

- `LIBSQL_SYNC_INTERVAL=60`

## Quick start

```bash
docker run --rm -p 8090:8090 \
  -e LIBSQL_DATABASE_URL=libsql://your-db-name.turso.io \
  -e LIBSQL_AUTH_TOKEN=your-auth-token \
  -e LIBSQL_REQUIRE_REMOTE=true \
  -v $(pwd)/pb_data:/pb/pb_data \
  fadlee/pocketbase-libsql:latest
```

Open:

```text
http://localhost:8090/_/
```

## Notes

- Linux container build targets `linux/amd64` and `linux/arm64`
- Embedded replica behavior depends on platform support inside build/runtime
- In production, set `LIBSQL_REQUIRE_REMOTE=true` to prevent accidental local SQLite fallback

## Source

- GitHub: `https://github.com/fadlee/pocketbase-libsql`
