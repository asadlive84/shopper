package api

import (
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *APIService) handleGRPCError(err error, contextMsg string, fields ...zap.Field) error {
	if err == nil {
		return nil
	}

	// Default fields for logging
	logFields := append([]zap.Field{zap.String("context", contextMsg)}, fields...)

	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.InvalidArgument:
			s.logger.Error("Input validation failed", append(logFields, zap.String("details", st.Message()))...)
			return fmt.Errorf("input validation failed: %s", st.Message())
		case codes.AlreadyExists:
			s.logger.Error("Resource already exists", append(logFields, zap.String("details", st.Message()))...)
			return fmt.Errorf("resource already exists: %s", st.Message())
		case codes.NotFound:
			s.logger.Error("Resource not found", append(logFields, zap.String("details", st.Message()))...)
			return fmt.Errorf("resource not found: %s", st.Message())
		case codes.Internal:
			s.logger.Error("Internal server error", append(logFields, zap.String("details", st.Message()))...)
			return fmt.Errorf("internal server error: %s", st.Message())
		case codes.Unavailable:
			s.logger.Error("Service unavailable", append(logFields, zap.String("details", st.Message()))...)
			return fmt.Errorf("service unavailable: %s", st.Message())
		default:
			s.logger.Error("Unknown gRPC error", append(logFields, zap.String("details", st.Message()))...)
			return fmt.Errorf("unknown error: %s", st.Message())
		}
	}

	// Non-gRPC error
	s.logger.Error("Unexpected error", append(logFields, zap.Error(err))...)
	return fmt.Errorf("unexpected error: %v", err)
}
