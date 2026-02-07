# Database Migration Scripts

This directory contains database migration scripts for the API Aggregator platform.

## Files

- `migrate.go` - Go-based migration script using GORM AutoMigrate
- `schema.sql` - SQL schema file for manual execution or reference

## Running Migrations

### Using Go Script (Recommended)

```bash
# Set database connection
export DATABASE_URL="host=localhost user=postgres password=postgres dbname=api_aggregator port=5432 sslmode=disable"

# Run migration
cd backend/scripts
go run migrate.go
```

### Using SQL File

```bash
# Connect to PostgreSQL and run the schema
psql -U postgres -d api_aggregator -f schema.sql
```

### Using Docker Compose

```bash
# From project root
docker-compose up -d postgres
docker-compose exec backend go run scripts/migrate.go
```

## Database Schema

The migration creates the following tables:

1. **users** - User accounts with authentication and quota management
2. **api_keys** - API keys for user authentication and rate limiting
3. **api_configs** - Third-party API provider configurations
4. **load_balancer_configs** - Load balancing strategy configurations per model
5. **request_logs** - API request logs for analytics and monitoring
6. **sign_in_records** - Daily sign-in records for quota rewards

## Indexes

The migration creates optimized indexes for:
- User lookups by username, email, status
- API key validation and user association
- API config filtering by type and priority
- Request log analytics queries
- Sign-in record daily checks

## Testing Migrations

### Automated Test Setup

The migration tests verify table creation, constraints, indexes, and data integrity.

**Linux/Mac:**
```bash
cd backend/scripts
chmod +x setup_test_db.sh
./setup_test_db.sh
```

**Windows:**
```cmd
cd backend\scripts
setup_test_db.bat
```

### Manual Test Setup

```bash
# Create test database
createdb -U postgres api_aggregator_test

# Set environment variable
export TEST_DATABASE_URL="host=localhost user=postgres password=postgres dbname=api_aggregator_test port=5432 sslmode=disable"

# Run tests
cd backend/scripts
go test -v
```

### Test Coverage

The migration tests verify:
- All tables are created successfully
- Unique constraints on username, email, and API keys
- Foreign key constraints between tables
- Default values for all fields
- JSONB storage for models and headers
- Soft delete functionality
- Index creation and effectiveness

## Notes

- All tables use soft deletes (deleted_at column)
- JSONB columns are used for flexible data storage (models, headers)
- Foreign key constraints ensure referential integrity
- Indexes are optimized for common query patterns
