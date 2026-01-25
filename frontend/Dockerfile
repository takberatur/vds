# Dockerfile (Bun + SvelteKit)
FROM oven/bun:1-alpine AS builder

WORKDIR /app

# Install dependencies
COPY package.json bun.lock ./
RUN bun install --frozen-lockfile

# Copy source code
COPY . .

# Build SvelteKit app
RUN bun run build

# Production stage
FROM oven/bun:1-alpine

WORKDIR /app

# Copy built app dan dependencies
# COPY --from=builder /app/node_modules node_modules
COPY --from=builder /app/build ./build
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/package.json ./package.json

COPY --from=builder /app/.env* ./

EXPOSE 3002

CMD ["bun", "./build/index.js"]
