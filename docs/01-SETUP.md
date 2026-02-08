# Setup Guide

## Prerequisites

- Go 1.21+
- Node.js 18+
- PostgreSQL 14+
- Docker (optional)

## Quick Setup

### 1. Clone Repository

```bash
git clone <repository-url>
cd liyali-gateway
```

### 2. Backend Setup

```bash
cd backend
cp .env.example .env
# Edit .env with your database credentials
go mod download
make migrate
make seed
make dev
```

Backend runs at: http://localhost:8080

### 3. Frontend Setup

```bash
cd frontend
cp .env.example .env.local
# Edit .env.local
npm install
npm run dev
```

Frontend runs at: http://localhost:3000

### 4. Admin Console Setup

```bash
cd admin-console
cp .env.example .env.local
# Edit .env.local
npm install
npm run dev
```

Admin console runs at: http://localhost:3001

## Environment Variables

### Backend (.env)

```env
DATABASE_URL=postgresql://user:pass@localhost:5432/liyali_gateway
JWT_SECRET=your-secret-key
PORT=8080
ENVIRONMENT=development
```

### Frontend (.env.local)

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_APP_URL=http://localhost:3000
```

### Admin Console (.env.local)

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
```

## Database Setup

### Create Database

```bash
createdb liyali_gateway
```

### Run Migrations

```bash
cd backend
make migrate
```

### Seed Data

```bash
make seed
```

### Default Credentials

- Email: `admin@example.com`
- Password: `password`

## Docker Setup (Optional)

```bash
docker-compose up -d
```

## Verification

1. Backend: http://localhost:8080/health
2. Frontend: http://localhost:3000
3. Admin: http://localhost:3001

## Troubleshooting

### Database Connection Failed

- Check PostgreSQL is running
- Verify DATABASE_URL is correct
- Ensure database exists

### Port Already in Use

- Change PORT in .env files
- Kill process using the port

### Module Not Found

- Run `go mod download` (backend)
- Run `npm install` (frontend/admin)

## Next Steps

- [Development Guide](./02-DEVELOPMENT.md)
- [API Documentation](./03-API.md)
- [Deployment Guide](./04-DEPLOYMENT.md)
