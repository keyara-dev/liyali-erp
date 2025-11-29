# Use the official Node.js image as the base image
FROM node:18-alpine AS base

# Install OpenSSL (required by Prisma)
RUN apk add --no-cache openssl

# Install pnpm
RUN npm install -g pnpm

# Set working directory
WORKDIR /app

# Copy workspace files
COPY package.json pnpm-workspace.yaml pnpm-lock.yaml ./
COPY .npmrc ./

# Install dependencies and build stage
FROM base AS build
COPY . .

# Accept build-time variable
ARG ADMIN_ENCRYPTION_KEY
ENV ADMIN_ENCRYPTION_KEY=$ADMIN_ENCRYPTION_KEY

# Install all dependencies with cache mount
RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install --no-frozen-lockfile

# Generate Prisma client using workspace filter
RUN pnpm --filter @bgs-tickety/database run prisma:generate

# Set environment variable to skip API calls during build
ENV NEXT_PUBLIC_SKIP_BUILD_STATIC_CHECK=true

# Build the admin application
RUN pnpm --filter admin build

# Debug: Check what's in the build output
RUN ls -la /app/admin/.next/

# Production stage
FROM node:18-alpine AS runner
WORKDIR /app

# Install OpenSSL (required by Prisma)
RUN apk add --no-cache openssl

# Create a non-root user
RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

# Copy the Next.js build and application files
COPY --from=build /app/admin/.next ./.next
COPY --from=build /app/admin/public ./public
COPY --from=build /app/admin/package.json ./package.json
COPY --from=build /app/admin/server.js ./server.js
COPY --from=build /app/node_modules ./node_modules

# Debug: Check what's copied
RUN ls -la /app/
RUN ls -la /app/.next/ || echo "No .next directory"

# Set ownership of the application files to nextjs user
USER root
RUN chown -R nextjs:nodejs /app
USER nextjs

# Set environment variables for Cloud Run
ENV NODE_ENV=production
ENV PORT=8080
ENV HOSTNAME=0.0.0.0

# Expose port
EXPOSE 8080

# The Next.js standalone build puts server.js in the root of standalone output
# Use exec form to ensure proper signal handling
CMD ["node", "server.js"]