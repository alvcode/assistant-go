package logging

import (
	"context"
)

const defaultEnv = "prod"
const TestsEnv = "dev"

type ctxLogger struct{}

func ContextWithLogger(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, ctxLogger{}, logger)
}

func LoggerFromContext(ctx context.Context) *Logger {
	if l, ok := ctx.Value(ctxLogger{}).(*Logger); ok {
		return l
	}
	return NewLogger(defaultEnv)
}
