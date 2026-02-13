package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail 错误详情
type ErrorDetail struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// SuccessWithMessage 带消息的成功响应
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: message,
		Data:    data,
	})
}

// Created 创建成功响应
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, data)
}

// BadRequest 400 错误
func BadRequest(c *gin.Context, message string, details ...string) {
	detail := ""
	if len(details) > 0 {
		detail = details[0]
	}
	c.JSON(http.StatusBadRequest, ErrorResponse{
		Error: ErrorDetail{
			Code:    400001,
			Message: message,
			Details: detail,
		},
	})
}

// Unauthorized 401 错误
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, ErrorResponse{
		Error: ErrorDetail{
			Code:    401001,
			Message: message,
		},
	})
}

// Forbidden 403 错误
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, ErrorResponse{
		Error: ErrorDetail{
			Code:    403001,
			Message: message,
		},
	})
}

// NotFound 404 错误
func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, ErrorResponse{
		Error: ErrorDetail{
			Code:    404001,
			Message: message,
		},
	})
}

// Conflict 409 错误
func Conflict(c *gin.Context, message string, details ...string) {
	detail := ""
	if len(details) > 0 {
		detail = details[0]
	}
	c.JSON(http.StatusConflict, ErrorResponse{
		Error: ErrorDetail{
			Code:    409001,
			Message: message,
			Details: detail,
		},
	})
}

// InternalError 500 错误
func InternalError(c *gin.Context, err interface{}) {
	var message string
	switch v := err.(type) {
	case error:
		message = v.Error()
	case string:
		message = v
	default:
		message = "internal server error"
	}
	
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Error: ErrorDetail{
			Code:    500001,
			Message: message,
		},
	})
}

// InternalErrorWithMessage 500 错误（自定义消息）
func InternalErrorWithMessage(c *gin.Context, message string, err error) {
	details := ""
	if err != nil {
		details = err.Error()
	}
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Error: ErrorDetail{
			Code:    500001,
			Message: message,
			Details: details,
		},
	})
}

// InternalErrorWithDetails 500 错误（带详情）
func InternalErrorWithDetails(c *gin.Context, message string, err error) {
	details := ""
	if err != nil {
		details = err.Error()
	}
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Error: ErrorDetail{
			Code:    500001,
			Message: message,
			Details: details,
		},
	})
}

// TooManyRequests 429 错误
func TooManyRequests(c *gin.Context, message string) {
	c.JSON(http.StatusTooManyRequests, ErrorResponse{
		Error: ErrorDetail{
			Code:    429001,
			Message: message,
		},
	})
}

// RequestTimeout 408 错误
func RequestTimeout(c *gin.Context, message string) {
	c.JSON(http.StatusRequestTimeout, ErrorResponse{
		Error: ErrorDetail{
			Code:    408001,
			Message: message,
		},
	})
}

// ValidationError 验证错误
func ValidationError(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, ErrorResponse{
		Error: ErrorDetail{
			Code:    400002,
			Message: "Validation error",
			Details: err.Error(),
		},
	})
}

// HandleError 统一错误处理
func HandleError(c *gin.Context, err error) {
	// 这里可以根据错误类型进行不同的处理
	// 暂时统一返回500错误
	InternalError(c, err.Error())
}

// Error 通用错误响应
func Error(c *gin.Context, statusCode int, code int, message string, err error) {
	details := ""
	if err != nil {
		details = err.Error()
		// 添加错误到 gin.Context，以便日志中间件记录
		_ = c.Error(err)
	}
	c.JSON(statusCode, ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

// ErrorFromError 从自定义错误创建响应
func ErrorFromError(c *gin.Context, err error) {
	// 添加错误到 gin.Context
	_ = c.Error(err)
	
	// 尝试转换为 AppError
	type appError interface {
		Error() string
		GetCode() int
		GetMessage() string
	}
	
	if appErr, ok := err.(appError); ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: ErrorDetail{
				Code:    appErr.GetCode(),
				Message: appErr.GetMessage(),
				Details: appErr.Error(),
			},
		})
		return
	}
	
	// 默认处理
	InternalError(c, err)
}
