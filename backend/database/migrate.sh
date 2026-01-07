#!/bin/bash

# Migration script for Liyali Gateway (Consolidated Version)
# Usage: ./migrate.sh [up|down|reset|seed|unseed|drop]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if .env file exists (go up one directory to backend root)
if [ ! -f "../.env" ]; then
    echo -e "${RED}Error: .env file not found${NC}"
    echo "Please create a .env file with database configuration in the backend directory"
    exit 1
fi

# Load environment variables (from parent directory)
source ../.env

# Default action
ACTION=${1:-up}

echo -e "${BLUE}🚀 Liyali Gateway Database Migration (Consolidated)${NC}"
echo "=================================================="
echo -e "${BLUE}Database: ${DB_NAME} | Host: ${DB_HOST}:${DB_PORT}${NC}"
echo ""

# Function to run SQL migration directly
run_sql_migration() {
    local sql_file=$1
    local description=$2
    
    echo -e "${YELLOW}📋 ${description}...${NC}"
    
    # Use psql to run the SQL file directly
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f "$sql_file" -v ON_ERROR_STOP=1
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ ${description} completed successfully!${NC}"
    else
        echo -e "${RED}❌ ${description} failed!${NC}"
        exit 1
    fi
}

case $ACTION in
    "up")
        echo -e "${GREEN}🔧 Running complete setup (schema + seed data)...${NC}"
        run_sql_migration "migrations/001_create_complete_schema_consolidated.up.sql" "Creating database schema"
        run_sql_migration "migrations/002_seed_initial_data.up.sql" "Seeding initial data"
        echo ""
        echo -e "${GREEN}🎉 Complete setup finished successfully!${NC}"
        echo -e "${BLUE}📊 Database now contains:${NC}"
        echo "   • Complete schema with all tables and indexes"
        echo "   • 2 organizations (Default + Demo Corporation)"
        echo "   • 5 users with different roles"
        echo "   • 5 vendors and 6 categories"
        echo "   • 4 sample budgets and 6 workflows"
        echo "   • 3 sample requisitions for testing"
        ;;
    "down")
        echo -e "${YELLOW}🔄 Running complete teardown (remove data + schema)...${NC}"
        echo -e "${RED}⚠️  This will REMOVE ALL DATA and DROP ALL TABLES! Are you sure? (y/N)${NC}"
        read -r response
        if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
            run_sql_migration "migrations/002_seed_initial_data.down.sql" "Removing seed data"
            run_sql_migration "migrations/001_create_complete_schema_consolidated.down.sql" "Dropping database schema"
            echo -e "${GREEN}✅ Complete teardown completed successfully!${NC}"
        else
            echo -e "${YELLOW}Migration cancelled${NC}"
        fi
        ;;
    "reset")
        echo -e "${YELLOW}🔄 Resetting database (complete fresh start)...${NC}"
        echo -e "${RED}⚠️  This will DROP ALL TABLES and recreate everything! Are you sure? (y/N)${NC}"
        read -r response
        if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
            echo -e "${YELLOW}Step 1: Removing existing data...${NC}"
            run_sql_migration "migrations/002_seed_initial_data.down.sql" "Removing seed data" || true
            echo -e "${YELLOW}Step 2: Dropping schema...${NC}"
            run_sql_migration "migrations/001_create_complete_schema_consolidated.down.sql" "Dropping database schema" || true
            echo -e "${YELLOW}Step 3: Creating fresh schema...${NC}"
            run_sql_migration "migrations/001_create_complete_schema_consolidated.up.sql" "Creating database schema"
            echo -e "${YELLOW}Step 4: Seeding fresh data...${NC}"
            run_sql_migration "migrations/002_seed_initial_data.up.sql" "Seeding initial data"
            echo ""
            echo -e "${GREEN}🎉 Database reset completed successfully!${NC}"
            echo -e "${BLUE}📊 Fresh database ready with complete sample data!${NC}"
        else
            echo -e "${YELLOW}Migration cancelled${NC}"
        fi
        ;;
    "seed")
        echo -e "${GREEN}🌱 Seeding database with sample data...${NC}"
        run_sql_migration "migrations/002_seed_initial_data.up.sql" "Seeding initial data"
        echo -e "${BLUE}📊 Sample data includes:${NC}"
        echo "   • 2 organizations with settings"
        echo "   • 5 users (admin, requester, approver, finance, manager)"
        echo "   • 5 vendors and 6 categories"
        echo "   • 4 budgets and 6 workflows"
        echo "   • 3 sample requisitions"
        ;;
    "unseed")
        echo -e "${YELLOW}🧹 Removing seed data...${NC}"
        echo -e "${RED}⚠️  This will REMOVE ALL SAMPLE DATA! Are you sure? (y/N)${NC}"
        read -r response
        if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
            run_sql_migration "migrations/002_seed_initial_data.down.sql" "Removing seed data"
            echo -e "${GREEN}✅ Seed data removed successfully!${NC}"
            echo -e "${BLUE}📊 Database now has empty tables (schema preserved)${NC}"
        else
            echo -e "${YELLOW}Operation cancelled${NC}"
        fi
        ;;
    "drop")
        echo -e "${YELLOW}💥 Emergency: Dropping all tables...${NC}"
        echo -e "${RED}⚠️  This will DROP ALL TABLES! Are you sure? (y/N)${NC}"
        read -r response
        if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
            run_sql_migration "migrations/000_drop_all_tables.up.sql" "Emergency table drop"
            echo -e "${GREEN}✅ All tables dropped successfully!${NC}"
        else
            echo -e "${YELLOW}Operation cancelled${NC}"
        fi
        ;;
    "schema")
        echo -e "${GREEN}🏗️  Creating database schema only (no data)...${NC}"
        run_sql_migration "migrations/001_create_complete_schema_consolidated.up.sql" "Creating database schema"
        echo -e "${BLUE}📊 Schema created with empty tables${NC}"
        ;;
    *)
        echo -e "${RED}Invalid action: $ACTION${NC}"
        echo "Usage: $0 [up|down|reset|seed|unseed|drop|schema]"
        echo ""
        echo -e "${BLUE}Actions:${NC}"
        echo "  ${GREEN}up${NC}     - Complete setup (schema + sample data) [DEFAULT]"
        echo "  ${YELLOW}down${NC}   - Complete teardown (remove data + drop schema)"
        echo "  ${YELLOW}reset${NC}  - Fresh start (drop + create + seed)"
        echo "  ${GREEN}seed${NC}   - Add sample data to existing schema"
        echo "  ${YELLOW}unseed${NC} - Remove sample data (keep schema)"
        echo "  ${GREEN}schema${NC} - Create schema only (no sample data)"
        echo "  ${RED}drop${NC}   - Emergency: drop all tables"
        echo ""
        echo -e "${BLUE}Examples:${NC}"
        echo "  ./migrate.sh up      # Complete setup for development"
        echo "  ./migrate.sh reset   # Fresh database reset"
        echo "  ./migrate.sh schema  # Production schema setup"
        echo "  ./migrate.sh seed    # Add sample data later"
        exit 1
        ;;
esac

echo ""
echo -e "${BLUE}🔗 Next steps:${NC}"
echo "   • Start the backend server: go run main.go"
echo "   • Check database with: psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME"
echo "   • View sample users in auth-users.md (if seeded)"