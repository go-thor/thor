package json

import (
	"encoding/json"
	"fmt"
)

// Codec is a JSON codec
type Codec struct{}

// New creates a new JSON codec
func New() *Codec {
	return &Codec{}
}

// Marshal marshals v into bytes using JSON
func (c *Codec) Marshal(v interface{}) ([]byte, error) {
	if v == nil {
		return nil, fmt.Errorf("cannot marshal nil interface")
	}
	return json.Marshal(v)
}

// Unmarshal unmarshals data into v using JSON
func (c *Codec) Unmarshal(data []byte, v interface{}) error {
	if v == nil {
		return fmt.Errorf("cannot unmarshal into nil interface")
	}
	return json.Unmarshal(data, v)
}

// Name returns the name of the codec
func (c *Codec) Name() string {
	return "json"
}
