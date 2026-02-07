@echo off
echo Starting test database...
docker-compose up -d postgres

echo Waiting for database to be ready...
timeout /t 10 /nobreak > nul

echo Creating test database...
docker exec api-aggregator-postgres psql -U postgres -c "DROP DATABASE IF EXISTS api_aggregator_test;"
docker exec api-aggregator-postgres psql -U postgres -c "CREATE DATABASE api_aggregator_test;"

echo Running tests in Docker...
docker run --rm ^
  --network api_v2_default ^
  -v "%cd%:/app" ^
  -w /app ^
  -e TEST_DATABASE_URL="host=api-aggregator-postgres user=postgres password=postgres dbname=api_aggregator_test port=5432 sslmode=disable" ^
  golang:1.23-alpine ^
  sh -c "go mod download && go test -v ./internal/service/..."

echo Tests completed!
