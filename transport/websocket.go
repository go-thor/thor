package transport

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-thor/thor"

	"github.com/gorilla/websocket"
)

type WebSocketTransport struct {
	codec    thor.Codec
	upgrader websocket.Upgrader
}

func NewWebSocketTransport(codec thor.Codec) *WebSocketTransport {
	return &WebSocketTransport{
		codec: codec,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     func(r *http.Request) bool { return true }, // 允许跨域
		},
	}
}

func (t *WebSocketTransport) ListenAndServe(addr string, handler thor.HandlerFunc) error {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := t.upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Printf("WebSocket upgrade error: %v\n", err)
			return
		}
		defer conn.Close()

		// 循环处理 WebSocket 连接的多个请求
		for {
			ctx := &thor.RPCContext{
				Ctx:      context.Background(),
				Request:  &thor.Request{},
				Metadata: make(map[string]string),
			}

			// 读取消息
			_, msg, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					fmt.Printf("ReadMessage error: %v\n", err)
				}
				return // 客户端关闭连接或其他错误
			}

			// 解码请求
			if err := t.codec.Decode(msg, &ctx.Request); err != nil {
				fmt.Printf("Decode error: %v\n", err)
				continue
			}

			// 处理请求
			if err := handler(ctx); err != nil {
				fmt.Printf("Handler error: %v\n", err)
			}

			// 发送响应
			respData, err := t.codec.Encode(ctx.Response)
			if err != nil {
				fmt.Printf("Encode error: %v\n", err)
				continue
			}
			if err := conn.WriteMessage(websocket.BinaryMessage, respData); err != nil {
				fmt.Printf("WriteMessage error: %v\n", err)
				return
			}
		}
	})
	return http.ListenAndServe(addr, nil)
}

func (t *WebSocketTransport) Dial(addr string) (thor.ClientConn, error) {
	conn, _, err := websocket.DefaultDialer.Dial("ws://"+addr+"/ws", nil)
	if err != nil {
		return nil, err
	}
	return &WebSocketConn{conn: conn, codec: t.codec}, nil
}

type WebSocketConn struct {
	conn  *websocket.Conn
	codec thor.Codec
}

func (c *WebSocketConn) Call(_ string, req interface{}, resp interface{}) error {
	// 发送请求
	reqData, err := c.codec.Encode(req)
	if err != nil {
		return err
	}
	if err := c.conn.WriteMessage(websocket.BinaryMessage, reqData); err != nil {
		return err
	}

	// 接收响应
	c.conn.SetReadDeadline(time.Now().Add(5 * time.Second)) // 设置超时
	_, respData, err := c.conn.ReadMessage()
	if err != nil {
		return err
	}

	return c.codec.Decode(respData, resp)
}

func (c *WebSocketConn) Close() error {
	return c.conn.Close()
}
