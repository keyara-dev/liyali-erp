# Development Guide

## Project Structure

```
liyali-gateway/
├── backend/          # Go API
├── frontend/         # User app
├── admin-console/    # Admin portal
├── docs/            # Documentation
└── scripts/         # Utilities
```

## Development Workflow

### Backend Development

```bash
cd backend

# Run with hot reload
make dev

# Run tests
make test

# Format code
make fmt

# Lint
make lint
```

### Frontend Development

```bash
cd frontend

# Development server
npm run dev

# Type check
npm run type-check

# Lint
npm run lint

# Build
npm run build
```

### Admin Console Development

```bash
cd admin-console

# Development server
npm run dev

# Type check
npx tsc --noEmit

# Build
npm run build
```

## Code Standards

### Backend (Go)

- Follow Go conventions
- Use `gofmt` for formatting
- Write tests for new features
- Document exported functions
- Use SQLC for database queries

### Frontend (TypeScript)

- Use TypeScript strictly
- Follow ESLint rules
- Use Prettier for formatting
- Component naming: PascalCase
- File naming: kebab-case

## Git Workflow

### Branches

- `main` - Production
- `develop` - Development
- `feature/*` - New features
- `fix/*` - Bug fixes

### Commit Messages

```
feat: add user management
fix: resolve login issue
docs: update API documentation
refactor: improve auth service
test: add workflow tests
```

## Testing

### Backend Tests

```bash
# All tests
go test ./...

# Specific package
go test ./services/...

# With coverage
go test ./... -cover
```

### Frontend Tests

```bash
# Run tests
npm test

# Watch mode
npm test -- --watch

# Coverage
npm test -- --coverage
```

## Database Changes

### Create Migration

```bash
cd backend
migrate create -ext sql -dir database/migrations -seq migration_name
```

### Run Migrations

```bash
make migrate
```

### Rollback

```bash
make migrate-down
```

### Update SQLC

```bash
sqlc generate
```

## API Development

### Adding New Endpoint

1. Define route in `backend/routes/`
2. Create handler in `backend/handlers/`
3. Add service logic in `backend/services/`
4. Create repository in `backend/repository/`
5. Write SQLC queries in `backend/database/queries/`
6. Add tests
7. Update API documentation

### Testing Endpoints

Use `backend/scripts/test_requests.http` with REST Client extension.

## Frontend Development

### Adding New Page

1. Create page in `frontend/src/app/`
2. Add server actions in `frontend/src/app/_actions/`
3. Create components in `frontend/src/components/`
4. Add types in `frontend/src/types/`
5. Update navigation

### State Management

- Use Zustand for global state
- Use React Query for server state
- Use React Hook Form for forms

## Debugging

### Backend

```bash
# Enable debug logging
export LOG_LEVEL=debug

# Use delve debugger
dlv debug
```

### Frontend

- Use React DevTools
- Use browser DevTools
- Check Network tab for API calls

## Performance

### Backend

- Use database indexes
- Implement caching
- Optimize queries
- Use connection pooling

### Frontend

- Use Next.js Image component
- Implement code splitting
- Lazy load components
- Optimize bundle size

## Security

- Never commit secrets
- Use environment variables
- Validate all inputs
- Sanitize user data
- Use parameterized queries
- Implement rate limiting

## Documentation

- Update docs with code changes
- Add code comments
- Write clear commit messages
- Document breaking changes

## Resources

- [Backend Docs](../backend/docs/)
- [Frontend Docs](../frontend/docs/)
- [API Reference](./03-API.md)
- [Deployment](./04-DEPLOYMENT.md)
