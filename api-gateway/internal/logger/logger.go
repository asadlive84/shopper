package logger

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

func NewLogger() *zap.Logger {
    config := zap.NewProductionConfig()
    config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
    logger, err := config.Build()
    if err != nil {
        panic("Failed to initialize logger: " + err.Error())
    }
    return logger
}