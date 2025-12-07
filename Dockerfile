# Build stage
FROM golang:1.25.4-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# COPY source code
COPY . .

# BUILD with CGO enabled
RUN CGO_ENABLED=1 go build -o godo ./cmd/api

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS and create non-root user
RUN apk add --no-cache ca-certificates \
    && adduser -D -g '' appuser

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/godo .

# Copy migrations
COPY --from=builder /app/migrations ./migrations

# Create data directory for Sqlite
RUN mkdir -p /data && chown appuser:appuser /data

# Switch to non-root user
USER appuser

EXPOSE 8080

CMD [ "./godo" ]
