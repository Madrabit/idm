package common

import (
	"context"
	"github.com/gofiber/contrib/fiberzap/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.Logger
}

func NewLogger(cfg Config) *Logger {
	var zapEncoderCfg = zapcore.EncoderConfig{
		TimeKey:          "timestamp",
		LevelKey:         "level",
		NameKey:          "logger",
		CallerKey:        "caller",
		FunctionKey:      zapcore.OmitKey,
		MessageKey:       "msg",
		StacktraceKey:    "stacktrace",
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeLevel:      zapcore.LowercaseLevelEncoder,
		EncodeTime:       zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000000"),
		EncodeDuration:   zapcore.MillisDurationEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder,
		ConsoleSeparator: "  ",
	}
	var zapCfg = zap.Config{
		Level:       zap.NewAtomicLevelAt(parseLogLevel(cfg.LogLevel)),
		Development: cfg.LogDevelopMode,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:      "json",
		EncoderConfig: zapEncoderCfg,
		// логируем сообщения и ошибки в консоль
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stdout"},
	}
	var logger = zap.Must(zapCfg.Build())
	logger.Info("logger construction succeeded")
	var created = &Logger{logger}
	created.setNewFiberZapLogger()
	return created
}

func (l *Logger) setNewFiberZapLogger() {
	var fiberZapLogger = fiberzap.NewLogger(fiberzap.LoggerConfig{
		SetLogger: l.Logger,
	})
	log.SetLogger(fiberZapLogger)
}

func parseLogLevel(level string) zapcore.Level {
	switch level {
	case "debug", "DEBUG":
		return zapcore.DebugLevel
	case "info", "INFO":
		return zapcore.InfoLevel
	case "warn", "WARN":
		return zapcore.WarnLevel
	case "error", "ERROR":
		return zapcore.ErrorLevel
	case "panic", "PANIC":
		return zapcore.PanicLevel
	case "fatal", "FATAL":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

var ridKey = "requestid"

func (l *Logger) DebugCtx(
	ctx fiber.Ctx,
	msg string,
	fields ...zap.Field,
) {
	rid := ctx.Get(fiber.HeaderXRequestID)
	if rid == "" {
		rid = ctx.Get("X-Request-ID")
	}
	if rid == "" {
		rid = uuid.NewString()
		ctx.Set("X-Request-ID", rid)
		ctx.Locals(ridKey, rid) // чтобы следующие миддлы/хендлеры увидели
	}

	if rid != "" {
		fields = append(fields, zap.String("request_id", rid))
	}
	l.Debug(msg, fields...)
}

func (l *Logger) ErrorCtx(
	ctx context.Context,
	msg string,
	fields ...zap.Field) {
	if rid, ok := ctx.Value(ridKey).(string); ok && rid != "" {
		fields = append(fields, zap.String(string(ridKey), rid))
	}
	l.Error(msg, fields...)
}
