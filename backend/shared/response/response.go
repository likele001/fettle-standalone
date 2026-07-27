package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	TraceID   string      `json:"trace_id,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	traceID, _ := c.Get("trace_id")
	c.JSON(http.StatusOK, Response{
		Code:      0,
		Message:   "success",
		Data:      data,
		TraceID:   getString(traceID),
		Timestamp: time.Now().Unix(),
	})
}

// SuccessWithMessage 成功响应带自定义消息
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	traceID, _ := c.Get("trace_id")
	c.JSON(http.StatusOK, Response{
		Code:      0,
		Message:   message,
		Data:      data,
		TraceID:   getString(traceID),
		Timestamp: time.Now().Unix(),
	})
}

// Error 错误响应
func Error(c *gin.Context, code int, message string) {
	traceID, _ := c.Get("trace_id")
	c.JSON(getHTTPStatus(code), Response{
		Code:      code,
		Message:   message,
		TraceID:   getString(traceID),
		Timestamp: time.Now().Unix(),
	})
}

// ErrorWithData 错误响应带数据
func ErrorWithData(c *gin.Context, code int, message string, data interface{}) {
	traceID, _ := c.Get("trace_id")
	c.JSON(getHTTPStatus(code), Response{
		Code:      code,
		Message:   message,
		Data:      data,
		TraceID:   getString(traceID),
		Timestamp: time.Now().Unix(),
	})
}

// PageData 分页数据
type PageData struct {
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	HasMore  bool        `json:"has_more"`
	List     interface{} `json:"list"`
}

// getString 安全转换字符串
func getString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// getHTTPStatus 根据业务码获取HTTP状态码
func getHTTPStatus(code int) int {
	switch {
	case code == 0:
		return http.StatusOK
	case code >= 1001 && code <= 1004:
		return http.StatusBadRequest
	case code == 1002:
		return http.StatusUnauthorized
	case code == 1003:
		return http.StatusForbidden
	case code == 1004:
		return http.StatusNotFound
	case code >= 2001 && code <= 2002:
		return http.StatusBadGateway
	case code == 3001:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}
