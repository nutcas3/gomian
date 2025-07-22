package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"sync"
)

var (
	defaultLogger *zap.Logger
	loggerMu      sync.RWMutex
	initialized   bool
)

type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo LogLevel = "info"
	LogLevelWarn LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

type Config struct {
	Level LogLevel

	Development bool

	Encoding string
}

func DefaultConfig() Config {
	return Config{
		Level:       LogLevelInfo,
		Development: false,
		Encoding:    "json",
	}
}

func Initialize(config Config) error {
	loggerMu.Lock()
	defer loggerMu.Unlock()

	var level zapcore.Level
	switch config.Level {
	case LogLevelDebug:
		level = zapcore.DebugLevel
	case LogLevelInfo:
		level = zapcore.InfoLevel
	case LogLevelWarn:
		level = zapcore.WarnLevel
	case LogLevelError:
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	zapConfig := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Development:      config.Development,
		Encoding:         config.Encoding,
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	if config.Development {
		zapConfig.EncoderConfig = zap.NewDevelopmentEncoderConfig()
	} else {
		zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	logger, err := zapConfig.Build()
	if err != nil {
		return err
	}

	defaultLogger = logger
	initialized = true
	return nil
}

func ensureInitialized() {
	loggerMu.RLock()
	init := initialized
	loggerMu.RUnlock()

	if !init {
		_ = Initialize(DefaultConfig())
	}
}

func Logger() *zap.Logger {
	ensureInitialized()
	loggerMu.RLock()
	defer loggerMu.RUnlock()
	return defaultLogger
}

func With(fields ...zap.Field) *zap.Logger {
	return Logger().With(fields...)
}

func Debug(msg string, fields ...zap.Field) {
	Logger().Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	Logger().Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Logger().Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Logger().Error(msg, fields...)
}

func Sync() error {
	loggerMu.RLock()
	defer loggerMu.RUnlock()
	
	if defaultLogger != nil {
		return defaultLogger.Sync()
	}
	return nil
}
