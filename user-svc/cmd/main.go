package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/asadlive84/shopper/user-svc/config"
	"github.com/asadlive84/shopper/user-svc/internal/adapters/db/mongodb"
	"github.com/asadlive84/shopper/user-svc/internal/adapters/db/postgresql"
	gc "github.com/asadlive84/shopper/user-svc/internal/adapters/grpc"
	"github.com/asadlive84/shopper/user-svc/internal/adapters/rabbitmq"
	"github.com/asadlive84/shopper/user-svc/internal/application/core"
	"github.com/asadlive84/shopper/user-svc/internal/logger"
	"github.com/asadlive84/shopper/user-svc/internal/monitoring"
	"github.com/asadlive84/shopper/user-svc/internal/tracing"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// ChainUnaryInterceptors combines multiple unary interceptors
func ChainUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		n := len(interceptors)
		if n == 0 {
			return handler(ctx, req)
		}

		// Build the interceptor chain
		chainedHandler := handler
		for i := n - 1; i >= 0; i-- {
			currentInterceptor := interceptors[i]
			nextHandler := chainedHandler
			chainedHandler = func(currentCtx context.Context, currentReq interface{}) (interface{}, error) {
				return currentInterceptor(currentCtx, currentReq, info, nextHandler)
			}
		}

		// Call the chained handler
		return chainedHandler(ctx, req)
	}
}

func main() {
	// Load configuration
	cfg := config.NewConfig()

	// Initialize logger
	zapLogger := logger.NewLogger()
	defer zapLogger.Sync()

	// Set OpenTelemetry global propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	// Initialize tracer (Jaeger)
	tp, err := tracing.InitTracer("user-service", cfg.JaegerEndpoint)
	if err != nil {
		zapLogger.Fatal("Failed to initialize tracer", zap.Error(err))
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			zapLogger.Error("Failed to shutdown tracer", zap.Error(err))
		}
	}()

	// Postgres database setup
	postgresDB, err := postgresql.Adapter(cfg.Postgres.DSN)
	if err != nil {
		zapLogger.Fatal("Failed to connect to Postgres", zap.Error(err))
	}

	// MongoDB setup
	mongoDB, err := mongodb.Adapter(cfg.MongoDB.URI, cfg.MongoDB.Database)
	if err != nil {
		zapLogger.Fatal("Failed to connect to MongoDB", zap.Error(err))
	}

	// RabbitMQ setup
	rabbitClient, err := rabbitmq.Adapter(cfg.RabbitMQ.URL, zapLogger)
	if err != nil {
		zapLogger.Fatal("Failed to connect to RabbitMQ", zap.Error(err))
	}
	defer rabbitClient.Close()

	err = rabbitClient.Setup(cfg.RabbitMQ.Exchange, cfg.RabbitMQ.Queue)
	if err != nil {
		zapLogger.Fatal("Failed to setup RabbitMQ", zap.Error(err))
	}

	// Start RabbitMQ consumer in a goroutine
	go func() {
		err := rabbitClient.Consume(cfg.RabbitMQ.Queue, func(ctx context.Context, msg string) {
			zapLogger.Info("Consumed message", zap.String("message", msg))
		})
		if err != nil {
			zapLogger.Error("Failed to start RabbitMQ consumer", zap.Error(err))
		}
	}()

	// Initialize core API service
	apiService := core.NewApplication(postgresDB, mongoDB, rabbitClient, zapLogger)
	// hydraClient := monitoring.NewHydraClient()
	// gRPC server with interceptors
	grpcOpts := []grpc.ServerOption{
		grpc.UnaryInterceptor(ChainUnaryInterceptors(
			monitoring.PrometheusInterceptor(),
			logger.LoggingInterceptor(zapLogger),
			monitoring.TracingInterceptor(),
			// monitoring.AuthInterceptor(hydraClient),
		)),
	}
	grpcServer := grpc.NewServer(grpcOpts...)

	// Start gRPC server in a goroutine
	go func() {
		zapLogger.Info("Starting User gRPC server", zap.String("port", cfg.GRPCPort))
		if err := gc.StartGRPCServer(cfg.GRPCPort, apiService, zapLogger, grpcServer); err != nil {
			zapLogger.Fatal("Failed to start gRPC server", zap.Error(err))
		}
	}()

	// Prometheus metrics server
	promServer := &http.Server{
		Addr:    ":9091",
		Handler: promhttp.Handler(),
	}
	go func() {
		zapLogger.Info("Starting Prometheus metrics server", zap.String("port", ":9091"))
		if err := promServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zapLogger.Fatal("Failed to start Prometheus server", zap.Error(err))
		}
	}()

	// Graceful shutdown handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	zapLogger.Info("Received shutdown signal, shutting down...")

	// Stop gRPC server gracefully
	grpcServer.GracefulStop()
	zapLogger.Info("gRPC server stopped")

	// Stop Prometheus server gracefully
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := promServer.Shutdown(shutdownCtx); err != nil {
		zapLogger.Error("Failed to shutdown Prometheus server", zap.Error(err))
	} else {
		zapLogger.Info("Prometheus server stopped")
	}

	zapLogger.Info("Service stopped gracefully")
}
