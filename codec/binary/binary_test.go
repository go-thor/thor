package binary

import (
	"testing"

	"github.com/go-thor/thor"
	"github.com/go-thor/thor/jsoncodec"
)

func TestBinaryCodec_Marshal_Unmarshal_Request(t *testing.T) {
	// 创建内部编解码器
	inner := jsoncodec.New()
	codec := New(inner)

	// 创建请求
	req := &thor.Request{
		ServiceMethod: "Test.Echo",
		Metadata: map[string]string{
			"key": "value",
		},
		Payload: []byte(`{"name":"thor"}`),
		Seq:     123,
	}

	// 编码
	data, err := codec.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// 解码
	req2 := &thor.Request{}
	err = codec.Unmarshal(data, req2)
	if err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	// 验证结果
	if req2.ServiceMethod != req.ServiceMethod {
		t.Errorf("ServiceMethod mismatch: got %s, want %s", req2.ServiceMethod, req.ServiceMethod)
	}
	if req2.Seq != req.Seq {
		t.Errorf("Seq mismatch: got %d, want %d", req2.Seq, req.Seq)
	}
	if req2.Metadata["key"] != req.Metadata["key"] {
		t.Errorf("Metadata mismatch: got %v, want %v", req2.Metadata, req.Metadata)
	}
	if string(req2.Payload) != string(req.Payload) {
		t.Errorf("Payload mismatch: got %s, want %s", string(req2.Payload), string(req.Payload))
	}
}

func TestBinaryCodec_Marshal_Unmarshal_Response(t *testing.T) {
	// 创建内部编解码器
	inner := jsoncodec.New()
	codec := New(inner)

	// 创建响应
	resp := &thor.Response{
		ServiceMethod: "Test.Echo",
		Metadata: map[string]string{
			"key": "value",
		},
		Reply: []byte(`{"result":"success"}`),
		Error: "some error",
		Seq:   123,
	}

	// 编码
	data, err := codec.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	// 解码
	resp2 := &thor.Response{}
	err = codec.Unmarshal(data, resp2)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// 验证结果
	if resp2.ServiceMethod != resp.ServiceMethod {
		t.Errorf("ServiceMethod mismatch: got %s, want %s", resp2.ServiceMethod, resp.ServiceMethod)
	}
	if resp2.Seq != resp.Seq {
		t.Errorf("Seq mismatch: got %d, want %d", resp2.Seq, resp.Seq)
	}
	if resp2.Metadata["key"] != resp.Metadata["key"] {
		t.Errorf("Metadata mismatch: got %v, want %v", resp2.Metadata, resp.Metadata)
	}
	if string(resp2.Reply) != string(resp.Reply) {
		t.Errorf("Reply mismatch: got %s, want %s", string(resp2.Reply), string(resp.Reply))
	}
	if resp2.Error != resp.Error {
		t.Errorf("Error mismatch: got %s, want %s", resp2.Error, resp.Error)
	}
}

func TestBinaryCodec_LargePayload(t *testing.T) {
	// 创建内部编解码器
	inner := jsoncodec.New()
	codec := New(inner)

	// 创建大负载
	payload := make([]byte, 1024*1024) // 1MB
	for i := range payload {
		payload[i] = byte(i % 256)
	}

	// 创建请求
	req := &thor.Request{
		ServiceMethod: "Test.Echo",
		Payload:       payload,
		Seq:           123,
	}

	// 编码
	data, err := codec.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// 解码
	req2 := &thor.Request{}
	err = codec.Unmarshal(data, req2)
	if err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	// 验证结果
	if len(req2.Payload) != len(req.Payload) {
		t.Errorf("Payload length mismatch: got %d, want %d", len(req2.Payload), len(req.Payload))
	}

	// 比较部分数据
	for i := 0; i < 100; i++ {
		if req2.Payload[i] != req.Payload[i] {
			t.Errorf("Payload data mismatch at index %d: got %d, want %d", i, req2.Payload[i], req.Payload[i])
			break
		}
	}
}

func TestBinaryCodec_InvalidData(t *testing.T) {
	// 创建内部编解码器
	inner := jsoncodec.New()
	codec := New(inner)

	// 测试无效数据
	testCases := []struct {
		name string
		data []byte
	}{
		{"Empty", []byte{}},
		{"TooShort", []byte{0x82, 0x74, 0x01}},
		{"InvalidMagic", []byte{0x12, 0x34, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &thor.Request{}
			err := codec.Unmarshal(tc.data, req)
			if err == nil {
				t.Errorf("Expected error for invalid data, got nil")
			}
		})
	}
}
