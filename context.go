package thor

import (
	"context"
	"fmt"
	"io"
)

// RPCContext 封装请求上下文
type RPCContext struct {
	Ctx      context.Context
	Request  *Request
	Response interface{}
	Metadata map[string]string
}

// Request RPC 请求结构
type Request struct {
	ServiceName string
	MethodName  string
	Params      interface{}
}

// Transport 协议接口
type Transport interface {
	ListenAndServe(addr string, handler HandlerFunc) error
	Dial(addr string) (ClientConn, error)
}

// ClientConn 客户端连接接口
type ClientConn interface {
	Call(method string, req interface{}, resp interface{}) error
	Close() error
}

// Codec 序列化/反序列化接口，支持流式解析
type Codec interface {
	Encode(v interface{}) ([]byte, error)
	Decode(data []byte, v interface{}) error
	EncodeStream(w io.Writer, v interface{}) error
	DecodeStream(r io.Reader, v interface{}) error
}

// Middleware 中间件接口
type Middleware func(next HandlerFunc) HandlerFunc

// HandlerFunc 服务处理函数
type HandlerFunc func(*RPCContext) error

// Errorf 格式化错误
func Errorf(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}
