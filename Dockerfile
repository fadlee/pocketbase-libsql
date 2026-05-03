# syntax=docker/dockerfile:1.7

FROM --platform=$TARGETPLATFORM golang:1.26-bookworm AS builder

ARG POCKETBASE_VERSION=v0.37.5

WORKDIR /src

RUN apt-get update \
    && apt-get install -y --no-install-recommends build-essential pkg-config ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN set -eux; \
    export CGO_ENABLED=1; \
    LDFLAGS="-s -w -X github.com/pocketbase/pocketbase.Version=${POCKETBASE_VERSION}"; \
    go build -trimpath -ldflags="${LDFLAGS}" -o /out/pocketbase-libsql .; \
    mkdir -p /out/pb/pb_data /out/pb/pb_hooks /out/pb/pb_migrations

FROM gcr.io/distroless/cc-debian12:nonroot

WORKDIR /pb

COPY --from=builder /out/pocketbase-libsql /usr/local/bin/pocketbase
COPY --chown=nonroot:nonroot --from=builder /out/pb /pb

VOLUME ["/pb/pb_data", "/pb/pb_hooks", "/pb/pb_migrations"]

EXPOSE 8090

ENTRYPOINT ["/usr/local/bin/pocketbase"]
CMD ["serve", "--http=0.0.0.0:8090"]
