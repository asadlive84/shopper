package logger

import (
	"context"
	"time"

	// "github.com/asadlive84/shopper/user-svc/internal/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func NewLogger() *zap.Logger {
	logger, _ := zap.NewProduction()
	return logger
}


// LoggingInterceptor logs gRPC requests and errors
func LoggingInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		start := time.Now()
		method := info.FullMethod

		// হ্যান্ডলার কল
		resp, err = handler(ctx, req)

		// লেটেন্সি
		latency := time.Since(start)

		// লগিং
		if err != nil {
			logger.Error("gRPC request failed",
				zap.String("method", method),
				zap.Duration("latency", latency),
				zap.Error(err),
			)
		} else {
			logger.Info("gRPC request succeeded",
				zap.String("method", method),
				zap.Duration("latency", latency),
			)
		}

		return resp, err
	}
}