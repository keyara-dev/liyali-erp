# Liyali Gateway - Makefile
# Deployment and build automation for all apps

.PHONY: help deploy deploy-all deploy-backend deploy-web deploy-admin build build-all build-backend build-web build-admin test test-backend test-web clean migrate indexes install check-env verify pre-deploy logs status open dev-backend dev-web dev-admin air

# Default target
help:
	@echo "Liyali Gateway - Available Commands"
	@echo ""
	@echo "Deployment:"
	@echo "  make deploy              - Deploy all apps (backend + web + admin)"
	@echo "  make deploy-backend      - Deploy backend only"
	@echo "  make deploy-web          - Deploy web frontend only"
	@echo "  make deploy-admin        - Deploy admin console only"
	@echo ""
	@echo "Build:"
	@echo "  make build               - Build all apps"
	@echo "  make build-backend       - Build backend only"
	@echo "  make build-web           - Build web frontend only"
	@echo "  make build-admin         - Build admin console only"
	@echo ""
	@echo "Database:"
	@echo "  make migrate             - Run database migrations"
	@echo "  make indexes             - Create concurrent indexes (performance optimization)"
	@echo ""
	@echo "  📖 Database Documentation:"
	@echo "     backend/database/README.md              - Database management guide"
	@echo "     backend/database/migrations/README.md   - Detailed migration guide"
	@echo "     backend/database/MIGRATION_SUMMARY.md   - Migration history"
	@echo ""
	@echo "Testing:"
	@echo "  make test                - Run all tests"
	@echo "  make test-backend        - Run backend tests"
	@echo "  make test-web            - Run web frontend tests"
	@echo ""
	@echo "Development:"
	@echo "  make dev-backend         - Run backend in dev mode"
	@echo "  make air                 - Run backend with air (hot-reload)"
	@echo "  make dev-web             - Run web frontend in dev mode"
	@echo "  make dev-admin           - Run admin console in dev mode"
	@echo ""
	@echo "Railway Management:"
	@echo "  make logs                - Tail logs (current linked service)"
	@echo "  make status              - Show deployment status"
	@echo "  make open                - Open the Railway dashboard"
	@echo ""
	@echo "Utilities:"
	@echo "  make clean               - Clean build artifacts"
	@echo "  make install             - Install dependencies"
	@echo "  make check-env           - Verify environment setup"
	@echo "  make verify              - Build + test"
	@echo "  make pre-deploy          - Full pre-deployment checks"

# ============================================================================
# DEPLOYMENT COMMANDS
# ============================================================================

# Deploy all apps
deploy: deploy-backend deploy-web deploy-admin
	@echo "✅ All apps deployed successfully!"

deploy-all: deploy

# Deploy backend (Railway — run `railway link` once per service first)
deploy-backend:
	@echo "🚀 Deploying backend to Railway..."
	@cd backend && railway up
	@echo "✅ Backend deployed!"

# Deploy web frontend
deploy-web:
	@echo "🚀 Deploying web frontend to Railway..."
	@cd frontend && railway up
	@echo "✅ Web frontend deployed!"

# Deploy admin console
deploy-admin:
	@echo "🚀 Deploying admin console to Railway..."
	@cd admin-console && railway up
	@echo "✅ Admin console deployed!"

# ============================================================================
# BUILD COMMANDS
# ============================================================================

# Build all apps
build: build-backend build-web build-admin
	@echo "✅ All apps built successfully!"

build-all: build

# Build backend
build-backend:
	@echo "🔨 Building backend..."
	@cd backend && go build -o liyali-backend .
	@echo "✅ Backend built: backend/liyali-backend"

# Build web frontend
build-web:
	@echo "🔨 Building web frontend..."
	@cd frontend && npm run build
	@echo "✅ Web frontend built: frontend/.next/"

# Build admin console
build-admin:
	@echo "🔨 Building admin console..."
	@cd admin-console && npm run build
	@echo "✅ Admin console built: admin-console/.next/"

# ============================================================================
# DATABASE COMMANDS
# ============================================================================

# Run database migrations
migrate:
	@echo "🗄️  Running database migrations..."
	@cd backend && export $$(cat .env | grep -v '^#' | xargs) && go run cmd/migrate/main.go
	@echo "✅ Migrations completed!"

# Create concurrent indexes (run separately, not in transaction)
indexes:
	@echo "📊 Creating concurrent indexes..."
	@cd backend && export $$(cat .env | grep -v '^#' | xargs) && psql $$DATABASE_URL -f database/scripts/create_concurrent_indexes.sql
	@echo "✅ Indexes created!"

# ============================================================================
# TESTING COMMANDS
# ============================================================================

# Run all tests
test: test-backend test-web
	@echo "✅ All tests passed!"

# Run backend tests
test-backend:
	@echo "🧪 Running backend tests..."
	@cd backend && go test ./...
	@echo "✅ Backend tests passed!"

# Run web frontend tests
test-web:
	@echo "🧪 Running web frontend tests..."
	@cd frontend && npm run build
	@echo "✅ Web frontend tests passed!"

# ============================================================================
# DEVELOPMENT COMMANDS
# ============================================================================

# Run backend in dev mode
dev-backend:
	@echo "🔧 Starting backend in dev mode..."
	@cd backend && go run main.go

# Run backend with air (hot-reload)
air:
	@echo "🔥 Starting backend with air (hot-reload)..."
	@cd backend && $$(go env GOPATH)/bin/air

# Run web frontend in dev mode
dev-web:
	@echo "🔧 Starting web frontend in dev mode..."
	@cd frontend && npm run dev

# Run admin console in dev mode
dev-admin:
	@echo "🔧 Starting admin console in dev mode..."
	@cd admin-console && npm run dev

# ============================================================================
# UTILITY COMMANDS
# ============================================================================

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -f backend/liyali-backend
	@rm -rf frontend/.next
	@rm -rf frontend/node_modules/.cache
	@rm -rf admin-console/.next
	@rm -rf admin-console/node_modules/.cache
	@echo "✅ Clean complete!"

# Install dependencies
install:
	@echo "📦 Installing dependencies..."
	@cd backend && go mod download
	@echo "📦 Installing air (hot-reload)..."
	@go install github.com/air-verse/air@latest
	@cd frontend && npm install
	@cd admin-console && npm install
	@echo "✅ Dependencies installed!"

# Check environment setup
check-env:
	@echo "🔍 Checking environment setup..."
	@echo "Backend .env:"
	@test -f backend/.env && echo "  ✅ backend/.env exists" || echo "  ❌ backend/.env missing"
	@echo "Frontend .env:"
	@test -f frontend/.env && echo "  ✅ frontend/.env exists" || echo "  ❌ frontend/.env missing"
	@echo "Admin Console .env:"
	@test -f admin-console/.env && echo "  ✅ admin-console/.env exists" || echo "  ❌ admin-console/.env missing"

# Verify builds
verify: build test
	@echo "✅ All builds verified!"

# Pre-deployment checks
pre-deploy: check-env verify migrate
	@echo "✅ Pre-deployment checks complete!"
	@echo "Ready to deploy!"

# ============================================================================
# RAILWAY MANAGEMENT COMMANDS
# ============================================================================

# Tail logs for the current linked Railway service
logs:
	@echo "📋 Tailing Railway logs (Ctrl+C to stop)..."
	@railway logs

# Show deployment status
status:
	@echo "📊 Railway status..."
	@railway status

# Open the Railway project dashboard
open:
	@echo "🌐 Opening Railway dashboard..."
	@railway open
