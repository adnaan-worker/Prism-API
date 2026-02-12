# Database Migration Scripts

This directory contains database migration and management scripts for the Prism API project.

## Quick Start

<<<<<<< Updated upstream
### 1. Run Database Migration
=======
- `migrate.go` - Go-based migration script using GORM AutoMigrate
- `init_admin.go` - Script to create or update admin user
- `schema.sql` - SQL schema file for manual execution or reference

## Running Migrations

### Using Go Script (Recommended)
>>>>>>> Stashed changes

```bash
# Set database connection (optional, defaults to localhost)
export DATABASE_URL="host=localhost user=postgres password=postgres dbname=api_aggregator port=5432 sslmode=disable"

# Run migration
cd backend/scripts
go run migrate_unified.go
```

### 2. Create Admin User

```bash
# Create admin user with default credentials
go run create_admin.go

# Default credentials:
# Username: admin
# Email: admin@example.com
# Password: admin123
```

## Available Scripts

### Migration Scripts

- **`migrate_unified.go`** - Main migration script that creates all tables and indexes
  - Creates all core tables (users, api_keys, api_configs, etc.)
  - Creates account pool tables (account_pools, account_credentials, account_pool_credentials)
  - Creates Kiro model mapping table
  - Initializes Kiro model mappings
  - Creates performance indexes

- **`migrate.go`** - Legacy migration script (deprecated, use migrate_unified.go instead)

### Management Scripts

- **`create_admin.go`** - Creates an admin user
  - Username: admin
  - Email: admin@example.com  
  - Password: admin123
  - Quota: 1,000,000 credits

- **`verify_models.go`** - Verifies Kiro model mappings in database

### Test Scripts

- **`migrate_test.go`** - Unit tests for migration functions
- **`setup_test_db.sh`** - Shell script to setup test database (Linux/Mac)
- **`setup_test_db.bat`** - Batch script to setup test database (Windows)

## Initializing Admin User

### Using Go Script (Recommended)

```bash
# Make sure DATABASE_URL is set (or it will use .env file)
cd backend/scripts
go run init_admin.go
```

The script will:
- Read admin credentials from `.env` file or environment variables
- Create admin user if it doesn't exist
- Update password if admin user already exists (with confirmation)

### Custom Admin Credentials

You can override the default credentials using environment variables:

```bash
export ADMIN_USERNAME=myadmin
export ADMIN_EMAIL=myadmin@example.com
export ADMIN_PASSWORD=mypassword123
cd backend/scripts
go run init_admin.go
```

### Default Credentials

If not specified, the script uses:
- Username: `admin`
- Email: `admin@example.com`
- Password: `admin123`

## Database Schema

### Core Tables

1. **users** - User accounts
2. **api_keys** - API keys for authentication
3. **api_configs** - API configuration (endpoints, models, etc.)
4. **request_logs** - Request history and logs
5. **sign_in_records** - Daily sign-in records

### Load Balancing & Pricing

6. **load_balancer_configs** - Load balancing strategies per model
7. **pricings** - Pricing configuration per model and API config

### Billing

8. **billing_transactions** - All billing operations (charges, refunds, etc.)

### Caching

9. **request_caches** - Request/response cache with semantic matching

### Account Pool System

10. **account_pools** - Account pool definitions
11. **account_credentials** - Individual account credentials
12. **account_pool_credentials** - Many-to-many relationship between pools and credentials
13. **kiro_model_mappings** - Model name to Kiro model ID mappings

## Environment Variables

- `DATABASE_URL` - PostgreSQL connection string
  - Default: `host=localhost user=postgres password=postgres dbname=api_aggregator port=5432 sslmode=disable`

## Migration Process

The unified migration script (`migrate_unified.go`) performs the following steps:

1. **Connect to Database** - Establishes connection using DATABASE_URL
2. **Auto-Migrate Models** - Uses GORM AutoMigrate for all model structs
3. **Create Association Tables** - Creates many-to-many relationship tables
4. **Create Custom Tables** - Creates Kiro model mapping table
5. **Initialize Data** - Populates Kiro model mappings
6. **Create Indexes** - Creates performance indexes

## Notes

- All timestamps use `TIMESTAMP` type with timezone support
- JSON fields use PostgreSQL `jsonb` type for better performance
- Foreign keys are properly defined with CASCADE delete where appropriate
- Indexes are created for frequently queried columns

## Troubleshooting

### Connection Failed

```bash
# Check PostgreSQL is running
pg_isready -h localhost -p 5432

# Test connection
psql -h localhost -U postgres -d api_aggregator
```

### Migration Failed

```bash
# Drop and recreate database (WARNING: deletes all data)
psql -h localhost -U postgres -c "DROP DATABASE IF EXISTS api_aggregator"
psql -h localhost -U postgres -c "CREATE DATABASE api_aggregator"

# Run migration again
go run migrate_unified.go
```

### Check Migration Status

```bash
# Connect to database
psql -h localhost -U postgres -d api_aggregator

# List all tables
\dt

# Check specific table structure
\d users
\d account_pools
\d kiro_model_mappings
```
