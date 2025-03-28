# Sonic JSON 编解码器

`sonic` 包提供了基于 [ByteDance Sonic](https://github.com/bytedance/sonic) 的高性能 JSON 编解码器实现，用于 Thor RPC 框架。

Sonic 是由字节跳动开发的高性能 JSON 库，针对 Go 语言进行了特殊优化，使用 JIT 编译技术，对各种常见 JSON 场景提供了极致的性能。

## 特点

- **极致性能**: 相比标准库更快的序列化和反序列化速度（最高可达 10 倍）
- **完全兼容**: 与标准库 `encoding/json` 完全兼容
- **JIT 加速**: 利用 JIT 编译技术实现高性能序列化和反序列化
- **高可靠性**: 经过字节跳动内部大量服务验证，可靠性和稳定性极高

## 性能对比

与标准库 JSON 编解码器相比，Sonic 具有显著性能优势：

- **序列化速度**: 平均快 3-10 倍
- **反序列化速度**: 平均快 3-10 倍
- **内存使用**: 更高效的内存管理

## 使用方法

### 基本用法

```go
import (
    "github.com/go-thor/thor"
    "github.com/go-thor/thor/codec/sonic"
    "github.com/go-thor/thor/transport/tcp"
)

// 创建 Sonic 编解码器
sonicCodec := sonic.New()

// 创建 TCP 传输
transport := tcp.New(
    tcp.WithAddress(":8080"),
    tcp.WithReadTimeout(5*time.Second),
    tcp.WithWriteTimeout(5*time.Second),
)

// 创建服务器
server := thor.NewServer(sonicCodec, transport)

// 注册服务
server.Register(new(YourService))

// 启动服务器
server.Serve()
```

### 配置选项

Sonic 编解码器提供了几个配置选项：

```go
// 使用紧凑输出模式（没有缩进和空白字符）
sonicCodec := sonic.New(sonic.WithCompactOutput(true))

// 配置 HTML 转义行为
sonicCodec := sonic.New(sonic.WithEscapeHTML(false))

// 组合多个选项
sonicCodec := sonic.New(
    sonic.WithCompactOutput(true),
    sonic.WithEscapeHTML(true),
)
```

### 客户端用法

```go
// 创建 Sonic 编解码器
sonicCodec := sonic.New()

// 创建 TCP 传输
transport := tcp.New(
    tcp.WithTarget("localhost:8080"),
    tcp.WithReadTimeout(5*time.Second),
    tcp.WithWriteTimeout(5*time.Second),
)

// 创建客户端
client := thor.NewClient(sonicCodec, transport)

// 发送请求
var reply YourReplyType
err := client.Call(ctx, "ServiceName.MethodName", request, &reply)
```

## 实现细节

Sonic 编解码器在内部利用了 ByteDance Sonic 库的高性能特性：

1. 使用 JIT 编译技术加速 JSON 序列化和反序列化
2. 采用了并行处理策略，充分利用多核 CPU 性能
3. 针对大型 JSON 结构进行了特殊优化
4. 支持流式处理，减少内存消耗

## 适用场景

Sonic 编解码器特别适合以下场景：

- 高性能 API 服务
- 大量 JSON 序列化/反序列化操作
- 要求极低延迟的实时服务
- 大规模微服务架构中的服务间通信

## 注意事项

1. Sonic 编解码器要求 Go 1.16 或更高版本
2. 目前仅支持 AMD64 架构，暂不支持 ARM 等其他架构
3. 为了最大化性能，建议在生产环境中使用 `WithCompactOutput(true)` 选项
4. Sonic 完全兼容 `encoding/json` 的标签和行为