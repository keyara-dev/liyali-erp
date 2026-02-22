# Makefile Usage Guide

Quick reference for using the Makefile commands.

---

## Quick Start

```bash
# Show all available commands
make help

# Deploy everything
make deploy

# Pre-deployment checks
make pre-deploy
```

---

## Deployment Commands

```bash
make deploy              # Deploy all apps (backend + web + admin)
make deploy-backend      # Deploy backend only
make deploy-web          # Deploy web frontend only
make deploy-admin        # Deploy admin console only
```

**How it works**: The Makefile runs flyctl deploy from the project root with explicit paths:

```bash
flyctl deploy --app liyali-gateway-api --config backend/fly.toml --dockerfile backend/Dockerfile
flyctl deploy --app liyali-gateway-frontend --config frontend/fly.toml --dockerfile frontend/Dockerfile
flyctl deploy --app liyali-admin-console --config admin-console/fly.toml --dockerfile admin-console/Dockerfile
```

---

## Build Commands

```bash
make build               # Build all apps
make build-backend       # Build backend only
make build-web           # Build web frontend only
make build-admin         # Build admin console only
```

---

## Testing Commands

```bash
make test                # Run all tests
make test-backend        # Run backend tests
make test-web            # Run web frontend tests
```

---

## Database Commands

```bash
make migrate             # Run database migrations
```

---

## Development Commands

```bash
make dev-backend         # Run backend in dev mode (http://localhost:8081)
make dev-web             # Run web frontend in dev mode (http://localhost:3000)
make dev-admin           # Run admin console in dev mode (http://localhost:3000)
```

---

## Utility Commands

```bash
make clean               # Clean build artifacts
make install             # Install all dependencies
make check-env           # Verify environment files exist
make verify              # Build + test all apps
make pre-deploy          # Full pre-deployment checks
```

---

## Common Workflows

### First Time Setup

```bash
# 1. Install dependencies
make install

# 2. Check environment
make check-env

# 3. Run migrations
make migrate

# 4. Build everything
make build

# 5. Run tests
make test
```

### Development

```bash
# Start backend
make dev-backend

# In another terminal, start frontend
make dev-web

# In another terminal, start admin console
make dev-admin
```

### Deployment

```bash
# 1. Pre-deployment checks
make pre-deploy

# 2. Deploy all apps
make deploy

# Or deploy individually
make deploy-backend
make deploy-web
make deploy-admin
```

### After Making Changes

```bash
# 1. Clean old builds
make clean

# 2. Rebuild
make build

# 3. Test
make test

# 4. Deploy
make deploy
```

---

## Tips

- Always run `make pre-deploy` before deploying to production
- Use `make check-env` to verify environment files are configured
- Run `make clean` if you encounter build issues
- Use individual deploy commands for faster deployments when only one app changed

---

**See also**: [DEPLOYMENT_GUIDE.md](DEPLOYMENT_GUIDE.md) for detailed deployment instructions
