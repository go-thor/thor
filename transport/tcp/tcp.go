package tcp

import (
	"context"
	"encoding/binary"
	"io"
	"log"
	"net"
	"sync"
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

// Transport is a TCP transport
type Transport struct {
	readTimeout  time.Duration
	writeTimeout time.Duration
	dialTimeout  time.Duration
	maxMsgSize   int
	address      string
	target       string

	listener net.Listener
	conns    map[string]net.Conn
	mu       sync.RWMutex
	closed   bool
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

// WithAddress sets the server listen address
func WithAddress(addr string) Option {
	return func(t *Transport) {
		t.address = addr
	}
}

// WithTarget sets the client target address
func WithTarget(target string) Option {
	return func(t *Transport) {
		t.target = target
	}
}

// New creates a new TCP transport
func New(opts ...Option) *Transport {
	t := &Transport{
		readTimeout:  DefaultReadTimeout,
		writeTimeout: DefaultWriteTimeout,
		dialTimeout:  DefaultDialTimeout,
		maxMsgSize:   DefaultMaxMessageSize,
		conns:        make(map[string]net.Conn),
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

// Send sends the message to the specified address
func (t *Transport) Send(ctx context.Context, addr string, message []byte) ([]byte, error) {
	if t.closed {
		return nil, errors.ErrServerClosed
	}

	if len(message) > t.maxMsgSize {
		return nil, errors.New(errors.ErrorCodeInvalidArgument, "message too large")
	}

	// 优先使用传入的 addr，如果没有则使用 t.target
	targetAddr := addr
	if targetAddr == "" {
		targetAddr = t.target
	}
	if targetAddr == "" {
		return nil, errors.New(errors.ErrorCodeInvalidArgument, "target address is required")
	}

	t.mu.RLock()
	conn, ok := t.conns[targetAddr]
	t.mu.RUnlock()

	if !ok {
		var err error
		conn, err = net.DialTimeout("tcp", targetAddr, t.dialTimeout)
		if err != nil {
			return nil, errors.Wrap(errors.ErrorCodeUnknown, err, "failed to dial")
		}

		t.mu.Lock()
		t.conns[targetAddr] = conn
		t.mu.Unlock()
	}

	// Check if context is done
	select {
	case <-ctx.Done():
		return nil, errors.Wrap(errors.ErrorCodeCancelled, ctx.Err(), "context cancelled")
	default:
	}

	// Send message size as 4-byte header
	sizeHeader := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeHeader, uint32(len(message)))

	// Set write deadline
	if t.writeTimeout > 0 {
		if err := conn.SetWriteDeadline(time.Now().Add(t.writeTimeout)); err != nil {
			return nil, errors.Wrap(errors.ErrorCodeUnknown, err, "failed to set write deadline")
		}
	}

	// Write header
	if _, err := conn.Write(sizeHeader); err != nil {
		t.mu.Lock()
		delete(t.conns, targetAddr)
		t.mu.Unlock()
		conn.Close()
		return nil, errors.Wrap(errors.ErrorCodeUnknown, err, "failed to write header")
	}

	// Write message
	if _, err := conn.Write(message); err != nil {
		t.mu.Lock()
		delete(t.conns, targetAddr)
		t.mu.Unlock()
		conn.Close()
		return nil, errors.Wrap(errors.ErrorCodeUnknown, err, "failed to write message")
	}

	// Set read deadline
	if t.readTimeout > 0 {
		if err := conn.SetReadDeadline(time.Now().Add(t.readTimeout)); err != nil {
			return nil, errors.Wrap(errors.ErrorCodeUnknown, err, "failed to set read deadline")
		}
	}

	// Read response size
	if _, err := io.ReadFull(conn, sizeHeader); err != nil {
		t.mu.Lock()
		delete(t.conns, targetAddr)
		t.mu.Unlock()
		conn.Close()
		return nil, errors.Wrap(errors.ErrorCodeUnknown, err, "failed to read header")
	}

	// Get response size
	respSize := binary.BigEndian.Uint32(sizeHeader)
	if respSize > uint32(t.maxMsgSize) {
		return nil, errors.New(errors.ErrorCodeInvalidArgument, "response too large")
	}

	// Read response
	resp := make([]byte, respSize)
	if _, err := io.ReadFull(conn, resp); err != nil {
		t.mu.Lock()
		delete(t.conns, targetAddr)
		t.mu.Unlock()
		conn.Close()
		return nil, errors.Wrap(errors.ErrorCodeUnknown, err, "failed to read response")
	}

	return resp, nil
}

// Listen starts listening for incoming messages
func (t *Transport) Listen(addr string, handler func(ctx context.Context, message []byte) ([]byte, error)) error {
	if t.closed {
		return errors.ErrServerClosed
	}

	// 优先使用传入的 addr，如果没有则使用 t.address
	listenAddr := addr
	if listenAddr == "" {
		listenAddr = t.address
	}
	if listenAddr == "" {
		return errors.New(errors.ErrorCodeInvalidArgument, "address is required")
	}

	var err error
	t.listener, err = net.Listen("tcp", listenAddr)
	if err != nil {
		return errors.Wrap(errors.ErrorCodeUnknown, err, "failed to listen")
	}

	for {
		conn, err := t.listener.Accept()
		if err != nil {
			if t.closed {
				return errors.ErrServerClosed
			}
			return errors.Wrap(errors.ErrorCodeUnknown, err, "failed to accept")
		}

		go t.handleConn(conn, handler)
	}
}

// handleConn handles a connection
func (t *Transport) handleConn(conn net.Conn, handler func(ctx context.Context, message []byte) ([]byte, error)) {
	defer conn.Close()

	for {
		// Set read deadline
		if t.readTimeout > 0 {
			if err := conn.SetReadDeadline(time.Now().Add(t.readTimeout)); err != nil {
				log.Printf("设置读取超时失败: %v", err)
				return
			}
		}

		// Read message size
		sizeHeader := make([]byte, 4)
		if _, err := io.ReadFull(conn, sizeHeader); err != nil {
			if err != io.EOF {
				log.Printf("读取消息头失败: %v", err)
			}
			return
		}

		// Get message size
		msgSize := binary.BigEndian.Uint32(sizeHeader)
		if msgSize > uint32(t.maxMsgSize) {
			log.Printf("消息过大: %d > %d", msgSize, t.maxMsgSize)
			return
		}

		log.Printf("接收消息大小: %d", msgSize)

		// Read message
		message := make([]byte, msgSize)
		if _, err := io.ReadFull(conn, message); err != nil {
			log.Printf("读取消息失败: %v", err)
			return
		}

		// Handle message
		ctx := context.Background()
		resp, err := handler(ctx, message)
		if err != nil {
			log.Printf("处理消息失败: %v", err)
			resp = []byte(err.Error())
		}

		log.Printf("响应大小: %d", len(resp))

		// Set write deadline
		if t.writeTimeout > 0 {
			if err := conn.SetWriteDeadline(time.Now().Add(t.writeTimeout)); err != nil {
				log.Printf("设置写入超时失败: %v", err)
				return
			}
		}

		// Send response size
		binary.BigEndian.PutUint32(sizeHeader, uint32(len(resp)))
		if _, err := conn.Write(sizeHeader); err != nil {
			log.Printf("写入响应头失败: %v", err)
			return
		}

		// Send response
		if _, err := conn.Write(resp); err != nil {
			log.Printf("写入响应失败: %v", err)
			return
		}

		log.Printf("响应已发送")
	}
}

// Close closes the transport
func (t *Transport) Close() error {
	if t.closed {
		return nil
	}

	t.closed = true

	if t.listener != nil {
		t.listener.Close()
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	for _, conn := range t.conns {
		conn.Close()
	}

	t.conns = make(map[string]net.Conn)
	return nil
}

// Name returns the name of the transport
func (t *Transport) Name() string {
	return "tcp"
}

// Make sure Transport implements pkg.Transport
var _ pkg.Transport = (*Transport)(nil)
