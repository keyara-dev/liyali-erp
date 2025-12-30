# Configuration Guide

This guide covers all configuration options for the Liyali Gateway frontend application.

## Environment Variables

### Required Variables

```bash
# .env.local
NEXT_PUBLIC_APP_URL=http://localhost:3000
NEXT_PUBLIC_API_URL=http://localhost:3001/api
AUTH_SECRET=your-32-character-secret-key-here
NODE_ENV=development
```

### Optional Variables

```bash
# Database (if using direct DB connection)
DATABASE_URL=postgresql://user:password@localhost:5432/liyali_gateway

# External Services
NEXT_PUBLIC_SENTRY_DSN=https://your-sentry-dsn
NEXT_PUBLIC_ANALYTICS_ID=your-analytics-id

# Feature Flags
NEXT_PUBLIC_ENABLE_OFFLINE_MODE=true
NEXT_PUBLIC_ENABLE_REAL_TIME=true
NEXT_PUBLIC_ENABLE_PDF_GENERATION=true

# Development
NEXT_PUBLIC_DEBUG_MODE=true
NEXT_PUBLIC_MOCK_API=true
```

### Environment-Specific Configurations

#### Development (.env.local)

```bash
NODE_ENV=development
NEXT_PUBLIC_APP_URL=http://localhost:3000
NEXT_PUBLIC_API_URL=http://localhost:3001/api
NEXT_PUBLIC_DEBUG_MODE=true
NEXT_PUBLIC_MOCK_API=true
AUTH_SECRET=dev-secret-key-32-characters-long
```

#### Staging (.env.staging)

```bash
NODE_ENV=production
NEXT_PUBLIC_APP_URL=https://staging.liyali.com
NEXT_PUBLIC_API_URL=https://api-staging.liyali.com/api
NEXT_PUBLIC_DEBUG_MODE=false
NEXT_PUBLIC_MOCK_API=false
AUTH_SECRET=staging-secret-key-32-characters-long
```

#### Production (.env.production)

```bash
NODE_ENV=production
NEXT_PUBLIC_APP_URL=https://app.liyali.com
NEXT_PUBLIC_API_URL=https://api.liyali.com/api
NEXT_PUBLIC_DEBUG_MODE=false
NEXT_PUBLIC_MOCK_API=false
AUTH_SECRET=production-secret-key-32-characters-long
```

## Next.js Configuration

### next.config.js

```javascript
/** @type {import('next').NextConfig} */
const nextConfig = {
  // Enable experimental features
  experimental: {
    // Enable React 19 features
    reactCompiler: true,
    // Enable Turbopack for faster builds
    turbo: {
      rules: {
        '*.svg': {
          loaders: ['@svgr/webpack'],
          as: '*.js',
        },
      },
    },
  },

  // Image optimization
  images: {
    domains: ['localhost', 'api.liyali.com'],
    formats: ['image/webp', 'image/avif'],
  },

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
    CUSTOM_KEY: process.env.CUSTOM_KEY,
  },

  // Bundle analyzer (development only)
  ...(process.env.ANALYZE === 'true' && {
    webpack: (config) => {
      config.plugins.push(
        new (require('@next/bundle-analyzer'))({
          enabled: true,
        })
      );
      return config;
    },
  }),
};

module.exports = nextConfig;
```

## TypeScript Configuration

### tsconfig.json

```json
{
  "compilerOptions": {
    "lib": ["dom", "dom.iterable", "es6"],
    "allowJs": true,
    "skipLibCheck": true,
    "strict": true,
    "noEmit": true,
    "esModuleInterop": true,
    "module": "esnext",
    "moduleResolution": "bundler",
    "resolveJsonModule": true,
    "isolatedModules": true,
    "jsx": "preserve",
    "incremental": true,
    "plugins": [
      {
        "name": "next"
      }
    ],
    "baseUrl": ".",
    "paths": {
      "@/*": ["./src/*"],
      "@/components/*": ["./src/components/*"],
      "@/lib/*": ["./src/lib/*"],
      "@/hooks/*": ["./src/hooks/*"],
      "@/types/*": ["./src/types/*"],
      "@/app/*": ["./src/app/*"]
    }
  },
  "include": [
    "next-env.d.ts",
    "**/*.ts",
    "**/*.tsx",
    ".next/types/**/*.ts"
  ],
  "exclude": ["node_modules"]
}
```

## Tailwind CSS Configuration

### tailwind.config.js

