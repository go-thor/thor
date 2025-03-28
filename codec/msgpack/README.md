# MessagePack 编解码器

`msgpack` 包提供了 [MessagePack](https://msgpack.org/) 格式的编解码器实现，用于 thor RPC 框架。

MessagePack 是一种高效的二进制序列化格式，类似于 JSON，但更小更快。它可以在不同语言之间交换数据，同时占用更少的空间。

## 特点

- **体积小**：MessagePack 数据相比 JSON 体积平均减少 30-40%
- **速度快**：序列化和反序列化性能优于 JSON
- **类型丰富**：支持整数、浮点数、字符串、数组、映射等类型
- **兼容 JSON 标签**：可以利用现有的 JSON 标签进行序列化

## 性能对比

与 JSON 编解码器相比，MessagePack 有以下性能优势：

- **序列化速度**：快约 2 倍
- **反序列化速度**：快约 3 倍
- **数据大小**：减少约 30-35%

## 使用方法

### 基本用法

```go
import (
    "github.com/go-thor/thor"
    "github.com/go-thor/thor/codec/binary"
    "github.com/go-thor/thor/codec/msgpack"
    "github.com/go-thor/thor/transport/tcp"
)

// 创建 MessagePack 编解码器
msgpackCodec := msgpack.New()

// 创建二进制协议编解码器，使用 MessagePack 作为内部编解码器
binaryCodec := binary.New(msgpackCodec)

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

### 客户端用法

```go
// 创建 MessagePack 编解码器
msgpackCodec := msgpack.New()

// 创建二进制协议编解码器，使用 MessagePack 作为内部编解码器
binaryCodec := binary.New(msgpackCodec)

// 创建 TCP 传输
transport := tcp.New(
    tcp.WithTarget("localhost:8080"),
    tcp.WithReadTimeout(5*time.Second),
    tcp.WithWriteTimeout(5*time.Second),
)

// 创建客户端
client := thor.NewClient(binaryCodec, transport)

// 发送请求
var reply YourReplyType
err := client.Call(ctx, "ServiceName.MethodName", request, &reply)
```

## 注意事项

1. MessagePack 编解码器是使用 `github.com/vmihailenco/msgpack/v5` 库实现的
2. 默认使用 `json` 标签，与 JSON 编解码器保持兼容
3. 对于需要跨语言通信的场景，MessagePack 是一个不错的选择
4. 如果只是在 Go 服务之间通信，可以考虑使用 Protobuf 获得更好的性能和类型安全