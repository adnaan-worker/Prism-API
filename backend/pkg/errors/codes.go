package errors

// 预定义错误
var (
	// 通用错误 (400xxx)
	ErrInvalidParam     = New(400001, "Invalid parameter")
	ErrInvalidRequest   = New(400002, "Invalid request")
	ErrValidationFailed = New(400003, "Validation failed")

	// 认证错误 (401xxx)
	ErrUnauthorized     = New(401001, "Unauthorized")
	ErrInvalidToken     = New(401002, "Invalid token")
	ErrTokenExpired     = New(401003, "Token expired")
	ErrInvalidPassword  = New(401004, "Invalid password")

	// 权限错误 (403xxx)
	ErrForbidden        = New(403001, "Forbidden")
	ErrInsufficientPerm = New(403002, "Insufficient permissions")

	// 资源错误 (404xxx)
	ErrNotFound         = New(404001, "Resource not found")
	ErrUserNotFound     = New(404002, "User not found")
	ErrAPIKeyNotFound   = New(404003, "API key not found")
	ErrAPIConfigNotFound = New(404004, "API config not found")
	ErrModelNotFound    = New(404005, "Model not found")

	// 冲突错误 (409xxx)
	ErrConflict         = New(409001, "Resource conflict")
	ErrUserExists       = New(409002, "User already exists")
	ErrAPIKeyExists     = New(409003, "API key already exists")
	ErrEmailExists      = New(409004, "Email already exists")
	ErrUsernameExists   = New(409005, "Username already exists")

	// 配额错误 (429xxx)
	ErrQuotaExceeded    = New(429001, "Quota exceeded")
	ErrRateLimitExceeded = New(429002, "Rate limit exceeded")

	// 服务器错误 (500xxx)
	ErrInternal         = New(500001, "Internal server error")
	ErrDatabase         = New(500002, "Database error")
	ErrCache            = New(500003, "Cache error")
	ErrExternal         = New(500004, "External service error")
	ErrEncryption       = New(500005, "Encryption error")
)