```javascript
import type { Config } from 'tailwindcss';

const config: Config = {
  darkMode: ['class'],
  content: [
    './src/pages/**/*.{js,ts,jsx,tsx,mdx}',
    './src/components/**/*.{js,ts,jsx,tsx,mdx}',
    './src/app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {
      colors: {
        border: 'hsl(var(--border))',
        input: 'hsl(var(--input))',
        ring: 'hsl(var(--ring))',
        background: 'hsl(var(--background))',
        foreground: 'hsl(var(--foreground))',
        primary: {
          DEFAULT: 'hsl(var(--primary))',
          foreground: 'hsl(var(--primary-foreground))',
        },
        secondary: {
          DEFAULT: 'hsl(var(--secondary))',
          foreground: 'hsl(var(--secondary-foreground))',
        },
        destructive: {
          DEFAULT: 'hsl(var(--destructive))',
          foreground: 'hsl(var(--destructive-foreground))',
        },
        muted: {
          DEFAULT: 'hsl(var(--muted))',
          foreground: 'hsl(var(--muted-foreground))',
        },
        accent: {
          DEFAULT: 'hsl(var(--accent))',
          foreground: 'hsl(var(--accent-foreground))',
        },
        popover: {
          DEFAULT: 'hsl(var(--popover))',
          foreground: 'hsl(var(--popover-foreground))',
        },
        card: {
          DEFAULT: 'hsl(var(--card))',
          foreground: 'hsl(var(--card-foreground))',
        },
      },
      borderRadius: {
        lg: 'var(--radius)',
        md: 'calc(var(--radius) - 2px)',
        sm: 'calc(var(--radius) - 4px)',
      },
      fontFamily: {
        sans: ['var(--font-sans)', 'system-ui', 'sans-serif'],
        mono: ['var(--font-mono)', 'monospace'],
      },
      animation: {
        'accordion-down': 'accordion-down 0.2s ease-out',
        'accordion-up': 'accordion-up 0.2s ease-out',
      },
    },
  },
  plugins: [
    require('tailwindcss-animate'),
    require('@tailwindcss/typography'),
    require('@tailwindcss/container-queries'),
  ],
};

export default config;
```

## ESLint Configuration

### .eslintrc.json

```json
{
  "extends": [
    "next/core-web-vitals",
    "@typescript-eslint/recommended",
    "prettier"
  ],
  "parser": "@typescript-eslint/parser",
  "plugins": ["@typescript-eslint"],
  "rules": {
    "@typescript-eslint/no-unused-vars": "error",
    "@typescript-eslint/no-explicit-any": "warn",
    "@typescript-eslint/prefer-const": "error",
    "react-hooks/exhaustive-deps": "warn",
    "react/no-unescaped-entities": "off",
    "@next/next/no-page-custom-font": "off"
  },
  "ignorePatterns": [
    "node_modules/",
    ".next/",
    "out/",
    "public/"
  ]
}
```

## Prettier Configuration

### .prettierrc

```json
{
  "semi": true,
  "trailingComma": "es5",
  "singleQuote": true,
  "printWidth": 80,
  "tabWidth": 2,
  "useTabs": false,
  "bracketSpacing": true,
  "arrowParens": "avoid",
  "endOfLine": "lf",
  "plugins": ["prettier-plugin-tailwindcss"]
}
```

## Application Configuration

### Session Configuration

```typescript
// src/lib/session-config.ts
export const SESSION_CONFIG = {
  SESSION_TTL: 30 * 60 * 1000,        // 30 minutes
  SCREEN_LOCK_COUNTDOWN: 90 * 1000,   // 90 seconds
  IDLE_TIMEOUT: 25 * 60 * 1000,       // 25 minutes
  REFRESH_THRESHOLD: 5 * 60 * 1000,   // 5 minutes before expiry
} as const;
```

### API Configuration

```typescript
// src/lib/api/config.ts
export const API_CONFIG = {
  BASE_URL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:3001/api',
  TIMEOUT: 30000, // 30 seconds
  RETRY_ATTEMPTS: 3,
  RETRY_DELAY: 1000, // 1 second
} as const;
```

### Query Configuration

```typescript
// src/lib/query-config.ts
export const QUERY_CONFIG = {
  STALE_TIME: 5 * 60 * 1000,      // 5 minutes
  GC_TIME: 10 * 60 * 1000,        // 10 minutes
  RETRY_ATTEMPTS: 3,
  RETRY_DELAY: (attemptIndex: number) => 
    Math.min(1000 * 2 ** attemptIndex, 30000),
} as const;
```

