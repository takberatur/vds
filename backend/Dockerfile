# Build Stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git make

# Install lux
RUN go install github.com/iawia002/lux@latest

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binaries (API, worker, and seed)
RUN go build -o bin/api ./cmd/api && \
    go build -o bin/worker ./cmd/worker && \
    go build -o bin/seed ./cmd/seed

# Final Stage
FROM alpine:3.19

WORKDIR /app

# Install runtime dependencies, yt-dlp, JS runtime, and migrate tool
RUN apk add --no-cache \
    ca-certificates \
    ffmpeg \
    python3 \
    py3-pip \
    nodejs \
    wget \
    curl \
    build-base \
    python3-dev \
    libffi-dev \
    openssl-dev \
    curl-dev \
    linux-headers \
    chromium && \
    pip3 install --force-reinstall --break-system-packages "yt-dlp[default]" curl-cffi && \
    python3 -c "import curl_cffi; print('curl_cffi installed successfully')" && \
    curl -L https://github.com/golang-migrate/migrate/releases/download/v4.19.1/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /usr/local/bin/migrate && \
    chmod +x /usr/local/bin/migrate

# Copy binaries dari builder
COPY --from=builder /app/bin/api ./api
COPY --from=builder /app/bin/worker ./worker
COPY --from=builder /app/bin/seed ./seed
COPY --from=builder /go/bin/lux /usr/local/bin/lux
COPY --from=builder /app/.env . 
# Note: In production, .env should be injected via docker-compose or k8s secrets

# Copy migrations if needed
COPY --from=builder /app/internal/repository/migrations ./internal/repository/migrations

# Expose port API
EXPOSE 8080

# Setup cron job to clean up temp files every 5 minutes
RUN echo "*/5 * * * * find /tmp -name '*.mp4' -mmin +5 -delete" >> /var/spool/cron/crontabs/root

# Create entrypoint script to run both cron and api
RUN echo '#!/bin/sh' > /entrypoint.sh && \
    echo 'crond' >> /entrypoint.sh && \
    echo 'exec "$@"' >> /entrypoint.sh && \
    chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]

# Default run API (overrideable in docker-compose)
CMD ["./api"]
