package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"time"

	"github.com/go-thor/thor"
	"github.com/go-thor/thor/codec/protobuf"
	pb "github.com/go-thor/thor/examples/protobuf_protocol/proto"
	"github.com/go-thor/thor/transport/tcp"
	"google.golang.org/protobuf/proto"
)

// ProtobufService 是一个示例服务，处理用户请求
type ProtobufService struct{}

// ProcessUser 处理用户请求并返回响应
func (s *ProtobufService) ProcessUser(ctx context.Context, req *pb.UserRequest) (*pb.UserResponse, error) {
	log.Printf("收到用户请求: ID=%d, 用户名=%s, 邮箱=%s", req.Id, req.Username, req.Email)

	// 创建响应
	resp := &pb.UserResponse{
		Id:        req.Id,
		Username:  req.Username,
		Email:     req.Email,
		CreateAt:  req.CreateAt,
		UpdatedAt: time.Now().Format(time.RFC3339),
		Tags:      req.Tags,
		Active:    req.Active,
		Message:   "处理成功",
		Properties: map[string]string{
			"last_login": time.Now().Format(time.RFC3339),
			"ip":         "127.0.0.1",
			"user_agent": "Mozilla/5.0",
		},
	}

	log.Printf("发送用户响应: ID=%d, 消息=%s", resp.Id, resp.Message)
	return resp, nil
}

// CustomProtobufCodec 是一个自定义的 Protobuf 编解码器，用于处理 thor.Request 和 thor.Response
type CustomProtobufCodec struct {
	protobufCodec *protobuf.Codec
}

// NewCustomProtobufCodec 创建一个新的自定义 Protobuf 编解码器
func NewCustomProtobufCodec() *CustomProtobufCodec {
	return &CustomProtobufCodec{
		protobufCodec: protobuf.New(),
	}
}

// Marshal 将对象序列化为二进制数据
func (c *CustomProtobufCodec) Marshal(v interface{}) ([]byte, error) {
	switch msg := v.(type) {
	case *thor.Request:
		// 处理请求
		// 将元数据转换为 Protobuf 消息
		metadata := &pb.Metadata{
			Data: msg.Metadata,
		}
		metadataBytes, err := proto.Marshal(metadata)
		if err != nil {
			return nil, fmt.Errorf("marshal metadata: %w", err)
		}

		// 构建请求内容
		var buf bytes.Buffer

		// 1. 服务方法名长度 (1 byte)
		methodLen := len(msg.ServiceMethod)
		if methodLen > 255 {
			return nil, fmt.Errorf("service method too long: %d > 255", methodLen)
		}
		if err := binary.Write(&buf, binary.BigEndian, uint8(methodLen)); err != nil {
			return nil, fmt.Errorf("write service method length: %w", err)
		}

		// 2. 服务方法名 (变长)
		if _, err := buf.WriteString(msg.ServiceMethod); err != nil {
			return nil, fmt.Errorf("write service method: %w", err)
		}

		// 3. 序列号 (8 bytes)
		if err := binary.Write(&buf, binary.BigEndian, msg.Seq); err != nil {
			return nil, fmt.Errorf("write sequence: %w", err)
		}

		// 4. 元数据长度 (4 bytes)
		if err := binary.Write(&buf, binary.BigEndian, uint32(len(metadataBytes))); err != nil {
			return nil, fmt.Errorf("write metadata length: %w", err)
		}

		// 5. 元数据 (变长)
		if _, err := buf.Write(metadataBytes); err != nil {
			return nil, fmt.Errorf("write metadata: %w", err)
		}

		// 6. 准备负载数据
		var payloadBytes []byte

		// 使用已有的负载数据
		if len(msg.Payload) > 0 {
			payloadBytes = msg.Payload
		} else if len(msg.Args) > 0 {
			// 使用已设置的Args
			payloadBytes = msg.Args
		} else {
			// 没有有效负载，使用空字节数组
			payloadBytes = []byte{}
		}

		// 7. 负载长度 (4 bytes)
		if err := binary.Write(&buf, binary.BigEndian, uint32(len(payloadBytes))); err != nil {
			return nil, fmt.Errorf("write payload length: %w", err)
		}

		// 8. 负载内容 (变长)
		if _, err := buf.Write(payloadBytes); err != nil {
			return nil, fmt.Errorf("write payload: %w", err)
		}

		return buf.Bytes(), nil

	case *thor.Response:
		// 处理响应
		// 将元数据转换为 Protobuf 消息
		metadata := &pb.Metadata{
			Data: msg.Metadata,
		}
		metadataBytes, err := proto.Marshal(metadata)
		if err != nil {
			return nil, fmt.Errorf("marshal metadata: %w", err)
		}

		// 构建响应内容
		var buf bytes.Buffer

		// 1. 服务方法名长度 (1 byte)
		methodLen := len(msg.ServiceMethod)
		if methodLen > 255 {
			return nil, fmt.Errorf("service method too long: %d > 255", methodLen)
		}
		if err := binary.Write(&buf, binary.BigEndian, uint8(methodLen)); err != nil {
			return nil, fmt.Errorf("write service method length: %w", err)
		}

		// 2. 服务方法名 (变长)
		if _, err := buf.WriteString(msg.ServiceMethod); err != nil {
			return nil, fmt.Errorf("write service method: %w", err)
		}

		// 3. 序列号 (8 bytes)
		if err := binary.Write(&buf, binary.BigEndian, msg.Seq); err != nil {
			return nil, fmt.Errorf("write sequence: %w", err)
		}

		// 4. 元数据长度 (4 bytes)
		if err := binary.Write(&buf, binary.BigEndian, uint32(len(metadataBytes))); err != nil {
			return nil, fmt.Errorf("write metadata length: %w", err)
		}

		// 5. 元数据 (变长)
		if _, err := buf.Write(metadataBytes); err != nil {
			return nil, fmt.Errorf("write metadata: %w", err)
		}

		// 6. 错误信息长度 (2 bytes)
		if err := binary.Write(&buf, binary.BigEndian, uint16(len(msg.Error))); err != nil {
			return nil, fmt.Errorf("write error length: %w", err)
		}

		// 7. 错误信息 (变长)
		if _, err := buf.WriteString(msg.Error); err != nil {
			return nil, fmt.Errorf("write error: %w", err)
		}

		// 8. 准备负载数据
		var payloadBytes []byte

		// 使用已有的负载数据
		if len(msg.Payload) > 0 {
			payloadBytes = msg.Payload
		} else if len(msg.Reply) > 0 {
			// 使用已设置的Reply
			payloadBytes = msg.Reply
		} else {
			// 没有有效负载，使用空字节数组
			payloadBytes = []byte{}
		}

		// 9. 负载长度 (4 bytes)
		if err := binary.Write(&buf, binary.BigEndian, uint32(len(payloadBytes))); err != nil {
			return nil, fmt.Errorf("write payload length: %w", err)
		}

		// 10. 负载内容 (变长)
		if _, err := buf.Write(payloadBytes); err != nil {
			return nil, fmt.Errorf("write payload: %w", err)
		}

		return buf.Bytes(), nil

	default:
		// 对于其他类型，直接使用protobuf编解码器
		if protoMsg, ok := v.(proto.Message); ok {
			return c.protobufCodec.Marshal(protoMsg)
		}
		return nil, fmt.Errorf("type %T is not supported", v)
	}
}

