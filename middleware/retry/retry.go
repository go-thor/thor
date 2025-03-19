package retry

import (
	"context"
	"log"
	"time"

	"github.com/go-thor/thor/pkg"
	"github.com/go-thor/thor/pkg/errors"
)

// Options 定义重试中间件的选项
type Options struct {
	// MaxRetries 是最大重试次数
	MaxRetries int
	// RetryInterval 是重试间隔
	RetryInterval time.Duration
	// RetryableErrors 定义哪些错误码应该重试
	RetryableErrors []string
	// OnRetry 在每次重试前调用的回调函数
	OnRetry func(ctx context.Context, attempt int, err error)
}

// Option 是配置重试中间件的函数
type Option func(*Options)

// WithMaxRetries 设置最大重试次数
func WithMaxRetries(max int) Option {
	return func(o *Options) {
		o.MaxRetries = max
	}
}

// WithRetryInterval 设置重试间隔
func WithRetryInterval(interval time.Duration) Option {
	return func(o *Options) {
		o.RetryInterval = interval
	}
}

// WithRetryableErrors 设置可重试的错误码
func WithRetryableErrors(codes ...string) Option {
	return func(o *Options) {
		o.RetryableErrors = codes
	}
}

// WithOnRetry 设置重试回调函数
func WithOnRetry(fn func(ctx context.Context, attempt int, err error)) Option {
	return func(o *Options) {
		o.OnRetry = fn
	}
}

// New 创建一个新的重试中间件
func New(opts ...Option) pkg.Middleware {
	options := &Options{
		MaxRetries:    3,
		RetryInterval: time.Second,
		RetryableErrors: []string{
			errors.ErrorCodeTimeout,
			errors.ErrorCodeDeadlineExceeded,
		},
		OnRetry: func(ctx context.Context, attempt int, err error) {
			log.Printf("重试调用 (尝试 %d): %v", attempt, err)
		},
	}

	for _, opt := range opts {
		opt(options)
	}

	return func(next pkg.Handler) pkg.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			var (
				resp interface{}
				err  error
			)

			// 尝试调用，最多重试指定次数
			for attempt := 0; attempt <= options.MaxRetries; attempt++ {
				// 如果不是第一次尝试，则等待指定的重试间隔
				if attempt > 0 {
					// 在重试前调用回调函数
					if options.OnRetry != nil {
						options.OnRetry(ctx, attempt, err)
					}

					// 等待重试间隔
					select {
					case <-ctx.Done():
						return nil, ctx.Err()
					case <-time.After(options.RetryInterval):
						// 继续重试
					}
				}

				// 执行实际调用
				resp, err = next(ctx, req)
				if err == nil {
					// 调用成功，返回响应
					return resp, nil
				}

				// 检查是否是可重试的错误
				errCode := errors.Code(err)
				shouldRetry := false
				for _, code := range options.RetryableErrors {
					if errCode == code {
						shouldRetry = true
						break
					}
				}

				// 如果不可重试或者已经达到最大重试次数，返回错误
				if !shouldRetry || attempt == options.MaxRetries {
					return nil, err
				}
			}

			// 不应该到达这里，但为了完整性返回最后的错误
			return nil, err
		}
	}
}
