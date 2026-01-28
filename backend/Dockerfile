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
FROM debian:bookworm-slim

WORKDIR /app

# Install runtime dependencies, yt-dlp, JS runtime, and migrate tool
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    wget \
    tar \
    gzip \
    ffmpeg \
    python3 \
    python3-venv \
    nodejs \
    npm \
    chromium \
    && rm -rf /var/lib/apt/lists/* && \
    python3 -m venv /opt/venv && \
    /opt/venv/bin/pip install --no-cache-dir -U pip setuptools wheel && \
    /opt/venv/bin/pip install --no-cache-dir -U "yt-dlp[default,curl-cffi]" && \
    /opt/venv/bin/python -c "import curl_cffi; print('curl_cffi installed successfully')" && \
    curl -L https://github.com/golang-migrate/migrate/releases/download/v4.19.1/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /usr/local/bin/migrate && \
    chmod +x /usr/local/bin/migrate

ENV PATH="/opt/venv/bin:${PATH}"

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

# Default run API (overrideable in docker-compose)
CMD ["./api"]
