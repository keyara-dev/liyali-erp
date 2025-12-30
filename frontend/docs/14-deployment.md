# Build & Deployment

Comprehensive guide for building, deploying, and maintaining the Liyali Gateway Frontend in production environments.

## Build Process

### Production Build

The application uses Next.js's optimized build process for production deployment.

```bash
# Install dependencies
pnpm install

# Run type checking
pnpm type-check

# Run linting
pnpm lint

# Build for production
pnpm build

# Start production server
pnpm start
```

### Build Configuration

```javascript
// next.config.js
/** @type {import('next').NextConfig} */
const nextConfig = {
  // Enable React Compiler for better performance
  experimental: {
    reactCompiler: true,
  },

  // Image optimization
  images: {
    domains: ['localhost', 'api.liyali.com', 'cdn.liyali.com'],
    formats: ['image/webp', 'image/avif'],
    minimumCacheTTL: 60,
  },

  // Compression
  compress: true,

  // Security headers
  async headers() {
    return [
      {
        source: '/(.*)',
        headers: [
          {
            key: 'X-Frame-Options',
            value: 'DENY',
          },
          {
            key: 'X-Content-Type-Options',
            value: 'nosniff',
          },
          {
            key: 'Referrer-Policy',
            value: 'origin-when-cross-origin',
          },
          {
            key: 'Permissions-Policy',
            value: 'camera=(), microphone=(), geolocation=()',
          },
        ],
      },
    ];
  },

  // Redirects
  async redirects() {
    return [
      {
        source: '/',
        destination: '/dashboard',
        permanent: false,
      },
    ];
  },

  // Environment variables validation
  env: {
    NEXT_PUBLIC_APP_VERSION: process.env.npm_package_version,
  },

  // Output configuration for different deployment targets
  output: process.env.BUILD_STANDALONE === 'true' ? 'standalone' : undefined,

  // Bundle analyzer (only in development)
  ...(process.env.ANALYZE === 'true' && {
    webpack: (config, { isServer }) => {
      if (!isServer) {
        config.plugins.push(
          new (require('@next/bundle-analyzer'))({
            enabled: true,
          })
        );
      }
      return config;
    },
  }),
};

module.exports = nextConfig;
```

### Build Optimization

#### Bundle Analysis

```bash
# Analyze bundle size
ANALYZE=true pnpm build

# This will open a visual representation of your bundle
# showing which modules are taking up the most space
```

#### Performance Monitoring

```typescript
// src/lib/performance.ts
export function reportWebVitals(metric: any) {
  if (process.env.NODE_ENV === 'production') {
    // Send to analytics service
    if (typeof window !== 'undefined' && window.gtag) {
      window.gtag('event', metric.name, {
        event_category: 'Web Vitals',
        event_label: metric.id,
        value: Math.round(metric.name === 'CLS' ? metric.value * 1000 : metric.value),
        non_interaction: true,
      });
    }
  }
}
```

## Environment Configuration

### Environment Variables

#### Development (.env.local)

```bash
# Application
NODE_ENV=development
NEXT_PUBLIC_APP_URL=http://localhost:3000
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1

# Authentication
AUTH_SECRET=dev-secret-key-32-characters-long
NEXTAUTH_URL=http://localhost:3000

# Features
NEXT_PUBLIC_ENABLE_OFFLINE=true
NEXT_PUBLIC_ENABLE_PWA=true
NEXT_PUBLIC_ENABLE_ANALYTICS=false

# Development
NEXT_PUBLIC_DEBUG_MODE=true
```

#### Staging (.env.staging)

```bash
# Application
NODE_ENV=production
NEXT_PUBLIC_APP_URL=https://staging.liyali.com
NEXT_PUBLIC_API_URL=https://api-staging.liyali.com/api/v1

# Authentication
AUTH_SECRET=staging-secret-key-32-characters-long
NEXTAUTH_URL=https://staging.liyali.com

# Features
NEXT_PUBLIC_ENABLE_OFFLINE=true
NEXT_PUBLIC_ENABLE_PWA=true
NEXT_PUBLIC_ENABLE_ANALYTICS=true

# Analytics
NEXT_PUBLIC_GA_MEASUREMENT_ID=G-XXXXXXXXXX
NEXT_PUBLIC_SENTRY_DSN=https://your-sentry-dsn

# Development
NEXT_PUBLIC_DEBUG_MODE=false
```

