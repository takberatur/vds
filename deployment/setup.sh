#!/bin/bash

# Exit on error
set -e

echo "🚀 Starting Deployment Setup..."

# 1. Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker could not be found. Please install Docker first."
    exit 1
fi

# 2. Build and Start Containers
echo "📦 Building and starting containers..."
docker-compose up -d --build

# 3. Wait for Database to be ready
echo "⏳ Waiting for database to be ready..."
sleep 10

# 4. Run Migrations (Optional - if you have migrate tool in container or local)
# Ideally, the app container should run migrations on startup or have a separate migrate service.
# For now, we assume the app handles it or we run it manually.

echo "✅ Deployment successful!"
echo "   - Backend API: http://localhost:8080"
# echo "   - Frontend Web: http://localhost:3000"
