package transport

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/go-thor/thor"
)

type UDPTransport struct {
	codec thor.Codec
}

func NewUDPTransport(codec thor.Codec) *UDPTransport {
	return &UDPTransport{codec: codec}
}

func (t *UDPTransport) ListenAndServe(addr string, handler thor.HandlerFunc) error {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	buf := make([]byte, 65535) // UDP 最大数据包大小
	for {
		n, clientAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Printf("ReadFromUDP error: %v\n", err)
			continue
		}

		// 处理单个数据包
		go func(data []byte, addr *net.UDPAddr) {
			ctx := &thor.RPCContext{Ctx: context.Background(), Metadata: make(map[string]string), Request: &thor.Request{}}
			if err := t.codec.Decode(data, &ctx.Request); err != nil {
				fmt.Printf("Decode error: %v\n", err)
				return
			}

			if err := handler(ctx); err != nil {
				fmt.Printf("Handler error: %v\n", err)
			}

			// 发送响应
			respData, err := t.codec.Encode(ctx.Response)
			if err != nil {
				fmt.Printf("Encode error: %v\n", err)
				return
			}
			if _, err := conn.WriteToUDP(respData, addr); err != nil {
				fmt.Printf("WriteToUDP error: %v\n", err)
			}
		}(buf[:n], clientAddr)
	}
}

func (t *UDPTransport) Dial(addr string) (thor.ClientConn, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return nil, err
	}
	return &UDPConn{conn: conn, codec: t.codec}, nil
}

type UDPConn struct {
	conn  *net.UDPConn
	codec thor.Codec
}

func (c *UDPConn) Call(_ string, req interface{}, resp interface{}) error {
	// 发送请求
	reqData, err := c.codec.Encode(req)
	if err != nil {
		return err
	}
	if _, err := c.conn.Write(reqData); err != nil {
		return err
	}

	// 接收响应
	buf := make([]byte, 65535)
	c.conn.SetReadDeadline(time.Now().Add(5 * time.Second)) // 设置超时
	n, _, err := c.conn.ReadFromUDP(buf)
	if err != nil {
		return err
	}

	return c.codec.Decode(buf[:n], resp)
}

func (c *UDPConn) Close() error {
	return c.conn.Close()
}
