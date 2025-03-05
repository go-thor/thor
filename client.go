package thor

import (
	"context"
	"fmt"
	"strings"
)

type Client struct {
	conn        ClientConn
	codec       Codec
	middlewares []Middleware // 中间件列表
}

func NewClient(transport Transport, addr string, codec Codec) (*Client, error) {
	conn, err := transport.Dial(addr)
	if err != nil {
		return nil, err
	}
	return &Client{
		conn:        conn,
		codec:       codec,
		middlewares: make([]Middleware, 0),
	}, nil
}

// Use 注册客户端中间件
func (c *Client) Use(mw Middleware) {
	c.middlewares = append(c.middlewares, mw)
}

func (c *Client) Call(route string, ctx context.Context, req interface{}, resp interface{}) error {
	parts := strings.Split(route, ".")
	if len(parts) != 2 {
		return fmt.Errorf("invalid route format: %s, expected ServiceName.MethodName", route)
	}
	serviceName, methodName := parts[0], parts[1]

	request := Request{
		ServiceName: serviceName,
		MethodName:  methodName,
		Params:      req,
	}

	rpcCtx := &RPCContext{
		Ctx:      ctx,
		Request:  &request,
		Metadata: make(map[string]string),
	}

	// 构建中间件链
	handler := func(ctx *RPCContext) error {
		return c.conn.Call("", ctx.Request, &ctx.Response)
	}
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		handler = c.middlewares[i](handler)
	}

	// 执行中间件链
	if err := handler(rpcCtx); err != nil {
		return err
	}

	// 将响应赋值给调用者提供的 resp
	*resp.(*interface{}) = rpcCtx.Response
	return nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}
