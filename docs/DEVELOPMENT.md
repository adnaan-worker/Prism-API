# Development Guide

## Prerequisites

- Go 1.23+
- Node.js 20+
- PostgreSQL 15+
- Redis 7+
- Git

## Project Structure

```
api-aggregator/
├── backend/              # Go backend service
│   ├── cmd/             # Application entry points
│   │   └── server/      # Main server
│   ├── config/          # Configuration management
│   ├── internal/        # Private application code
│   │   ├── adapter/     # API adapters (OpenAI, Anthropic, Gemini)
│   │   ├── api/         # HTTP handlers
│   │   ├── loadbalancer/# Load balancing strategies
│   │   ├── middleware/  # HTTP middleware
│   │   ├── models/      # Data models
│   │   ├── repository/  # Data access layer
│   │   └── service/     # Business logic
│   ├── pkg/             # Public libraries
│   │   └── redis/       # Redis client
│   └── scripts/         # Database migrations
├── portal/              # User portal (React)
│   ├── src/
│   │   ├── components/  # Reusable components
│   │   ├── layouts/     # Page layouts
│   │   ├── lib/         # Utilities and API client
│   │   ├── pages/       # Page components
│   │   ├── router/      # Route configuration
│   │   ├── services/    # API services
│   │   └── types/       # TypeScript types
│   └── public/          # Static assets
├── admin/               # Admin dashboard (React)
│   └── src/             # Similar structure to portal
├── docs/                # Documentation
└── docker-compose.yml   # Docker configuration
```

## Backend Development

### Setup

1. Install dependencies:
```bash
cd backend
go mod download
```

2. Set up environment variables:
```bash
cp .env.example .env
# Edit .env with your local settings
```

3. Start PostgreSQL and Redis:
```bash
docker-compose up -d postgres redis
```

4. Run database migrations:
```bash
go run scripts/migrate.go
```

5. Start the server:
```bash
go run cmd/server/main.go
```

The backend will be available at http://localhost:8080

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run specific test
go test -v ./internal/service -run TestAuthService
```

### Code Structure

#### Models
Define data structures in `internal/models/`:
```go
type User struct {
    ID        uint      `gorm:"primaryKey"`
    Username  string    `gorm:"uniqueIndex;not null"`
    Email     string    `gorm:"uniqueIndex;not null"`
    Password  string    `gorm:"not null"`
    Quota     int       `gorm:"default:10000"`
    IsAdmin   bool      `gorm:"default:false"`
    CreatedAt time.Time
}
```

#### Repository
Data access in `internal/repository/`:
```go
type UserRepository interface {
    Create(user *models.User) error
    FindByID(id uint) (*models.User, error)
    FindByUsername(username string) (*models.User, error)
}
```

#### Service
Business logic in `internal/service/`:
```go
type AuthService struct {
    userRepo  repository.UserRepository
    jwtSecret string
}

func (s *AuthService) Register(req RegisterRequest) (*models.User, error) {
    // Business logic here
}
```

#### Handler
HTTP handlers in `internal/api/`:
```go
func (h *AuthHandler) Register(c *gin.Context) {
    var req RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    // Call service
}
```

### Adding New Features

1. **Define Model** in `internal/models/`
2. **Create Repository** interface and implementation in `internal/repository/`
3. **Implement Service** in `internal/service/`
4. **Create Handler** in `internal/api/`
5. **Register Routes** in `cmd/server/main.go`
6. **Write Tests** for each layer

### API Adapters

To add a new AI provider:

1. Create adapter in `internal/adapter/`:
```go
type NewProviderAdapter struct{}

func (a *NewProviderAdapter) ConvertRequest(req *models.UnifiedRequest) (interface{}, error) {
    // Convert to provider format
}

func (a *NewProviderAdapter) ConvertResponse(resp interface{}) (*models.UnifiedResponse, error) {
    // Convert from provider format
}
```

2. Register in adapter factory:
```go
func NewAdapter(apiType string) Adapter {
    switch apiType {
    case "newprovider":
        return &NewProviderAdapter{}
    // ...
    }
}
```

### Load Balancing

Implement new strategy in `internal/loadbalancer/`:
```go
type CustomLoadBalancer struct {
    configs []*models.APIConfig
}

