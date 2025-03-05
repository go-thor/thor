package codec

import (
	"encoding/json"
	"io"
	"github.com/go-thor/thor"
)

type JSONCodec struct{}

func NewJSONCodec() thor.Codec {
	return &JSONCodec{}
}

func (c *JSONCodec) Encode(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (c *JSONCodec) Decode(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (c *JSONCodec) EncodeStream(w io.Writer, v interface{}) error {
	enc := json.NewEncoder(w)
	return enc.Encode(v)
}

func (c *JSONCodec) DecodeStream(r io.Reader, v interface{}) error {
	dec := json.NewDecoder(r)
	return dec.Decode(v)
}
