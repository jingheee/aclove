package jsonutil

import (
	"bytes"
	"testing"

	"github.com/bytedance/sonic"
)

// sonicConfig 是用于测试的 sonic 配置
var sonicConfig = sonic.Config{
	UseInt64:              true,
	NoNullSliceOrMap:      true,
	ValidateString:        true,
	DisallowUnknownFields: false,
}.Froze()

func TestInt64Marshal(t *testing.T) {
	tests := []struct {
		name     string
		value    Int64
		expected string
	}{
		{
			name:     "small value",
			value:    123,
			expected: `"123"`,
		},
		{
			name:     "snowflake-like ID",
			value:    1234567890123456789,
			expected: `"1234567890123456789"`,
		},
		{
			name:     "negative value",
			value:    -456,
			expected: `"-456"`,
		},
		{
			name:     "zero",
			value:    0,
			expected: `"0"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := sonicConfig.Marshal(tt.value)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}
			if string(result) != tt.expected {
				t.Errorf("Marshal() = %s, want %s", string(result), tt.expected)
			}
		})
	}
}

func TestInt64Unmarshal(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected Int64
	}{
		{
			name:     "string value",
			json:     `"1234567890123456789"`,
			expected: 1234567890123456789,
		},
		{
			name:     "number value",
			json:     `1234567890123456789`,
			expected: 1234567890123456789,
		},
		{
			name:     "zero as string",
			json:     `"0"`,
			expected: 0,
		},
		{
			name:     "zero as number",
			json:     `0`,
			expected: 0,
		},
		{
			name:     "negative as string",
			json:     `"-456"`,
			expected: -456,
		},
		{
			name:     "negative as number",
			json:     `-456`,
			expected: -456,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result Int64
			err := sonicConfig.Unmarshal([]byte(tt.json), &result)
			if err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Unmarshal() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestInt64InStruct(t *testing.T) {
	type TestStruct struct {
		ID   Int64  `json:"id"`
		Name string `json:"name"`
	}

	t.Run("marshal", func(t *testing.T) {
		s := TestStruct{ID: 1234567890123456789, Name: "test"}
		data, err := sonicConfig.Marshal(s)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		expected := `{"id":"1234567890123456789","name":"test"}`
		if string(data) != expected {
			t.Errorf("Marshal() = %s, want %s", string(data), expected)
		}
	})

	t.Run("unmarshal from string", func(t *testing.T) {
		json := `{"id":"1234567890123456789","name":"test"}`
		var s TestStruct
		err := sonicConfig.Unmarshal([]byte(json), &s)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}
		if s.ID != 1234567890123456789 {
			t.Errorf("ID = %d, want 1234567890123456789", s.ID)
		}
		if s.Name != "test" {
			t.Errorf("Name = %s, want test", s.Name)
		}
	})

	t.Run("unmarshal from number", func(t *testing.T) {
		json := `{"id":1234567890123456789,"name":"test"}`
		var s TestStruct
		err := sonicConfig.Unmarshal([]byte(json), &s)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}
		if s.ID != 1234567890123456789 {
			t.Errorf("ID = %d, want 1234567890123456789", s.ID)
		}
	})
}

func TestInt64Pointer(t *testing.T) {
	type TestStruct struct {
		ID *Int64 `json:"id,omitempty"`
	}

	t.Run("marshal with value", func(t *testing.T) {
		id := NewInt64(1234567890123456789)
		s := TestStruct{ID: &id}
		data, err := sonicConfig.Marshal(s)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}
		expected := `{"id":"1234567890123456789"}`
		if string(data) != expected {
			t.Errorf("Marshal() = %s, want %s", string(data), expected)
		}
	})

	t.Run("marshal with nil", func(t *testing.T) {
		s := TestStruct{ID: nil}
		data, err := sonicConfig.Marshal(s)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}
		expected := `{}`
		if string(data) != expected {
			t.Errorf("Marshal() = %s, want %s", string(data), expected)
		}
	})

	t.Run("unmarshal to pointer from string", func(t *testing.T) {
		json := `{"id":"1234567890123456789"}`
		var s TestStruct
		err := sonicConfig.Unmarshal([]byte(json), &s)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}
		if s.ID == nil {
			t.Fatal("ID is nil")
		}
		if *s.ID != 1234567890123456789 {
			t.Errorf("ID = %d, want 1234567890123456789", *s.ID)
		}
	})
}

func TestInt64Methods(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		i := Int64(1234567890123456789)
		if i.String() != "1234567890123456789" {
			t.Errorf("String() = %s, want 1234567890123456789", i.String())
		}
	})

	t.Run("Int64", func(t *testing.T) {
		i := Int64(1234567890123456789)
		if i.Int64() != 1234567890123456789 {
			t.Errorf("Int64() = %d, want 1234567890123456789", i.Int64())
		}
	})

	t.Run("Int64Ptr", func(t *testing.T) {
		i := Int64(1234567890123456789)
		ptr := i.Int64Ptr()
		if ptr == nil {
			t.Fatal("Int64Ptr() returned nil")
		}
		if *ptr != 1234567890123456789 {
			t.Errorf("*Int64Ptr() = %d, want 1234567890123456789", *ptr)
		}
	})

	t.Run("NewInt64", func(t *testing.T) {
		i := NewInt64(1234567890123456789)
		if i != 1234567890123456789 {
			t.Errorf("NewInt64() = %d, want 1234567890123456789", i)
		}
	})

	t.Run("NewInt64Ptr", func(t *testing.T) {
		ptr := NewInt64Ptr(1234567890123456789)
		if ptr == nil {
			t.Fatal("NewInt64Ptr() returned nil")
		}
		if *ptr != 1234567890123456789 {
			t.Errorf("*NewInt64Ptr() = %d, want 1234567890123456789", *ptr)
		}
	})
}

func TestSonicMarshal(t *testing.T) {
	type TestData struct {
		ID    Int64    `json:"id"`
		Items []string `json:"items"`
	}

	data := TestData{
		ID:    1234567890123456789,
		Items: []string{"a", "b", "c"},
	}

	result, err := Marshal(data)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	expected := `{"id":"1234567890123456789","items":["a","b","c"]}`
	if string(result) != expected {
		t.Errorf("Marshal() = %s, want %s", string(result), expected)
	}
}

func TestSonicUnmarshal(t *testing.T) {
	type TestData struct {
		ID    Int64    `json:"id"`
		Items []string `json:"items"`
	}

	json := `{"id":"1234567890123456789","items":["a","b","c"]}`

	var data TestData
	err := Unmarshal([]byte(json), &data)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if data.ID != 1234567890123456789 {
		t.Errorf("ID = %d, want 1234567890123456789", data.ID)
	}
	if len(data.Items) != 3 {
		t.Errorf("len(Items) = %d, want 3", len(data.Items))
	}
}

func TestJSONReader(t *testing.T) {
	type TestData struct {
		ID   Int64  `json:"id"`
		Name string `json:"name"`
	}

	json := `{"id":"1234567890123456789","name":"test"}`
	reader := bytes.NewReader([]byte(json))

	var data TestData
	err := JSONReader(reader, &data)
	if err != nil {
		t.Fatalf("JSONReader failed: %v", err)
	}

	if data.ID != 1234567890123456789 {
		t.Errorf("ID = %d, want 1234567890123456789", data.ID)
	}
	if data.Name != "test" {
		t.Errorf("Name = %s, want test", data.Name)
	}
}

func TestJSONWriter(t *testing.T) {
	type TestData struct {
		ID   Int64  `json:"id"`
		Name string `json:"name"`
	}

	data := TestData{
		ID:   1234567890123456789,
		Name: "test",
	}

	var buf bytes.Buffer
	err := JSONWriter(&buf, data)
	if err != nil {
		t.Fatalf("JSONWriter failed: %v", err)
	}

	// sonic 会在末尾添加换行符
	result := buf.String()
	expected := `{"id":"1234567890123456789","name":"test"}` + "\n"
	if result != expected {
		t.Errorf("JSONWriter() = %s, want %s", result, expected)
	}
}

// BenchmarkSonicMarshal 测试 sonic 的序列化性能
func BenchmarkSonicMarshal(b *testing.B) {
	type TestData struct {
		ID    Int64    `json:"id"`
		Name  string   `json:"name"`
		Items []string `json:"items"`
	}

	data := TestData{
		ID:    1234567890123456789,
		Name:  "test",
		Items: []string{"a", "b", "c", "d", "e"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Marshal(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSonicUnmarshal 测试 sonic 的反序列化性能
func BenchmarkSonicUnmarshal(b *testing.B) {
	type TestData struct {
		ID    Int64    `json:"id"`
		Name  string   `json:"name"`
		Items []string `json:"items"`
	}

	json := []byte(`{"id":"1234567890123456789","name":"test","items":["a","b","c","d","e"]}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var data TestData
		err := Unmarshal(json, &data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkStandardMarshal 测试标准库的序列化性能（用于对比）
func BenchmarkStandardMarshal(b *testing.B) {
	encodingJSON := sonic.ConfigStd.Marshal

	type TestData struct {
		ID    int64    `json:"id"`
		Name  string   `json:"name"`
		Items []string `json:"items"`
	}

	data := TestData{
		ID:    1234567890123456789,
		Name:  "test",
		Items: []string{"a", "b", "c", "d", "e"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := encodingJSON(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkStandardUnmarshal 测试标准库的反序列化性能（用于对比）
func BenchmarkStandardUnmarshal(b *testing.B) {
	encodingJSON := sonic.ConfigStd.Unmarshal

	type TestData struct {
		ID    int64    `json:"id"`
		Name  string   `json:"name"`
		Items []string `json:"items"`
	}

	json := []byte(`{"id":1234567890123456789,"name":"test","items":["a","b","c","d","e"]}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var data TestData
		err := encodingJSON(json, &data)
		if err != nil {
			b.Fatal(err)
		}
	}
}
