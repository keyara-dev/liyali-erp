# Deployment Guide

## Quick Start

### Local Development

```bash
# Backend
cd backend
cp .env.example .env  # Configure environment
make db-reset         # Setup database
make run              # Start server at :8080

# Frontend
cd frontend
cp .env.example .env.local
npm install && npm run dev
```

### Docker

```bash
docker-compose up --build
# Backend: http://localhost:8080
# Frontend: http://localhost:3000
```

---

## Environment Variables

### Backend (.env)

```bash
# Required
DATABASE_URL=postgresql://user:pass@localhost:5432/liyali_gateway?sslmode=disable
JWT_SECRET=your-32-char-secret-here
PORT=8080
CORS_ALLOWED_ORIGINS=http://localhost:3000

# Optional
ENV=development
LOG_LEVEL=debug
```

### Frontend (.env.local)

```bash
# Required
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXTAUTH_SECRET=your-32-char-secret-here
NEXTAUTH_URL=http://localhost:3000
DATABASE_URL=postgresql://user:pass@localhost:5432/liyali_gateway

# Optional
NODE_ENV=development
```

### Generate Secrets

```bash
openssl rand -base64 32
```

---

## Fly.io Deployment (Demo/Staging)

### Setup (15 min)

```bash
# Install CLI
curl -L https://fly.io/install.sh | sh
flyctl auth login

# Deploy Backend
cd backend
flyctl launch --no-deploy
flyctl secrets set DATABASE_URL="postgresql://..." JWT_SECRET="..." CORS_ALLOWED_ORIGINS="*" PORT="8080"
flyctl deploy

# Deploy Frontend
cd ../frontend
flyctl launch --no-deploy
flyctl secrets set NEXT_PUBLIC_API_URL="https://backend.fly.dev" NEXTAUTH_SECRET="..." NEXTAUTH_URL="https://frontend.fly.dev"
flyctl deploy

# Update CORS with frontend URL
cd ../backend
flyctl secrets set CORS_ALLOWED_ORIGINS="https://frontend.fly.dev"
```

### Commands

```bash
flyctl logs              # View logs
flyctl status            # Check status
flyctl apps restart      # Restart app
flyctl secrets list      # List secrets
```

---

## Google Cloud Run (Production)

### Prerequisites

1. GCP project with billing enabled
2. Enable APIs: Cloud Run, Artifact Registry, Cloud Build
3. Create service account with roles: `run.admin`, `artifactregistry.writer`, `iam.serviceAccountUser`
4. Download service account key as `gcp-key.json`

### GitHub Secrets

```
GCP_SA_KEY              # Content of gcp-key.json
GCP_PROJECT_ID          # Your project ID
GCP_REGION              # e.g., us-central1
BACKEND_SERVICE_NAME    # e.g., liyali-backend
FRONTEND_SERVICE_NAME   # e.g., liyali-frontend
DATABASE_URL            # PostgreSQL connection string
JWT_SECRET              # Generated secret
CORS_ALLOWED_ORIGINS    # Frontend URL
NEXT_PUBLIC_API_URL     # Backend URL
NEXTAUTH_SECRET         # Generated secret
NEXTAUTH_URL            # Frontend URL
```

### Deploy

Push to `main` branch triggers automatic deployment via GitHub Actions.

```bash
git push origin main
```

### Verify

```bash
curl https://backend-url.run.app/health
curl https://frontend-url.run.app
```

---

## Database

### Migrations

```bash
cd backend
make db-migrate     # Run migrations
make db-reset       # Reset and reseed
```

### Manual Migration

```bash
psql $DATABASE_URL < database/migrations/001_init_system.up.sql
psql $DATABASE_URL < database/migrations/002_seed_data.up.sql
```

---

## Troubleshooting

| Issue | Solution |
|-------|----------|
| CORS errors | Update `CORS_ALLOWED_ORIGINS` with frontend URL |
| DB connection fails | Check `DATABASE_URL` format includes `?sslmode=require` for remote |
| Token errors | Verify `JWT_SECRET` is set and 32+ chars |
| Build fails | Check Docker builds locally: `docker build -t test .` |

### Logs

```bash
# Fly.io
flyctl logs -a app-name

# GCP
gcloud run services logs read service-name --region=us-central1
```

---

## Test Credentials

```
admin@liyali.com / Admin@123
```