func (lb *CustomLoadBalancer) SelectConfig() (*models.APIConfig, error) {
    // Your selection logic
}
```

## Frontend Development

### Portal Setup

1. Install dependencies:
```bash
cd portal
npm install
```

2. Configure environment:
```bash
cp .env.example .env
# Edit .env
```

3. Start development server:
```bash
npm run dev
```

Portal will be available at http://localhost:3000

### Admin Setup

Same as portal, but in the `admin/` directory. Admin will be available at http://localhost:3002

### Project Structure

```
src/
├── components/       # Reusable UI components
├── layouts/         # Page layouts (DashboardLayout, etc.)
├── lib/             # Utilities
│   ├── api.ts       # Axios instance
│   └── queryClient.ts # TanStack Query client
├── pages/           # Page components
├── router/          # React Router configuration
├── services/        # API service functions
├── styles/          # Global styles
└── types/           # TypeScript type definitions
```

### Adding New Pages

1. **Create Page Component** in `src/pages/`:
```tsx
export default function NewPage() {
  return (
    <div>
      <h1>New Page</h1>
    </div>
  );
}
```

2. **Add Route** in `src/router/index.tsx`:
```tsx
{
  path: '/new-page',
  element: <NewPage />
}
```

3. **Add Navigation** in layout component

### API Integration

Use TanStack Query for data fetching:

```tsx
// Define service function
export const fetchData = async () => {
  const response = await api.get('/endpoint');
  return response.data;
};

// Use in component
const { data, isLoading, error } = useQuery({
  queryKey: ['data'],
  queryFn: fetchData
});
```

For mutations:

```tsx
const mutation = useMutation({
  mutationFn: createData,
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['data'] });
  }
});
```

### Styling

Use Tailwind CSS for styling:

```tsx
<div className="flex items-center justify-between p-4 bg-white rounded-lg shadow">
  <h2 className="text-xl font-semibold">Title</h2>
</div>
```

Use Ant Design components:

```tsx
import { Button, Table, Modal } from 'antd';

<Button type="primary" onClick={handleClick}>
  Click Me
</Button>
```

## Database Migrations

### Creating Migrations

1. Define schema in `backend/scripts/schema.sql`
2. Update `migrate.go` if needed
3. Run migration:
```bash
go run scripts/migrate.go
```

### Migration Best Practices

- Always backup before running migrations
- Test migrations on development database first
- Make migrations reversible when possible
- Document schema changes

## Testing

### Backend Testing

```bash
# Unit tests
go test ./internal/service/...

# Integration tests
go test ./internal/api/...

# Property-based tests
go test ./internal/... -run Property
```

### Frontend Testing

```bash
# Run tests
npm test

# Run with coverage
npm test -- --coverage
```

## Code Quality

### Backend

Use Go tools:

```bash
# Format code
go fmt ./...

# Lint code
golangci-lint run

# Vet code
go vet ./...
```

### Frontend

```bash
# Lint
npm run lint

# Format
npm run format

# Type check
npm run type-check
```

## Debugging

### Backend

Use VS Code debugger or Delve:

```bash
dlv debug cmd/server/main.go
```

### Frontend

Use browser DevTools and React DevTools extension.

## Environment Variables

### Backend

```env
DATABASE_URL=postgres://...
REDIS_URL=redis://...
JWT_SECRET=secret
PORT=8080
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
REDIS_POOL_SIZE=10
```

### Frontend

```env
VITE_API_BASE_URL=http://localhost:8080/api
VITE_APP_NAME=API聚合平台
```

## Common Tasks

### Add New API Endpoint

1. Define request/response types
2. Create repository method
3. Implement service method
4. Create handler
5. Register route
6. Write tests
7. Update API documentation

### Add New Frontend Feature

1. Create service function
2. Create page/component
3. Add route
4. Update navigation
5. Test functionality

## Performance Tips

### Backend

- Use connection pooling
- Implement caching with Redis
- Use database indexes
- Optimize queries with GORM
- Use goroutines for concurrent operations

### Frontend

- Use React.memo for expensive components
- Implement virtual scrolling for large lists
- Use TanStack Query caching
- Lazy load routes and components
- Optimize bundle size

## Troubleshooting

### Backend won't start

- Check PostgreSQL is running
- Check Redis is running
- Verify environment variables
- Check port 8080 is available

### Frontend build fails

- Clear node_modules and reinstall
- Check Node.js version
- Verify environment variables

### Database connection errors

- Verify DATABASE_URL format
- Check PostgreSQL is accessible
- Verify credentials

## Resources

- [Go Documentation](https://golang.org/doc/)
- [Gin Framework](https://gin-gonic.com/)
- [GORM](https://gorm.io/)
- [React Documentation](https://react.dev/)
- [Ant Design](https://ant.design/)
- [TanStack Query](https://tanstack.com/query/)
