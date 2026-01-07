package logger

import (
	"context"

	"github.com/natefinch/lumberjack"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Options struct {
	Path       string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
}

func Logger(options *Options) *zap.Logger {
	log := &lumberjack.Logger{
		Filename:   options.Path,
		MaxSize:    options.MaxSize,
		MaxBackups: options.MaxBackups,
		MaxAge:     options.MaxAge,
		Compress:   options.Compress,
	}
	writer := zapcore.AddSync(log)
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		writer,
		zap.InfoLevel,
	)
	return zap.New(core)
}

func WithTraceContext(ctx context.Context, logger *zap.Logger) *zap.Logger {
	spanCtx := trace.SpanContextFromContext(ctx)
	if !spanCtx.IsValid() {
		return logger
	}

	return logger.With(
		zap.String("trace_id", spanCtx.TraceID().String()),
		zap.String("span_id", spanCtx.SpanID().String()),
	)
}
