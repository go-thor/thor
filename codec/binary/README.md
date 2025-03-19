# Thor 二进制协议

本文档描述了 Thor RPC 框架的二进制协议设计及实现。

## 协议格式

二进制协议采用 TLV (Type-Length-Value) 格式设计，具有以下特点：

1. 固定长度的协议头，包含协议标识、版本信息、消息类型等
2. 优化的内存对齐结构，提高读写效率
3. 变长字段使用长度+内容的格式，并添加填充实现对齐
4. 支持元数据传输
5. 支持心跳消息
6. 支持二进制负载
7. 支持多种压缩算法

### 协议头格式

```
+--------+--------+--------+--------+
|  Magic Number   |    Version     |  2 bytes (1+1)
+--------+--------+--------+--------+
| Message Type    |    Flags       |  2 bytes (1+1) - 消息类型+标志
+--------+--------+--------+--------+
|             Sequence ID          |  8 bytes - 序列号(aligned)
+--------+--------+--------+--------+
```

其中：
- Magic Number: 1 字节魔数 (0x74，'t' 的 ASCII 码)
- Version: 1 字节版本号
- Message Type: 1 字节消息类型
- Flags: 1 字节标志位，包含编码类型（低4位）和压缩类型（高4位）
- Sequence ID: 8 字节序列号，8字节对齐

### 请求消息格式

请求消息在协议头之后，包含：

```
+--------+--------+--------+--------+
|      Service Method Length       |  2 bytes - 服务方法名长度(aligned)
+--------+--------+--------+--------+
|          Service Method          |  变长 - 服务方法名(padded to 4 bytes)
+--------+--------+--------+--------+
|         Metadata Length          |  4 bytes - 元数据长度(aligned)
+--------+--------+--------+--------+
|            Metadata              |  变长 - 元数据(padded to 8 bytes)
+--------+--------+--------+--------+
|           Payload Length         |  4 bytes - 负载长度(aligned)
+--------+--------+--------+--------+
|             Payload              |  变长 - 负载数据
+--------+--------+--------+--------+
```

### 响应消息格式

响应消息在协议头之后，包含：

```
+--------+--------+--------+--------+
|      Service Method Length       |  2 bytes - 服务方法名长度(aligned)
+--------+--------+--------+--------+
|          Service Method          |  变长 - 服务方法名(padded to 4 bytes)
+--------+--------+--------+--------+
|         Metadata Length          |  4 bytes - 元数据长度(aligned)
+--------+--------+--------+--------+
|            Metadata              |  变长 - 元数据(padded to 8 bytes)
+--------+--------+--------+--------+
|           Error Length           |  2 bytes - 错误信息长度(aligned)
+--------+--------+--------+--------+
|             Error                |  变长 - 错误信息(padded to 4 bytes)
+--------+--------+--------+--------+
|           Payload Length         |  4 bytes - 负载长度(aligned)
+--------+--------+--------+--------+
|             Payload              |  变长 - 负载数据
+--------+--------+--------+--------+
```

### 心跳消息格式

心跳消息在协议头之后，包含：

```
+--------+--------+--------+--------+
|            Timestamp             |  8 bytes - 时间戳(aligned)
+--------+--------+--------+--------+
```

## 常量定义

```go
const (
    // 魔数，用于标识协议
    MagicNumber uint8 = 0x74 // 't' in ASCII

    // 协议版本
    Version uint8 = 0x01

    // 消息类型
    TypeRequest   uint8 = 0x01
    TypeResponse  uint8 = 0x02
    TypeHeartbeat uint8 = 0x03 // 心跳类型

    // 编码类型（低4位）
    EncodingRaw      uint8 = 0x00 // 原始二进制
    EncodingJSON     uint8 = 0x01 // JSON
    EncodingProtobuf uint8 = 0x02 // Protobuf
    EncodingMsgpack  uint8 = 0x03 // MessagePack
    EncodingCustom   uint8 = 0x0F // 自定义编码
    EncodingMask     uint8 = 0x0F // 编码掩码

    // 压缩标志（高4位）
    CompressNone   uint8 = 0x00 // 不压缩
    CompressGzip   uint8 = 0x10 // Gzip压缩
    CompressSnappy uint8 = 0x20 // Snappy压缩
    CompressLZ4    uint8 = 0x30 // LZ4压缩
    CompressMask   uint8 = 0xF0 // 压缩掩码
)
```

## 使用方式

二进制编解码器需要一个内部编解码器来处理负载和元数据的序列化，如：

```go
// 创建内部编解码器 (用于序列化/反序列化负载)
jsonCodec := jsoncodec.New()

// 创建二进制编解码器 (用于处理整个消息)
binaryCodec := binary.New(jsonCodec)

// 设置压缩算法（可选）
binaryCodec.SetCompression(binary.CompressGzip)

// 设置自定义编解码器（可选）
binaryCodec.SetCustomCodec(customCodec)

// 创建TCP传输
transport := tcp.New(
    tcp.WithAddress(":8888"),
    tcp.WithReadTimeout(5*time.Second),
    tcp.WithWriteTimeout(5*time.Second),
)

// 创建服务器或客户端
server := thor.NewServer(binaryCodec, transport)
// 或
client := thor.NewClient(binaryCodec, transport)
```

## 特性

1. **高效的二进制传输**：相比 JSON 等文本格式，二进制格式更紧凑，传输更高效
2. **内存对齐优化**：所有字段按照4/8字节边界对齐，提高访问效率
3. **协议自描述**：通过魔数和版本号可以进行协议识别和版本兼容性检查
4. **安全检查**：包含长度检查，防止缓冲区溢出
5. **心跳支持**：提供心跳消息类型，支持连接保活
6. **支持大型负载**：负载长度使用 4 字节无符号整数，最大支持 4GB 数据
7. **压缩支持**：内置多种压缩算法，自动处理压缩和解压缩
8. **自定义编码**：支持注册自定义编解码器
9. **可扩展**：预留标志位，便于功能扩展

## 限制

1. 服务方法名长度不能超过 65535 字符
2. 错误信息长度不能超过 65535 字符
3. 默认使用内部编解码器对元数据进行序列化
4. 压缩和自定义编码功能需要额外依赖库支持

## 性能优化建议

1. **内存对齐**：所有字段都进行了内存对齐，减少CPU访问开销
2. **字段填充**：变长字段添加填充，确保后续字段都在对齐边界上
3. **选择性压缩**：小数据不建议开启压缩，可能会导致数据膨胀
4. **自定义编码**：对于特定业务数据，可以实现自定义编码器提高性能
5. **对象池**：在高并发场景下，使用对象池减少内存分配
6. **直接内存**：考虑使用unsafe包直接操作内存，进一步提高性能 