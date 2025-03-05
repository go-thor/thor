package transport

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/go-thor/thor"
)

type HTTPTransport struct {
	codec thor.Codec
}

func NewHTTPTransport(codec thor.Codec) *HTTPTransport {
	return &HTTPTransport{codec: codec}
}

func (t *HTTPTransport) ListenAndServe(addr string, handler thor.HandlerFunc) error {
	http.HandleFunc("/rpc", func(w http.ResponseWriter, r *http.Request) {
		ctx := &thor.RPCContext{
			Ctx:      r.Context(),     // 初始化 Ctx
			Request:  &thor.Request{}, // 初始化 Request
			Metadata: make(map[string]string),
		}
		for k, v := range r.Header {
			ctx.Metadata[k] = v[0]
		}
		if err := t.codec.DecodeStream(r.Body, &ctx.Request); err != nil {
			fmt.Printf("Decode error: %v\n", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := handler(ctx); err != nil {
			fmt.Printf("Handler error: %v\n", err)
		}
		t.codec.EncodeStream(w, ctx.Response)
	})
	return http.ListenAndServe(addr, nil)
}

func (t *HTTPTransport) Dial(addr string) (thor.ClientConn, error) {
	return &HTTPConn{addr: addr, codec: t.codec}, nil
}

type HTTPConn struct {
	addr  string
	codec thor.Codec
}

func (c *HTTPConn) Call(_ string, req interface{}, resp interface{}) error {
	data, err := c.codec.Encode(req)
	if err != nil {
		return err
	}
	respBody, err := http.Post("http://"+c.addr+"/rpc", "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer respBody.Body.Close()
	return c.codec.DecodeStream(respBody.Body, resp)
}

func (c *HTTPConn) Close() error {
	return nil
}
