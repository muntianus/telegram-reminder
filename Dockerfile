# syntax=docker/dockerfile:1

FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o bot

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/bot ./bot
CMD ["./bot"]

