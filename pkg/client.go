package pkg

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"sync"
	"time"

	"github.com/go-thor/thor/pkg/errors"
	"github.com/go-thor/thor/pkg/jsoncodec"
)

// DefaultTimeout is the default timeout for client calls
const DefaultTimeout = 10 * time.Second

// clientRequest represents a request from the client to the server
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

// ClientOption is a function that configures a Client
type ClientOption func(*Options)

// NewClient creates a new client
func NewClient(codec Codec, transport Transport) *DefaultClient {
	return &DefaultClient{
		codec:       codec,
		transport:   transport,
		seq:         0,
		pending:     make(map[uint64]*Call),
		middlewares: make([]Middleware, 0),
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
	select {
	case <-ctx.Done():
		return ctx.Err()
	case call := <-call.Done:
		return call.Error
	}
}

// Go invokes the function asynchronously
func (c *DefaultClient) Go(ctx context.Context, serviceMethod string, args interface{}, reply interface{}, done chan *Call) *Call {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		call := &Call{
			ServiceMethod: serviceMethod,
			Args:          args,
			Reply:         reply,
			Error:         errors.New(errors.ErrorCodeUnknown, "client closed"),
			Done:          done,
		}
		if done != nil {
			done <- call
		}
		return call
	}

	seq := c.seq
	c.seq++

	// Create call
	call := &Call{
		ServiceMethod: serviceMethod,
		Args:          args,
		Reply:         reply,
		Seq:           seq,
		Done:          done,
	}

	// Add to pending
	c.pending[seq] = call
	c.mu.Unlock()

	// 创建基础处理函数
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		// Marshal payload
		argData, err := c.codec.Marshal(req)
		if err != nil {
			return nil, err
		}

		// 创建实际请求
		rpcReq := &Request{
			ServiceMethod: serviceMethod,
			Seq:           seq,
			Payload:       argData,
		}

		// 创建 JSON 编解码器用于内部消息
		jsonCodec := jsoncodec.New()

		// Marshal request
		reqData, err := jsonCodec.Marshal(rpcReq)
		if err != nil {
			return nil, err
		}

		// Send request
		respData, err := c.transport.Send(ctx, "", reqData)
		if err != nil {
			return nil, err
		}

		// Unmarshal response
		var resp Response
		err = jsonCodec.Unmarshal(respData, &resp)
		if err != nil {
			return nil, err
		}

		// Check if call was successfully executed
		if resp.Error != "" {
			return nil, errors.New(resp.Error, resp.Error)
		}

		// Create reply instance
		replyValue := reflect.New(reflect.TypeOf(reply).Elem())

		// Unmarshal response into reply
		err = c.codec.Unmarshal(resp.Payload, replyValue.Interface())
		if err != nil {
			return nil, err
		}

		return replyValue.Interface(), nil
	}

	// 应用中间件（倒序应用，先添加的中间件最后执行）
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		handler = c.middlewares[i](handler)
	}

	// 异步调用
	go func() {
		result, err := handler(ctx, args)
		if err != nil {
			call.Error = err
		} else if result != nil {
			// 将结果复制到 reply
			reflect.ValueOf(reply).Elem().Set(reflect.ValueOf(result).Elem())
		}
		c.finish(call)
	}()

	return call
}

// Close closes the client
func (c *DefaultClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true
	return c.transport.Close()
}

// send sends a call to the server
func (c *DefaultClient) send(call *Call) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if client is closed
	if c.closed {
		call.Error = errors.ErrClientClosed
		call.done()
		return
	}

	// Create request
	req := &Request{
		ServiceMethod: call.ServiceMethod,
		Seq:           call.Seq,
		Metadata:      call.Metadata,
	}

	// Marshal args
	var err error
	req.Payload, err = c.codec.Marshal(call.Args)
	if err != nil {
		call.Error = err
		call.done()
		return
	}

	// Marshal request
	reqData, err := c.codec.Marshal(req)
	if err != nil {
		call.Error = err
		call.done()
		return
	}

	// Create context with timeout
	ctx := call.Context
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, DefaultTimeout)
		defer cancel()
	}

	// Send request
	respData, err := c.transport.Send(ctx, "", reqData)
	if err != nil {
		call.Error = err
		call.done()
		return
	}

	// Handle response
	go c.handleResponse(respData, call)
}

// handleResponse handles a response from the server
func (c *DefaultClient) handleResponse(respData []byte, call *Call) {
	// 创建 JSON 编解码器用于内部消息
	jsonCodec := jsoncodec.New()

	// Unmarshal response
	var resp Response
	err := jsonCodec.Unmarshal(respData, &resp)
	if err != nil {
		call.Error = err
		c.finish(call)
		return
	}

	log.Printf("收到响应: %+v", resp)

	// Check if response is for this call
	c.mu.Lock()
	pendingCall, ok := c.pending[resp.Seq]
	if !ok {
		c.mu.Unlock()
		log.Printf("未找到待处理的调用: %d", resp.Seq)
		return
	}

	// Remove call from pending
	delete(c.pending, resp.Seq)
	c.mu.Unlock()

	// Check if call was successfully executed
	if resp.Error != "" {
		pendingCall.Error = errors.New(resp.Error, resp.Error)
		c.finish(pendingCall)
		log.Printf("调用出错: %s", resp.Error)
		return
	}

	log.Printf("响应负载长度: %d", len(resp.Payload))

	// Unmarshal response into reply
	err = c.codec.Unmarshal(resp.Payload, pendingCall.Reply)
	if err != nil {
		pendingCall.Error = err
		log.Printf("反序列化响应失败: %v", err)
	} else {
		log.Printf("响应解析成功: %+v", pendingCall.Reply)
	}

	c.finish(pendingCall)
}

// finish finishes a call
func (c *DefaultClient) finish(call *Call) {
	call.done()
}

// done finishes a call and signals it's done
func (call *Call) done() {
	select {
	case call.Done <- call:
		// ok
	default:
		// We don't want to block here. It's the caller's responsibility to make
		// sure the channel has enough buffer space. For now we'll just drop
		// the message.
		fmt.Println("thor: done channel is full")
	}
}
