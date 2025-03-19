package timeout

import (
	"context"
	"time"

	"github.com/go-thor/thor/pkg"
	"github.com/go-thor/thor/pkg/errors"
)

// WithTimeout creates a timeout middleware with the specified duration
func WithTimeout(timeout time.Duration) pkg.Middleware {
	return func(next pkg.Handler) pkg.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// 创建带超时的上下文
			timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			// 用带超时的上下文调用下一个处理器
			resp, err := next(timeoutCtx, req)
			if err != nil {
				// 检查是否是超时错误
				if timeoutCtx.Err() == context.DeadlineExceeded {
					return nil, errors.Newf(errors.ErrorCodeDeadlineExceeded, "RPC调用超时，超时设置: %v", timeout)
				}
			}
			return resp, err
		}
	}
}
