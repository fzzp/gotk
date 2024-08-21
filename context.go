package gotk

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type ContextKey string

var (
	VersionCtxKey    ContextKey = "gotk_version"
	RequestIDCtxKey  ContextKey = "gotk_request_id"
	ReadJSONMaxBytes            = 2 << 20 // 默认2MB
)

// SetApiVersion 设置版本信息到上下文
func SetVersionCtx(next http.Handler, version string) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), VersionCtxKey, version)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

// SetRequestIDCtx 设置request id，是一个v4 uuid
// 如果 uuid.NewRandom() 报错，创建一个随机字符串
func SetRequestIDCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rid string
		id, err := uuid.NewRandom()
		if err != nil {
			rid = RandomString(10) + strconv.Itoa(int(time.Now().UnixMilli()))
		} else {
			rid = id.String()
		}
		ctx := context.WithValue(r.Context(), RequestIDCtxKey, rid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
