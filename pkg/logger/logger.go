package logger

import (
	"github.com/natefinch/lumberjack"
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
