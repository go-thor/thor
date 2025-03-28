package msgpack

import (
	"bytes"
	"fmt"

	"github.com/vmihailenco/msgpack/v5"
)

// Codec 是 MessagePack 编解码器
type Codec struct{}

// New 创建一个新的 MessagePack 编解码器
func New() *Codec {
	return &Codec{}
}

// Marshal 将对象编码为 MessagePack 格式的字节数组
func (c *Codec) Marshal(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	enc.SetCustomStructTag("json") // 使用 json 标签以保持与 JSON 编解码器的兼容性

	if err := enc.Encode(v); err != nil {
		return nil, fmt.Errorf("msgpack encode: %w", err)
	}

	return buf.Bytes(), nil
}

// Unmarshal 将 MessagePack 格式的字节数组解码为对象
func (c *Codec) Unmarshal(data []byte, v interface{}) error {
	dec := msgpack.NewDecoder(bytes.NewReader(data))
	dec.SetCustomStructTag("json") // 使用 json 标签以保持与 JSON 编解码器的兼容性

	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("msgpack decode: %w", err)
	}

	return nil
}

// Name 返回编解码器的名称
func (c *Codec) Name() string {
	return "msgpack"
}
