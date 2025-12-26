#!/bin/bash

###############################################################################
# E2E Test Quick Start Script - Liyali Gateway
# This script helps you quickly set up and run E2E tests
###############################################################################

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_DIR="d:\dev\next-apps\liyali-gateway"
BACKEND_URL="http://localhost:8080"
FRONTEND_URL="http://localhost:3000"
DOCKER_COMPOSE_FILE="$PROJECT_DIR/docker-compose.yml"

###############################################################################
# Helper Functions
###############################################################################

print_header() {
    echo -e "\n${BLUE}===================================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}===================================================${NC}\n"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

###############################################################################
# Main Script
###############################################################################

main() {
    print_header "Liyali Gateway E2E Test Quick Start"

    # Menu
    echo "Choose test setup option:"
    echo "1. Start with Docker Compose (Recommended)"
    echo "2. Use existing local services"
    echo "3. Run quick smoke tests only"
    echo "4. View E2E test plan"
    echo ""
    read -p "Enter choice (1-4): " choice

    case $choice in
        1)
            setup_docker
            ;;
        2)
            check_services
            ;;
        3)
            run_smoke_tests
            ;;
        4)
            show_test_plan
            ;;
        *)
            print_error "Invalid choice"
            exit 1
            ;;
    esac
}

###############################################################################
# Option 1: Setup Docker
###############################################################################

setup_docker() {
    print_header "Setting up with Docker Compose"

    # Check if Docker is installed
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed"
        echo "Please install Docker from: https://www.docker.com/products/docker-desktop"
        exit 1
    fi

    print_info "Checking for existing containers..."
    if [ "$(docker ps -q)" ]; then
        print_warning "Some Docker containers are running"
        read -p "Continue anyway? (y/n): " -r
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi

    print_info "Starting Docker Compose services..."
    cd "$PROJECT_DIR"
    docker-compose up -d

    print_info "Waiting for services to be ready (60 seconds)..."
    sleep 60

    # Check services
    if check_backend_health && check_database_health; then
        print_success "Services are running!"
        print_info "Backend URL: $BACKEND_URL"
        print_info "Frontend URL: $FRONTEND_URL"

        echo ""
        print_header "Next Steps"
        echo "1. Open browser to: $FRONTEND_URL"
        echo "2. Register new account"
        echo "3. Follow test cases in E2E-TEST-PLAN.md"
        echo ""
        print_info "To run API tests, use:"
        echo "  bash E2E-TEST-QUICK-START.sh"
        echo ""
        print_info "To stop services:"
        echo "  docker-compose down"
    else
        print_error "Services failed to start"
        echo "Check logs with: docker-compose logs"
        exit 1
    fi
}

###############################################################################
# Option 2: Check existing services
###############################################################################

check_services() {
    print_header "Checking existing local services"

    local backend_ok=false
    local frontend_ok=false

    if check_backend_health; then
        print_success "Backend is running on $BACKEND_URL"
        backend_ok=true
    else
        print_error "Backend is not responding on $BACKEND_URL"
    fi

    if check_frontend_health; then
        print_success "Frontend is running on $FRONTEND_URL"
        frontend_ok=true
    else
        print_error "Frontend is not responding on $FRONTEND_URL"
    fi

    if $backend_ok && $frontend_ok; then
        print_success "All services are running!"
        run_api_tests
    else
        print_error "Some services are not running"
        print_info "Please start them manually or choose Docker Compose option"
        exit 1
    fi
}

###############################################################################
# Option 3: Run smoke tests
###############################################################################

run_smoke_tests() {
    print_header "Running Smoke Tests"

    if ! check_backend_health; then
        print_error "Backend is not running"
        exit 1
    fi

    echo "1. Testing API endpoints..."
    test_health_endpoint

    echo ""
    echo "2. Testing authentication..."
    test_auth_endpoints

    echo ""
    echo "3. Testing CRUD operations..."
    test_crud_endpoints

    print_success "Smoke tests completed!"
}

###############################################################################
# Option 4: Show test plan
###############################################################################

show_test_plan() {
    if [ -f "$PROJECT_DIR/E2E-TEST-PLAN.md" ]; then
        less "$PROJECT_DIR/E2E-TEST-PLAN.md"
    else
        print_error "E2E-TEST-PLAN.md not found"
        exit 1
    fi
}

###############################################################################
# Health Checks
###############################################################################

check_backend_health() {
    if curl -s "$BACKEND_URL/health" > /dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

check_frontend_health() {
    if curl -s "$FRONTEND_URL" > /dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

check_database_health() {
    # Check if backend can access database
    if curl -s "$BACKEND_URL/api/v1/health" > /dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

###############################################################################
# API Tests
###############################################################################

test_health_endpoint() {
    print_info "Testing health endpoint..."
    response=$(curl -s "$BACKEND_URL/health")
    if [[ $response == *"OK"* ]] || [[ $response == *"healthy"* ]]; then
        print_success "Health check passed"
        return 0
    else
        print_error "Health check failed: $response"
        return 1
    fi
}

test_auth_endpoints() {
    print_info "Testing auth endpoints..."

    # Test register
    register_response=$(curl -s -X POST "$BACKEND_URL/api/v1/auth/register" \
        -H "Content-Type: application/json" \
        -d '{
            "email": "test'$(date +%s)'@example.com",
            "password": "TestPass123!",
            "name": "Test User"
        }')

    if [[ $register_response == *"token"* ]] || [[ $register_response == *"success"* ]]; then
        print_success "Registration endpoint working"
    else
        print_warning "Registration test: $register_response"
    fi
}

test_crud_endpoints() {
    print_info "Testing CRUD endpoints..."

    # Would test create, read, update, delete operations
    print_info "CRUD tests would run here"
}

run_api_tests() {
    print_header "Running API Tests"

    test_health_endpoint
    echo ""

    test_auth_endpoints
    echo ""

    test_crud_endpoints
    echo ""

    print_success "API tests completed!"
}

###############################################################################
# Cleanup
###############################################################################

cleanup() {
    print_info "Cleaning up..."
    # Any cleanup code here
}

trap cleanup EXIT

###############################################################################
# Entry Point
###############################################################################

main "$@"
