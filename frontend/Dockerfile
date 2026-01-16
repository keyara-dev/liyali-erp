# Production Dockerfile for Frontend
# Base stage with common dependencies
FROM node:20-alpine AS base

# Install OpenSSL (required by Prisma) and other dependencies
RUN apk add --no-cache libc6-compat openssl

WORKDIR /app

# Copy package files
COPY package.json package-lock.json* ./

# Dependencies stage
FROM base AS deps

# Install dependencies with cache mount for faster builds
RUN --mount=type=cache,id=npm,target=/root/.npm \
    npm ci --only=production --ignore-scripts && \
    npm cache clean --force

# Builder stage
FROM base AS builder

WORKDIR /app

# Copy dependencies from deps stage
COPY --from=deps /app/node_modules ./node_modules

# Copy all source files
COPY . .

# Set build-time environment variables
ARG NEXT_PUBLIC_API_URL
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL

# Disable telemetry during build
ENV NEXT_TELEMETRY_DISABLED=1

# Skip API calls during build if needed
ENV NEXT_PUBLIC_SKIP_BUILD_STATIC_CHECK=true

# Build the application
RUN npm run build

# Debug: Verify build output
RUN echo "Build completed. Checking output..." && \
    ls -la .next/ && \
    ls -la .next/standalone/ || echo "No standalone directory found"

# Production runner stage
FROM node:20-alpine AS runner

WORKDIR /app

# Install OpenSSL (required by Prisma at runtime)
RUN apk add --no-cache openssl

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

# Create non-root user for security
RUN addgroup --system --gid 1001 nodejs && \
    adduser --system --uid 1001 nextjs

# Copy public assets
COPY --from=builder /app/public ./public

# Copy Next.js standalone build output
# The standalone output includes all necessary files
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

# Debug: Verify copied files
RUN echo "Files copied to runner stage:" && \
    ls -la /app/ && \
    ls -la /app/.next/ || echo "No .next directory"

# Set proper ownership
USER root
RUN chown -R nextjs:nodejs /app
USER nextjs

# Set environment variables for Cloud Run
ENV PORT=3000
ENV HOSTNAME="0.0.0.0"

# Expose port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD node -e "require('http').get('http://localhost:3000/api/health', (r) => {process.exit(r.statusCode === 200 ? 0 : 1)})"

# Use exec form to ensure proper signal handling
CMD ["node", "server.js"]
