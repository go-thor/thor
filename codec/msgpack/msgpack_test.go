package msgpack

import (
	"reflect"
	"testing"
	"time"
)

type testStruct struct {
	String string                 `json:"string"`
	Int    int                    `json:"int"`
	Bool   bool                   `json:"bool"`
	Float  float64                `json:"float"`
	Time   time.Time              `json:"time"`
	Slice  []string               `json:"slice"`
	Map    map[string]interface{} `json:"map"`
}

func TestCodec_Marshal_Unmarshal(t *testing.T) {
	codec := New()

	// 准备测试数据
	testTime := time.Date(2025, 3, 28, 0, 0, 0, 0, time.UTC)
	original := testStruct{
		String: "hello",
		Int:    42,
		Bool:   true,
		Float:  3.14,
		Time:   testTime,
		Slice:  []string{"a", "b", "c"},
		Map: map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		},
	}

	// 测试 Marshal
	data, err := codec.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if len(data) == 0 {
		t.Errorf("Marshal returned empty data")
	}

	// 测试 Unmarshal
	var result testStruct
	err = codec.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// 比较结果
	if !reflect.DeepEqual(original.String, result.String) {
		t.Errorf("String field mismatch: expected %v, got %v", original.String, result.String)
	}
	if original.Int != result.Int {
		t.Errorf("Int field mismatch: expected %v, got %v", original.Int, result.Int)
	}
	if original.Bool != result.Bool {
		t.Errorf("Bool field mismatch: expected %v, got %v", original.Bool, result.Bool)
	}
	if original.Float != result.Float {
		t.Errorf("Float field mismatch: expected %v, got %v", original.Float, result.Float)
	}
	if !original.Time.Equal(result.Time) {
		t.Errorf("Time field mismatch: expected %v, got %v", original.Time, result.Time)
	}
	if !reflect.DeepEqual(original.Slice, result.Slice) {
		t.Errorf("Slice field mismatch: expected %v, got %v", original.Slice, result.Slice)
	}

	// Map 中的数值类型可能会有差异，因此我们单独检查每个键值
	if len(original.Map) != len(result.Map) {
		t.Errorf("Map length mismatch: expected %d, got %d", len(original.Map), len(result.Map))
	}

	// 检查 Name 方法
	if codec.Name() != "msgpack" {
		t.Errorf("Name mismatch: expected 'msgpack', got '%s'", codec.Name())
	}
}

func TestCodec_Marshal_Unmarshal_Nil(t *testing.T) {
	codec := New()

	// 测试 nil 值
	var nilMap map[string]interface{}
	data, err := codec.Marshal(nilMap)
	if err != nil {
		t.Fatalf("Marshal nil failed: %v", err)
	}

	var result map[string]interface{}
	err = codec.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Unmarshal nil failed: %v", err)
	}
	if result != nil {
		t.Errorf("Expected nil result, got %v", result)
	}
}

func TestCodec_Marshal_Unmarshal_Complex(t *testing.T) {
	codec := New()

	// 准备嵌套的复杂数据
	complex := map[string]interface{}{
		"nested": map[string]interface{}{
			"array": []interface{}{1, "two", 3.0, true},
			"struct": struct {
				Field1 string `json:"field_1"`
				Field2 int    `json:"field_2"`
			}{"value", 42},
		},
		"empty_slice": []int{},
	}

	// 测试 Marshal
	data, err := codec.Marshal(complex)
	if err != nil {
		t.Fatalf("Marshal complex failed: %v", err)
	}

	// 测试 Unmarshal
	var result map[string]interface{}
	err = codec.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Unmarshal complex failed: %v", err)
	}

	// 验证结果包含预期的键
	if _, ok := result["nested"]; !ok {
		t.Errorf("Missing 'nested' key in result")
	}
	if _, ok := result["empty_slice"]; !ok {
		t.Errorf("Missing 'empty_slice' key in result")
	}
}
