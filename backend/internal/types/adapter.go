package types

import (
	"context"
	"net/http"
)

// Adapter 定义适配器接口（独立包避免循环依赖）
type Adapter interface {
	Call(ctx context.Context, req interface{}) (interface{}, error)
	CallStream(ctx context.Context, req interface{}) (*http.Response, error)
}
