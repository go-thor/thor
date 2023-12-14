package proto

import (
	"github.com/go-thor/thor/codec"
	"google.golang.org/protobuf/proto"
)

type coder struct{}

func NewCoder() codec.Coder {
	return &coder{}
}

func (t coder) String() string {
	return "proto"
}

func (t coder) Marshal(v interface{}) ([]byte, error) {
	return proto.Marshal(v.(proto.Message))
}

func (t coder) Unmarshal(d []byte, v interface{}) error {
	return proto.Unmarshal(d, v.(proto.Message))
}
