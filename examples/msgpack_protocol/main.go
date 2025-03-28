package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-thor/thor"
	"github.com/go-thor/thor/codec/binary"
	"github.com/go-thor/thor/codec/msgpack"
	"github.com/go-thor/thor/transport/tcp"
)

// 定义服务
type MsgPackService struct{}

// 定义请求和响应结构
type UserRequest struct {
	ID       int       `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	CreateAt time.Time `json:"create_at"`
	Tags     []string  `json:"tags"`
	Active   bool      `json:"active"`
}

type UserResponse struct {
	ID         int               `json:"id"`
	Username   string            `json:"username"`
	Email      string            `json:"email"`
	CreateAt   time.Time         `json:"create_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
	Tags       []string          `json:"tags"`
	Active     bool              `json:"active"`
	Properties map[string]string `json:"properties"`
	Message    string            `json:"message"`
}

// 实现服务方法
func (s *MsgPackService) ProcessUser(ctx context.Context, req *UserRequest) (*UserResponse, error) {
	log.Printf("收到用户请求: ID=%d, Username=%s", req.ID, req.Username)
	log.Printf("请求详情: %+v", req)

	// 创建响应对象
	resp := &UserResponse{
		ID:        req.ID,
		Username:  req.Username,
		Email:     req.Email,
		CreateAt:  req.CreateAt,
		UpdatedAt: time.Now(),
		Tags:      append(req.Tags, "processed"),
		Active:    req.Active,
		Properties: map[string]string{
			"ip":       "127.0.0.1",
			"platform": "web",
			"device":   "desktop",
		},
		Message: fmt.Sprintf("用户 %s 已成功处理", req.Username),
	}

	log.Printf("发送响应: %+v", resp)

	return resp, nil
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
	// 创建 MessagePack 编解码器 (用于序列化/反序列化负载)
	msgpackCodec := msgpack.New()

	// 创建二进制编解码器 (用于处理整个消息)
	binaryCodec := binary.New(msgpackCodec)

	// 创建TCP传输
	transport := tcp.New(
		tcp.WithAddress(":8899"),
		tcp.WithReadTimeout(5*time.Second),
		tcp.WithWriteTimeout(5*time.Second),
	)

	// 创建服务器
	server := thor.NewServer(binaryCodec, transport)

	// 注册服务
	err := server.Register(new(MsgPackService))
	if err != nil {
		log.Fatalf("注册服务失败: %v", err)
	}

	log.Println("MessagePack 服务器启动在 :8899")

	// 启动服务器
	if err := server.Serve(); err != nil {
		log.Fatalf("服务器错误: %v", err)
	}
}

func startClient() error {
	// 创建 MessagePack 编解码器 (用于序列化/反序列化负载)
	msgpackCodec := msgpack.New()

	// 创建二进制编解码器 (用于处理整个消息)
	binaryCodec := binary.New(msgpackCodec)

	// 创建TCP传输
	transport := tcp.New(
		tcp.WithTarget("localhost:8899"),
		tcp.WithReadTimeout(5*time.Second),
		tcp.WithWriteTimeout(5*time.Second),
	)

	// 创建客户端
	client := thor.NewClient(binaryCodec, transport)

	// 创建请求
	req := &UserRequest{
		ID:       1001,
		Username: "user123",
		Email:    "user123@example.com",
		CreateAt: time.Now().Add(-24 * time.Hour), // 假设用户创建于昨天
		Tags:     []string{"new", "premium"},
		Active:   true,
	}

	// 创建响应对象
	var resp UserResponse

	// 发送请求
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("发送 MessagePack 请求...")
	err := client.Call(ctx, "MsgPackService.ProcessUser", req, &resp)
	if err != nil {
		return fmt.Errorf("调用失败: %w", err)
	}

	log.Printf("收到响应: %+v", resp)
	log.Printf("响应消息: %s", resp.Message)
	log.Printf("响应属性: %v", resp.Properties)

	// 关闭客户端
	if err := client.Close(); err != nil {
		return fmt.Errorf("关闭客户端失败: %w", err)
	}

	return nil
}
