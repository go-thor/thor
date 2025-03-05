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
			req, _ := ctx.Request.(*thor.Request)
			fmt.Printf("Server: Received Request: %s.%s, params: %v\n", req.ServiceName, req.MethodName, req.Params)
			err := next(ctx)
			fmt.Printf("Server: Send response: %s.%s, duration: %v, error: %v\n", req.ServiceName, req.MethodName, time.Since(start), err)
			return err
		}
	}
}

// ClientLoggerMiddleware 客户端日志中间件
func ClientLoggerMiddleware() thor.Middleware {
	return func(next thor.HandlerFunc) thor.HandlerFunc {
		return func(ctx *thor.RPCContext) error {
			req, _ := ctx.Request.(*thor.Request)
			fmt.Printf("Client: Sending request: %s.%s, params: %v\n", req.ServiceName, req.MethodName, req.Params)
			start := time.Now()
			err := next(ctx)
			fmt.Printf("Client: Received response for %s.%s, duration: %v, response: %v, error: %v\n",
				req.ServiceName, req.MethodName, time.Since(start), ctx.Response, err)
			return err
		}
	}
}
