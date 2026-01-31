package models

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ResponseCode 响应状态码类型
type ResponseCode int

const (
	CodeSuccess ResponseCode = 0
	CodeError   ResponseCode = 1
)

// Response 通用响应结构体
type Response struct {
	Code    ResponseCode `json:"code"`              // 业务状态码: 0-成功, 非0-失败
	Message string       `json:"message"`           // 提示信息
	Data    any          `json:"data,omitempty"`    // 响应数据
}

// NewSuccessResponse 创建成功响应
func NewSuccessResponse(data any) *Response {
	return &Response{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	}
}

// NewSuccessResponseWithMsg 创建带自定义消息的成功响应
func NewSuccessResponseWithMsg(message string, data any) *Response {
	return &Response{
		Code:    CodeSuccess,
		Message: message,
		Data:    data,
	}
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(message string) *Response {
	return &Response{
		Code:    CodeError,
		Message: message,
		Data:    nil,
	}
}

// NewErrorResponseWithCode 创建带自定义状态码的错误响应
func NewErrorResponseWithCode(code ResponseCode, message string) *Response {
	return &Response{
		Code:    code,
		Message: message,
		Data:    nil,
	}
}

// JSONSuccess 返回成功JSON响应
func JSONSuccess(c *gin.Context, data any) {
	c.JSON(http.StatusOK, NewSuccessResponse(data))
}

// JSONSuccessWithMsg 返回带自定义消息的成功JSON响应
func JSONSuccessWithMsg(c *gin.Context, message string, data any) {
	c.JSON(http.StatusOK, NewSuccessResponseWithMsg(message, data))
}

// JSONCreated 返回创建成功响应 (201)
func JSONCreated(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, NewSuccessResponse(data))
}

// JSONError 返回错误JSON响应
func JSONError(c *gin.Context, httpStatus int, message string) {
	c.JSON(httpStatus, NewErrorResponse(message))
}

// JSONBadRequest 返回400错误响应
func JSONBadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, NewErrorResponse(message))
}

// JSONUnauthorized 返回401错误响应
func JSONUnauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, NewErrorResponse(message))
}

// JSONForbidden 返回403错误响应
func JSONForbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, NewErrorResponse(message))
}

// JSONNotFound 返回404错误响应
func JSONNotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, NewErrorResponse(message))
}

// JSONTooManyRequests 返回429错误响应
func JSONTooManyRequests(c *gin.Context, message string) {
	c.JSON(http.StatusTooManyRequests, NewErrorResponse(message))
}

// JSONInternalError 返回500错误响应
func JSONInternalError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, NewErrorResponse(message))
}
