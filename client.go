package thor

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-thor/thor/errors"
	"github.com/go-thor/thor/jsoncodec"
)

// DefaultTimeout is the default timeout for client calls
const DefaultTimeout = 10 * time.Second

// clientRequest represents an RPC request
type clientRequest struct {
	ServiceMethod string
	Seq           uint64
	Metadata      map[string]string
	next          *clientRequest
}

// DefaultClient is the default implementation of Client
type DefaultClient struct {
	codec       Codec
	transport   Transport
	mu          sync.Mutex
	seq         uint64
	pending     map[uint64]*Call
	closed      bool
	middlewares []Middleware
}

// ClientOptions defines options for client
type ClientOptions struct {
	Timeout time.Duration
}

// ClientOption defines a function that sets options for client
type ClientOption func(*ClientOptions)

// NewClient creates a new client
func NewClient(codec Codec, transport Transport) *DefaultClient {
	return &DefaultClient{
		codec:     codec,
		transport: transport,
		pending:   make(map[uint64]*Call),
	}
}

// Use adds middleware to the client
func (c *DefaultClient) Use(middleware ...Middleware) {
	c.middlewares = append(c.middlewares, middleware...)
}

// Call invokes the named function, waits for it to complete, and returns its error status
func (c *DefaultClient) Call(ctx context.Context, serviceMethod string, args interface{}, reply interface{}) error {
	return c.CallWithMetadata(ctx, serviceMethod, args, reply, nil)
}

// CallWithMetadata is like Call but with additional metadata
func (c *DefaultClient) CallWithMetadata(ctx context.Context, serviceMethod string, args interface{}, reply interface{}, metadata map[string]string) error {
	call := c.Go(ctx, serviceMethod, args, reply, make(chan *Call, 1))
	call.Metadata = metadata

	// Wait for the call to complete
	<-call.Done

	return call.Error
}

// Go invokes the function asynchronously
func (c *DefaultClient) Go(ctx context.Context, serviceMethod string, args interface{}, reply interface{}, done chan *Call) *Call {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		if done == nil {
			done = make(chan *Call, 1)
		}
		call := &Call{
			ServiceMethod: serviceMethod,
			Args:          args,
			Reply:         reply,
			Error:         errors.ErrClientClosed,
			Done:          done,
		}
		call.done()
		return call
	}

	// Increment sequence number
	seq := c.seq
	c.seq++

	// Create call
	call := &Call{
		ServiceMethod: serviceMethod,
		Args:          args,
		Reply:         reply,
		Done:          done,
		Context:       ctx,
		Seq:           seq,
	}

	// Add call to pending map
	c.pending[seq] = call
	c.mu.Unlock()

	// Send the request
	go c.send(call)

	return call
}

// send sends the RPC request
func (c *DefaultClient) send(call *Call) {
	// Marshal the request
	reqData, err := c.codec.Marshal(call.Args)
	if err != nil {
		c.mu.Lock()
		call = c.pending[call.Seq]
		delete(c.pending, call.Seq)
		c.mu.Unlock()
		if call != nil {
			call.Error = fmt.Errorf("marshal request: %w", err)
			call.done()
		}
		return
	}

	// Create JSON codec for internal message
	jsonCodec := jsoncodec.New()

	// Create request
	rpcReq := &Request{
		ServiceMethod: call.ServiceMethod,
		Seq:           call.Seq,
		Metadata:      call.Metadata,
		Payload:       reqData,
		Args:          reqData,
	}

	// Marshal the request
	reqData, err = jsonCodec.Marshal(rpcReq)
	if err != nil {
		c.mu.Lock()
		call = c.pending[call.Seq]
		delete(c.pending, call.Seq)
		c.mu.Unlock()
		if call != nil {
			call.Error = fmt.Errorf("marshal request: %w", err)
			call.done()
		}
		return
	}

	// Get context with timeout
	ctx := call.Context
	if ctx == nil {
		ctx = context.Background()
	}
	if deadline, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, DefaultTimeout)
		defer cancel()
	} else {
		log.Printf("截止时间: %v", deadline)
	}

	// Send the request
	respData, err := c.transport.Send(ctx, "", reqData)
	if err != nil {
		c.mu.Lock()
		call = c.pending[call.Seq]
		delete(c.pending, call.Seq)
		c.mu.Unlock()
		if call != nil {
			call.Error = err
			call.done()
		}
		return
	}

	// Handle the response
	go c.handleResponse(respData, call)
}

// Close closes the client
func (c *DefaultClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return errors.ErrClientClosed
	}
	c.closed = true

	// Terminate pending calls
	for _, call := range c.pending {
		call.Error = errors.ErrClientClosed
		call.done()
	}

	return c.transport.Close()
}

// handleResponse handles the response
func (c *DefaultClient) handleResponse(respData []byte, call *Call) {
	// Unmarshal the response
	var resp Response
	jsonCodec := jsoncodec.New()
	log.Printf("客户端收到的原始响应数据长度: %d", len(respData))
	err := jsonCodec.Unmarshal(respData, &resp)
	if err != nil {
		c.mu.Lock()
		call = c.pending[call.Seq]
		delete(c.pending, call.Seq)
		c.mu.Unlock()
		if call != nil {
			call.Error = fmt.Errorf("unmarshal response: %w", err)
			call.done()
		}
		return
	}
	log.Printf("客户端解析后的响应对象: %+v", resp)

	// Check for errors
	if resp.Error != "" {
		c.mu.Lock()
		call = c.pending[resp.Seq]
		delete(c.pending, resp.Seq)
		c.mu.Unlock()
		if call != nil {
			call.Error = errors.New(errors.ErrorCodeUnknown, resp.Error)
			call.done()
		}
		return
	}

	// 优先使用Reply字段，如果为空则使用Payload字段
	responseData := resp.Reply
	if len(responseData) == 0 {
		responseData = resp.Payload
	}
	log.Printf("客户端用于反序列化的响应数据长度: %d", len(responseData))

	// Unmarshal the reply
	err = c.codec.Unmarshal(responseData, call.Reply)
	if err != nil {
		c.mu.Lock()
		call = c.pending[call.Seq]
		delete(c.pending, call.Seq)
		c.mu.Unlock()
		if call != nil {
			call.Error = fmt.Errorf("unmarshal reply: %w", err)
			call.done()
		}
		return
	}
	log.Printf("客户端最终的响应对象: %+v", call.Reply)

	// Call is complete
	c.mu.Lock()
	pendingCall := c.pending[resp.Seq]
	delete(c.pending, resp.Seq)
	c.mu.Unlock()

	// Finish the call
	if pendingCall != nil {
		c.finish(pendingCall)
	}
}

// finish finishes the call
func (c *DefaultClient) finish(call *Call) {
	// Call is done
	call.done()
}

// done marks the call as done
func (call *Call) done() {
	select {
	case call.Done <- call:
		// ok
	default:
		// We don't want to block here. It's the caller's responsibility to make
		// sure the channel has enough buffer space. See comment in Go().
		log.Println("rpc: discarding Call reply due to insufficient Done chan capacity")
	}
}
