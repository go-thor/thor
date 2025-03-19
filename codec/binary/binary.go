package binary

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"

	"github.com/go-thor/thor"
)

const (
	// 魔数，用于标识协议
	MagicNumber uint16 = 0x8274 // 'th' in ASCII

	// 协议版本
	Version uint8 = 0x01

	// 消息类型
	TypeRequest  uint8 = 0x01
	TypeResponse uint8 = 0x02

	// 编码类型
	EncodingRaw      uint8 = 0x00 // 原始二进制
	EncodingJSON     uint8 = 0x01 // JSON
	EncodingProtobuf uint8 = 0x02 // Protobuf
	EncodingMsgpack  uint8 = 0x03 // MessagePack
)

// Codec 是二进制编解码器
type Codec struct {
	// 内部编解码器，用于处理负载的序列化和反序列化
	innerCodec thor.Codec
}

// New 创建一个新的二进制编解码器
func New(innerCodec thor.Codec) *Codec {
	if innerCodec == nil {
		panic("innerCodec cannot be nil")
	}
	return &Codec{
		innerCodec: innerCodec,
	}
}

// 协议头格式 (14 bytes + 变长字段):
// +--------+--------+--------+--------+
// |           Magic Number           |  2 bytes 魔数 (0x8274)
// +--------+--------+--------+--------+
// |            Version                |  1 byte 版本号
// +--------+--------+--------+--------+
// |           Message Type           |  1 byte 消息类型 (请求/响应)
// +--------+--------+--------+--------+
// |            Encoding             |  1 byte 编码类型
// +--------+--------+--------+--------+
// |             Sequence ID           |  8 bytes 序列号
// +--------+--------+--------+--------+
// |         Service Method Length     |  1 byte 服务方法名长度
// +--------+--------+--------+--------+
// |         Service Method         |  变长 服务方法名
// +--------+--------+--------+--------+
// |         Metadata Length         |  4 bytes 元数据长度
// +--------+--------+--------+--------+
// |            Metadata            |  变长 元数据
// +--------+--------+--------+--------+
// |           Payload Length         |  4 bytes 负载长度
// +--------+--------+--------+--------+
// |             Payload             |  变长 负载数据
// +--------+--------+--------+--------+

