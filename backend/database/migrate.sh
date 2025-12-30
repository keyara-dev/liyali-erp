#!/bin/bash

# Migration script for Liyali Gateway
# Usage: ./migrate.sh [up|down|reset]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
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

echo -e "${YELLOW}Liyali Gateway Database Migration${NC}"
echo "=================================="

case $ACTION in
    "up")
        echo -e "${GREEN}Running UP migration...${NC}"
        go run run_migration.go migrations/001_create_complete_schema.up.sql
        echo -e "${GREEN}✅ Migration completed successfully!${NC}"
        ;;
    "down")
        echo -e "${YELLOW}Running DOWN migration...${NC}"
        echo -e "${RED}⚠️  This will DROP ALL TABLES! Are you sure? (y/N)${NC}"
        read -r response
        if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
            go run run_migration.go migrations/001_create_complete_schema.down.sql
            echo -e "${GREEN}✅ Rollback completed successfully!${NC}"
        else
            echo -e "${YELLOW}Migration cancelled${NC}"
        fi
        ;;
    "reset")
        echo -e "${YELLOW}Resetting database (DOWN + UP)...${NC}"
        echo -e "${RED}⚠️  This will DROP ALL TABLES and recreate them! Are you sure? (y/N)${NC}"
        read -r response
        if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
            echo -e "${YELLOW}Step 1: Running DOWN migration...${NC}"
            go run run_migration.go migrations/001_create_complete_schema.down.sql
            echo -e "${YELLOW}Step 2: Running UP migration...${NC}"
            go run run_migration.go migrations/001_create_complete_schema.up.sql
            echo -e "${GREEN}✅ Database reset completed successfully!${NC}"
        else
            echo -e "${YELLOW}Migration cancelled${NC}"
        fi
        ;;
    "drop")
        echo -e "${YELLOW}Dropping all tables...${NC}"
        echo -e "${RED}⚠️  This will DROP ALL TABLES! Are you sure? (y/N)${NC}"
        read -r response
        if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
            go run run_migration.go migrations/000_drop_all_tables.up.sql
            echo -e "${GREEN}✅ All tables dropped successfully!${NC}"
        else
            echo -e "${YELLOW}Operation cancelled${NC}"
        fi
        ;;
    *)
        echo -e "${RED}Invalid action: $ACTION${NC}"
        echo "Usage: $0 [up|down|reset|drop]"
        echo ""
        echo "Actions:"
        echo "  up    - Run the UP migration (create tables)"
        echo "  down  - Run the DOWN migration (drop tables)"
        echo "  reset - Run DOWN then UP (complete reset)"
        echo "  drop  - Drop all tables using drop script"
        exit 1
        ;;
esac