#### Production (.env.production)

```bash
# Application
NODE_ENV=production
NEXT_PUBLIC_APP_URL=https://app.liyali.com
NEXT_PUBLIC_API_URL=https://api.liyali.com/api/v1

# Authentication
AUTH_SECRET=production-secret-key-32-characters-long
NEXTAUTH_URL=https://app.liyali.com

# Features
NEXT_PUBLIC_ENABLE_OFFLINE=true
NEXT_PUBLIC_ENABLE_PWA=true
NEXT_PUBLIC_ENABLE_ANALYTICS=true

# Analytics
NEXT_PUBLIC_GA_MEASUREMENT_ID=G-XXXXXXXXXX
NEXT_PUBLIC_SENTRY_DSN=https://your-sentry-dsn

# Development
NEXT_PUBLIC_DEBUG_MODE=false
```

### Environment Validation

```typescript
// src/lib/env.ts
import { z } from 'zod';

const envSchema = z.object({
  NODE_ENV: z.enum(['development', 'staging', 'production']),
  NEXT_PUBLIC_APP_URL: z.string().url(),
  NEXT_PUBLIC_API_URL: z.string().url(),
  AUTH_SECRET: z.string().min(32),
  NEXTAUTH_URL: z.string().url(),
  NEXT_PUBLIC_ENABLE_OFFLINE: z.string().transform(val => val === 'true'),
  NEXT_PUBLIC_ENABLE_PWA: z.string().transform(val => val === 'true'),
  NEXT_PUBLIC_ENABLE_ANALYTICS: z.string().transform(val => val === 'true'),
});

export const env = envSchema.parse({
  NODE_ENV: process.env.NODE_ENV,
  NEXT_PUBLIC_APP_URL: process.env.NEXT_PUBLIC_APP_URL,
  NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL,
  AUTH_SECRET: process.env.AUTH_SECRET,
  NEXTAUTH_URL: process.env.NEXTAUTH_URL,
  NEXT_PUBLIC_ENABLE_OFFLINE: process.env.NEXT_PUBLIC_ENABLE_OFFLINE,
  NEXT_PUBLIC_ENABLE_PWA: process.env.NEXT_PUBLIC_ENABLE_PWA,
  NEXT_PUBLIC_ENABLE_ANALYTICS: process.env.NEXT_PUBLIC_ENABLE_ANALYTICS,
});
```

## Docker Deployment

### Multi-stage Dockerfile

```dockerfile
# Dockerfile
FROM node:18-alpine AS base

# Install dependencies only when needed
FROM base AS deps
RUN apk add --no-cache libc6-compat
WORKDIR /app

# Install dependencies based on the preferred package manager
COPY package.json pnpm-lock.yaml* ./
RUN corepack enable pnpm && pnpm i --frozen-lockfile

# Rebuild the source code only when needed
FROM base AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .

# Environment variables for build
ARG NEXT_PUBLIC_APP_URL
ARG NEXT_PUBLIC_API_URL
ARG AUTH_SECRET
ARG NEXTAUTH_URL

ENV NEXT_PUBLIC_APP_URL=$NEXT_PUBLIC_APP_URL
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL
ENV AUTH_SECRET=$AUTH_SECRET
ENV NEXTAUTH_URL=$NEXTAUTH_URL

# Disable telemetry during build
ENV NEXT_TELEMETRY_DISABLED 1

# Build the application
RUN corepack enable pnpm && pnpm build

# Production image, copy all the files and run next
FROM base AS runner
WORKDIR /app

ENV NODE_ENV production
ENV NEXT_TELEMETRY_DISABLED 1

RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

COPY --from=builder /app/public ./public

# Set the correct permission for prerender cache
RUN mkdir .next
RUN chown nextjs:nodejs .next

# Automatically leverage output traces to reduce image size
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

USER nextjs

EXPOSE 3000

ENV PORT 3000
ENV HOSTNAME "0.0.0.0"

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:3000/api/health || exit 1

CMD ["node", "server.js"]
```

### Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  frontend:
    build:
      context: .
      dockerfile: Dockerfile
      args:
        NEXT_PUBLIC_APP_URL: ${NEXT_PUBLIC_APP_URL}
        NEXT_PUBLIC_API_URL: ${NEXT_PUBLIC_API_URL}
        AUTH_SECRET: ${AUTH_SECRET}
        NEXTAUTH_URL: ${NEXTAUTH_URL}
    ports:
      - "3000:3000"
    environment:
      - NODE_ENV=production
    volumes:
      - ./logs:/app/logs
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:3000/api/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/nginx/ssl
    depends_on:
      - frontend
    restart: unless-stopped
```

### Build Scripts

```json
{
  "scripts": {
    "docker:build": "docker build -t liyali-frontend .",
    "docker:run": "docker run -p 3000:3000 liyali-frontend",
    "docker:compose:up": "docker-compose up -d",
    "docker:compose:down": "docker-compose down",
    "docker:compose:logs": "docker-compose logs -f frontend"
  }
}
```

## Vercel Deployment

### Vercel Configuration

```json
{
  "version": 2,
  "builds": [
    {
      "src": "package.json",
      "use": "@vercel/next"
    }
  ],
  "env": {
    "AUTH_SECRET": "@auth-secret",
    "NEXTAUTH_URL": "@nextauth-url"
  },
  "build": {
    "env": {
      "NEXT_PUBLIC_APP_URL": "https://app.liyali.com",
      "NEXT_PUBLIC_API_URL": "https://api.liyali.com/api/v1"
    }
  },
  "functions": {
    "src/app/api/**/*.ts": {
      "maxDuration": 30
    }
  },
  "headers": [
    {
      "source": "/(.*)",
      "headers": [
        {
          "key": "X-Frame-Options",
          "value": "DENY"
        },
        {
          "key": "X-Content-Type-Options",
          "value": "nosniff"
        }
      ]
    }
  ],
  "redirects": [
    {
      "source": "/",
      "destination": "/dashboard",
      "permanent": false
    }
  ]
}
```

### Deployment Commands

```bash
# Install Vercel CLI
npm i -g vercel

# Login to Vercel
vercel login

# Deploy to preview
vercel

# Deploy to production
vercel --prod

# Set environment variables
vercel env add AUTH_SECRET production
vercel env add NEXTAUTH_URL production
```

## AWS Deployment

### S3 + CloudFront (Static Export)

```javascript
// next.config.js for static export
/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'export',
  trailingSlash: true,
  images: {
    unoptimized: true,
  },
  // Disable features that require server
  experimental: {
    appDir: true,
  },
};

module.exports = nextConfig;
```

#### Build and Deploy Script

```bash
#!/bin/bash
# deploy-s3.sh

# Build the application
pnpm build

# Sync to S3
aws s3 sync out/ s3://your-bucket-name --delete

# Invalidate CloudFront cache
aws cloudfront create-invalidation --distribution-id YOUR_DISTRIBUTION_ID --paths "/*"

echo "Deployment complete!"
```

### EC2 Deployment

```bash
#!/bin/bash
# deploy-ec2.sh

# Update system
sudo yum update -y

# Install Node.js
curl -fsSL https://rpm.nodesource.com/setup_18.x | sudo bash -
sudo yum install -y nodejs

# Install PM2
sudo npm install -g pm2

# Clone repository
git clone https://github.com/your-org/liyali-gateway.git
cd liyali-gateway/frontend

# Install dependencies
npm install

# Build application
npm run build

# Start with PM2
pm2 start ecosystem.config.js --env production

# Save PM2 configuration
pm2 save
pm2 startup
```

#### PM2 Configuration

```javascript
// ecosystem.config.js
module.exports = {
  apps: [
    {
      name: 'liyali-frontend',
      script: 'node_modules/next/dist/bin/next',
      args: 'start',
      cwd: './',
      instances: 'max',
      exec_mode: 'cluster',
      env: {
        NODE_ENV: 'development',
        PORT: 3000,
      },
      env_production: {
        NODE_ENV: 'production',
        PORT: 3000,
        NEXT_PUBLIC_APP_URL: 'https://app.liyali.com',
        NEXT_PUBLIC_API_URL: 'https://api.liyali.com/api/v1',
      },
      error_file: './logs/err.log',
      out_file: './logs/out.log',
      log_file: './logs/combined.log',
      time: true,
    },
  ],
};
```

## Nginx Configuration

### Reverse Proxy Setup

```nginx
# nginx.conf
upstream frontend {
    server localhost:3000;
}

