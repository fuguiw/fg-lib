package log

import (
	"context"
	"log/slog"
)

type ctxKey string

const (
	TraceIDKey ctxKey = "trace_id"
	ReqIDKey   ctxKey = "request_id"
	UserIDKey  ctxKey = "user_id"
)

func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

func WithRequestID(ctx context.Context, reqID string) context.Context {
	return context.WithValue(ctx, ReqIDKey, reqID)
}

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// contextFields returns a list of slog.Attr from the context.
func contextFields(ctx context.Context) []slog.Attr {
	var attrs []slog.Attr
	if v, ok := ctx.Value(TraceIDKey).(string); ok {
		attrs = append(attrs, slog.String("trace_id", v))
	}
	if v, ok := ctx.Value(ReqIDKey).(string); ok {
		attrs = append(attrs, slog.String("request_id", v))
	}
	if v, ok := ctx.Value(UserIDKey).(string); ok {
		attrs = append(attrs, slog.String("user_id", v))
	}
	return attrs
}
