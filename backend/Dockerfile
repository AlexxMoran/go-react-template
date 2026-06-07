# syntax=docker/dockerfile:1

# ── build the API binary ──────────────────────────────────────────────────────
FROM golang:1.26-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/api ./cmd/api

# ── goose migration tool ──────────────────────────────────────────────────────
FROM golang:1.26-alpine AS goose
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go install github.com/pressly/goose/v3/cmd/goose@latest

# ── migration runner image (one-shot) ─────────────────────────────────────────
FROM alpine:3.20 AS migrate
COPY --from=goose /go/bin/goose /usr/local/bin/goose
COPY db/migrations /migrations
ENV GOOSE_MIGRATION_DIR=/migrations
ENTRYPOINT ["goose"]
CMD ["up"]

# ── final API image (distroless, non-root) ────────────────────────────────────
FROM gcr.io/distroless/static-debian12:nonroot AS api
COPY --from=build /out/api /api
EXPOSE 3000
USER nonroot:nonroot
ENTRYPOINT ["/api"]
