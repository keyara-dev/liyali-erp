# Backend Setup Instructions

Run these commands in order to set up the Go backend:

## Step 1: Initialize Go Module

```bash
cd /Users/cre8tor/Documents/Code/liyali-gateway/backend
go mod init github.com/cozyCodr/liyali-gateway
```

## Step 2: Create Project Structure

```bash
# Create directories
mkdir -p cmd/api
mkdir -p internal/{config,database/{migrations,queries},handlers,middleware,services,rbac}
mkdir -p tests/{unit,integration}
mkdir -p scripts
```

## Step 3: Install Dependencies

```bash
# Core dependencies
go get github.com/gofiber/fiber/v3
go get github.com/jackc/pgx/v5/pgxpool
go get golang.org/x/crypto/bcrypt
go get github.com/golang-jwt/jwt/v5
go get github.com/rs/zerolog

# Validation
go get github.com/go-playground/validator/v10

# Email
go get github.com/sendgrid/sendgrid-go

# Testing
go get github.com/stretchr/testify

# Environment variables
go get github.com/joho/godotenv
```

## Step 4: Install sqlc

```bash
# Install sqlc
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Verify installation
sqlc version
```

## Step 5: Create Environment File

```bash
# Copy and edit .env file
cat > .env << 'EOF'
DATABASE_URL=postgresql://postgres:password@localhost:5432/liyali_db?sslmode=disable
JWT_SECRET=your-secret-key-change-in-production-min-32-characters
PORT=3001
NODE_ENV=development
SENDGRID_API_KEY=your-sendgrid-api-key
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001
EOF
```

## Step 6: Create sqlc Configuration

```bash
cat > sqlc.yaml << 'EOF'
version: "2"
sql:
  - engine: "postgresql"
    queries: "internal/database/queries"
    schema: "internal/database/migrations"
    gen:
      go:
        package: "db"
        out: "internal/db"
        sql_package: "pgx/v5"
        emit_json_tags: true
        emit_prepared_queries: false
        emit_interface: true
        emit_exact_table_names: false
        emit_empty_slices: true
        emit_all_enum_values: true
EOF
```

## Next Steps

After running these commands, you're ready to:
1. Create database migrations (see auth-rbac-quickstart.md Step 1)
2. Create sqlc queries (see auth-rbac-quickstart.md Step 1)
3. Generate sqlc code with `sqlc generate`
4. Start building the auth service

**Reference**: See `docs/auth-rbac-quickstart.md` for complete implementation guide
