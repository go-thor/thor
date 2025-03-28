package codec_test

import (
	"testing"
	"time"

	"github.com/go-thor/thor/codec/json"
	"github.com/go-thor/thor/codec/msgpack"
)

// 测试用的复杂数据结构
type BenchmarkData struct {
	ID        int               `json:"id" msgpack:"id"`
	Name      string            `json:"name" msgpack:"name"`
	Email     string            `json:"email" msgpack:"email"`
	CreatedAt time.Time         `json:"created_at" msgpack:"created_at"`
	UpdatedAt time.Time         `json:"updated_at" msgpack:"updated_at"`
	Active    bool              `json:"active" msgpack:"active"`
	Score     float64           `json:"score" msgpack:"score"`
	Tags      []string          `json:"tags" msgpack:"tags"`
	Metadata  map[string]string `json:"metadata" msgpack:"metadata"`
}

// 准备测试数据
func prepareTestData() *BenchmarkData {
	return &BenchmarkData{
		ID:        12345,
		Name:      "Test User",
		Email:     "test.user@example.com",
		CreatedAt: time.Date(2025, 3, 28, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Now(),
		Active:    true,
		Score:     98.7654,
		Tags:      []string{"tag1", "tag2", "tag3", "tag4", "tag5"},
		Metadata: map[string]string{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
			"key4": "value4",
			"key5": "value5",
		},
	}
}

// Marshal 基准测试
func BenchmarkMarshal(b *testing.B) {
	data := prepareTestData()

	// JSON 编解码器
	jsonCodec := json.New()

	// MessagePack 编解码器
	msgpackCodec := msgpack.New()

	b.ResetTimer()

	b.Run("JSON", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := jsonCodec.Marshal(data)
			if err != nil {
				b.Fatalf("JSON Marshal failed: %v", err)
			}
		}
	})

	b.Run("MessagePack", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := msgpackCodec.Marshal(data)
			if err != nil {
				b.Fatalf("MessagePack Marshal failed: %v", err)
			}
		}
	})
}

// Unmarshal 基准测试
func BenchmarkUnmarshal(b *testing.B) {
	data := prepareTestData()

	// JSON 编解码器
	jsonCodec := json.New()
	jsonData, _ := jsonCodec.Marshal(data)

	// MessagePack 编解码器
	msgpackCodec := msgpack.New()
	msgpackData, _ := msgpackCodec.Marshal(data)

	b.ResetTimer()

	b.Run("JSON", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var result BenchmarkData
			err := jsonCodec.Unmarshal(jsonData, &result)
			if err != nil {
				b.Fatalf("JSON Unmarshal failed: %v", err)
			}
		}
	})

	b.Run("MessagePack", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var result BenchmarkData
			err := msgpackCodec.Unmarshal(msgpackData, &result)
			if err != nil {
				b.Fatalf("MessagePack Unmarshal failed: %v", err)
			}
		}
	})
}

// 数据大小比较测试
func TestDataSize(t *testing.T) {
	data := prepareTestData()

	// JSON 编解码器
	jsonCodec := json.New()
	jsonData, _ := jsonCodec.Marshal(data)

	// MessagePack 编解码器
	msgpackCodec := msgpack.New()
	msgpackData, _ := msgpackCodec.Marshal(data)

	t.Logf("JSON 数据大小: %d 字节", len(jsonData))
	t.Logf("MessagePack 数据大小: %d 字节", len(msgpackData))
	t.Logf("MessagePack 数据大小相比 JSON 减少了: %.2f%%", (1-float64(len(msgpackData))/float64(len(jsonData)))*100)
}
