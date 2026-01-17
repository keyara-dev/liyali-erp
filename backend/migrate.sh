#!/bin/bash

# Migration script for Fly.io deployment
# This script runs database migrations before starting the main application

set -e

echo "🚀 Starting database migration process..."

# Check if required environment variables are set
if [ -z "$DB_HOST" ] || [ -z "$DB_NAME" ] || [ -z "$DB_USER" ] || [ -z "$DB_PASSWORD" ]; then
    echo "❌ Missing required database environment variables"
    echo "Required: DB_HOST, DB_NAME, DB_USER, DB_PASSWORD"
    exit 1
fi

# Construct database URL
DB_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT:-5432}/${DB_NAME}?sslmode=${DB_SSL_MODE:-require}"

echo "📊 Database connection details:"
echo "  Host: $DB_HOST"
echo "  Port: ${DB_PORT:-5432}"
echo "  Database: $DB_NAME"
echo "  User: $DB_USER"
echo "  SSL Mode: ${DB_SSL_MODE:-require}"

# Test database connection
echo "🔍 Testing database connection..."
if ! psql "$DB_URL" -c "SELECT 1;" > /dev/null 2>&1; then
    echo "❌ Failed to connect to database"
    exit 1
fi
echo "✅ Database connection successful"

# Check if tables already exist
echo "🔍 Checking if migrations are needed..."
TABLES_EXIST=$(psql "$DB_URL" -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name IN ('users', 'organizations', 'vendors');" 2>/dev/null || echo "0")

if [ "$TABLES_EXIST" -ge 3 ]; then
    echo "✅ Database tables already exist, skipping migration"
    exit 0
fi

echo "📋 Running database migrations..."

# Run the migration SQL file
MIGRATION_FILE="/app/database/migrations/001_init_system.up.sql"

if [ ! -f "$MIGRATION_FILE" ]; then
    echo "❌ Migration file not found: $MIGRATION_FILE"
    exit 1
fi

echo "📄 Applying migration: $MIGRATION_FILE"
if psql "$DB_URL" -f "$MIGRATION_FILE"; then
    echo "✅ Migration completed successfully"
else
    echo "❌ Migration failed"
    exit 1
fi

# Verify migration
echo "🔍 Verifying migration..."
TABLES_AFTER=$(psql "$DB_URL" -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name IN ('users', 'organizations', 'vendors', 'requisitions', 'budgets');" 2>/dev/null || echo "0")

if [ "$TABLES_AFTER" -ge 5 ]; then
    echo "✅ Migration verification successful"
else
    echo "❌ Migration verification failed - expected tables not found"
    exit 1
fi

echo "🎉 Database migration process completed successfully!"