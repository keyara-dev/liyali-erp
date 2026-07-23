# Deployment (Fly.io)

| | Value |
|---|---|
| App name | `liyali-admin-console` |
| Region | `jnb` (Johannesburg) |
| URL | `https://liyali-admin-console.fly.dev` |
| Port | `3001` |

## Deploy

```bash
# From repo root
make deploy-admin

# Or directly
cd admin-console && flyctl deploy
```

## Secrets

`NEXT_PUBLIC_API_URL` is a **build arg** (baked at build time) — set in `fly.toml [build.args]`, not as a secret.

```bash
# If AUTH_SECRET is needed
flyctl secrets set AUTH_SECRET=<value> --app liyali-admin-console
```

## Logs & Status

```bash
flyctl logs --app liyali-admin-console
flyctl status --app liyali-admin-console
flyctl ssh console --app liyali-admin-console
```

## Troubleshooting

**App won't start**
```bash
flyctl logs --app liyali-admin-console   # check for build/runtime errors
```

**CORS errors**
Ensure `CORS_ALLOWED_ORIGINS` on the backend includes `https://liyali-admin-console.fly.dev`:
```bash
flyctl secrets set CORS_ALLOWED_ORIGINS="https://liyali-gateway-frontend.fly.dev,https://liyali-admin-console.fly.dev" --app liyali-gateway-api
```

**Can't log in (401)**
- Confirm the user has `is_super_admin = true` in the database
- Check `NEXT_PUBLIC_API_URL` build arg points to the correct backend URL
