# Liyali Gateway Frontend Documentation

A modern, responsive procurement management frontend built with Next.js 15, React 19, and TypeScript.

## 🏗️ Architecture Overview

The Liyali Gateway Frontend is built using **modern React patterns** with a focus on performance, accessibility, and developer experience.

### Core Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        App Router (Next.js 15)                 │
├─────────────────────────────────────────────────────────────────┤
│  Pages & Layouts - File-based routing with nested layouts      │
├─────────────────────────────────────────────────────────────────┤
│  Server Actions - Server-side data mutations                   │
├─────────────────────────────────────────────────────────────────┤
│  Components - Reusable UI components with shadcn/ui            │
├─────────────────────────────────────────────────────────────────┤
│  Hooks - Custom React hooks for data fetching & state          │
├─────────────────────────────────────────────────────────────────┤
│  State Management - Zustand + React Query + Context            │
├─────────────────────────────────────────────────────────────────┤
│  Storage Layer - IndexedDB with offline-first approach         │
└─────────────────────────────────────────────────────────────────┘
```

### Key Features

- **Next.js 15 App Router** - Modern file-based routing with server components
- **React 19** - Latest React features with concurrent rendering
- **TypeScript** - Full type safety across the application
- **Offline-First Architecture** - Works without internet connection
- **Real-time Updates** - Live data synchronization
- **Responsive Design** - Mobile-first approach with Tailwind CSS
- **Accessibility** - WCAG 2.1 AA compliant components
- **Performance Optimized** - Code splitting, lazy loading, and caching

## 📚 Documentation Structure

### Getting Started
- [Quick Start Guide](./01-quick-start.md) - Get up and running in 5 minutes
- [Installation](./02-installation.md) - Detailed setup instructions
- [Configuration](./03-configuration.md) - Environment and build configuration

### Architecture & Design
- [System Architecture](./04-architecture.md) - Detailed architecture overview
- [Component Design System](./05-component-system.md) - UI component patterns
- [State Management](./06-state-management.md) - Data flow and state patterns

### Core Features
- [Authentication & Authorization](./07-auth.md) - User authentication system
- [Offline-First Strategy](./08-offline-first.md) - Offline functionality and sync
- [Real-time Features](./09-real-time.md) - Live updates and notifications
- [PDF Generation](./10-pdf-generation.md) - Document generation system

### Development
- [Development Guide](./11-development.md) - Local development setup
- [Testing Strategy](./12-testing.md) - Testing approaches and tools
- [Component Library](./13-component-library.md) - Reusable component documentation

### Deployment & Operations
- [Build & Deployment](./14-deployment.md) - Production deployment
- [Performance Optimization](./15-performance.md) - Performance best practices
- [Troubleshooting](./16-troubleshooting.md) - Common issues and solutions

## 🚀 Quick Start

```bash
# Clone and setup
git clone <repository-url>
cd liyali-gateway/frontend

# Install dependencies
pnpm install

# Setup environment
cp .env.example .env.local
# Edit .env.local with your settings

# Run development server
pnpm dev
```

## 🔧 Technology Stack

### Core Framework
- **Next.js 15** - React framework with App Router
- **React 19** - UI library with latest features
- **TypeScript** - Type-safe JavaScript

### UI & Styling
- **Tailwind CSS 4** - Utility-first CSS framework
- **shadcn/ui** - High-quality component library
- **Radix UI** - Accessible component primitives
- **Framer Motion** - Animation library
- **Lucide React** - Icon library

### State Management
- **Zustand** - Lightweight state management
- **TanStack Query** - Server state management
- **React Context** - Component state sharing

### Data & Storage
- **IndexedDB (idb)** - Client-side database
- **React Hook Form** - Form state management
- **Zod** - Schema validation

### Development Tools
- **ESLint** - Code linting
- **Prettier** - Code formatting
- **TypeScript** - Type checking

## 📊 System Capabilities

### Document Management
- **Requisitions** - Purchase request workflow
- **Purchase Orders** - Vendor order management
- **Payment Vouchers** - Payment processing
- **GRN (Goods Received Notes)** - Inventory receiving
- **Budgets** - Budget allocation and tracking

### User Experience
- **Responsive Design** - Works on all device sizes
- **Dark/Light Mode** - Theme switching
- **Offline Support** - Works without internet
- **Real-time Updates** - Live data synchronization
- **Progressive Web App** - Installable web app

### Performance Features
- **Code Splitting** - Lazy loading of components
- **Image Optimization** - Next.js image optimization
- **Caching Strategy** - Multi-layer caching
- **Bundle Analysis** - Performance monitoring

## 🎯 Key Patterns

### Component Architecture
- **Server Components** - Default for static content
- **Client Components** - For interactive features
- **Compound Components** - Complex UI patterns
- **Render Props** - Flexible component composition

### Data Flow
- **Server Actions** - Server-side mutations
- **React Query** - Client-side caching
- **Optimistic Updates** - Immediate UI feedback
- **Background Sync** - Offline-to-online sync

### Code Organization
- **Feature-based Structure** - Organized by functionality
- **Barrel Exports** - Clean import statements
- **Type-first Development** - Types define interfaces
- **Custom Hooks** - Reusable logic extraction

## 🤝 Contributing

Please read our [Development Guide](./11-development.md) for details on our code standards and development process.

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🆘 Support

- **Documentation**: Check the docs in this folder
- **Component Library**: See [Component Library](./13-component-library.md)
- **Troubleshooting**: See [Troubleshooting Guide](./16-troubleshooting.md)
- **Performance**: See [Performance Guide](./15-performance.md)