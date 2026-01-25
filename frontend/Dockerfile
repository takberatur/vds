# Dockerfile (Bun + SvelteKit)
FROM oven/bun:1-alpine AS builder

WORKDIR /app

# Install dependencies
COPY package.json bun.lock ./
RUN bun install --frozen-lockfile --production=false

# Copy source code
COPY . .

# Build SvelteKit app
ENV NODE_ENV=production
RUN bun run build

# Production stage
FROM oven/bun:1-alpine

WORKDIR /app

ENV NODE_ENV=production
ENV PORT=3002
ENV HOST=0.0.0.0

# Copy built app dan dependencies
COPY --from=builder /app/build ./build
COPY --from=builder /app/package.json ./package.json

COPY --from=builder /app/static ./static
COPY --from=builder /app/build/client ./build/client
COPY --from=builder /app/build/server ./build/server

# Install ONLY production dependencies
RUN bun install --frozen-lockfile --production

# Copy .env if exists
COPY --from=builder /app/.env ./.env

EXPOSE 3002

CMD ["bun", "build/entry.js"]
