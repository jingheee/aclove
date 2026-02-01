package jsonutil

import (
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
)

// sonicAPIForRender 是用于渲染的 sonic API 实例
var sonicAPIForRender = sonic.Config{
	UseInt64:         true,
	NoNullSliceOrMap: true,
	ValidateString:   true,
}.Froze()

// JSONRenderer 是自定义的 JSON 渲染器，使用 sonic 替代 goccy/go-json
type JSONRenderer struct {
	Data any
}

// Render 将数据序列化为 JSON 并写入响应
func (r JSONRenderer) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)
	bytes, err := sonicAPIForRender.Marshal(r.Data)
	if err != nil {
		return err
	}
	_, err = w.Write(bytes)
	return err
}

// WriteContentType 写入 Content-Type 头
func (r JSONRenderer) WriteContentType(w http.ResponseWriter) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = []string{"application/json; charset=utf-8"}
	}
}

// IndentedJSONRenderer 是带缩进的 JSON 渲染器
type IndentedJSONRenderer struct {
	Data any
}

// Render 将数据序列化为带缩进的 JSON 并写入响应
func (r IndentedJSONRenderer) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)
	bytes, err := sonicAPIForRender.MarshalIndent(r.Data, "", "    ")
	if err != nil {
		return err
	}
	_, err = w.Write(bytes)
	return err
}

// WriteContentType 写入 Content-Type 头
func (r IndentedJSONRenderer) WriteContentType(w http.ResponseWriter) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = []string{"application/json; charset=utf-8"}
	}
}

// JSON 发送 JSON 响应
func JSON(c *gin.Context, status int, obj any) {
	c.Render(status, JSONRenderer{Data: obj})
}

// IndentedJSON 发送带缩进的 JSON 响应
func IndentedJSON(c *gin.Context, status int, obj any) {
	c.Render(status, IndentedJSONRenderer{Data: obj})
}

// JSON200 发送 200 OK JSON 响应
func JSON200(c *gin.Context, obj any) {
	JSON(c, http.StatusOK, obj)
}

// JSON201 发送 201 Created JSON 响应
func JSON201(c *gin.Context, obj any) {
	JSON(c, http.StatusCreated, obj)
}

// JSON400 发送 400 Bad Request JSON 响应
func JSON400(c *gin.Context, obj any) {
	JSON(c, http.StatusBadRequest, obj)
}

// JSON401 发送 401 Unauthorized JSON 响应
func JSON401(c *gin.Context, obj any) {
	JSON(c, http.StatusUnauthorized, obj)
}

// JSON403 发送 403 Forbidden JSON 响应
func JSON403(c *gin.Context, obj any) {
	JSON(c, http.StatusForbidden, obj)
}

// JSON404 发送 404 Not Found JSON 响应
func JSON404(c *gin.Context, obj any) {
	JSON(c, http.StatusNotFound, obj)
}

// JSON500 发送 500 Internal Server Error JSON 响应
func JSON500(c *gin.Context, obj any) {
	JSON(c, http.StatusInternalServerError, obj)
}

// JSON502 发送 502 Bad Gateway JSON 响应
func JSON502(c *gin.Context, obj any) {
	JSON(c, http.StatusBadGateway, obj)
}

// JSON503 发送 503 Service Unavailable JSON 响应
func JSON503(c *gin.Context, obj any) {
	JSON(c, http.StatusServiceUnavailable, obj)
}

// JSON504 发送 504 Gateway Timeout JSON 响应
func JSON504(c *gin.Context, obj any) {
	JSON(c, http.StatusGatewayTimeout, obj)
}

// StreamJSON 流式发送 JSON 响应
func StreamJSON(c *gin.Context, status int, obj any) error {
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.Status(status)
	return sonicAPIForRender.NewEncoder(c.Writer).Encode(obj)
}