## Feature Flags

### Feature Toggle System

```typescript
// src/lib/features.ts
export const FEATURES = {
  OFFLINE_MODE: process.env.NEXT_PUBLIC_ENABLE_OFFLINE_MODE === 'true',
  REAL_TIME: process.env.NEXT_PUBLIC_ENABLE_REAL_TIME === 'true',
  PDF_GENERATION: process.env.NEXT_PUBLIC_ENABLE_PDF_GENERATION === 'true',
  DEBUG_MODE: process.env.NEXT_PUBLIC_DEBUG_MODE === 'true',
  MOCK_API: process.env.NEXT_PUBLIC_MOCK_API === 'true',
} as const;

export function isFeatureEnabled(feature: keyof typeof FEATURES): boolean {
  return FEATURES[feature];
}
```

### Usage in Components

```typescript
import { isFeatureEnabled } from '@/lib/features';

export function MyComponent() {
  if (!isFeatureEnabled('OFFLINE_MODE')) {
    return null;
  }

  return <div>Offline features enabled</div>;
}
```

## Theme Configuration

### CSS Variables

```css
/* src/app/globals.css */
@layer base {
  :root {
    --background: 0 0% 100%;
    --foreground: 222.2 84% 4.9%;
    --card: 0 0% 100%;
    --card-foreground: 222.2 84% 4.9%;
    --popover: 0 0% 100%;
    --popover-foreground: 222.2 84% 4.9%;
    --primary: 221.2 83.2% 53.3%;
    --primary-foreground: 210 40% 98%;
    --secondary: 210 40% 96%;
    --secondary-foreground: 222.2 84% 4.9%;
    --muted: 210 40% 96%;
    --muted-foreground: 215.4 16.3% 46.9%;
    --accent: 210 40% 96%;
    --accent-foreground: 222.2 84% 4.9%;
    --destructive: 0 84.2% 60.2%;
    --destructive-foreground: 210 40% 98%;
    --border: 214.3 31.8% 91.4%;
    --input: 214.3 31.8% 91.4%;
    --ring: 221.2 83.2% 53.3%;
    --radius: 0.5rem;
  }

  .dark {
    --background: 222.2 84% 4.9%;
    --foreground: 210 40% 98%;
    --card: 222.2 84% 4.9%;
    --card-foreground: 210 40% 98%;
    --popover: 222.2 84% 4.9%;
    --popover-foreground: 210 40% 98%;
    --primary: 217.2 91.2% 59.8%;
    --primary-foreground: 222.2 84% 4.9%;
    --secondary: 217.2 32.6% 17.5%;
    --secondary-foreground: 210 40% 98%;
    --muted: 217.2 32.6% 17.5%;
    --muted-foreground: 215 20.2% 65.1%;
    --accent: 217.2 32.6% 17.5%;
    --accent-foreground: 210 40% 98%;
    --destructive: 0 62.8% 30.6%;
    --destructive-foreground: 210 40% 98%;
    --border: 217.2 32.6% 17.5%;
    --input: 217.2 32.6% 17.5%;
    --ring: 224.3 76.3% 94.1%;
  }
}
```

## Build Configuration

### Production Build

```bash
# Build optimization
npm run build

# Analyze bundle size
ANALYZE=true npm run build

# Static export
npm run export
```

### Docker Configuration

```dockerfile
# Dockerfile
FROM node:18-alpine AS base

# Install dependencies only when needed
FROM base AS deps
RUN apk add --no-cache libc6-compat
WORKDIR /app

COPY package.json package-lock.json* ./
RUN npm ci

# Rebuild the source code only when needed
FROM base AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .

ENV NEXT_TELEMETRY_DISABLED 1

RUN npm run build

# Production image, copy all the files and run next
FROM base AS runner
WORKDIR /app

ENV NODE_ENV production
ENV NEXT_TELEMETRY_DISABLED 1

RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

COPY --from=builder /app/public ./public

COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

USER nextjs

EXPOSE 3000

ENV PORT 3000
ENV HOSTNAME "0.0.0.0"

CMD ["node", "server.js"]
```

## Deployment Configuration

### Vercel

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
    "NEXT_PUBLIC_API_URL": "@api-url"
  }
}
```

### Environment-Specific Builds

```bash
# Development
npm run dev

# Staging
NODE_ENV=staging npm run build
npm run start

# Production
NODE_ENV=production npm run build
npm run start
```

This configuration provides a solid foundation for development, staging, and production environments while maintaining flexibility for customization.