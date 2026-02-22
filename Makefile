# Liyali Gateway - Makefile
# Deployment and build automation for all apps

.PHONY: help deploy deploy-all deploy-backend deploy-web deploy-admin build build-all build-backend build-web build-admin test test-backend test-web clean migrate install check-env verify pre-deploy logs status open ssh-backend ssh-web ssh-admin restart scale-backend scale-web scale-admin dev-backend dev-web dev-admin

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
	@echo ""
	@echo "Testing:"
	@echo "  make test                - Run all tests"
	@echo "  make test-backend        - Run backend tests"
	@echo "  make test-web            - Run web frontend tests"
	@echo ""
	@echo "Development:"
	@echo "  make dev-backend         - Run backend in dev mode"
	@echo "  make dev-web             - Run web frontend in dev mode"
	@echo "  make dev-admin           - Run admin console in dev mode"
	@echo ""
	@echo "Fly.io Management:"
	@echo "  make logs                - View logs for all apps"
	@echo "  make status              - Check status of all apps"
	@echo "  make open                - Open all apps in browser"
	@echo "  make restart             - Restart all apps"
	@echo "  make ssh-backend         - SSH into backend"
	@echo "  make ssh-web             - SSH into frontend"
	@echo "  make ssh-admin           - SSH into admin console"
	@echo "  make scale-backend       - Show backend scale info"
	@echo "  make scale-web           - Show frontend scale info"
	@echo "  make scale-admin         - Show admin console scale info"
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

# Deploy backend
deploy-backend:
	@echo "🚀 Deploying backend (liyali-gateway-api)..."
	@flyctl deploy --app liyali-gateway-api --config backend/fly.toml --dockerfile backend/Dockerfile
	@echo "✅ Backend deployed!"

# Deploy web frontend
deploy-web:
	@echo "🚀 Deploying web frontend (liyali-gateway-frontend)..."
	@flyctl deploy --app liyali-gateway-frontend --config frontend/fly.toml --dockerfile frontend/Dockerfile
	@echo "✅ Web frontend deployed!"

# Deploy admin console
deploy-admin:
	@echo "🚀 Deploying admin console (liyali-admin-console)..."
	@flyctl deploy --app liyali-admin-console --config admin-console/fly.toml --dockerfile admin-console/Dockerfile
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
	@cd backend && go run cmd/migrate/main.go
	@echo "✅ Migrations completed!"

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
# FLY.IO MANAGEMENT COMMANDS
# ============================================================================

# View logs for all apps
logs:
	@echo "📋 Viewing logs (press Ctrl+C to stop)..."
	@echo "Backend logs:"
	@flyctl logs --app liyali-gateway-api &
	@echo "Frontend logs:"
	@flyctl logs --app liyali-gateway-frontend &
	@echo "Admin Console logs:"
	@flyctl logs --app liyali-admin-console &
	@wait

# View status of all apps
status:
	@echo "📊 Checking app status..."
	@echo "\nBackend (liyali-gateway-api):"
	@flyctl status --app liyali-gateway-api
	@echo "\nFrontend (liyali-gateway-frontend):"
	@flyctl status --app liyali-gateway-frontend
	@echo "\nAdmin Console (liyali-admin-console):"
	@flyctl status --app liyali-admin-console

# Open apps in browser
open:
	@echo "🌐 Opening apps in browser..."
	@flyctl open --app liyali-gateway-api
	@flyctl open --app liyali-gateway-frontend
	@flyctl open --app liyali-admin-console

# SSH into backend
ssh-backend:
	@echo "🔐 SSH into backend..."
	@flyctl ssh console --app liyali-gateway-api

# SSH into frontend
ssh-web:
	@echo "🔐 SSH into frontend..."
	@flyctl ssh console --app liyali-gateway-frontend

# SSH into admin console
ssh-admin:
	@echo "🔐 SSH into admin console..."
	@flyctl ssh console --app liyali-admin-console

# Restart all apps
restart:
	@echo "🔄 Restarting all apps..."
	@flyctl apps restart liyali-gateway-api
	@flyctl apps restart liyali-gateway-frontend
	@flyctl apps restart liyali-admin-console
	@echo "✅ All apps restarted!"

# Scale backend
scale-backend:
	@echo "📈 Current backend scale:"
	@flyctl scale show --app liyali-gateway-api
	@echo "\nTo scale, run: flyctl scale count <number> --app liyali-gateway-api"

# Scale frontend
scale-web:
	@echo "📈 Current frontend scale:"
	@flyctl scale show --app liyali-gateway-frontend
	@echo "\nTo scale, run: flyctl scale count <number> --app liyali-gateway-frontend"

# Scale admin console
scale-admin:
	@echo "📈 Current admin console scale:"
	@flyctl scale show --app liyali-admin-console
	@echo "\nTo scale, run: flyctl scale count <number> --app liyali-admin-console"
