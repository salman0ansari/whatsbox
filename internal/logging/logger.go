package logging

import (
	"os"
	"path/filepath"

	"github.com/salman0ansari/whatsbox/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *zap.Logger

func Setup(cfg *config.Config) error {
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(cfg.LogLevel)); err != nil {
		level = zapcore.InfoLevel
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var encoder zapcore.Encoder
	if cfg.LogFormat == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	var cores []zapcore.Core

	// Stdout/stderr output
	if cfg.LogOutput == "stdout" || cfg.LogOutput == "both" {
		consoleCore := zapcore.NewCore(
			encoder,
			zapcore.AddSync(os.Stdout),
			level,
		)
		cores = append(cores, consoleCore)
	}

	// File output
	if cfg.LogOutput == "file" || cfg.LogOutput == "both" {
		// Ensure log directory exists
		logDir := filepath.Dir(cfg.LogFilePath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return err
		}

		fileWriter := &lumberjack.Logger{
			Filename:   cfg.LogFilePath,
			MaxSize:    cfg.LogFileMaxSize, // MB
			MaxBackups: cfg.LogFileMaxBackups,
			MaxAge:     30, // days
			Compress:   true,
		}

		fileCore := zapcore.NewCore(
			encoder,
			zapcore.AddSync(fileWriter),
			level,
		)
		cores = append(cores, fileCore)
	}

	core := zapcore.NewTee(cores...)
	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return nil
}

func Sync() {
	if Logger != nil {
		_ = Logger.Sync()
	}
}

// Helper functions for common log operations
func Info(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	Logger.Debug(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Logger.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Logger.Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	Logger.Fatal(msg, fields...)
}

func With(fields ...zap.Field) *zap.Logger {
	return Logger.With(fields...)
}
