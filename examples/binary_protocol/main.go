package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-thor/thor"
	"github.com/go-thor/thor/codec/binary"
	"github.com/go-thor/thor/jsoncodec"
	"github.com/go-thor/thor/transport/tcp"
)

// 定义服务
type EchoService struct{}

// 定义请求和响应结构
type EchoRequest struct {
	Message string `json:"message"`
}

type EchoResponse struct {
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
}

// 实现服务方法
func (s *EchoService) Echo(ctx context.Context, req *EchoRequest) (*EchoResponse, error) {
	log.Printf("收到请求: %s", req.Message)
	return &EchoResponse{
		Message: "Echo: " + req.Message,
		Time:    time.Now(),
	}, nil
}

func main() {
	// 创建服务器
	go startServer()

	// 等待服务器启动
	time.Sleep(time.Second)

	// 创建客户端并发送请求
	err := startClient()
	if err != nil {
		log.Fatalf("客户端错误: %v", err)
	}
}

func startServer() {
	// 创建内部编解码器 (用于序列化/反序列化负载)
	jsonCodec := jsoncodec.New()

	// 创建二进制编解码器 (用于处理整个消息)
	binaryCodec := binary.New(jsonCodec)

	// 创建TCP传输
	transport := tcp.New(
		tcp.WithAddress(":8888"),
		tcp.WithReadTimeout(5*time.Second),
		tcp.WithWriteTimeout(5*time.Second),
	)

	// 创建服务器
	server := thor.NewServer(binaryCodec, transport)

	// 注册服务
	err := server.Register(new(EchoService))
	if err != nil {
		log.Fatalf("注册服务失败: %v", err)
	}

	log.Println("服务器启动在 :8888")

	// 启动服务器
	if err := server.Serve(); err != nil {
		log.Fatalf("服务器错误: %v", err)
	}
}

func startClient() error {
	// 创建内部编解码器 (用于序列化/反序列化负载)
	jsonCodec := jsoncodec.New()

	// 创建二进制编解码器 (用于处理整个消息)
	binaryCodec := binary.New(jsonCodec)

	// 创建TCP传输
	transport := tcp.New(
		tcp.WithTarget("localhost:8888"),
		tcp.WithReadTimeout(5*time.Second),
		tcp.WithWriteTimeout(5*time.Second),
	)

	// 创建客户端
	client := thor.NewClient(binaryCodec, transport)

	// 创建请求和响应
	req := &EchoRequest{Message: "Hello, Binary Protocol!"}
	var resp EchoResponse

	// 发送请求
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("发送请求...")
	err := client.Call(ctx, "EchoService.Echo", req, &resp)
	if err != nil {
		return fmt.Errorf("调用失败: %w", err)
	}

	log.Printf("收到响应: %+v", resp)

	// 关闭客户端
	if err := client.Close(); err != nil {
		return fmt.Errorf("关闭客户端失败: %w", err)
	}

	return nil
}
