package transport

import (
	"context"
	"fmt"
	"io"
	"net"

	"github.com/go-thor/thor"
)

type TCPTransport struct {
	codec thor.Codec
}

func NewTCPTransport(codec thor.Codec) *TCPTransport {
	return &TCPTransport{codec: codec}
}

func (t *TCPTransport) ListenAndServe(addr string, handler thor.HandlerFunc) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		// 在一个 goroutine 中处理同一个连接的多个请求
		go t.handleConnection(conn, handler)
	}
}

// handleConnection 处理单个 TCP 连接，支持多个请求
func (t *TCPTransport) handleConnection(conn net.Conn, handler thor.HandlerFunc) {
	defer conn.Close()
	for {
		ctx := &thor.RPCContext{Ctx: context.Background(), Metadata: make(map[string]string), Request: &thor.Request{}}
		if err := t.codec.DecodeStream(conn, &ctx.Request); err != nil {
			if err != io.EOF {
				fmt.Printf("Decode error: %v\n", err)
			}
			return // 连接关闭或出错，退出循环
		}
		if err := handler(ctx); err != nil {
			fmt.Printf("Handler error: %v\n", err)
		}
		if err := t.codec.EncodeStream(conn, ctx.Response); err != nil {
			fmt.Printf("Encode error: %v\n", err)
			return
		}
	}
}

func (t *TCPTransport) Dial(addr string) (thor.ClientConn, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &TCPConn{conn: conn, codec: t.codec}, nil
}

type TCPConn struct {
	conn  net.Conn
	codec thor.Codec
}

func (c *TCPConn) Call(_ string, req interface{}, resp interface{}) error {
	if err := c.codec.EncodeStream(c.conn, req); err != nil {
		return err
	}
	return c.codec.DecodeStream(c.conn, resp)
}

func (c *TCPConn) Close() error {
	return c.conn.Close()
}