server {
    listen 80;
    server_name app.liyali.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name app.liyali.com;

    # SSL Configuration
    ssl_certificate /etc/nginx/ssl/cert.pem;
    ssl_certificate_key /etc/nginx/ssl/key.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;

    # Security Headers
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;

    # Gzip Compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_proxied expired no-cache no-store private must-revalidate auth;
    gzip_types
        text/plain
        text/css
        text/xml
        text/javascript
        application/javascript
        application/xml+rss
        application/json;

    # Static Assets
    location /_next/static/ {
        alias /app/.next/static/;
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    location /static/ {
        alias /app/public/;
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    # Main Application
    location / {
        proxy_pass http://frontend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
        proxy_read_timeout 86400;
    }

    # Health Check
    location /health {
        access_log off;
        return 200 "healthy\n";
        add_header Content-Type text/plain;
    }
}
```

## CI/CD Pipeline

### GitHub Actions

```yaml
# .github/workflows/deploy.yml
name: Deploy to Production

on:
  push:
    branches: [main]
  workflow_dispatch:

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '18'
          cache: 'pnpm'
      
      - name: Install dependencies
        run: pnpm install
      
      - name: Run tests
        run: pnpm test
      
      - name: Run type check
        run: pnpm type-check
      
      - name: Run linting
        run: pnpm lint

  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '18'
          cache: 'pnpm'
      
      - name: Install dependencies
        run: pnpm install
      
      - name: Build application
        run: pnpm build
        env:
          NEXT_PUBLIC_APP_URL: ${{ secrets.NEXT_PUBLIC_APP_URL }}
          NEXT_PUBLIC_API_URL: ${{ secrets.NEXT_PUBLIC_API_URL }}
          AUTH_SECRET: ${{ secrets.AUTH_SECRET }}
          NEXTAUTH_URL: ${{ secrets.NEXTAUTH_URL }}
      
      - name: Upload build artifacts
        uses: actions/upload-artifact@v3
        with:
          name: build-files
          path: .next/

  deploy:
    needs: build
    runs-on: ubuntu-latest
    environment: production
    steps:
      - uses: actions/checkout@v4
      
      - name: Download build artifacts
        uses: actions/download-artifact@v3
        with:
          name: build-files
          path: .next/
      
      - name: Deploy to Vercel
        uses: amondnet/vercel-action@v25
        with:
          vercel-token: ${{ secrets.VERCEL_TOKEN }}
          vercel-org-id: ${{ secrets.VERCEL_ORG_ID }}
          vercel-project-id: ${{ secrets.VERCEL_PROJECT_ID }}
          vercel-args: '--prod'
```

### GitLab CI/CD

```yaml
# .gitlab-ci.yml
stages:
  - test
  - build
  - deploy

variables:
  NODE_VERSION: "18"

cache:
  paths:
    - node_modules/
    - .next/cache/

test:
  stage: test
  image: node:${NODE_VERSION}-alpine
  script:
    - corepack enable pnpm
    - pnpm install
    - pnpm test
    - pnpm type-check
    - pnpm lint
  artifacts:
    reports:
      coverage: coverage/
    expire_in: 1 hour

build:
  stage: build
  image: node:${NODE_VERSION}-alpine
  script:
    - corepack enable pnpm
    - pnpm install
    - pnpm build
  artifacts:
    paths:
      - .next/
      - public/
    expire_in: 1 hour
  only:
    - main

deploy:
  stage: deploy
  image: alpine:latest
  before_script:
    - apk add --no-cache curl
  script:
    - curl -X POST $DEPLOY_WEBHOOK_URL
  only:
    - main
  when: manual
```

## Monitoring and Logging

### Application Monitoring

```typescript
// src/lib/monitoring.ts
import * as Sentry from '@sentry/nextjs';

// Initialize Sentry
if (process.env.NEXT_PUBLIC_SENTRY_DSN) {
  Sentry.init({
    dsn: process.env.NEXT_PUBLIC_SENTRY_DSN,
    environment: process.env.NODE_ENV,
    tracesSampleRate: process.env.NODE_ENV === 'production' ? 0.1 : 1.0,
    beforeSend(event) {
      // Filter out non-critical errors in production
      if (process.env.NODE_ENV === 'production') {
        if (event.exception) {
          const error = event.exception.values?.[0];
          if (error?.type === 'ChunkLoadError') {
            return null; // Don't send chunk load errors
          }
        }
      }
      return event;
    },
  });
}

// Custom error tracking
export function trackError(error: Error, context?: Record<string, any>) {
  console.error('Application Error:', error);
  
  if (process.env.NODE_ENV === 'production') {
    Sentry.captureException(error, {
      contexts: { custom: context },
    });
  }
}

// Performance tracking
export function trackPerformance(name: string, duration: number) {
  if (process.env.NODE_ENV === 'production') {
    Sentry.addBreadcrumb({
      category: 'performance',
      message: `${name} took ${duration}ms`,
      level: 'info',
    });
  }
}
```

### Health Check Endpoint

```typescript
// src/app/api/health/route.ts
import { NextResponse } from 'next/server';

export async function GET() {
  try {
    // Check database connection
    // Check external API availability
    // Check other critical services
    
    return NextResponse.json({
      status: 'healthy',
      timestamp: new Date().toISOString(),
      version: process.env.NEXT_PUBLIC_APP_VERSION,
      environment: process.env.NODE_ENV,
    });
  } catch (error) {
    return NextResponse.json(
      {
        status: 'unhealthy',
        error: error instanceof Error ? error.message : 'Unknown error',
        timestamp: new Date().toISOString(),
      },
      { status: 503 }
    );
  }
}
```

## Security Considerations

### Content Security Policy

```typescript
// src/middleware.ts
import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

export function middleware(request: NextRequest) {
  const response = NextResponse.next();

  // Content Security Policy
  const csp = [
    "default-src 'self'",
    "script-src 'self' 'unsafe-eval' 'unsafe-inline' https://www.googletagmanager.com",
    "style-src 'self' 'unsafe-inline'",
    "img-src 'self' data: https:",
    "font-src 'self' data:",
    "connect-src 'self' https://api.liyali.com",
    "frame-ancestors 'none'",
  ].join('; ');

  response.headers.set('Content-Security-Policy', csp);
  response.headers.set('X-Frame-Options', 'DENY');
  response.headers.set('X-Content-Type-Options', 'nosniff');
  response.headers.set('Referrer-Policy', 'origin-when-cross-origin');

  return response;
}

export const config = {
  matcher: [
    '/((?!api|_next/static|_next/image|favicon.ico).*)',
  ],
};
```

### Environment Security

```bash
# Use secrets management for sensitive data
# Never commit .env files to version control
# Rotate secrets regularly
# Use different secrets for each environment
# Implement proper access controls
```

## Performance Optimization

### Caching Strategy

```typescript
// src/lib/cache.ts
export const CACHE_STRATEGIES = {
  // Static assets - cache for 1 year
  STATIC: 'public, max-age=31536000, immutable',
  
  // API responses - cache for 5 minutes
  API: 'public, max-age=300, stale-while-revalidate=60',
  
  // User-specific data - no cache
  PRIVATE: 'private, no-cache, no-store, must-revalidate',
  
  // HTML pages - cache for 1 hour
  HTML: 'public, max-age=3600, stale-while-revalidate=300',
};
```

### Bundle Optimization

```javascript
// next.config.js
const nextConfig = {
  // Enable SWC minification
  swcMinify: true,
  
  // Optimize images
  images: {
    formats: ['image/webp', 'image/avif'],
  },
  
  // Webpack optimizations
  webpack: (config, { dev, isServer }) => {
    if (!dev && !isServer) {
      // Optimize bundle splitting
      config.optimization.splitChunks = {
        chunks: 'all',
        cacheGroups: {
          vendor: {
            test: /[\\/]node_modules[\\/]/,
            name: 'vendors',
            chunks: 'all',
          },
        },
      };
    }
    return config;
  },
};
```

This deployment guide provides comprehensive instructions for deploying the Liyali Gateway Frontend to various platforms while maintaining security, performance, and reliability standards.