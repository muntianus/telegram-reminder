# syntax=docker/dockerfile:1.4

FROM --platform=${BUILDPLATFORM} golang:1.24-alpine AS builder
ARG TARGETOS
ARG TARGETARCH
WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -o /bot ./cmd/bot

FROM gcr.io/distroless/static
WORKDIR /app
COPY --from=builder /bot /app/bot
COPY --from=builder /app/tasks.yml /app/tasks.yml
USER nonroot:nonroot
ENTRYPOINT ["/app/bot"]

