# Frontend Documentation

Essential documentation for the Liyali Gateway frontend application.

## Architecture

Built with Next.js 14, TypeScript, TailwindCSS, and TanStack Query.

## Key Systems

- **[Authentication](./authentication.md)** - JWT auth, session management, organization context
- **[Notifications](./notifications.md)** - Real-time notifications with React Query
- **[Workflows](./workflows.md)** - Document approval workflows and task management

## Patterns

- **Server Actions** - All API calls use server actions with `authenticatedApiClient`
- **React Query** - State management and caching for server state
- **Zustand** - Client-side state management (organization store)
- **Component Architecture** - Reusable UI components with proper TypeScript

## Development

```bash
npm run dev          # Start development server
npm run build        # Build for production
npm run type-check   # TypeScript validation
```
