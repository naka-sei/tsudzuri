# syntax=docker/dockerfile:1

# Allow build to pass GO_VERSION as an argument
ARG GO_VERSION=latest

# Builder stage
FROM golang:${GO_VERSION}-alpine AS builder
WORKDIR /app
RUN apk add --no-cache make gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make build

# Development stage
FROM golang:${GO_VERSION}-alpine AS dev
WORKDIR /app
RUN apk add --no-cache make gcc musl-dev

# Install air for hot reload
RUN go install github.com/air-verse/air@v1.62.0

# Runtime stage
FROM alpine:latest AS runtime
WORKDIR /app
RUN apk add --no-cache ca-certificates

COPY --from=builder /app/.bin/api /app/api
COPY --from=builder /app/migrations /app/migrations

EXPOSE 8080
CMD ["/app/api"]