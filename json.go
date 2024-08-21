package gotk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type JSONContextKey string

var (
	VersionCtxKey    JSONContextKey = "gotk_version"
	RequestIDCtxKey  JSONContextKey = "gotk_request_id"
	ReadJSONMaxBytes                = 2 << 20 // 默认2MB
)

// JSON 响应数据
type JSON struct {
	BizCode   string `json:"bizCode"`             // 业务编码
	Message   string `json:"message"`             // 客户消息
	Data      any    `json:"data"`                // 任意数据
	Version   string `json:"version,omitempty"`   // 版本信息
	RequestID string `json:"requestId,omitempty"` // 请求Id，做简单的链路追踪
}

// NewJSON 创建一个JSON结构，一般用于响应http请求
func NewJSON(data any, bizCode, message, version, requestId string) *JSON {
	return &JSON{
		BizCode:   bizCode,
		Message:   message,
		Data:      data,
		Version:   version,
		RequestID: requestId,
	}
}

// FormCtxGetJSON 从context里找Version和RequestID，并返回一个JSON结构体
func FormCtxGetJSON(ctx context.Context, data any, bizCode, message string) *JSON {
	version, _ := ctx.Value(VersionCtxKey).(string)
	requestId, _ := ctx.Value(RequestIDCtxKey).(string)
	return &JSON{
		BizCode:   bizCode,
		Message:   message,
		Data:      data,
		Version:   version,
		RequestID: requestId,
	}
}

// ReadJSON 读取入参，绑定到 dst 上
func ReadJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	// 限制请求体大小
	r.Body = http.MaxBytesReader(w, r.Body, int64(ReadJSONMaxBytes))

	// 使用请求体创建一个解码器
	dec := json.NewDecoder(r.Body)

	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		switch {
		case errors.As(err, &syntaxError):
			return errors.New("请输入JSON格式请求体")
		case errors.Is(err, io.EOF):
			return errors.New("请求体不能为空")
		case strings.Contains(err.Error(), "http: request body too large"):
			return fmt.Errorf("请求体大小不能超过 %d MB", ReadJSONMaxBytes)
		default:
			log.Println("app.readJSON 未知错误，请检查参数: ", err.Error())
			return errors.New("未知错误，请检查参数")
		}
	}

	return nil
}

// WriteJSON 写入响应
//
// 如果 apiErr 为 nil，则使用内置ApiError, http statuscode 为 500
//
// 如果执行有错误，会追加到 apiErr 到 exception 并返回
func WriteJSON(w http.ResponseWriter, r *http.Request, apiErr *ApiError, data interface{}, headers ...http.Header) *ApiError {
	if apiErr == nil {
		apiErr = &ApiError{
			message:    http.StatusText(http.StatusInternalServerError),
			bizCode:    strconv.Itoa(http.StatusInternalServerError),
			exception:  errors.New("apiErr *ApiError 不能为nil"),
			statusCode: http.StatusInternalServerError,
		}
	}

	js := FormCtxGetJSON(r.Context(), data, apiErr.bizCode, apiErr.message)

	buf, err := json.Marshal(js)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Status", http.StatusText(http.StatusInternalServerError))
		apiErr.AsException(err, "json无法解析data数据")
		return apiErr
	}

	// 添加请求头，如果有 headers => []http.Header => []map[string][]string
	if len(headers) > 0 {
		for i := range headers {
			for key, val := range headers[i] {
				w.Header()[key] = val
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Status", http.StatusText(apiErr.StatusCode()))
	w.WriteHeader(apiErr.StatusCode())

	_, err = w.Write(buf)
	if err != nil {
		// 当进入这里，说明无法写入响应，没有任何数据返回 [请求方看到该网页无法正常运作]
		return apiErr.AsException(err, "写入response错误")
	}

	return nil
}
