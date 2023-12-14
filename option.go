package thor

import (
	"github.com/go-thor/thor/logger"
	"github.com/go-thor/thor/server"
)

type (
	// Options new server options
	Options struct {
		startupTimeout  int
		shutdownTimeout int
		log             logger.Logger
		servers         []server.Server
	}

	// Option setter
	Option func(ops *Options)
)

// WithLog with service id.
func WithLog(l logger.Logger) Option {
	return func(o *Options) { o.log = l }
}

// WithStartupTimeout app startup timeout
func WithStartupTimeout(timeout int) Option {
	return func(ops *Options) { ops.startupTimeout = timeout }
}

// WithShutdownTimeout app shutdown timeout
func WithShutdownTimeout(timeout int) Option {
	return func(ops *Options) { ops.shutdownTimeout = timeout }
}

// WithServer set servers
func WithServer(boxes ...server.Server) Option {
	return func(ops *Options) { ops.servers = boxes }
}
