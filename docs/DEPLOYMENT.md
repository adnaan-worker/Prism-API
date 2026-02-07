# Deployment Guide

## Prerequisites

- Docker 20.10+
- Docker Compose 2.0+
- (Optional) Domain name for production deployment

## Quick Start with Docker Compose

### 1. Clone the Repository

```bash
git clone <repository-url>
cd api-aggregator
```

### 2. Configure Environment Variables

Copy the example environment file and update it with your settings:

```bash
cp .env.example .env
```

Edit `.env` and set the following variables:

```env
# IMPORTANT: Change this in production!
JWT_SECRET=your-secure-random-secret-key

# Database (default values work for Docker Compose)
DATABASE_URL=postgres://postgres:postgres@postgres:5432/api_aggregator?sslmode=disable

# Redis (default values work for Docker Compose)
REDIS_URL=redis://redis:6379

# Server
PORT=8080

# Optional: Create initial admin user
ADMIN_USERNAME=admin
ADMIN_EMAIL=admin@example.com
ADMIN_PASSWORD=change-this-password
```

### 3. Start All Services

```bash
docker-compose up -d
```

This will start:
- PostgreSQL database (port 5432)
- Redis cache (port 6379)
- Backend API (port 8080)
- User Portal (port 3000)
- Admin Dashboard (port 3001)
- Nginx reverse proxy (port 80)

### 4. Verify Services

Check that all services are running:

```bash
docker-compose ps
```

Test the backend health endpoint:

```bash
curl http://localhost:8080/health
```

### 5. Access the Applications

- **User Portal**: http://localhost:3000
- **Admin Dashboard**: http://localhost:3001
- **Backend API**: http://localhost:8080/api

## Production Deployment

### 1. Security Configuration

#### Generate Secure JWT Secret

```bash
openssl rand -base64 32
```

Update your `.env` file with the generated secret.

#### Update Admin Credentials

Change the default admin credentials in `.env`:

```env
ADMIN_USERNAME=your-admin-username
ADMIN_EMAIL=your-admin@example.com
ADMIN_PASSWORD=your-secure-password
```

### 2. Domain Configuration

Update `nginx.conf` to use your domain names:

```nginx
server {
    listen 80;
    server_name yourdomain.com;
    # ... rest of configuration
}

server {
    listen 80;
    server_name admin.yourdomain.com;
    # ... rest of configuration
}
```

### 3. SSL/TLS Configuration

For production, you should use HTTPS. Here's how to set up SSL with Let's Encrypt:

#### Install Certbot

```bash
sudo apt-get update
sudo apt-get install certbot python3-certbot-nginx
```

#### Obtain SSL Certificates

```bash
sudo certbot --nginx -d yourdomain.com -d admin.yourdomain.com
```

#### Update docker-compose.yml

Add SSL certificate volumes to the nginx service:

```yaml
nginx:
  image: nginx:alpine
  volumes:
    - ./nginx.conf:/etc/nginx/nginx.conf:ro
    - /etc/letsencrypt:/etc/letsencrypt:ro
```

### 4. Database Backup

Set up automated database backups:

```bash
# Create backup script
cat > backup.sh << 'EOF'
#!/bin/bash
BACKUP_DIR="/backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
docker exec api-aggregator-postgres pg_dump -U postgres api_aggregator > "$BACKUP_DIR/backup_$TIMESTAMP.sql"
# Keep only last 7 days of backups
find $BACKUP_DIR -name "backup_*.sql" -mtime +7 -delete
EOF

chmod +x backup.sh

# Add to crontab (daily at 2 AM)
(crontab -l 2>/dev/null; echo "0 2 * * * /path/to/backup.sh") | crontab -
```

### 5. Monitoring

#### View Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f backend
docker-compose logs -f postgres
docker-compose logs -f redis
```

#### Resource Usage

```bash
docker stats
```

### 6. Performance Tuning

#### Database Connection Pool

Adjust in `.env`:

```env
DB_MAX_OPEN_CONNS=50
DB_MAX_IDLE_CONNS=10
DB_CONN_MAX_LIFETIME=10m
```

#### Redis Connection Pool

```env
REDIS_POOL_SIZE=20
REDIS_MIN_IDLE_CONN=5
```

#### Server Timeouts

```env
SERVER_READ_TIMEOUT=15s
SERVER_WRITE_TIMEOUT=15s
REQUEST_TIMEOUT=60s
```

## Scaling

### Horizontal Scaling

To scale the backend service:

```bash
docker-compose up -d --scale backend=3
```

Update nginx.conf to load balance across multiple backend instances:

```nginx
upstream backend {
    server backend:8080;
    server backend:8080;
    server backend:8080;
}
```

### Database Scaling

For production workloads, consider:
- Using managed PostgreSQL (AWS RDS, Google Cloud SQL, etc.)
- Setting up read replicas
- Implementing connection pooling with PgBouncer

### Redis Scaling

For high-traffic scenarios:
- Use Redis Cluster for horizontal scaling
- Use managed Redis (AWS ElastiCache, Redis Cloud, etc.)
- Implement Redis Sentinel for high availability

## Maintenance

### Update Services

```bash
# Pull latest images
docker-compose pull

# Restart services
docker-compose up -d
```

### Database Migration

```bash
# Run migrations
docker exec api-aggregator-backend ./server migrate
```

### Clear Redis Cache

```bash
docker exec api-aggregator-redis redis-cli FLUSHALL
```

## Troubleshooting

### Backend Can't Connect to Database

Check database logs:
```bash
docker-compose logs postgres
```

Verify connection string in `.env`.

### High Memory Usage

Check container stats:
```bash
docker stats
```

Adjust resource limits in `docker-compose.yml`:

```yaml
backend:
  deploy:
    resources:
      limits:
        memory: 512M
      reservations:
        memory: 256M
```

### Rate Limiting Issues

Check Redis connection:
```bash
docker exec api-aggregator-redis redis-cli ping
```

### Nginx 502 Bad Gateway

Check if backend is running:
```bash
docker-compose ps backend
curl http://localhost:8080/health
```

## Health Checks

All services include health checks. View health status:

```bash
docker-compose ps
```

Healthy services will show "healthy" in the status column.

## Rollback

If something goes wrong:

```bash
# Stop all services
docker-compose down

# Restore database from backup
docker exec -i api-aggregator-postgres psql -U postgres api_aggregator < /backups/backup_YYYYMMDD_HHMMSS.sql

# Start services
docker-compose up -d
```

## Support

For issues and questions:
- Check logs: `docker-compose logs`
- Review documentation in `/docs`
- Check GitHub issues
