#!/bin/bash

# Setup test database for migration tests
# This script creates a test database and runs the migration tests

set -e

echo "Setting up test database..."

# Database configuration
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-postgres}
DB_NAME="api_aggregator_test"

# Set PostgreSQL password for commands
export PGPASSWORD=$DB_PASSWORD

# Check if PostgreSQL is running
if ! pg_isready -h $DB_HOST -p $DB_PORT -U $DB_USER > /dev/null 2>&1; then
    echo "Error: PostgreSQL is not running on $DB_HOST:$DB_PORT"
    echo "Please start PostgreSQL first:"
    echo "  docker-compose up -d postgres"
    exit 1
fi

# Drop test database if exists
echo "Dropping existing test database (if any)..."
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -c "DROP DATABASE IF EXISTS $DB_NAME;" postgres

# Create test database
echo "Creating test database..."
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -c "CREATE DATABASE $DB_NAME;" postgres

echo "Test database created successfully!"
echo ""
echo "Running migration tests..."
echo ""

# Set test database URL
export TEST_DATABASE_URL="host=$DB_HOST user=$DB_USER password=$DB_PASSWORD dbname=$DB_NAME port=$DB_PORT sslmode=disable"

# Run tests
cd "$(dirname "$0")"
go test -v

echo ""
echo "Tests completed!"
