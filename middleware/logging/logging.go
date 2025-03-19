package logging

import (
	"context"
	"log"
	"time"

	"github.com/go-thor/thor/pkg"
)

// Middleware is a logging middleware
type Middleware struct{}

// New creates a new logging middleware
func New() pkg.Middleware {
	return func(next pkg.Handler) pkg.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			start := time.Now()

			// Call the next handler
			resp, err := next(ctx, req)

			// Log the request
			if err != nil {
				log.Printf("[ERROR] request=%v error=%v time=%v", req, err, time.Since(start))
			} else {
				log.Printf("[INFO] request=%v response=%v time=%v", req, resp, time.Since(start))
			}

			return resp, err
		}
	}
}
