package gotk

import (
	"context"
	"io"
	"log/slog"

	"gopkg.in/natefinch/lumberjack.v2"
)

type SlogHandle func(ctx context.Context, r slog.Record) error

// LogHandler 定义一个LogHandler，方便扩展和自定义一些功能
type LogHandler struct {
	slog.Handler
	handlers []*SlogHandle
}

// RegisterSlogHandler 按照注册顺序执行
func (lh *LogHandler) RegisterSlogHandler(handlers ...*SlogHandle) {
	if lh.handlers == nil {
		lh.handlers = make([]*SlogHandle, 0)
	}
	for _, fn := range handlers {
		lh.handlers = append(lh.handlers, fn)
	}
}

// Handle 如果ctx有 request_id 附加到日志输出
func (lh *LogHandler) Handle(ctx context.Context, r slog.Record) error {
	// requestID 需要在gin路由中间件设置上去
	requestID, ok := ctx.Value(RequestIDCtxKey).(string)

	// 附加到slog属性上
	if ok && requestID != "" {
		r.AddAttrs(slog.String("request_id", requestID))
	}

	for _, fn := range lh.handlers {
		err := (*fn)(ctx, r)
		if err != nil {
			return err
		}
	}

	return lh.Handler.Handle(ctx, r)
}

// NewLogger 创建JSON格式日志，并设置为全局默认的slog;level=(DEBUG,INFO,WARN,ERROR); output 日志输出位置
func NewLogger(level string, AddSource bool, output io.Writer) {
	logLevel := convertLevel(level)

	handler := slog.NewJSONHandler(
		output,
		&slog.HandlerOptions{
			AddSource: AddSource, // 暂时不添加日志输出位置，此项目大部分都是同一个地方输出
			Level:     logLevel,
		})

	myHandler := LogHandler{Handler: handler}

	l := slog.New(&myHandler)

	// 设为全局默认 slog 实例
	slog.SetDefault(l)
}

// DefaultOutput 默认日志输出
func DefaultOutput(filename string) io.Writer {
	return &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    500, // megabytes
		MaxBackups: 28,
		MaxAge:     30,   //days
		Compress:   true, // disabled by default
	}
}

// convertLevel 从字符串转换成slog.Leveler
func convertLevel(level string) slog.Leveler {
	switch level {
	case "DEBUG":
		return slog.LevelDebug // -4
	case "INFO":
		return slog.LevelInfo // 0
	case "WARN":
		return slog.LevelWarn // 4
	case "ERROR":
		return slog.LevelError // 8
	default:
		return slog.LevelInfo // 0
	}
}
