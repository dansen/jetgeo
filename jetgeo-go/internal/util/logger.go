package util

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger 根据 LOG_LEVEL (debug/info) 创建 logger
func NewLogger() *zap.Logger {
	level := zapcore.InfoLevel
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		switch strings.ToLower(v) {
		case "debug":
			level = zapcore.DebugLevel
		case "info":
			level = zapcore.InfoLevel
		}
	}
	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(level)
	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	return logger
}
