# API Documentation

## Base URL

```
http://localhost:8080/api
```

## Authentication

Most endpoints require authentication using JWT tokens or API keys.

### JWT Authentication
Include the JWT token in the Authorization header:
```
Authorization: Bearer <your-jwt-token>
```

### API Key Authentication
Include the API key in the Authorization header:
```
Authorization: Bearer <your-api-key>
```

## Endpoints

### Authentication

#### Register
```http
POST /api/auth/register
Content-Type: application/json

{
  "username": "string",
  "email": "string",
  "password": "string"
}
```

Response:
```json
{
  "id": 1,
  "username": "string",
  "email": "string",
  "quota": 10000,
  "is_admin": false
}
```

#### Login
```http
POST /api/auth/login
Content-Type: application/json

{
  "username": "string",
  "password": "string"
}
```

Response:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "username": "string",
    "email": "string",
    "quota": 10000,
    "is_admin": false
  }
}
```

### User Management

#### Get User Info
```http
GET /api/user/info
Authorization: Bearer <jwt-token>
```

Response:
```json
{
  "id": 1,
  "username": "string",
  "email": "string",
  "quota": 10000,
  "used_quota": 500,
  "is_admin": false
}
```

#### Daily Sign-in
```http
POST /api/user/signin
Authorization: Bearer <jwt-token>
```

Response:
```json
{
  "message": "Sign-in successful",
  "quota_added": 1000,
  "new_quota": 11000
}
```

### API Keys

#### List API Keys
```http
GET /api/apikeys
Authorization: Bearer <jwt-token>
```

Response:
```json
[
  {
    "id": 1,
    "name": "My API Key",
    "key": "sk-xxxxxxxxxxxxxxxx",
    "rate_limit": 60,
    "created_at": "2024-01-01T00:00:00Z"
  }
]
```

#### Create API Key
```http
POST /api/apikeys
Authorization: Bearer <jwt-token>
Content-Type: application/json

{
  "name": "My API Key"
}
```

Response:
```json
{
  "id": 1,
  "name": "My API Key",
  "key": "sk-xxxxxxxxxxxxxxxx",
  "rate_limit": 60,
  "created_at": "2024-01-01T00:00:00Z"
}
```

#### Delete API Key
```http
DELETE /api/apikeys/:id
Authorization: Bearer <jwt-token>
```

### Admin Endpoints

All admin endpoints require JWT authentication with admin privileges.

#### Get Users List
```http
GET /api/admin/users?page=1&page_size=10&search=username
Authorization: Bearer <admin-jwt-token>
```

Response:
```json
{
  "users": [
    {
      "id": 1,
      "username": "string",
      "email": "string",
      "quota": 10000,
      "used_quota": 500,
      "is_admin": false,
      "is_active": true,
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "total": 100,
  "page": 1,
  "page_size": 10
}
```

#### Get User Details
```http
GET /api/admin/users/:id
Authorization: Bearer <admin-jwt-token>
```

#### Update User Status
```http
PUT /api/admin/users/:id/status
Authorization: Bearer <admin-jwt-token>
Content-Type: application/json

{
  "is_active": true
}
```

#### Update User Quota
```http
PUT /api/admin/users/:id/quota
Authorization: Bearer <admin-jwt-token>
Content-Type: application/json

{
  "quota": 20000
}
```

#### API Configuration Management

##### List API Configs
```http
GET /api/admin/api-configs
Authorization: Bearer <admin-jwt-token>
```

##### Create API Config
```http
POST /api/admin/api-configs
Authorization: Bearer <admin-jwt-token>
Content-Type: application/json

{
  "name": "OpenAI Config",
  "type": "openai",
  "api_key": "sk-...",
  "base_url": "https://api.openai.com/v1",
  "models": ["gpt-4", "gpt-3.5-turbo"],
  "weight": 1,
  "is_active": true
}
```

##### Update API Config
```http
PUT /api/admin/api-configs/:id
Authorization: Bearer <admin-jwt-token>
Content-Type: application/json
```

##### Delete API Config
```http
DELETE /api/admin/api-configs/:id
Authorization: Bearer <admin-jwt-token>
```

##### Batch Operations
```http
POST /api/admin/api-configs/batch/delete
POST /api/admin/api-configs/batch/activate
POST /api/admin/api-configs/batch/deactivate
Authorization: Bearer <admin-jwt-token>
Content-Type: application/json

{
  "ids": [1, 2, 3]
}
```

#### Statistics

##### Get Overview Stats
```http
GET /api/admin/stats/overview
Authorization: Bearer <admin-jwt-token>
```

Response:
```json
{
  "total_users": 100,
  "active_users": 80,
  "total_requests": 10000,
  "total_tokens": 5000000,
  "request_trend": [
    {"date": "2024-01-01", "count": 100},
    {"date": "2024-01-02", "count": 150}
  ],
  "model_usage": [
    {"model": "gpt-4", "count": 500},
    {"model": "gpt-3.5-turbo", "count": 300}
  ]
}
```

#### Request Logs

##### Get Logs
```http
GET /api/admin/logs?page=1&page_size=20&user_id=1&model=gpt-4&status=success&start_time=2024-01-01&end_time=2024-01-31
Authorization: Bearer <admin-jwt-token>
```

Response:
```json
{
  "logs": [
    {
      "id": 1,
      "user_id": 1,
      "username": "user1",
      "model": "gpt-4",
      "prompt_tokens": 100,
      "completion_tokens": 50,
      "total_tokens": 150,
      "status": "success",
      "error_message": "",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "total": 1000,
  "page": 1,
  "page_size": 20
}
```

##### Export Logs
```http
GET /api/admin/logs/export?start_time=2024-01-01&end_time=2024-01-31
Authorization: Bearer <admin-jwt-token>
```

Returns CSV file.

## Error Responses

All endpoints return errors in the following format:

```json
{
  "error": "Error message",
  "code": 400001,
  "message": "Detailed error description"
}
```

### Common Error Codes

- `400001` - Invalid request parameters
- `401001` - Unauthorized (missing or invalid token)
- `403001` - Forbidden (insufficient permissions)
- `404001` - Resource not found
- `429001` - Rate limit exceeded
- `500001` - Internal server error

## Rate Limiting

API keys are rate-limited to 60 requests per minute by default. When the rate limit is exceeded, the API returns a 429 status code with error code `429001`.
