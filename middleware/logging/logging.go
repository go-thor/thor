package logging

import (
	"context"
	"log"
	"time"

	"github.com/go-thor/thor"
)

// Middleware struct for logging middleware
type Middleware struct{}

// New creates a new logging middleware
func New() thor.Middleware {
	return func(ctx context.Context, req *thor.Request, next func(context.Context, *thor.Request) (*thor.Response, error)) (*thor.Response, error) {
		start := time.Now()
		log.Printf("Request: %s start", req.ServiceMethod)

		resp, err := next(ctx, req)

		if err != nil {
			log.Printf("Request: %s error: %v, took: %v", req.ServiceMethod, err, time.Since(start))
			return nil, err
		}

		log.Printf("Request: %s completed, took: %v", req.ServiceMethod, time.Since(start))
		return resp, nil
	}
}
