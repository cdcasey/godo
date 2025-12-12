# Build stage
# bookworm is required because Alpine uses Musl libc, which is incompatible
# with the pre-compiled glibc dependencies in the Turso (libsql) driver.
FROM golang:1.25.4-bookworm AS builder

# Install build dependencies
# RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Install templ
RUN go install github.com/a-h/templ/cmd/templ@latest

# COPY source code
COPY . .

# Generate templ files and build with CGO enabled
RUN templ generate
RUN CGO_ENABLED=1 go build -o godo ./cmd/api

# Runtime stage
# FROM alpine:latest
FROM debian:bookworm-slim

# Install ca-certificates for HTTPS and create non-root user
# RUN apk add --no-cache ca-certificates \
#     && adduser -D -g '' appuser
RUN apt-get update && apt-get install -y --no-install-recommends \
	ca-certificates \
	&& rm -rf /var/lib/apt/lists/* \
	&& useradd --create-home --shell /bin/false appuser

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