// Unmarshal 将二进制数据反序列化为对象
func (c *CustomProtobufCodec) Unmarshal(data []byte, v interface{}) error {
	switch msg := v.(type) {
	case *thor.Request:
		// 处理请求
		if len(data) < 1 {
			return fmt.Errorf("data too short: %d < 1", len(data))
		}

		buf := bytes.NewReader(data)

		// 1. 读取服务方法名长度 (1 byte)
		var methodLen uint8
		if err := binary.Read(buf, binary.BigEndian, &methodLen); err != nil {
			return fmt.Errorf("read service method length: %w", err)
		}

		// 2. 读取服务方法名 (变长)
		methodBytes := make([]byte, methodLen)
		if _, err := buf.Read(methodBytes); err != nil {
			return fmt.Errorf("read service method: %w", err)
		}
		msg.ServiceMethod = string(methodBytes)

		// 3. 读取序列号 (8 bytes)
		if err := binary.Read(buf, binary.BigEndian, &msg.Seq); err != nil {
			return fmt.Errorf("read sequence: %w", err)
		}

		// 4. 读取元数据长度 (4 bytes)
		var metadataLen uint32
		if err := binary.Read(buf, binary.BigEndian, &metadataLen); err != nil {
			return fmt.Errorf("read metadata length: %w", err)
		}

		// 5. 读取元数据 (变长)
		metadataBytes := make([]byte, metadataLen)
		if _, err := buf.Read(metadataBytes); err != nil {
			return fmt.Errorf("read metadata: %w", err)
		}

		// 解析元数据
		if metadataLen > 0 {
			metadata := &pb.Metadata{}
			if err := proto.Unmarshal(metadataBytes, metadata); err != nil {
				return fmt.Errorf("unmarshal metadata: %w", err)
			}
			msg.Metadata = metadata.Data
		} else {
			msg.Metadata = make(map[string]string)
		}

		// 6. 读取负载长度 (4 bytes)
		var payloadLen uint32
		if err := binary.Read(buf, binary.BigEndian, &payloadLen); err != nil {
			return fmt.Errorf("read payload length: %w", err)
		}

		// 7. 读取负载 (变长)
		payloadBytes := make([]byte, payloadLen)
		if _, err := buf.Read(payloadBytes); err != nil {
			return fmt.Errorf("read payload: %w", err)
		}

		// 存储原始负载
		msg.Payload = payloadBytes
		msg.Args = payloadBytes

		return nil

	case *thor.Response:
		// 处理响应
		if len(data) < 1 {
			return fmt.Errorf("data too short: %d < 1", len(data))
		}

		buf := bytes.NewReader(data)

		// 1. 读取服务方法名长度 (1 byte)
		var methodLen uint8
		if err := binary.Read(buf, binary.BigEndian, &methodLen); err != nil {
			return fmt.Errorf("read service method length: %w", err)
		}

		// 2. 读取服务方法名 (变长)
		methodBytes := make([]byte, methodLen)
		if _, err := buf.Read(methodBytes); err != nil {
			return fmt.Errorf("read service method: %w", err)
		}
		msg.ServiceMethod = string(methodBytes)

		// 3. 读取序列号 (8 bytes)
		if err := binary.Read(buf, binary.BigEndian, &msg.Seq); err != nil {
			return fmt.Errorf("read sequence: %w", err)
		}

		// 4. 读取元数据长度 (4 bytes)
		var metadataLen uint32
		if err := binary.Read(buf, binary.BigEndian, &metadataLen); err != nil {
			return fmt.Errorf("read metadata length: %w", err)
		}

		// 5. 读取元数据 (变长)
		metadataBytes := make([]byte, metadataLen)
		if _, err := buf.Read(metadataBytes); err != nil {
			return fmt.Errorf("read metadata: %w", err)
		}

		// 解析元数据
		if metadataLen > 0 {
			metadata := &pb.Metadata{}
			if err := proto.Unmarshal(metadataBytes, metadata); err != nil {
				return fmt.Errorf("unmarshal metadata: %w", err)
			}
			msg.Metadata = metadata.Data
		} else {
			msg.Metadata = make(map[string]string)
		}

		// 6. 读取错误信息长度 (2 bytes)
		var errorLen uint16
		if err := binary.Read(buf, binary.BigEndian, &errorLen); err != nil {
			return fmt.Errorf("read error length: %w", err)
		}

		// 7. 读取错误信息 (变长)
		errorBytes := make([]byte, errorLen)
		if _, err := buf.Read(errorBytes); err != nil {
			return fmt.Errorf("read error: %w", err)
		}
		msg.Error = string(errorBytes)

		// 8. 读取负载长度 (4 bytes)
		var payloadLen uint32
		if err := binary.Read(buf, binary.BigEndian, &payloadLen); err != nil {
			return fmt.Errorf("read payload length: %w", err)
		}

		// 9. 读取负载 (变长)
		payloadBytes := make([]byte, payloadLen)
		if _, err := buf.Read(payloadBytes); err != nil {
			return fmt.Errorf("read payload: %w", err)
		}

		// 存储原始负载
		msg.Payload = payloadBytes
		msg.Reply = payloadBytes

		return nil

	default:
		// 对于其他类型，直接使用protobuf编解码器
		if _, ok := v.(proto.Message); ok {
			return c.protobufCodec.Unmarshal(data, v)
		}
		return fmt.Errorf("type %T is not supported", v)
	}
}

