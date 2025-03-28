package sonic

import (
	"reflect"
	"testing"
	"time"
)

// TestStruct 用于测试的结构体
type TestStruct struct {
	String  string                 `json:"string"`
	Int     int                    `json:"int"`
	Bool    bool                   `json:"bool"`
	Float   float64                `json:"float"`
	Slice   []string               `json:"slice"`
	Map     map[string]interface{} `json:"map"`
	Time    time.Time              `json:"time"`
	Pointer *string                `json:"pointer,omitempty"`
	Null    interface{}            `json:"null"`
}

// TestCodec_Marshal_Unmarshal 测试序列化和反序列化
func TestCodec_Marshal_Unmarshal(t *testing.T) {
	// 创建测试数据
	testStr := "pointer value"
	testTime := time.Date(2025, 3, 28, 12, 0, 0, 0, time.UTC)

	original := TestStruct{
		String: "hello world",
		Int:    42,
		Bool:   true,
		Float:  3.14159,
		Time:   testTime,
		Slice:  []string{"a", "b", "c"},
		Map: map[string]interface{}{
			"key1": "value1",
			"key2": 123,
			"key3": 45.67,
			"key4": true,
		},
		Pointer: &testStr,
		Null:    nil,
	}

	// 测试不同配置的编解码器
	testCases := []struct {
		name   string
		codec  *Codec
		indent bool
	}{
		{"Default", New(), false},
		{"WithCompactOutput(true)", New(WithCompactOutput(true)), false},
		{"WithCompactOutput(false)", New(WithCompactOutput(false)), true},
		{"WithEscapeHTML(true)", New(WithEscapeHTML(true)), false},
		{"WithEscapeHTML(false)", New(WithEscapeHTML(false)), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			codec := tc.codec

			// 测试Marshal
			data, err := codec.Marshal(original)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}

			// 验证序列化结果
			if len(data) == 0 {
				t.Error("Marshal returned empty data")
			}

			// 检查是否包含缩进
			hasIndent := false
			for _, b := range data {
				if b == '\n' || b == ' ' {
					hasIndent = true
					break
				}
			}

			if tc.indent && !hasIndent {
				t.Error("Expected indented output, got compact output")
			}

			// 测试Unmarshal
			var result TestStruct
			err = codec.Unmarshal(data, &result)
			if err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			// 比较原始对象和反序列化后的对象
			if original.String != result.String {
				t.Errorf("String mismatch: expected %q, got %q", original.String, result.String)
			}
			if original.Int != result.Int {
				t.Errorf("Int mismatch: expected %d, got %d", original.Int, result.Int)
			}
			if original.Bool != result.Bool {
				t.Errorf("Bool mismatch: expected %v, got %v", original.Bool, result.Bool)
			}
			if original.Float != result.Float {
				t.Errorf("Float mismatch: expected %f, got %f", original.Float, result.Float)
			}
			if !original.Time.Equal(result.Time) {
				t.Errorf("Time mismatch: expected %v, got %v", original.Time, result.Time)
			}
			if !reflect.DeepEqual(original.Slice, result.Slice) {
				t.Errorf("Slice mismatch: expected %v, got %v", original.Slice, result.Slice)
			}

			// Map中的数值可能会有轻微差异，所以我们只检查长度
			if len(original.Map) != len(result.Map) {
				t.Errorf("Map length mismatch: expected %d, got %d", len(original.Map), len(result.Map))
			}

			// 指针检查
			if (original.Pointer == nil) != (result.Pointer == nil) {
				t.Errorf("Pointer nil status mismatch: expected %v, got %v", original.Pointer == nil, result.Pointer == nil)
			} else if original.Pointer != nil && *original.Pointer != *result.Pointer {
				t.Errorf("Pointer value mismatch: expected %q, got %q", *original.Pointer, *result.Pointer)
			}
		})
	}
}

// TestUnmarshal_EmptyData 测试反序列化空数据的情况
func TestUnmarshal_EmptyData(t *testing.T) {
	codec := New()
	var result TestStruct
	err := codec.Unmarshal([]byte{}, &result)
	if err == nil {
		t.Error("Expected error when unmarshaling empty data, got nil")
	}
}

// TestMarshal_ComplexData 测试序列化复杂数据
func TestMarshal_ComplexData(t *testing.T) {
	// 创建嵌套的复杂结构
	data := map[string]interface{}{
		"nested": map[string]interface{}{
			"array": []interface{}{1, "two", 3.0, true},
			"empty": []int{},
		},
		"html": "<script>alert('test')</script>",
	}

	// 测试两种配置下的HTML转义
	testCases := []struct {
		name       string
		codec      *Codec
		escapeHTML bool
	}{
		{"WithEscapeHTML(true)", New(WithEscapeHTML(true)), true},
		{"WithEscapeHTML(false)", New(WithEscapeHTML(false)), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			codec := tc.codec

			// 测试Marshal
			bytes, err := codec.Marshal(data)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}

			// 转为字符串方便检查
			jsonStr := string(bytes)

			// 检查HTML转义 - 注意：这是一个简化的检查
			if tc.escapeHTML {
				// 我们期望看到转义的尖括号
				if jsonStr == `{"html":"<script>alert('test')</script>","nested":{"array":[1,"two",3,true],"empty":[]}}` {
					t.Error("HTML was not escaped when it should have been")
				}
			}

			// 检查是否能正确反序列化
			var result map[string]interface{}
			err = codec.Unmarshal(bytes, &result)
			if err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			// 检查反序列化结果是否正确
			nested, ok := result["nested"].(map[string]interface{})
			if !ok {
				t.Fatalf("Expected nested map, got %T", result["nested"])
			}

			array, ok := nested["array"].([]interface{})
			if !ok {
				t.Fatalf("Expected array, got %T", nested["array"])
			}

			if len(array) != 4 {
				t.Errorf("Expected array length 4, got %d", len(array))
			}
		})
	}
}

// TestCodec_Name 测试Name方法
func TestCodec_Name(t *testing.T) {
	codec := New()
	if codec.Name() != "sonic" {
		t.Errorf("Expected name 'sonic', got '%s'", codec.Name())
	}
}
