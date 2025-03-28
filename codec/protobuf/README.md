# Protobuf 编解码器

`protobuf` 包提供了 [Protocol Buffers](https://developers.google.com/protocol-buffers) 格式的编解码器实现，用于 thor RPC 框架。

Protocol Buffers 是一种由 Google 开发的、高效、结构化的二进制序列化格式，用于在不同系统间交换数据。

## 特点

- **高效率**：Protobuf 提供比 JSON 和 XML 更小的数据体积和更快的解析速度
- **强类型**：结构化的 Schema 定义，带来更可靠的类型安全
- **向前兼容**：支持字段添加和删除，不会破坏已有的序列化/反序列化逻辑
- **多语言支持**：可生成多种编程语言的代码，实现跨语言通信

## 性能对比

基于基准测试结果，Protobuf 与 JSON 和 MessagePack 相比：

- **序列化速度**：
  - Protobuf 比 JSON 快约 5%
  - MessagePack 比 Protobuf 快约 90%

- **反序列化速度**：
  - Protobuf 比 JSON 快约 70%
  - MessagePack 比 Protobuf 快约 60%

- **数据大小**：
  - Protobuf 比 JSON 小约 34%
  - Protobuf 比 MessagePack 小约 2%

## 使用方法

### 1. 定义消息结构

首先，需要创建 `.proto` 文件定义消息结构：

```protobuf
syntax = "proto3";

package example;
option go_package = "github.com/your-username/your-repo/example";

message Person {
  string name = 1;
  int32 id = 2;
  string email = 3;
  repeated string phone_numbers = 4;
}
```

### 2. 生成 Go 代码

```bash
protoc --go_out=. --go_opt=paths=source_relative path/to/your.proto
```

### 3. 服务端用法

```go
import (
    "github.com/go-thor/thor"
    "github.com/go-thor/thor/codec/binary"
    "github.com/go-thor/thor/codec/protobuf"
    "github.com/go-thor/thor/transport/tcp"
)

// 创建 Protobuf 编解码器
protobufCodec := protobuf.New()

// 创建二进制协议编解码器，使用 Protobuf 作为内部编解码器
binaryCodec := binary.New(protobufCodec)

// 创建 TCP 传输
transport := tcp.New(
    tcp.WithAddress(":8080"),
    tcp.WithReadTimeout(5*time.Second),
    tcp.WithWriteTimeout(5*time.Second),
)

// 创建服务器
server := thor.NewServer(binaryCodec, transport)

// 注册服务
server.Register(new(YourService))

// 启动服务器
server.Serve()
```

### 4. 客户端用法

```go
// 创建 Protobuf 编解码器
protobufCodec := protobuf.New()

// 创建二进制协议编解码器，使用 Protobuf 作为内部编解码器
binaryCodec := binary.New(protobufCodec)

// 创建 TCP 传输
transport := tcp.New(
    tcp.WithTarget("localhost:8080"),
    tcp.WithReadTimeout(5*time.Second),
    tcp.WithWriteTimeout(5*time.Second),
)

// 创建客户端
client := thor.NewClient(binaryCodec, transport)

// 准备请求和响应对象
req := &example.Person{
    Name: "John Doe",
    Id: 1234,
    Email: "john.doe@example.com",
    PhoneNumbers: []string{"123-456-7890"},
}
var resp YourResponseProtoType

// 发送请求
err := client.Call(ctx, "ServiceName.MethodName", req, &resp)
```

## 注意事项

1. 使用 Protobuf 编解码器需要提前定义 `.proto` 文件并生成相应的 Go 代码
2. 只有实现了 `proto.Message` 接口的类型才能使用 Protobuf 编解码器序列化/反序列化
3. 对于需要高性能和强类型安全的服务间通信，Protobuf 是首选
4. 对于跨语言的通信场景，Protobuf 提供了良好的支持和丰富的文档