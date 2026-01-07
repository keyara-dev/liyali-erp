# Setup Guide

## Prerequisites

- **Go**: 1.21 or higher
- **Node.js**: 18 or higher  
- **PostgreSQL**: 14 or higher
- **Git**: Latest version

## Database Setup

### 1. Create Database
```sql
CREATE DATABASE liyali_gateway;
CREATE USER liyali_user WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE liyali_gateway TO liyali_user;
```

### 2. Environment Configuration
```bash
# Backend (.env)
DATABASE_URL=postgres://liyali_user:your_password@localhost:5432/liyali_gateway
JWT_SECRET=your-super-secret-jwt-key-here
PORT=8080

# Frontend (.env.local)
NEXT_PUBLIC_API_URL=http://localhost:8080
```

### 3. Run Migrations
```bash
cd backend/database
./migrate.sh up
```

## Backend Setup

```bash
cd backend
go mod tidy
go run main.go
```

Server starts on `http://localhost:8080`

## Frontend Setup

```bash
cd frontend
npm install
npm run dev
```

Application available at `http://localhost:3000`

## Default Users

After database seeding, these accounts are available:

| Role | Email | Password | Organization |
|------|-------|----------|--------------|
| Admin | `admin@liyali.com` | `admin123` | Default Org |
| Requester | `requester@demo.com` | `admin123` | Demo Corp |
| Approver | `approver@demo.com` | `admin123` | Demo Corp |
| Finance | `finance@demo.com` | `admin123` | Demo Corp |
| Manager | `manager@demo.com` | `admin123` | Demo Corp |

## Verification

### Backend Health Check
```bash
curl http://localhost:8080/health
# Expected: {"status": "ok"}
```

### Database Verification
```sql
-- Check tables
SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';
-- Expected: 30+ tables

-- Check users
SELECT email, name, role FROM users;
-- Expected: 5 users
```

### Frontend Test
1. Navigate to `http://localhost:3000`
2. Login with any user from the table above
3. Verify dashboard loads with data

### API Testing with cURL

#### Login Test
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@liyali.com","password":"admin123"}'
```

#### Create Requisition Test
```bash
curl -X POST http://localhost:8080/api/v1/requisitions \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "X-Organization-ID: YOUR_ORG_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Office Supplies",
    "description": "Monthly supplies order",
    "department": "Operations",
    "priority": "medium",
    "items": [{"description": "Paper", "quantity": 10, "unitPrice": 5.00, "amount": 50.00}],
    "totalAmount": 50.00,
    "currency": "USD"
  }'
```

## Troubleshooting

### Database Connection Issues
- Verify PostgreSQL is running
- Check connection string in `.env`
- Ensure database and user exist

### Migration Failures
```bash
# Reset database
cd backend/database
./migrate.sh reset
```

### Port Conflicts
- Backend: Change `PORT` in `.env`
- Frontend: Use `npm run dev -- -p 3001`

## Production Deployment

### Environment Variables
```bash
# Production backend
DATABASE_URL=your-production-db-url
JWT_SECRET=your-production-jwt-secret
ENVIRONMENT=production

# Production frontend
NEXT_PUBLIC_API_URL=https://your-api-domain.com
```

### Build Commands
```bash
# Backend
cd backend
go build -o liyali-gateway

# Frontend
cd frontend
npm run build
npm start
```

## Docker Setup (Optional)

```bash
# Start services
docker-compose up -d

# Run migrations
docker-compose exec backend ./migrate.sh up
```

---

**Next**: See [Authentication Guide](AUTH.md) for user management