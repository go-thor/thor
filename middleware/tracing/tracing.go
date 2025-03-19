package tracing

import (
	"context"

	"github.com/go-thor/thor/pkg"
)

// Option is a middleware option
type Option func(*options)

type options struct {
	tracer Tracer
}

// Tracer defines the interface for a tracer
type Tracer interface {
	// StartSpan starts a new span
	StartSpan(ctx context.Context, operation string) (context.Context, Span)
}

// Span defines the interface for a span
type Span interface {
	// Finish finishes the span
	Finish()
	// SetTag sets a tag on the span
	SetTag(key string, value interface{})
	// LogFields logs fields on the span
	LogFields(fields ...Field)
}

// Field defines a log field
type Field struct {
	Key   string
	Value interface{}
}

// NoopTracer is a no-op tracer
type NoopTracer struct{}

// StartSpan starts a new span
func (t *NoopTracer) StartSpan(ctx context.Context, operation string) (context.Context, Span) {
	return ctx, &NoopSpan{}
}

// NoopSpan is a no-op span
type NoopSpan struct{}

// Finish finishes the span
func (s *NoopSpan) Finish() {}

// SetTag sets a tag on the span
func (s *NoopSpan) SetTag(key string, value interface{}) {}

// LogFields logs fields on the span
func (s *NoopSpan) LogFields(fields ...Field) {}

// WithTracer sets the tracer for the middleware
func WithTracer(tracer Tracer) Option {
	return func(o *options) {
		o.tracer = tracer
	}
}

// New creates a new tracing middleware
func New(opts ...Option) pkg.Middleware {
	options := &options{
		tracer: &NoopTracer{},
	}

	for _, opt := range opts {
		opt(options)
	}

	return func(next pkg.Handler) pkg.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// Get the service method from the context
			serviceMethod, _ := ctx.Value("service_method").(string)

			// Start a new span
			spanCtx, span := options.tracer.StartSpan(ctx, serviceMethod)
			defer span.Finish()

			// Set span tags
			span.SetTag("service_method", serviceMethod)

			// Call the next handler with the span context
			resp, err := next(spanCtx, req)

			// Set error tag if there's an error
			if err != nil {
				span.SetTag("error", true)
				span.LogFields(Field{Key: "error.message", Value: err.Error()})
			}

			return resp, err
		}
	}
}
