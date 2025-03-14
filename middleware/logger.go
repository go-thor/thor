package middleware

import (
	"fmt"
	"time"

	"github.com/go-thor/thor"
)

func Logger() thor.Middleware {
	return func(next thor.HandlerFunc) thor.HandlerFunc {
		return func(ctx *thor.RPCContext) error {
			start := time.Now()
			fmt.Printf("Server: Received Request: %s.%s, params: %v\n", ctx.Request.ServiceName, ctx.Request.MethodName, ctx.Request.Params)
			err := next(ctx)
			fmt.Printf("Server: Send response: %s.%s, duration: %v, error: %v\n", ctx.Request.ServiceName, ctx.Request.MethodName, time.Since(start), err)
			return err
		}
	}
}

// ClientLoggerMiddleware 客户端日志中间件
func ClientLoggerMiddleware() thor.Middleware {
	return func(next thor.HandlerFunc) thor.HandlerFunc {
		return func(ctx *thor.RPCContext) error {
			fmt.Printf("Client: Sending request: %s.%s, params: %v\n", ctx.Request.ServiceName, ctx.Request.MethodName, ctx.Request.Params)
			start := time.Now()
			err := next(ctx)
			fmt.Printf("Client: Received response for %s.%s, duration: %v, response: %v, error: %v\n",
				ctx.Request.ServiceName, ctx.Request.MethodName, time.Since(start), ctx.Response, err)
			return err
		}
	}
}
