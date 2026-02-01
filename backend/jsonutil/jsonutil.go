package jsonutil

import (
	"io"
	"strconv"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
)

var (
	// sonicAPI 是 sonic 的 API 实例
	sonicAPI = sonic.Config{
		UseInt64:              true,
		NoNullSliceOrMap:      true,
		ValidateString:        true,
		DisallowUnknownFields: false,
	}.Froze()

	// Marshal 使用 sonic 序列化 JSON，性能比标准库高 3-5 倍
	Marshal = sonicAPI.Marshal

	// Unmarshal 使用 sonic 反序列化 JSON
	Unmarshal = sonicAPI.Unmarshal

	// MarshalIndent 使用 sonic 格式化序列化 JSON
	MarshalIndent = sonicAPI.MarshalIndent

	// NewEncoder 创建 sonic 编码器
	NewEncoder = sonicAPI.NewEncoder

	// NewDecoder 创建 sonic 解码器
	NewDecoder = sonicAPI.NewDecoder
)

// H 返回一个 gin.H 的快捷方式，用于构建 JSON 响应
func H(data any) gin.H {
	return gin.H{"data": data}
}

// Int64 是一个自定义类型，用于将 int64 序列化为字符串，并支持从字符串或数字反序列化
type Int64 int64

// MarshalJSON 将 int64 序列化为 JSON 字符串
// sonic 会自动识别并调用此方法
func (i Int64) MarshalJSON() ([]byte, error) {
	return []byte(`"` + strconv.FormatInt(int64(i), 10) + `"`), nil
}

// UnmarshalJSON 从 JSON 字符串或数字反序列化为 int64
// sonic 会自动识别并调用此方法
func (i *Int64) UnmarshalJSON(data []byte) error {
	// sonic 已经解析了 JSON，data 是原始值
	// 如果是字符串，data 会包含引号
	if len(data) >= 2 && data[0] == '"' && data[len(data)-1] == '"' {
		str := string(data[1 : len(data)-1])
		val, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return err
		}
		*i = Int64(val)
		return nil
	}

	// 如果是数字
	val, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return err
	}
	*i = Int64(val)
	return nil
}

// String 返回字符串表示
func (i Int64) String() string {
	return strconv.FormatInt(int64(i), 10)
}

// Int64 返回 int64 值
func (i Int64) Int64() int64 {
	return int64(i)
}

// Int64Ptr 返回 *int64
func (i Int64) Int64Ptr() *int64 {
	val := int64(i)
	return &val
}

// NewInt64 从 int64 创建 Int64
func NewInt64(val int64) Int64 {
	return Int64(val)
}

// NewInt64Ptr 从 int64 创建 *Int64
func NewInt64Ptr(val int64) *Int64 {
	i := Int64(val)
	return &i
}

// JSONReader 从 io.Reader 读取并反序列化 JSON
func JSONReader(r io.Reader, obj any) error {
	return sonicAPI.NewDecoder(r).Decode(obj)
}

// JSONWriter 将对象序列化为 JSON 并写入 io.Writer
func JSONWriter(w io.Writer, obj any) error {
	return sonicAPI.NewEncoder(w).Encode(obj)
}