// Marshal 将对象编码为字节数组
func (c *Codec) Marshal(v interface{}) ([]byte, error) {
	var buf bytes.Buffer

	// 写入协议头
	// 1. 魔数 (2 bytes)
	if err := binary.Write(&buf, binary.BigEndian, MagicNumber); err != nil {
		return nil, fmt.Errorf("write magic number: %w", err)
	}

	// 2. 版本 (1 byte)
	if err := binary.Write(&buf, binary.BigEndian, Version); err != nil {
		return nil, fmt.Errorf("write version: %w", err)
	}

	switch msg := v.(type) {
	case *thor.Request:
		// 3. 消息类型 (1 byte) - 请求
		if err := binary.Write(&buf, binary.BigEndian, TypeRequest); err != nil {
			return nil, fmt.Errorf("write message type: %w", err)
		}

		// 4. 编码类型 (1 byte) - 暂时固定为 EncodingRaw
		if err := binary.Write(&buf, binary.BigEndian, getEncodingType(c.innerCodec)); err != nil {
			return nil, fmt.Errorf("write encoding type: %w", err)
		}

		// 5. 序列号 (8 bytes)
		if err := binary.Write(&buf, binary.BigEndian, msg.Seq); err != nil {
			return nil, fmt.Errorf("write sequence: %w", err)
		}

		// 6. 服务方法名长度 (1 byte)
		methodLen := len(msg.ServiceMethod)
		if methodLen > 255 {
			return nil, fmt.Errorf("service method too long: %d > 255", methodLen)
		}
		if err := binary.Write(&buf, binary.BigEndian, uint8(methodLen)); err != nil {
			return nil, fmt.Errorf("write service method length: %w", err)
		}

		// 7. 服务方法名 (变长)
		if _, err := buf.WriteString(msg.ServiceMethod); err != nil {
			return nil, fmt.Errorf("write service method: %w", err)
		}

		// 8. 元数据
		metadataBytes, err := c.innerCodec.Marshal(msg.Metadata)
		if err != nil {
			return nil, fmt.Errorf("marshal metadata: %w", err)
		}

		// 8.1 元数据长度 (4 bytes)
		if err := binary.Write(&buf, binary.BigEndian, uint32(len(metadataBytes))); err != nil {
			return nil, fmt.Errorf("write metadata length: %w", err)
		}

		// 8.2 元数据内容 (变长)
		if _, err := buf.Write(metadataBytes); err != nil {
			return nil, fmt.Errorf("write metadata: %w", err)
		}

		// 9. 负载
		var payloadBytes []byte
		if len(msg.Payload) > 0 {
			payloadBytes = msg.Payload
		} else if len(msg.Args) > 0 {
			payloadBytes = msg.Args
		}

		// 9.1 负载长度 (4 bytes)
		if err := binary.Write(&buf, binary.BigEndian, uint32(len(payloadBytes))); err != nil {
			return nil, fmt.Errorf("write payload length: %w", err)
		}

		// 9.2 负载内容 (变长)
		if _, err := buf.Write(payloadBytes); err != nil {
			return nil, fmt.Errorf("write payload: %w", err)
		}

	case *thor.Response:
		// 3. 消息类型 (1 byte) - 响应
		if err := binary.Write(&buf, binary.BigEndian, TypeResponse); err != nil {
			return nil, fmt.Errorf("write message type: %w", err)
		}

		// 4. 编码类型 (1 byte)
		if err := binary.Write(&buf, binary.BigEndian, getEncodingType(c.innerCodec)); err != nil {
			return nil, fmt.Errorf("write encoding type: %w", err)
		}

		// 5. 序列号 (8 bytes)
		if err := binary.Write(&buf, binary.BigEndian, msg.Seq); err != nil {
			return nil, fmt.Errorf("write sequence: %w", err)
		}

		// 6. 服务方法名长度 (1 byte)
		methodLen := len(msg.ServiceMethod)
		if methodLen > 255 {
			return nil, fmt.Errorf("service method too long: %d > 255", methodLen)
		}
		if err := binary.Write(&buf, binary.BigEndian, uint8(methodLen)); err != nil {
			return nil, fmt.Errorf("write service method length: %w", err)
		}

		// 7. 服务方法名 (变长)
		if _, err := buf.WriteString(msg.ServiceMethod); err != nil {
			return nil, fmt.Errorf("write service method: %w", err)
		}

		// 8. 元数据
		metadataBytes, err := c.innerCodec.Marshal(msg.Metadata)
		if err != nil {
			return nil, fmt.Errorf("marshal metadata: %w", err)
		}

		// 8.1 元数据长度 (4 bytes)
		if err := binary.Write(&buf, binary.BigEndian, uint32(len(metadataBytes))); err != nil {
			return nil, fmt.Errorf("write metadata length: %w", err)
		}

		// 8.2 元数据内容 (变长)
		if _, err := buf.Write(metadataBytes); err != nil {
			return nil, fmt.Errorf("write metadata: %w", err)
		}

		// 9. 错误信息长度 (2 bytes)
		if err := binary.Write(&buf, binary.BigEndian, uint16(len(msg.Error))); err != nil {
			return nil, fmt.Errorf("write error length: %w", err)
		}

		// 10. 错误信息 (变长)
		if _, err := buf.WriteString(msg.Error); err != nil {
			return nil, fmt.Errorf("write error: %w", err)
		}

		// 11. 负载
		var payloadBytes []byte
		if len(msg.Payload) > 0 {
			payloadBytes = msg.Payload
		} else if len(msg.Reply) > 0 {
			payloadBytes = msg.Reply
		}

		// 11.1 负载长度 (4 bytes)
		if err := binary.Write(&buf, binary.BigEndian, uint32(len(payloadBytes))); err != nil {
			return nil, fmt.Errorf("write payload length: %w", err)
		}

		// 11.2 负载内容 (变长)
		if _, err := buf.Write(payloadBytes); err != nil {
			return nil, fmt.Errorf("write payload: %w", err)
		}

	default:
		// 对于其他类型，使用内部编解码器序列化
		return c.innerCodec.Marshal(v)
	}

	return buf.Bytes(), nil
}

