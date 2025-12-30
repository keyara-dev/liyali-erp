# Quick Start Guide

Get the Liyali Gateway Frontend up and running in 5 minutes.

## Prerequisites

- **Node.js 18+** - Latest LTS version recommended
- **pnpm** - Package manager (recommended over npm/yarn)
- **Git** - Version control

## 1. Clone and Setup

```bash
# Clone the repository
git clone <repository-url>
cd liyali-gateway/frontend

# Install dependencies
pnpm install
```

## 2. Environment Configuration

```bash
# Copy environment template
cp .env.example .env.local

# Edit configuration
nano .env.local
```

**Required Environment Variables:**
```env
# API Configuration
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
NEXT_PUBLIC_APP_URL=http://localhost:3000

# Authentication
AUTH_SECRET=your-super-secret-auth-key-change-in-production-min-32-chars
NEXTAUTH_URL=http://localhost:3000

# Features
NEXT_PUBLIC_ENABLE_OFFLINE=true
NEXT_PUBLIC_ENABLE_PWA=true
NEXT_PUBLIC_ENABLE_ANALYTICS=false

# Development
NODE_ENV=development
```

## 3. Run the Application

```bash
# Start development server
pnpm dev
```

You should see:
```
▲ Next.js 15.0.7
- Local:        http://localhost:3000
- Environments: .env.local

✓ Starting...
✓ Ready in 2.3s
```

## 4. Verify Installation

### Access the Application
Open [http://localhost:3000](http://localhost:3000) in your browser.

### Login with Demo Account
```
Email: admin@liyali.com
Password: admin123
```

### Test Key Features

1. **Dashboard** - View overview metrics
2. **Requisitions** - Create a new purchase request
3. **Offline Mode** - Disconnect internet and test offline functionality
4. **Dark Mode** - Toggle theme in the header
5. **Mobile View** - Test responsive design

## 5. Development Workflow

### File Structure
```
frontend/
├── src/
│   ├── app/                 # Next.js App Router pages
│   ├── components/          # Reusable UI components
│   ├── hooks/              # Custom React hooks
│   ├── lib/                # Utilities and configurations
│   └── types/              # TypeScript type definitions
├── public/                 # Static assets
└── docs/                   # Documentation
```

### Key Commands
```bash
# Development
pnpm dev              # Start dev server
pnpm build            # Build for production
pnpm start            # Start production server
pnpm lint             # Run ESLint
pnpm type-check       # Run TypeScript check

# Testing
pnpm test             # Run tests
pnpm test:watch       # Run tests in watch mode
pnpm test:coverage    # Run tests with coverage

# Analysis
pnpm analyze          # Analyze bundle size
pnpm lighthouse       # Run Lighthouse audit
```

## 6. First Steps

### Create Your First Component
```tsx
// src/components/my-component.tsx
import { Button } from '@/components/ui/button';

export function MyComponent() {
  return (
    <div className="p-4">
      <h2 className="text-2xl font-bold">Hello World</h2>
      <Button onClick={() => alert('Clicked!')}>
        Click me
      </Button>
    </div>
  );
}
```

### Add a New Page
```tsx
// src/app/my-page/page.tsx
import { MyComponent } from '@/components/my-component';

export default function MyPage() {
  return (
    <div className="container mx-auto py-8">
      <h1 className="text-3xl font-bold mb-8">My Page</h1>
      <MyComponent />
    </div>
  );
}
```

### Create a Custom Hook
```tsx
// src/hooks/use-my-data.ts
import { useQuery } from '@tanstack/react-query';

export function useMyData() {
  return useQuery({
    queryKey: ['my-data'],
    queryFn: async () => {
      // Fetch data logic
      return { message: 'Hello from hook!' };
    },
  });
}
```

## 7. Understanding the Architecture

### App Router Structure
```
src/app/
├── (auth)/              # Authentication pages
│   ├── login/
│   └── register/
├── (private)/           # Protected pages
│   ├── dashboard/
│   ├── requisitions/
│   └── settings/
├── layout.tsx           # Root layout
└── page.tsx            # Home page
```

### Component Organization
```
src/components/
├── ui/                  # Base UI components (shadcn/ui)
├── layout/             # Layout components
├── auth/               # Authentication components
├── workflows/          # Workflow-specific components
└── [feature]/          # Feature-specific components
```

### State Management
- **Server State**: TanStack Query for API data
- **Client State**: Zustand for global state
- **Form State**: React Hook Form for forms
- **URL State**: Next.js router for navigation state

## 8. Key Features Overview

### Offline-First Architecture
The app works offline by default:
- Data is cached in IndexedDB
- Actions are queued when offline
- Automatic sync when back online

### Real-time Updates
- Live notifications
- Real-time data synchronization
- Optimistic UI updates

### Responsive Design
- Mobile-first approach
- Adaptive layouts
- Touch-friendly interactions

### Accessibility
- WCAG 2.1 AA compliant
- Keyboard navigation
- Screen reader support
- High contrast mode

## 9. Next Steps

### Development
- **Read the [Development Guide](./11-development.md)** - Learn development patterns
- **Explore [Component System](./05-component-system.md)** - Understand UI components
- **Study [State Management](./06-state-management.md)** - Learn data flow patterns

### Customization
- **Modify Theme** - Update Tailwind configuration
- **Add Components** - Create new UI components
- **Extend Features** - Add new functionality

### Deployment
- **Build for Production** - `pnpm build`
- **Deploy to Vercel** - Connect GitHub repository
- **Configure Environment** - Set production environment variables

## Troubleshooting

### Common Issues

**Port Already in Use**
```bash
# Kill process using port 3000
lsof -ti:3000 | xargs kill -9

# Or use different port
pnpm dev -- --port 3001
```

**Module Not Found**
```bash
# Clear cache and reinstall
rm -rf node_modules .next
pnpm install
```

**TypeScript Errors**
```bash
# Run type check
pnpm type-check

# Clear TypeScript cache
rm -rf .next/types
```

**Build Errors**
```bash
# Check for ESLint errors
pnpm lint

# Fix auto-fixable issues
pnpm lint --fix
```

For more troubleshooting, see [Troubleshooting Guide](./16-troubleshooting.md).

## What's Next?

You now have a fully functional Liyali Gateway Frontend! The system includes:

- ✅ **Modern React Architecture** with Next.js 15 and React 19
- ✅ **Offline-First Functionality** with IndexedDB storage
- ✅ **Real-time Updates** with optimistic UI
- ✅ **Responsive Design** with Tailwind CSS
- ✅ **Type Safety** with TypeScript
- ✅ **Accessibility** with WCAG compliance

Explore the other documentation files to learn more about specific features and advanced configuration options.