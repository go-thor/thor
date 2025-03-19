package protobuf

import (
	"fmt"

	"google.golang.org/protobuf/proto"
)

// Codec is a protobuf codec
type Codec struct{}

// New creates a new protobuf codec
func New() *Codec {
	return &Codec{}
}

// Marshal marshals v into bytes using protobuf
func (c *Codec) Marshal(v interface{}) ([]byte, error) {
	if message, ok := v.(proto.Message); ok {
		return proto.Marshal(message)
	}
	return nil, fmt.Errorf("type %T is not proto.Message", v)
}

// Unmarshal unmarshals data into v using protobuf
func (c *Codec) Unmarshal(data []byte, v interface{}) error {
	if message, ok := v.(proto.Message); ok {
		return proto.Unmarshal(data, message)
	}
	return fmt.Errorf("type %T is not proto.Message", v)
}

// Name returns the name of the codec
func (c *Codec) Name() string {
	return "protobuf"
}
