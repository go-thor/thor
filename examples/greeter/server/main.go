package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/go-thor/thor/codec/protobuf"
	"github.com/go-thor/thor/examples/greeter/proto"
	"github.com/go-thor/thor/middleware/logging"
	"github.com/go-thor/thor/pkg"
	"github.com/go-thor/thor/transport/tcp"
)

// GreeterService implements the Greeter service
type GreeterService struct{}

// SayHello implements the SayHello method
func (s *GreeterService) SayHello(ctx context.Context, req *proto.HelloRequest) (*proto.HelloResponse, error) {
	log.Printf("收到请求: %v", req.Name)

	return nil, errors.New("test error")

	return &proto.HelloResponse{Message: fmt.Sprintf("你好, %s!", req.Name)}, nil
}

// SayGoodbye implements the SayGoodbye method
func (s *GreeterService) SayGoodbye(ctx context.Context, req *proto.GoodbyeRequest) (*proto.GoodbyeResponse, error) {
	log.Printf("收到请求: %v", req.Name)
	return &proto.GoodbyeResponse{Message: fmt.Sprintf("再见, %s!", req.Name)}, nil
}

func main() {
	// 创建 TCP 传输层
	transport := tcp.New(
		tcp.WithAddress(":50051"),
	)

	// 创建 protobuf 编解码器
	codec := protobuf.New()

	// 创建服务器
	server := pkg.NewServer(codec, transport)

	// 添加日志中间件
	server.Use(logging.New())

	// 注册 GreeterService
	err := proto.RegisterGreeterServer(server, &GreeterService{})
	if err != nil {
		log.Fatalf("注册服务失败: %v", err)
	}

	// 启动服务器
	log.Println("启动 RPC 服务器在 :50051")
	if err := server.Serve(); err != nil {
		log.Fatalf("启动服务失败: %v", err)
	}
}