// Unmarshal 将字节数组解码为对象
func (c *Codec) Unmarshal(data []byte, v interface{}) error {
	if len(data) < 13 { // 最小长度: 魔数(2) + 版本(1) + 类型(1) + 编码(1) + 序列号(8)
		return fmt.Errorf("data too short: %d < 13", len(data))
	}

	buf := bytes.NewBuffer(data)

	// 1. 读取魔数 (2 bytes)
	var magic uint16
	if err := binary.Read(buf, binary.BigEndian, &magic); err != nil {
		return fmt.Errorf("read magic number: %w", err)
	}
	if magic != MagicNumber {
		return fmt.Errorf("invalid magic number: 0x%x != 0x%x", magic, MagicNumber)
	}

	// 2. 读取版本 (1 byte)
	var version uint8
	if err := binary.Read(buf, binary.BigEndian, &version); err != nil {
		return fmt.Errorf("read version: %w", err)
	}
	if version != Version {
		return fmt.Errorf("unsupported version: %d != %d", version, Version)
	}

	// 3. 读取消息类型 (1 byte)
	var msgType uint8
	if err := binary.Read(buf, binary.BigEndian, &msgType); err != nil {
		return fmt.Errorf("read message type: %w", err)
	}

	// 4. 读取编码类型 (1 byte)
	var encoding uint8
	if err := binary.Read(buf, binary.BigEndian, &encoding); err != nil {
		return fmt.Errorf("read encoding type: %w", err)
	}

	// 5. 读取序列号 (8 bytes)
	var seq uint64
	if err := binary.Read(buf, binary.BigEndian, &seq); err != nil {
		return fmt.Errorf("read sequence: %w", err)
	}

	// 6. 读取服务方法名长度 (1 byte)
	var methodLen uint8
	if err := binary.Read(buf, binary.BigEndian, &methodLen); err != nil {
		return fmt.Errorf("read service method length: %w", err)
	}

	// 7. 读取服务方法名 (变长)
	methodBytes := make([]byte, methodLen)
	if _, err := io.ReadFull(buf, methodBytes); err != nil {
		return fmt.Errorf("read service method: %w", err)
	}
	serviceMethod := string(methodBytes)

	// 8. 读取元数据长度 (4 bytes)
	var metadataLen uint32
	if err := binary.Read(buf, binary.BigEndian, &metadataLen); err != nil {
		return fmt.Errorf("read metadata length: %w", err)
	}

	// 9. 读取元数据 (变长)
	metadataBytes := make([]byte, metadataLen)
	if _, err := io.ReadFull(buf, metadataBytes); err != nil {
		return fmt.Errorf("read metadata: %w", err)
	}

	metadata := make(map[string]string)
	if metadataLen > 0 {
		if err := c.innerCodec.Unmarshal(metadataBytes, &metadata); err != nil {
			return fmt.Errorf("unmarshal metadata: %w", err)
		}
	}

	switch msgType {
	case TypeRequest:
		req, ok := v.(*thor.Request)
		if !ok {
			return fmt.Errorf("expected *thor.Request, got %T", v)
		}

		// 10. 读取负载长度 (4 bytes)
		var payloadLen uint32
		if err := binary.Read(buf, binary.BigEndian, &payloadLen); err != nil {
			return fmt.Errorf("read payload length: %w", err)
		}

		// 11. 读取负载 (变长)
		payload := make([]byte, payloadLen)
		if _, err := io.ReadFull(buf, payload); err != nil {
			return fmt.Errorf("read payload: %w", err)
		}

		// 填充请求对象
		req.ServiceMethod = serviceMethod
		req.Metadata = metadata
		req.Seq = seq
		req.Payload = payload
		req.Args = payload // 同时设置 Args 和 Payload 确保兼容性

	case TypeResponse:
		resp, ok := v.(*thor.Response)
		if !ok {
			return fmt.Errorf("expected *thor.Response, got %T", v)
		}

		// 10. 读取错误信息长度 (2 bytes)
		var errorLen uint16
		if err := binary.Read(buf, binary.BigEndian, &errorLen); err != nil {
			return fmt.Errorf("read error length: %w", err)
		}

		// 11. 读取错误信息 (变长)
		errorBytes := make([]byte, errorLen)
		if _, err := io.ReadFull(buf, errorBytes); err != nil {
			return fmt.Errorf("read error: %w", err)
		}
		errorMsg := string(errorBytes)

		// 12. 读取负载长度 (4 bytes)
		var payloadLen uint32
		if err := binary.Read(buf, binary.BigEndian, &payloadLen); err != nil {
			return fmt.Errorf("read payload length: %w", err)
		}

		// 13. 读取负载 (变长)
		payload := make([]byte, payloadLen)
		if _, err := io.ReadFull(buf, payload); err != nil {
			return fmt.Errorf("read payload: %w", err)
		}

		// 填充响应对象
		resp.ServiceMethod = serviceMethod
		resp.Metadata = metadata
		resp.Seq = seq
		resp.Error = errorMsg
		resp.Payload = payload
		resp.Reply = payload // 同时设置 Reply 和 Payload 确保兼容性

	default:
		return fmt.Errorf("unsupported message type: %d", msgType)
	}

	return nil
}

// Name 返回编解码器的名称
func (c *Codec) Name() string {
	return "binary-" + c.innerCodec.Name()
}

// getEncodingType 根据内部编解码器获取编码类型
func getEncodingType(codec thor.Codec) uint8 {
	name := codec.Name()
	switch {
	case strings.Contains(name, "json"):
		return EncodingJSON
	case strings.Contains(name, "protobuf"):
		return EncodingProtobuf
	case strings.Contains(name, "msgpack"):
		return EncodingMsgpack
	default:
		return EncodingRaw
	}
}
