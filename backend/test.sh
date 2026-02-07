#!/bin/bash

# Start postgres if not running
docker-compose up -d postgres

# Wait for postgres to be ready
echo "Waiting for database to be ready..."
sleep 5

# Create test database
docker exec api-aggregator-postgres psql -U postgres -c "DROP DATABASE IF EXISTS api_aggregator_test;" 2>/dev/null || true
docker exec api-aggregator-postgres psql -U postgres -c "CREATE DATABASE api_aggregator_test;"

# Run tests
docker run --rm \
  --network api_v2_default \
  -v "$(pwd):/app" \
  -w /app \
  -e TEST_DATABASE_URL="host=api-aggregator-postgres user=postgres password=postgres dbname=api_aggregator_test port=5432 sslmode=disable" \
  golang:1.23-alpine \
  sh -c "go mod download && go test -v ./internal/service/..."

echo "Tests completed!"
