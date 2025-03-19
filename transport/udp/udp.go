package udp

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/go-thor/thor"
	"github.com/go-thor/thor/errors"
)

const (
	// DefaultReadTimeout is the default read timeout
	DefaultReadTimeout = 5 * time.Second
	// DefaultWriteTimeout is the default write timeout
	DefaultWriteTimeout = 5 * time.Second
	// DefaultMaxMessageSize is the default maximum message size (64KB, UDP packet limit)
	DefaultMaxMessageSize = 64 * 1024
	// DefaultBufferSize is the default buffer size
	DefaultBufferSize = 65536
)

// Transport is a UDP transport
type Transport struct {
	readTimeout  time.Duration
	writeTimeout time.Duration
	maxMsgSize   int
	bufferSize   int

	conn   *net.UDPConn
	mu     sync.RWMutex
	closed bool
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

// WithMaxMessageSize sets the maximum message size
func WithMaxMessageSize(size int) Option {
	return func(t *Transport) {
		t.maxMsgSize = size
	}
}

// WithBufferSize sets the buffer size
func WithBufferSize(size int) Option {
	return func(t *Transport) {
		t.bufferSize = size
	}
}

// New creates a new UDP transport
func New(opts ...Option) *Transport {
	t := &Transport{
		readTimeout:  DefaultReadTimeout,
		writeTimeout: DefaultWriteTimeout,
		maxMsgSize:   DefaultMaxMessageSize,
		bufferSize:   DefaultBufferSize,
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

// Send sends the message to the specified address
// Note: UDP is connectionless, so this just sends the message and doesn't wait for a response.
// For RPC, it's generally better to use TCP or HTTP which are connection-oriented.
func (t *Transport) Send(ctx context.Context, addr string, message []byte) ([]byte, error) {
	if t.closed {
		return nil, errors.ErrServerClosed
	}

	if len(message) > t.maxMsgSize {
		return nil, errors.New(errors.ErrorCodeInvalidArgument, "message too large")
	}

	// Check if context is done
	select {
	case <-ctx.Done():
		return nil, errors.Wrap(errors.ErrorCodeCancelled, ctx.Err(), "context cancelled")
	default:
	}

	// Resolve address
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, errors.Wrap(errors.ErrorCodeUnknown, err, "failed to resolve address")
	}

	// Create a connection
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return nil, errors.Wrap(errors.ErrorCodeUnknown, err, "failed to dial")
	}
	defer conn.Close()

	// Set write deadline
	if t.writeTimeout > 0 {
		if err := conn.SetWriteDeadline(time.Now().Add(t.writeTimeout)); err != nil {
			return nil, errors.Wrap(errors.ErrorCodeUnknown, err, "failed to set write deadline")
		}
	}

	// Write message
	if _, err := conn.Write(message); err != nil {
		return nil, errors.Wrap(errors.ErrorCodeUnknown, err, "failed to write message")
	}

	// Since UDP is connectionless, we can't wait for a response in the same connection.
	// For a real RPC system using UDP, you'd need to implement a protocol with request IDs,
	// retries, timeouts, etc. For simplicity, we'll just return an empty response.
	return nil, errors.New(errors.ErrorCodeUnknown, "UDP is connectionless and does not support response handling in Send")
}

// Listen starts listening for incoming messages
func (t *Transport) Listen(addr string, handler func(ctx context.Context, message []byte) ([]byte, error)) error {
	if t.closed {
		return errors.ErrServerClosed
	}

	// Resolve address
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return errors.Wrap(errors.ErrorCodeUnknown, err, "failed to resolve address")
	}

	// Create a connection
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return errors.Wrap(errors.ErrorCodeUnknown, err, "failed to listen")
	}
	t.conn = conn

	// Set buffer size
	if err := conn.SetReadBuffer(t.bufferSize); err != nil {
		return errors.Wrap(errors.ErrorCodeUnknown, err, "failed to set read buffer")
	}
	if err := conn.SetWriteBuffer(t.bufferSize); err != nil {
		return errors.Wrap(errors.ErrorCodeUnknown, err, "failed to set write buffer")
	}

	// Handle messages
	buffer := make([]byte, t.maxMsgSize)
	for {
		if t.closed {
			return errors.ErrServerClosed
		}

		// Set read deadline
		if t.readTimeout > 0 {
			if err := conn.SetReadDeadline(time.Now().Add(t.readTimeout)); err != nil {
				return errors.Wrap(errors.ErrorCodeUnknown, err, "failed to set read deadline")
			}
		}

		// Read message
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			if t.closed {
				return errors.ErrServerClosed
			}
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// Timeout, just continue
				continue
			}
			return errors.Wrap(errors.ErrorCodeUnknown, err, "failed to read message")
		}

		// Copy message to avoid race conditions
		message := make([]byte, n)
		copy(message, buffer[:n])

		// Handle message asynchronously
		go func(message []byte, addr *net.UDPAddr) {
			ctx := context.Background()
			resp, err := handler(ctx, message)
			if err != nil {
				// Log error (in a real implementation)
				return
			}

			// Set write deadline
			if t.writeTimeout > 0 {
				if err := conn.SetWriteDeadline(time.Now().Add(t.writeTimeout)); err != nil {
					// Log error (in a real implementation)
					return
				}
			}

			// In UDP, we need to send the response back to the client
			if _, err := conn.WriteToUDP(resp, addr); err != nil {
				// Log error (in a real implementation)
				return
			}
		}(message, addr)
	}
}

// Close closes the transport
func (t *Transport) Close() error {
	if t.closed {
		return nil
	}

	t.closed = true

	if t.conn != nil {
		return t.conn.Close()
	}
	return nil
}

// Name returns the name of the transport
func (t *Transport) Name() string {
	return "udp"
}

// Make sure Transport implements pkg.Transport
var _ pkg.Transport = (*Transport)(nil)
