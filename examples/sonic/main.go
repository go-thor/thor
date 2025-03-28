package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-thor/thor"
	"github.com/go-thor/thor/codec/sonic"
	"github.com/go-thor/thor/transport/tcp"
)

// UserService 是一个用户服务
type UserService struct{}

// UserRequest 是用户请求
type UserRequest struct {
	ID int `json:"id"`
}

// UserResponse 是用户响应
type UserResponse struct {
	ID       int                    `json:"id"`
	Name     string                 `json:"name"`
	Email    string                 `json:"email"`
	Created  time.Time              `json:"created"`
	Modified time.Time              `json:"modified"`
	Tags     []string               `json:"tags"`
	Metadata map[string]interface{} `json:"metadata"`
}

// GetUser 获取用户信息
func (s *UserService) GetUser(ctx context.Context, req *UserRequest) (*UserResponse, error) {
	// 模拟从数据库获取用户
	if req.ID <= 0 {
		return nil, fmt.Errorf("invalid user ID: %d", req.ID)
	}

	// 创建响应
	resp := &UserResponse{
		ID:       req.ID,
		Name:     fmt.Sprintf("User %d", req.ID),
		Email:    fmt.Sprintf("user%d@example.com", req.ID),
		Created:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Modified: time.Now(),
		Tags:     []string{"active", "verified", "premium"},
		Metadata: map[string]interface{}{
			"lastLogin":  time.Now().Add(-24 * time.Hour),
			"loginCount": 42,
			"preferences": map[string]interface{}{
				"theme":         "dark",
				"notifications": true,
				"language":      "zh-CN",
			},
		},
	}

	return resp, nil
}

// 启动服务器
func startServer() {
	// 创建Sonic编解码器（使用紧凑输出和HTML转义）
	sonicCodec := sonic.New(
		sonic.WithCompactOutput(true),
		sonic.WithEscapeHTML(true),
	)

	// 创建TCP传输
	transport := tcp.New(
		tcp.WithAddress(":8080"),
		tcp.WithReadTimeout(5*time.Second),
		tcp.WithWriteTimeout(5*time.Second),
	)

	// 创建服务器
	server := thor.NewServer(sonicCodec, transport)

	// 注册服务
	if err := server.Register(new(UserService)); err != nil {
		log.Fatalf("注册服务失败: %v", err)
	}

	// 启动服务器
	fmt.Println("服务器已启动，监听端口 :8080...")
	if err := server.Serve(); err != nil {
		log.Fatalf("服务器错误: %v", err)
	}
}

// 启动客户端
func startClient() {
	// 创建Sonic编解码器（使用紧凑输出和HTML转义）
	sonicCodec := sonic.New(
		sonic.WithCompactOutput(true),
		sonic.WithEscapeHTML(true),
	)

	// 创建TCP传输
	transport := tcp.New(
		tcp.WithTarget("localhost:8080"),
		tcp.WithReadTimeout(5*time.Second),
		tcp.WithWriteTimeout(5*time.Second),
	)

	// 创建客户端
	client := thor.NewClient(sonicCodec, transport)
	defer client.Close()

	// 创建请求上下文
	ctx := context.Background()

	// 发送请求
	req := &UserRequest{ID: 1}
	resp := &UserResponse{}

	fmt.Println("发送请求...")
	if err := client.Call(ctx, "UserService.GetUser", req, resp); err != nil {
		log.Fatalf("调用失败: %v", err)
	}

	// 打印响应
	fmt.Println("收到响应:")
	fmt.Printf("用户ID: %d\n", resp.ID)
	fmt.Printf("用户名: %s\n", resp.Name)
	fmt.Printf("邮箱: %s\n", resp.Email)
	fmt.Printf("创建时间: %s\n", resp.Created)
	fmt.Printf("修改时间: %s\n", resp.Modified)
	fmt.Printf("标签: %v\n", resp.Tags)
	fmt.Printf("元数据: %v\n", resp.Metadata)
}

func main() {
	// 启动服务器 (后台运行)
	go startServer()

	// 等待服务器启动
	time.Sleep(time.Second)

	// 启动客户端
	startClient()
}
