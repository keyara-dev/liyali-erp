# Liyali Gateway

Modern business operations platform for procurement, workflow automation, and team collaboration.

## Project Structure

```
liyali-gateway/
├── backend/              # Go backend API
├── frontend/             # Next.js user application
├── admin-console/        # Next.js admin portal
├── docs/                 # Shared documentation
└── scripts/              # Utility scripts
```

## Quick Start

### Prerequisites

- Go 1.21+
- Node.js 18+
- PostgreSQL 14+
- Docker (optional)

### Backend

```bash
cd backend
go mod download
make migrate
make seed
make dev
```

Access at: http://localhost:8080

### Frontend

```bash
cd frontend
npm install
npm run dev
```

Access at: http://localhost:3000

### Admin Console

```bash
cd admin-console
npm install
npm run dev
```

Access at: http://localhost:3001

## Documentation

### Backend

- [Quick Start](./backend/docs/01-quick-start.md)
- [API Reference](./backend/docs/13-api-reference.md)
- [Database](./backend/docs/DATABASE_IMPLEMENTATION.md)
- [Testing](./backend/docs/TESTING.md)

### Frontend

- [Quick Start](./frontend/docs/01-quick-start.md)
- [Architecture](./frontend/docs/04-architecture.md)
- [SEO](./frontend/docs/SEO.md)

### Admin Console

- [README](./admin-console/docs/README.md)

### Deployment

- [Deployment Guide](./docs/DEPLOYMENT.md)
- [Fly.io Guide](./docs/FLY_IO_DEPLOYMENT_GUIDE.md)
- [Admin Console Deployment](./ADMIN_CONSOLE_DEPLOYMENT_SETUP.md)
- [📚 Admin Console Deployment Index](./ADMIN_CONSOLE_DEPLOYMENT_INDEX.md) - Start here for admin console deployment

## Features

### Core Features

- ✅ Multi-tenant architecture
- ✅ Role-based access control (RBAC)
- ✅ Workflow automation
- ✅ Document management
- ✅ Real-time notifications
- ✅ Audit logging

### Subscription System

- ✅ Multiple subscription plans
- ✅ Feature-based access control
- ✅ Usage tracking
- ✅ Trial management

### Admin Features

- ✅ User management
- ✅ Organization management
- ✅ System settings
- ✅ Feature flags
- ✅ Analytics dashboard

## Tech Stack

### Backend

- **Language**: Go 1.21
- **Framework**: Fiber
- **Database**: PostgreSQL + SQLC
- **Auth**: JWT
- **Docs**: OpenAPI/Swagger

### Frontend

- **Framework**: Next.js 14
- **UI**: Tailwind CSS + shadcn/ui
- **State**: Zustand + React Query
- **Forms**: React Hook Form + Zod

### Admin Console

- **Framework**: Next.js 14
- **UI**: Tailwind CSS + shadcn/ui
- **State**: React Query

## Development

### Environment Variables

Create `.env` files in each directory:

**Backend** (`.env`):

```env
DATABASE_URL=postgresql://user:pass@localhost:5432/liyali_gateway
JWT_SECRET=your-secret-key
PORT=8080
```

**Frontend** (`.env.local`):

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_APP_URL=http://localhost:3000
```

**Admin Console** (`.env.local`):

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
```

### Running Tests

**Backend**:

```bash
cd backend
go test ./...
```

**Frontend**:

```bash
cd frontend
npm test
```

## Deployment

### Docker Compose

```bash
docker-compose up -d
```

### Individual Services

**Backend**:

```bash
cd backend
docker build -t liyali-backend .
docker run -p 8080:8080 liyali-backend
```

**Frontend**:

```bash
cd frontend
docker build -t liyali-frontend .
docker run -p 3000:3000 liyali-frontend
```

### Fly.io

See [Fly.io Deployment Guide](./FLY_IO_DEPLOYMENT_GUIDE.md)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Write/update tests
5. Submit a pull request

## License

Proprietary - All rights reserved

## Support

For support, email support@liyali.com or open an issue.
