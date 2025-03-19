package main

import (
	"context"
	"log"
	"time"

	"github.com/go-thor/thor/codec/protobuf"
	"github.com/go-thor/thor/examples/greeter/proto"
	"github.com/go-thor/thor/middleware/logging"
	"github.com/go-thor/thor/middleware/retry"
	"github.com/go-thor/thor/middleware/timeout"
	"github.com/go-thor/thor/pkg"
	"github.com/go-thor/thor/transport/tcp"
)

func main() {
	// 创建 TCP 传输层
	transport := tcp.New(
		tcp.WithTarget("localhost:50051"),
	)

	// 创建 protobuf 编解码器
	codec := protobuf.New()

	// 创建客户端
	client := pkg.NewClient(codec, transport)
	defer client.Close()

	// 添加客户端中间件
	// 1. 日志中间件
	client.Use(logging.New())

	// 2. 超时中间件 - 设置全局超时为3秒
	client.Use(timeout.WithTimeout(3 * time.Second))

	// 3. 重试中间件 - 最多重试2次，间隔500毫秒
	client.Use(retry.New(
		retry.WithMaxRetries(2),
		retry.WithRetryInterval(500*time.Millisecond),
	))

	// 创建 Greeter 客户端
	greeterClient := proto.NewGreeterClient(client)

	// 创建上下文
	ctx := context.Background()

	// 调用 SayHello
	helloResp, err := greeterClient.SayHello(ctx, &proto.HelloRequest{Name: "Thor"})
	if err != nil {
		log.Fatalf("调用 SayHello 失败: %v", err)
	}
	log.Printf("SayHello 响应: %s", helloResp.Message)

	// 调用 SayGoodbye
	goodbyeResp, err := greeterClient.SayGoodbye(ctx, &proto.GoodbyeRequest{Name: "Thor"})
	if err != nil {
		log.Fatalf("调用 SayGoodbye 失败: %v", err)
	}
	log.Printf("SayGoodbye 响应: %s", goodbyeResp.Message)
}
