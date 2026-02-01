package jsonutil

import (
	"bytes"
	"errors"
	"io"
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// sonicAPIForBinding 是用于绑定的 sonic API 实例
var sonicAPIForBinding = sonic.Config{
	UseInt64:              true,
	UseNumber:             false,
	NoNullSliceOrMap:      true,
	ValidateString:        true,
	DisallowUnknownFields: false,
}.Froze()

// JSONBinding 是自定义的 JSON 绑定器，使用 sonic 替代标准库
type JSONBinding struct{}

var _ binding.Binding = (*JSONBinding)(nil)

// Name 返回绑定器的名称
func (b *JSONBinding) Name() string {
	return "json"
}

// Bind 从请求体反序列化 JSON 数据到对象
func (b *JSONBinding) Bind(req *http.Request, obj any) error {
	if req == nil || req.Body == nil {
		return errors.New("invalid request")
	}
	return decodeJSON(req.Body, obj)
}

// BindBody 从字节数组反序列化 JSON 数据到对象
func (b *JSONBinding) BindBody(body []byte, obj any) error {
	return sonicAPIForBinding.Unmarshal(body, obj)
}

// decodeJSON 从 io.Reader 解码 JSON
func decodeJSON(r io.Reader, obj any) error {
	// 读取所有数据
	body, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	// 使用 sonic 反序列化
	if err := sonicAPIForBinding.Unmarshal(body, obj); err != nil {
		return err
	}

	return validate(obj)
}

// validate 验证对象
func validate(obj any) error {
	if binding.Validator == nil {
		return nil
	}
	return binding.Validator.ValidateStruct(obj)
}

// EnableCustomJSONBinding 启用自定义 JSON 绑定器
func EnableCustomJSONBinding() {
	binding.JSON = &JSONBinding{}
}

// BindJSON 是 gin.Context.BindJSON 的替代函数，使用 sonic 绑定器
func BindJSON(c *gin.Context, obj any) error {
	return (&JSONBinding{}).Bind(c.Request, obj)
}

// ShouldBindJSON 是 gin.Context.ShouldBindJSON 的替代函数，使用 sonic 绑定器
func ShouldBindJSON(c *gin.Context, obj any) error {
	return (&JSONBinding{}).Bind(c.Request, obj)
}

// UnmarshalBodyJSON 从请求体反序列化 JSON
func UnmarshalBodyJSON(c *gin.Context, obj any) error {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return err
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	return sonicAPIForBinding.Unmarshal(body, obj)
}
