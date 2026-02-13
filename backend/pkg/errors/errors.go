package errors

import (
	"fmt"
)

// AppError 应用错误
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
	Err     error  `json:"-"`
}

// Error 实现 error 接口
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap 实现 errors.Unwrap
func (e *AppError) Unwrap() error {
	return e.Err
}

// GetCode 获取错误码
func (e *AppError) GetCode() int {
	return e.Code
}

// GetMessage 获取错误消息
func (e *AppError) GetMessage() string {
	return e.Message
}

// WithDetails 添加详情
func (e *AppError) WithDetails(details string) *AppError {
	return &AppError{
		Code:    e.Code,
		Message: e.Message,
		Details: details,
		Err:     e.Err,
	}
}

// New 创建新错误
func New(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// Wrap 包装错误（支持2个或3个参数）
func Wrap(err error, args ...interface{}) *AppError {
	if len(args) == 1 {
		// 只有message参数
		message, ok := args[0].(string)
		if !ok {
			message = "unknown error"
		}
		return &AppError{
			Code:    500001,
			Message: message,
			Err:     err,
		}
	} else if len(args) >= 2 {
		// 有code和message参数
		code, ok := args[0].(int)
		if !ok {
			code = 500001
		}
		message, ok := args[1].(string)
		if !ok {
			message = "unknown error"
		}
		return &AppError{
			Code:    code,
			Message: message,
			Err:     err,
		}
	}
	
	// 默认情况
	return &AppError{
		Code:    500001,
		Message: err.Error(),
		Err:     err,
	}
}

// NewValidationError 创建验证错误
func NewValidationError(message string, details map[string]string) *AppError {
	detailsStr := ""
	for k, v := range details {
		if detailsStr != "" {
			detailsStr += "; "
		}
		detailsStr += fmt.Sprintf("%s: %s", k, v)
	}
	return &AppError{
		Code:    400002,
		Message: message,
		Details: detailsStr,
	}
}

// NewNotFoundError 创建未找到错误
func NewNotFoundError(message string) *AppError {
	return &AppError{
		Code:    404001,
		Message: message,
	}
}

// NewConflictError 创建冲突错误
func NewConflictError(message string) *AppError {
	return &AppError{
		Code:    409001,
		Message: message,
	}
}

// Is 判断错误类型
func Is(err error, target *AppError) bool {
	if err == nil || target == nil {
		return false
	}
	
	appErr, ok := err.(*AppError)
	if !ok {
		return false
	}
	
	return appErr.Code == target.Code
}
