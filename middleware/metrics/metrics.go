package metrics

import (
	"context"
	"sync"
	"time"

	"github.com/go-thor/thor/pkg"
)

// Option is a middleware option
type Option func(*options)

type options struct {
	reporter Reporter
}

// Reporter defines the interface for a metrics reporter
type Reporter interface {
	// ReportLatency reports the latency of a request
	ReportLatency(service, method string, latency time.Duration)
	// ReportRequest reports a request
	ReportRequest(service, method string)
	// ReportError reports an error
	ReportError(service, method string, err error)
}

// DefaultReporter is the default reporter
type DefaultReporter struct {
	mu sync.RWMutex

	// Metrics
	requestCount map[string]int64
	errorCount   map[string]int64
	latencies    map[string][]time.Duration
}

// NewDefaultReporter creates a new default reporter
func NewDefaultReporter() *DefaultReporter {
	return &DefaultReporter{
		requestCount: make(map[string]int64),
		errorCount:   make(map[string]int64),
		latencies:    make(map[string][]time.Duration),
	}
}

// ReportLatency reports the latency of a request
func (r *DefaultReporter) ReportLatency(service, method string, latency time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := service + "." + method
	r.latencies[key] = append(r.latencies[key], latency)
}

// ReportRequest reports a request
func (r *DefaultReporter) ReportRequest(service, method string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := service + "." + method
	r.requestCount[key]++
}

// ReportError reports an error
func (r *DefaultReporter) ReportError(service, method string, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := service + "." + method
	r.errorCount[key]++
}

// GetRequestCount gets the request count for a service method
func (r *DefaultReporter) GetRequestCount(service, method string) int64 {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := service + "." + method
	return r.requestCount[key]
}

// GetErrorCount gets the error count for a service method
func (r *DefaultReporter) GetErrorCount(service, method string) int64 {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := service + "." + method
	return r.errorCount[key]
}

// GetAverageLatency gets the average latency for a service method
func (r *DefaultReporter) GetAverageLatency(service, method string) time.Duration {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := service + "." + method
	latencies := r.latencies[key]
	if len(latencies) == 0 {
		return 0
	}

	var sum time.Duration
	for _, latency := range latencies {
		sum += latency
	}

	return sum / time.Duration(len(latencies))
}

// WithReporter sets the reporter for the middleware
func WithReporter(reporter Reporter) Option {
	return func(o *options) {
		o.reporter = reporter
	}
}

// New creates a new metrics middleware
func New(opts ...Option) pkg.Middleware {
	options := &options{
		reporter: NewDefaultReporter(),
	}

	for _, opt := range opts {
		opt(options)
	}

	return func(next pkg.Handler) pkg.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// Get the service and method from the context
			serviceMethod, _ := ctx.Value("service_method").(string)
			service, method := splitServiceMethod(serviceMethod)

			// Report the request
			options.reporter.ReportRequest(service, method)

			// Start the timer
			start := time.Now()

			// Call the next handler
			resp, err := next(ctx, req)

			// Report the latency
			latency := time.Since(start)
			options.reporter.ReportLatency(service, method, latency)

			// Report the error
			if err != nil {
				options.reporter.ReportError(service, method, err)
			}

			return resp, err
		}
	}
}

// splitServiceMethod splits a service method into service and method
func splitServiceMethod(serviceMethod string) (string, string) {
	for i := 0; i < len(serviceMethod); i++ {
		if serviceMethod[i] == '.' {
			return serviceMethod[:i], serviceMethod[i+1:]
		}
	}
	return serviceMethod, ""
}
