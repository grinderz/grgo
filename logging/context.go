package logging

import (
	"context"

	"go.uber.org/zap"

	"github.com/grinderz/grgo/libctx"
)

var (
	contextKey           = libctx.Key("zap") //nolint:gochecknoglobals
	defaultContextLogger = zap.NewNop()      //nolint:gochecknoglobals
)

func FromContext(ctx context.Context) *zap.Logger {
	log, ok := ctx.Value(contextKey).(*zap.Logger)
	if !ok {
		return defaultContextLogger
	}

	return log
}

func ToContext(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, contextKey, logger)
}
