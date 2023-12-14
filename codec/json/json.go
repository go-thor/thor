package json

import (
	"encoding/json"

	"github.com/go-thor/thor/codec"
	jsonIter "github.com/json-iterator/go"
)

type coder struct{}

var (
	jsonIterMarshler = jsonIter.ConfigCompatibleWithStandardLibrary
)

func NewCoder() codec.Coder {
	return &coder{}
}

func (j coder) String() string {
	return "json"
}

func (coder) Marshal(v interface{}) ([]byte, error) {
	switch m := v.(type) {
	case json.Marshaler:
		return m.MarshalJSON()
	default:
		return jsonIterMarshler.Marshal(m)
	}
}

func (coder) Unmarshal(data []byte, v interface{}) error {
	switch m := v.(type) {
	case json.Unmarshaler:
		return m.UnmarshalJSON(data)
	default:
		return jsonIterMarshler.Unmarshal(data, m)
	}
}
