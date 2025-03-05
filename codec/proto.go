package codec

import (
	"bytes"
	"io"

	"github.com/go-thor/thor"
	"google.golang.org/protobuf/proto"
)

type ProtoCodec struct {
	buffer bytes.Buffer
}

func NewProtoCodec() thor.Codec {
	return &ProtoCodec{}
}

func (c *ProtoCodec) Encode(v interface{}) ([]byte, error) {
	msg, ok := v.(proto.Message)
	if !ok {
		return nil, thor.Errorf("value is not a proto.Message")
	}
	return proto.Marshal(msg)
}

func (c *ProtoCodec) Decode(data []byte, v interface{}) error {
	msg, ok := v.(proto.Message)
	if !ok {
		return thor.Errorf("value is not a proto.Message")
	}
	return proto.Unmarshal(data, msg)
}

func (c *ProtoCodec) EncodeStream(w io.Writer, v interface{}) error {
	msg, ok := v.(proto.Message)
	if !ok {
		return thor.Errorf("value is not a proto.Message")
	}
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func (c *ProtoCodec) DecodeStream(r io.Reader, v interface{}) error {
	msg, ok := v.(proto.Message)
	if !ok {
		return thor.Errorf("value is not a proto.Message")
	}
	c.buffer.Reset()
	_, err := io.Copy(&c.buffer, r)
	if err != nil {
		return err
	}
	return proto.Unmarshal(c.buffer.Bytes(), msg)
}
