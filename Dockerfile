# syntax=docker/dockerfile:1
FROM oven/bun:1@sha256:3aa040e5445efa09d75bb428a64ba50db13e2b3ac9b6b0c7f145615f62b2ee18 AS frontend-builder

WORKDIR /build/frontend

RUN --mount=type=bind,source=package.json,target=package.json \
    --mount=type=bind,source=bun.lock,target=bun.lock \
    --mount=type=cache,target=/root/.bun/install/cache \
    bun install --frozen-lockfile

COPY . .
RUN bun run build

FROM golang:1.27@sha256:205c7636d74e0ffbe3d007162df7d74f859f5a8c8322dfb9f907977686c126cf AS mod

WORKDIR /app

RUN --mount=type=bind,source=go.mod,target=go.mod \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=cache,target=/go/pkg/mod \
    go mod download

FROM golang:1.27@sha256:205c7636d74e0ffbe3d007162df7d74f859f5a8c8322dfb9f907977686c126cf AS backend-builder

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

VOLUME /data
ENV PARSE_DMARC_CONFIG=/app/config.json \
    DATABASE_PATH=/data/parse-dmarc.db \
    SERVER_HOST=0.0.0.0

COPY --from=backend-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=backend-builder /app/parse-dmarc /usr/local/bin/parse-dmarc
COPY --from=backend-builder --chmod=0777 /data /data
COPY --from=backend-builder /tmp /tmp

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/parse-dmarc"]
CMD ["--config=/app/config.json"]