// Name 返回编解码器的名称
func (c *CustomProtobufCodec) Name() string {
	return "custom_protobuf"
}

// 启动服务器
func startServer() {
	// 创建自定义的Protobuf编解码器
	codec := NewCustomProtobufCodec()

	// 创建TCP传输
	transport := tcp.New(
		tcp.WithAddress(":8799"),
	)

	// 创建服务器
	server := thor.NewServer(codec, transport)

	// 注册服务
	if err := server.Register(new(ProtobufService)); err != nil {
		log.Fatalf("注册服务失败: %v", err)
	}

	// 启动服务器
	log.Println("Protobuf服务器启动在端口 8799...")
	if err := server.Serve(); err != nil {
		log.Fatalf("服务器错误: %v", err)
	}
}

// 启动客户端
func startClient() {
	// 等待服务器启动
	time.Sleep(time.Second)

	// 创建自定义的Protobuf编解码器
	codec := NewCustomProtobufCodec()

	// 创建TCP传输
	transport := tcp.New(
		tcp.WithTarget("localhost:8799"),
	)

	// 创建客户端
	client := thor.NewClient(codec, transport)

	// 创建请求
	req := &pb.UserRequest{
		Id:       1001,
		Username: "张三",
		Email:    "zhangsan@example.com",
		CreateAt: time.Now().Format(time.RFC3339),
		Tags:     []string{"VIP", "新用户"},
		Active:   true,
	}

	// 创建响应对象
	resp := &pb.UserResponse{}

	// 添加元数据
	metadata := map[string]string{
		"client_version": "1.0.0",
		"client_id":      "test-client",
	}

	// 发送请求
	log.Println("发送Protobuf请求...")
	err := client.CallWithMetadata(context.Background(), "ProtobufService.ProcessUser", req, resp, metadata)
	if err != nil {
		log.Fatalf("客户端错误: %v", err)
	}

	// 输出响应
	log.Printf("收到响应: ID=%d, 用户名=%s, 消息=%s", resp.Id, resp.Username, resp.Message)
	log.Printf("属性: %v", resp.Properties)
}

func main() {
	// 启动服务器 (后台运行)
	go startServer()

	// 启动客户端
	startClient()

	fmt.Println("示例运行完成")
}
