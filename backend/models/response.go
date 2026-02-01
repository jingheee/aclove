package models

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"io.lazydoge/aclove/jsonutil"
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
	jsonutil.JSON200(c, NewSuccessResponse(data))
}

// JSONSuccessWithMsg 返回带自定义消息的成功JSON响应
func JSONSuccessWithMsg(c *gin.Context, message string, data any) {
	jsonutil.JSON200(c, NewSuccessResponseWithMsg(message, data))
}

// JSONCreated 返回创建成功响应 (201)
func JSONCreated(c *gin.Context, data any) {
	jsonutil.JSON201(c, NewSuccessResponse(data))
}

// JSONError 返回错误JSON响应
func JSONError(c *gin.Context, httpStatus int, message string) {
	jsonutil.JSON(c, httpStatus, NewErrorResponse(message))
}

// JSONBadRequest 返回400错误响应
func JSONBadRequest(c *gin.Context, message string) {
	jsonutil.JSON400(c, NewErrorResponse(message))
}

// JSONUnauthorized 返回401错误响应
func JSONUnauthorized(c *gin.Context, message string) {
	jsonutil.JSON401(c, NewErrorResponse(message))
}

// JSONForbidden 返回403错误响应
func JSONForbidden(c *gin.Context, message string) {
	jsonutil.JSON403(c, NewErrorResponse(message))
}

// JSONNotFound 返回404错误响应
func JSONNotFound(c *gin.Context, message string) {
	jsonutil.JSON404(c, NewErrorResponse(message))
}

// JSONTooManyRequests 返回429错误响应
func JSONTooManyRequests(c *gin.Context, message string) {
	jsonutil.JSON(c, http.StatusTooManyRequests, NewErrorResponse(message))
}

// JSONInternalError 返回500错误响应
func JSONInternalError(c *gin.Context, message string) {
	jsonutil.JSON500(c, NewErrorResponse(message))
}

// JSONServiceUnavailable 返回503错误响应
func JSONServiceUnavailable(c *gin.Context, message string) {
	jsonutil.JSON(c, http.StatusServiceUnavailable, NewErrorResponse(message))
}
