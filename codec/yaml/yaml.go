package yaml

import (
	"github.com/ghodss/yaml"
	"github.com/go-thor/thor/codec"
)

type coder struct{}

func NewCoder() codec.Coder {
	return &coder{}
}

func (y coder) String() string {
	return "yaml"
}

func (y coder) Marshal(v interface{}) ([]byte, error) {
	return yaml.Marshal(v)
}

func (y coder) Unmarshal(d []byte, v interface{}) error {
	return yaml.Unmarshal(d, v)
}
