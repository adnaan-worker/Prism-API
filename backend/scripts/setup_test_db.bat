@echo off
REM Setup test database for migration tests (Windows)
REM This script creates a test database and runs the migration tests

setlocal

echo Setting up test database...

REM Database configuration
if "%DB_HOST%"=="" set DB_HOST=localhost
if "%DB_PORT%"=="" set DB_PORT=5432
if "%DB_USER%"=="" set DB_USER=postgres
if "%DB_PASSWORD%"=="" set DB_PASSWORD=postgres
set DB_NAME=api_aggregator_test

REM Set PostgreSQL password
set PGPASSWORD=%DB_PASSWORD%

REM Check if PostgreSQL is running
pg_isready -h %DB_HOST% -p %DB_PORT% -U %DB_USER% >nul 2>&1
if errorlevel 1 (
    echo Error: PostgreSQL is not running on %DB_HOST%:%DB_PORT%
    echo Please start PostgreSQL first:
    echo   docker-compose up -d postgres
    exit /b 1
)

REM Drop test database if exists
echo Dropping existing test database (if any)...
psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -c "DROP DATABASE IF EXISTS %DB_NAME%;" postgres

REM Create test database
echo Creating test database...
psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -c "CREATE DATABASE %DB_NAME%;" postgres

echo Test database created successfully!
echo.
echo Running migration tests...
echo.

REM Set test database URL
set TEST_DATABASE_URL=host=%DB_HOST% user=%DB_USER% password=%DB_PASSWORD% dbname=%DB_NAME% port=%DB_PORT% sslmode=disable

REM Run tests
cd /d "%~dp0"
go test -v

echo.
echo Tests completed!

endlocal
