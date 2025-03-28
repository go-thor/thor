# Sonic JSON 编解码器示例

这个示例展示了如何使用 ByteDance Sonic 编解码器在 Thor RPC 框架中构建一个简单的 RPC 服务。

## 功能特点

- 使用 Sonic 作为 JSON 编解码器，提供高性能的 JSON 序列化和反序列化
- 实现了一个简单的用户服务及其 RPC 方法
- 包含服务器和客户端两部分代码，可单独运行
- 演示了 Thor RPC 框架的基本用法

## 如何运行

### 服务端

```bash
go run main.go
```

### 客户端

```bash
go run main.go client
```

## 代码结构

- `main.go` - 包含服务器和客户端的实现
  - `UserService` - 一个简单的服务实现
  - `startServer()` - 启动 RPC 服务器
  - `startClient()` - 启动 RPC 客户端进行测试

## 注意事项

1. 确保先启动服务器，再启动客户端
2. 服务器默认监听 `:8080` 端口
3. 客户端连接到 `localhost:8080`