# syntax=docker/dockerfile:1.7

FROM --platform=$BUILDPLATFORM golang:1.26-bookworm AS builder

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
ARG POCKETBASE_VERSION=v0.37.5

WORKDIR /src

RUN apt-get update \
    && apt-get install -y --no-install-recommends build-essential pkg-config ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN set -eux; \
    export GOOS=${TARGETOS:-linux}; \
    case "${TARGETARCH}" in \
      amd64) export GOARCH=amd64 ;; \
      arm64) export GOARCH=arm64 ;; \
      *) echo "unsupported TARGETARCH: ${TARGETARCH}"; exit 1 ;; \
    esac; \
    export CGO_ENABLED=1; \
    LDFLAGS="-s -w -X github.com/pocketbase/pocketbase.Version=${POCKETBASE_VERSION}"; \
    go build -trimpath -ldflags="${LDFLAGS}" -o /out/pocketbase-libsql .

FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates tzdata \
    && rm -rf /var/lib/apt/lists/* \
    && useradd --system --create-home --home-dir /pb pb

WORKDIR /pb

COPY --from=builder /out/pocketbase-libsql /usr/local/bin/pocketbase

RUN mkdir -p /pb/pb_data /pb/pb_hooks /pb/pb_migrations \
    && chown -R pb:pb /pb

VOLUME ["/pb/pb_data", "/pb/pb_hooks", "/pb/pb_migrations"]

EXPOSE 8090

USER pb

ENTRYPOINT ["/usr/local/bin/pocketbase"]
CMD ["serve", "--http=0.0.0.0:8090"]
