# Installation Guide

This guide covers setting up the Liyali Gateway frontend development environment.

## Prerequisites

### Required Software

- **Node.js**: Version 18.17 or higher (LTS recommended)
- **npm**: Version 9.0 or higher (comes with Node.js)
- **Git**: For version control

### Recommended Tools

- **VS Code**: With recommended extensions
- **Chrome DevTools**: For debugging
- **React Developer Tools**: Browser extension
- **TanStack Query DevTools**: Built into the app

## Installation Steps

### 1. Clone the Repository

```bash
git clone <repository-url>
cd liyali-gateway/frontend
```

### 2. Install Dependencies

```bash
npm install
```

This installs all dependencies including:
- Next.js 15 with React 19
- TypeScript and type definitions
- Tailwind CSS 4 and shadcn/ui components
- TanStack Query for state management
- Development tools and linters

### 3. Environment Setup

Copy the environment template:

```bash
cp .env.example .env.local
```

Configure environment variables:

```bash
# .env.local
NEXT_PUBLIC_APP_URL=http://localhost:3000
NEXT_PUBLIC_API_URL=http://localhost:3001/api
AUTH_SECRET=your-32-character-secret-key-here
NODE_ENV=development
```

### 4. Verify Installation

Start the development server:

```bash
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) in your browser.

You should see the Liyali Gateway login page.

## Development Dependencies

### Core Framework

```json
{
  "next": "^15.0.0",
  "react": "^19.0.0",
  "react-dom": "^19.0.0",
  "typescript": "^5.0.0"
}
```

### UI and Styling

```json
{
  "@tailwindcss/typography": "^0.5.15",
  "tailwindcss": "^4.0.0",
  "@radix-ui/react-*": "Various Radix UI components",
  "lucide-react": "^0.460.0",
  "next-themes": "^0.4.3"
}
```

### State Management

```json
{
  "@tanstack/react-query": "^5.59.16",
  "@tanstack/react-query-devtools": "^5.59.16",
  "zustand": "^5.0.1"
}
```

### Forms and Validation

```json
{
  "react-hook-form": "^7.53.2",
  "@hookform/resolvers": "^3.9.1",
  "zod": "^3.23.8"
}
```

### Utilities

```json
{
  "date-fns": "^4.1.0",
  "clsx": "^2.1.1",
  "class-variance-authority": "^0.7.1",
  "sonner": "^1.7.1"
}
```

## VS Code Setup

### Recommended Extensions

Create `.vscode/extensions.json`:

```json
{
  "recommendations": [
    "bradlc.vscode-tailwindcss",
    "ms-vscode.vscode-typescript-next",
    "esbenp.prettier-vscode",
    "ms-vscode.vscode-eslint",
    "formulahendry.auto-rename-tag",
    "christian-kohler.path-intellisense",
    "ms-vscode.vscode-json"
  ]
}
```

### Workspace Settings

Create `.vscode/settings.json`:

```json
{
  "typescript.preferences.importModuleSpecifier": "relative",
  "editor.formatOnSave": true,
  "editor.defaultFormatter": "esbenp.prettier-vscode",
  "editor.codeActionsOnSave": {
    "source.fixAll.eslint": "explicit"
  },
  "tailwindCSS.experimental.classRegex": [
    ["cva\\(([^)]*)\\)", "[\"'`]([^\"'`]*).*?[\"'`]"],
    ["cx\\(([^)]*)\\)", "(?:'|\"|`)([^']*)(?:'|\"|`)"]
  ]
}
```

## Package Scripts

### Development

```bash
npm run dev          # Start development server
npm run dev:turbo    # Start with Turbopack (faster)
```

### Building

```bash
npm run build        # Build for production
npm run start        # Start production server
npm run export       # Export static site
```

### Code Quality

```bash
npm run lint         # Run ESLint
npm run lint:fix     # Fix ESLint issues
npm run type-check   # Run TypeScript compiler
npm run format       # Format with Prettier
```

### Testing

```bash
npm run test         # Run tests
npm run test:watch   # Run tests in watch mode
npm run test:coverage # Run tests with coverage
```

## Troubleshooting

### Common Issues

#### Node Version Mismatch

```bash
# Check Node version
node --version

# Use nvm to switch versions
nvm use 18
nvm install 18.17.0
```

#### Port Already in Use

```bash
# Kill process on port 3000
npx kill-port 3000

# Or use different port
npm run dev -- -p 3001
```

#### Module Resolution Issues

```bash
# Clear npm cache
npm cache clean --force

# Delete node_modules and reinstall
rm -rf node_modules package-lock.json
npm install
```

#### TypeScript Errors

```bash
# Restart TypeScript server in VS Code
Ctrl+Shift+P -> "TypeScript: Restart TS Server"

# Check TypeScript configuration
npm run type-check
```

### Environment Issues

#### Missing Environment Variables

Ensure all required variables are set in `.env.local`:

```bash
# Check if variables are loaded
npm run dev
# Look for "Environment loaded" in console
```

#### CORS Issues

If connecting to a backend API:

```bash
# Backend must allow frontend origin
Access-Control-Allow-Origin: http://localhost:3000
```

### Build Issues

#### Out of Memory

```bash
# Increase Node memory limit
export NODE_OPTIONS="--max-old-space-size=4096"
npm run build
```

#### Static Export Issues

```bash
# Check for dynamic imports or server-side code
npm run build
npm run export
```

## Next Steps

After successful installation:

1. **Read the [Quick Start Guide](./01-quick-start.md)** for basic usage
2. **Review the [Architecture Guide](./04-architecture.md)** to understand the codebase
3. **Check the [Development Guide](./11-development.md)** for coding standards
4. **Set up your IDE** with recommended extensions and settings

## Getting Help

- **Documentation**: Check other guides in this `docs/` folder
- **Issues**: Create GitHub issues for bugs or feature requests
- **Development**: Join the development team chat
- **Code Review**: Follow the pull request process

The installation should be straightforward, but don't hesitate to ask for help if you encounter issues.