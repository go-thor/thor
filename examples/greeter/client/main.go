package main

import (
	"context"
	"log"

	"github.com/go-thor/thor"
	"github.com/go-thor/thor/codec/protobuf"
	"github.com/go-thor/thor/examples/greeter/proto"
	"github.com/go-thor/thor/middleware/logging"
	"github.com/go-thor/thor/transport/tcp"
)

func main() {
	// 创建 TCP 传输层
	transport := tcp.New(
		tcp.WithTarget("localhost:50052"),
	)

	// 创建 protobuf 编解码器
	codec := protobuf.New()

	// 创建客户端
	client := thor.NewClient(codec, transport)
	defer client.Close()

	// 添加客户端中间件
	// 1. 日志中间件
	client.Use(logging.New())

	// 创建 Greeter 客户端
	greeterClient := proto.NewGreeterClient(client)

	// 创建上下文
	ctx := context.Background()

	// 调用 SayHello
	helloResp, err := greeterClient.SayHello(ctx, &proto.HelloRequest{Name: "Thor"})
	if err != nil {
		log.Fatalf("调用 SayHello 失败: %v", err)
	}
	log.Printf("SayHello 响应: %+v", helloResp)
	if helloResp != nil && helloResp.Message != "" {
		log.Printf("SayHello 响应消息: %s", helloResp.Message)
	} else {
		log.Printf("SayHello 响应为空或消息为空")
	}

	// 调用 SayGoodbye
	goodbyeResp, err := greeterClient.SayGoodbye(ctx, &proto.GoodbyeRequest{Name: "Thor"})
	if err != nil {
		log.Fatalf("调用 SayGoodbye 失败: %v", err)
	}
	log.Printf("SayGoodbye 响应: %+v", goodbyeResp)
	if goodbyeResp != nil && goodbyeResp.Message != "" {
		log.Printf("SayGoodbye 响应消息: %s", goodbyeResp.Message)
	} else {
		log.Printf("SayGoodbye 响应为空或消息为空")
	}
}
