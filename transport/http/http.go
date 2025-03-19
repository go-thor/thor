package http

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/go-thor/thor/pkg"
	"github.com/go-thor/thor/pkg/errors"
)

const (
	// DefaultReadTimeout is the default read timeout
	DefaultReadTimeout = 30 * time.Second
	// DefaultWriteTimeout is the default write timeout
	DefaultWriteTimeout = 30 * time.Second
	// DefaultDialTimeout is the default dial timeout
	DefaultDialTimeout = 10 * time.Second
	// DefaultMaxMessageSize is the default maximum message size (10MB)
	DefaultMaxMessageSize = 10 * 1024 * 1024
)

// Transport is an HTTP transport
type Transport struct {
	readTimeout  time.Duration
	writeTimeout time.Duration
	dialTimeout  time.Duration
	maxMsgSize   int

	server   *http.Server
	client   *http.Client
	basePath string
}

// Option is a transport option
type Option func(*Transport)

// WithReadTimeout sets the read timeout
func WithReadTimeout(timeout time.Duration) Option {
	return func(t *Transport) {
		t.readTimeout = timeout
	}
}

// WithWriteTimeout sets the write timeout
func WithWriteTimeout(timeout time.Duration) Option {
	return func(t *Transport) {
		t.writeTimeout = timeout
	}
}

// WithDialTimeout sets the dial timeout
func WithDialTimeout(timeout time.Duration) Option {
	return func(t *Transport) {
		t.dialTimeout = timeout
	}
}

// WithMaxMessageSize sets the maximum message size
func WithMaxMessageSize(size int) Option {
	return func(t *Transport) {
		t.maxMsgSize = size
	}
}

// WithBasePath sets the base path for the HTTP server
func WithBasePath(path string) Option {
	return func(t *Transport) {
		t.basePath = path
	}
}

// New creates a new HTTP transport
func New(opts ...Option) *Transport {
	t := &Transport{
		readTimeout:  DefaultReadTimeout,
		writeTimeout: DefaultWriteTimeout,
		dialTimeout:  DefaultDialTimeout,
		maxMsgSize:   DefaultMaxMessageSize,
		basePath:     "/rpc",
	}

	for _, opt := range opts {
		opt(t)
	}

	t.client = &http.Client{
		Timeout: t.readTimeout + t.writeTimeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
			DialContext: (&net.Dialer{
				Timeout:   t.dialTimeout,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxConnsPerHost:       0,
			ResponseHeaderTimeout: t.readTimeout,
		},
	}

	return t
}

// Send sends the message to the specified address
func (t *Transport) Send(ctx context.Context, addr string, message []byte) ([]byte, error) {
	if len(message) > t.maxMsgSize {
		return nil, errors.New(errors.ErrorCodeInvalidArgument, "message too large")
	}

	url := "http://" + addr + t.basePath
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(message))
	if err != nil {
		return nil, errors.Wrap(errors.ErrorCodeUnknown, err, "failed to create request")
	}

	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(errors.ErrorCodeUnknown, err, "failed to do request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(errors.ErrorCodeUnknown, "unexpected status code: "+resp.Status)
	}

	// Read response with a size limit
	limiter := io.LimitReader(resp.Body, int64(t.maxMsgSize))
	respBody, err := io.ReadAll(limiter)
	if err != nil {
		return nil, errors.Wrap(errors.ErrorCodeUnknown, err, "failed to read response")
	}

	if len(respBody) > t.maxMsgSize {
		return nil, errors.New(errors.ErrorCodeInvalidArgument, "response too large")
	}

	return respBody, nil
}

// Listen starts listening for incoming messages
func (t *Transport) Listen(addr string, handler func(ctx context.Context, message []byte) ([]byte, error)) error {
	mux := http.NewServeMux()
	mux.HandleFunc(t.basePath, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Read request with a size limit
		limiter := io.LimitReader(r.Body, int64(t.maxMsgSize))
		reqBody, err := io.ReadAll(limiter)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if len(reqBody) > t.maxMsgSize {
			http.Error(w, "Request too large", http.StatusRequestEntityTooLarge)
			return
		}

		// Handle request
		resp, err := handler(r.Context(), reqBody)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Set content type
		w.Header().Set("Content-Type", "application/octet-stream")

		// Write response
		if _, err := w.Write(resp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	t.server = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  t.readTimeout,
		WriteTimeout: t.writeTimeout,
	}

	return t.server.ListenAndServe()
}

// Close closes the transport
func (t *Transport) Close() error {
	if t.server != nil {
		return t.server.Close()
	}
	return nil
}

// Name returns the name of the transport
func (t *Transport) Name() string {
	return "http"
}

// Make sure Transport implements pkg.Transport
var _ pkg.Transport = (*Transport)(nil)
