# PowerShell script to run tests in Docker

Write-Host "Starting PostgreSQL container..." -ForegroundColor Green
docker-compose up -d postgres

Write-Host "Waiting for database to be ready..." -ForegroundColor Yellow
Start-Sleep -Seconds 10

Write-Host "Creating test database..." -ForegroundColor Green
docker exec api-aggregator-postgres psql -U postgres -c "DROP DATABASE IF EXISTS api_aggregator_test;" 2>$null
docker exec api-aggregator-postgres psql -U postgres -c "CREATE DATABASE api_aggregator_test;"

Write-Host "Running tests..." -ForegroundColor Green
docker run --rm `
  --network api_v2_default `
  -v "${PWD}:/app" `
  -w /app `
  -e TEST_DATABASE_URL="host=api-aggregator-postgres user=postgres password=postgres dbname=api_aggregator_test port=5432 sslmode=disable" `
  golang:1.23-alpine `
  sh -c "go mod download && go test -v ./internal/service/..."

Write-Host "Tests completed!" -ForegroundColor Green
