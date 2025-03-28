package sonic

import (
	"fmt"

	"github.com/bytedance/sonic"
)

// Codec 是基于ByteDance Sonic的高性能JSON编解码器
type Codec struct {
	// 是否使用紧凑输出
	compactOutput bool
	// 是否转义HTML标签
	escapeHTML bool
}

// Option 是Sonic编解码器的配置选项
type Option func(*Codec)

// WithCompactOutput 设置紧凑输出（不包含额外空格）
func WithCompactOutput(compact bool) Option {
	return func(c *Codec) {
		c.compactOutput = compact
	}
}

// WithEscapeHTML 设置是否转义HTML标签
func WithEscapeHTML(escape bool) Option {
	return func(c *Codec) {
		c.escapeHTML = escape
	}
}

// New 创建一个新的Sonic编解码器
func New(options ...Option) *Codec {
	// 默认配置
	c := &Codec{
		compactOutput: true,
		escapeHTML:    true,
	}

	// 应用选项
	for _, option := range options {
		option(c)
	}

	return c
}

// Marshal 将对象编码为JSON格式的字节数组
func (c *Codec) Marshal(v interface{}) ([]byte, error) {
	if c.compactOutput {
		return sonic.ConfigFastest.Marshal(v)
	}

	// 使用标准配置（含缩进）
	return sonic.ConfigDefault.Marshal(v)
}

// Unmarshal 将JSON格式的字节数组解码为对象
func (c *Codec) Unmarshal(data []byte, v interface{}) error {
	if len(data) == 0 {
		return fmt.Errorf("empty data")
	}

	return sonic.ConfigDefault.Unmarshal(data, v)
}

// Name 返回编解码器的名称
func (c *Codec) Name() string {
	return "sonic"
}
