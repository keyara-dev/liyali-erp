# Deployment (Fly.io)

All three apps are deployed to Fly.io in the `jnb` (Johannesburg) region.

| App | Fly app name | URL |
|---|---|---|
| Backend | `liyali-gateway-api` | `https://liyali-gateway-api.fly.dev` |
| Frontend | `liyali-gateway-frontend` | `https://liyali-gateway-frontend.fly.dev` |
| Admin Console | `liyali-gateway-admin` | `https://liyali-gateway-admin.fly.dev` |

## Deploy

```bash
# From repo root — deploys all three in sequence
make deploy

# Individual
make deploy-backend
make deploy-frontend
make deploy-admin
```

## Secrets

Set once via CLI, persisted by Fly.io:

```bash
# Backend
flyctl secrets set JWT_SECRET=<value> --app liyali-gateway-api
flyctl secrets set DATABASE_URL=<value> --app liyali-gateway-api
flyctl secrets set CORS_ALLOWED_ORIGINS=<value> --app liyali-gateway-api

# Frontend
flyctl secrets set AUTH_SECRET=<value> --app liyali-gateway-frontend
# BASE_URL is set in frontend/fly.toml [env] — not a secret
```

## Run Migrations in Production

```bash
flyctl ssh console --app liyali-gateway-api -C "go run database/migrate_all.go"
```

For a full reset (drops all data):
```bash
flyctl ssh console --app liyali-gateway-api -C "go run database/migrate_all.go --reset"
```

## Logs & Status

```bash
flyctl logs --app liyali-gateway-api
flyctl status --app liyali-gateway-api
flyctl ssh console --app liyali-gateway-api   # SSH into running machine
```

## Database

```bash
# Connect via proxy
flyctl proxy 5432 --app liyali-gateway-db
psql postgresql://postgres:<pass>@localhost:5432/liyali_gateway

# Backup
flyctl postgres backup create --app liyali-gateway-db
```

## VM Sizing (current)

All apps: `shared-cpu-1x`, `512MB RAM`, `min_machines_running = 1`, `auto_stop = false`.

Scale up if needed:
```bash
flyctl scale vm shared-cpu-2x --memory 1024 --app liyali-gateway-api
```
