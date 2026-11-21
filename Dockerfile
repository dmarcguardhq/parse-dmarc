# syntax=docker/dockerfile:1
FROM oven/bun:1 AS frontend-builder

WORKDIR /build/frontend

RUN --mount=type=bind,source=package.json,target=package.json \
    --mount=type=bind,source=bun.lock,target=bun.lock \
    --mount=type=cache,target=/root/.bun/install/cache \
    bun install --frozen-lockfile

COPY . .
RUN bun run build

FROM golang:1.27 AS mod

WORKDIR /app

RUN --mount=type=bind,source=go.mod,target=go.mod \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=cache,target=/go/pkg/mod \
    go mod download

FROM golang:1.27 AS backend-builder

ARG VERSION=dev
ARG COMMIT=none
ARG DATE=unknown
ARG BUILT_BY=docker

WORKDIR /app

ENV CGO_ENABLED=0

COPY . .
COPY --from=frontend-builder /build/frontend/dist ./internal/api/dist

RUN --mount=type=cache,target=/go/pkg/mod \
    go build -ldflags="-s -w -extldflags '-static' -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE} -X main.builtBy=${BUILT_BY}" -trimpath -o parse-dmarc . && \
    mkdir -p /data /tmp

FROM scratch AS final

ENV PARSE_DMARC_CONFIG=/app/config.json \
    DATABASE_PATH=/data/parse-dmarc.db \
    SERVER_HOST=0.0.0.0

COPY --from=backend-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=backend-builder /app/parse-dmarc /usr/local/bin/parse-dmarc
# The process runs as an unprivileged user and writes only to /data: the
# database and its write-ahead log. /data is owned by that user so a fresh
# named volume starts out writable; the root filesystem may be read-only.
COPY --from=backend-builder --chown=65532:65532 --chmod=0750 /data /data
COPY --from=backend-builder --chmod=1777 /tmp /tmp
VOLUME /data

USER 65532:65532
EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/parse-dmarc"]
CMD ["--config=/app/config.json"]
