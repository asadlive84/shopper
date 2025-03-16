package monitoring

import (
	"time"

	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

// PrometheusInterceptor is a unary interceptor for monitoring gRPC requests
func PrometheusInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		start := time.Now()
		method := info.FullMethod // gRPC মেথড নাম (যেমন "/user.UserService/CreateUser")
		endpoint := method        // সরলীকরণের জন্য মেথডকে এন্ডপয়েন্ট হিসেবে ব্যবহার

		// পরবর্তী হ্যান্ডলার কল করো
		resp, err = handler(ctx, req)

		// লেটেন্সি এবং স্ট্যাটাস গণনা
		latency := time.Since(start).Seconds()
		status := "success"
		if err != nil {
			status = "error"
		}

		// মেট্রিক্স রেকর্ড
		RequestCounter.WithLabelValues(method, endpoint, status).Inc()
		RequestLatency.WithLabelValues(method, endpoint).Observe(latency)

		return resp, err
	}
}

const tracerName = "user-service"

// TracingInterceptor adds tracing to gRPC requests
func TracingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// ট্রেসার পাওয়া
		tracer := otel.Tracer(tracerName)

		// কনটেক্সট থেকে প্যারেন্ট স্প্যান পাওয়া, নতুন স্প্যান শুরু
		ctx, span := tracer.Start(ctx, info.FullMethod, trace.WithAttributes(
			attribute.String("grpc.method", info.FullMethod),
		))
		defer span.End()

		// হ্যান্ডলার কল
		resp, err = handler(ctx, req)

		// এরর থাকলে স্প্যানে রেকর্ড
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}

		return resp, err
	}
}
