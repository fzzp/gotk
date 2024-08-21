package gotk

import (
	"fmt"
	"strings"
)

// ApiError 一个错误结构体，统一返回错误结构，包含可读错误信息、自定义错误编码和原始错误信息
type ApiError struct {
	message    string // 可读错误信息
	bizCode    string // 自定义唯一业务状态码
	exception  error  // 原始错误信息
	statusCode int    // HTTP status code
}

// 收集自定义业务状态码，保持每个状态码都是唯一的。
var bizCodeMap = map[string]struct{}{}

// NewApiError 创建一个ApiError, 如果 bizCode 已存在会触发panic;
// bizCode 是自定义业务编码;
// msg 是人类可读错误提示;
// statusCode 是 HTTP status code;
func NewApiError(statusCode int, bizCode, msg string) *ApiError {
	if _, exists := bizCodeMap[bizCode]; exists {
		panic(fmt.Sprintf("bizCode(%s) already exist, please replace it", bizCode))
	}
	bizCodeMap[bizCode] = struct{}{}

	return &ApiError{
		message:    msg,
		bizCode:    bizCode,
		statusCode: statusCode,
	}
}

// Error 实现 error 类型接口
func (a *ApiError) Error() string {
	if a.exception != nil {
		return fmt.Sprintf("[statusCode: %d, bizCode: %s, message: %s, exception: %v]", a.statusCode, a.bizCode, a.message, a.exception)
	}
	return fmt.Sprintf("[statusCode: %d, bizCode: %s, message: %s]", a.statusCode, a.bizCode, a.message)
}

// Unwrap 解开，提供给 errors.Is 和 errors.As 使用
func (a *ApiError) Unwrap() error {
	return a.exception
}

// BizCode 返回自定义业务错误码
func (a *ApiError) BizCode() string {
	return a.bizCode
}

// Message 返回可读错误信息
func (a *ApiError) Message() string {
	return a.message
}

// StatusCode 返回HTTP Status Code
func (a *ApiError) StatusCode() int {
	return a.statusCode
}

// AsMessage 修改消息
// 返回一个新的 ApiError 指针
func (a *ApiError) AsMessage(msg string) *ApiError {
	return &ApiError{
		bizCode:    a.bizCode,
		message:    msg,
		exception:  a.exception,
		statusCode: a.statusCode,
	}
}

// AsException 添加/追加错误, 返回一个新的 ApiError 指针
func (a *ApiError) AsException(err error, msgs ...string) *ApiError {
	var e error
	apiErr, ok := err.(*ApiError)
	if a.exception == nil {
		if ok {
			e = fmt.Errorf("%w", apiErr.exception)
		} else {
			e = fmt.Errorf("%w", err)
		}
	} else {
		if ok {
			e = fmt.Errorf("%w; %w", a.exception, apiErr.exception)
		} else {
			e = fmt.Errorf("%w; %w", a.exception, err)
		}
	}

	newErr := &ApiError{
		bizCode:    a.bizCode,
		message:    a.message,
		statusCode: a.statusCode,
		exception:  e,
	}
	if len(msgs) > 0 {
		newErr.message = strings.Join(msgs, ",")
	}
	return newErr
}
