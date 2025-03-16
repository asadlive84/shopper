package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/asadlive84/shopper/api-gateway/config"
	"github.com/asadlive84/shopper/api-gateway/internal/adapters/grpc"
	ad "github.com/asadlive84/shopper/api-gateway/internal/adapters/http"
	"github.com/asadlive84/shopper/api-gateway/internal/application/core/api"
	"github.com/asadlive84/shopper/api-gateway/internal/event"
	"github.com/asadlive84/shopper/api-gateway/internal/logger"
	"github.com/asadlive84/shopper/api-gateway/internal/rabbitmq"
	"github.com/asadlive84/shopper/api-gateway/internal/tracing"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"
)

func main() {
	cfg := config.NewConfig()

	zapLogger := logger.NewLogger()
	defer zapLogger.Sync()

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	tp, err := tracing.InitTracer("api-gateway", cfg.JaegerEndpoint)
	if err != nil {
		zapLogger.Fatal("Failed to initialize tracer", zap.Error(err))
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			zapLogger.Error("Failed to shutdown tracer", zap.Error(err))
		}
	}()

	rabbitClient, err := rabbitmq.NewRabbitMQClient(cfg.RabbitMQ.URL, zapLogger)
	if err != nil {
		zapLogger.Fatal("Failed to connect to RabbitMQ", zap.Error(err))
	}
	defer rabbitClient.Close()

	err = rabbitClient.Setup(cfg.RabbitMQ.Exchange, cfg.RabbitMQ.Queue)
	if err != nil {
		zapLogger.Fatal("Failed to setup RabbitMQ", zap.Error(err))
	}


	
	zapLogger.Info("Connecting to User gRPC service", zap.String("addr", cfg.GRPCUserServiceAddr))
	grpcClient := grpc.NewUserGRPCClient(cfg.GRPCUserServiceAddr)
	
	eventManager := event.NewEventManager(zapLogger, rabbitClient, grpcClient)
	go eventManager.ConsumeEvents(cfg.RabbitMQ.Queue)


	apiService := api.NewAPIService(grpcClient, cfg.JWTSecret, zapLogger)

	r := gin.Default()
	handler := ad.NewHandler(apiService, cfg.JWTSecret, zapLogger, rabbitClient, grpcClient)
	handler.SetupRoutes(r)

	// HTTP server with graceful shutdown
	httpServer := &http.Server{
		Addr:    cfg.HTTPPort,
		Handler: r,
	}
	go func() {
		zapLogger.Info("Starting HTTP server", zap.String("port", cfg.HTTPPort))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zapLogger.Fatal("Failed to run HTTP server", zap.Error(err))
		}
	}()

	// Prometheus metrics server
	promServer := &http.Server{
		Addr:    ":9092",
		Handler: promhttp.Handler(),
	}
	go func() {
		zapLogger.Info("Starting Prometheus metrics server", zap.String("port", ":9092"))
		if err := promServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zapLogger.Fatal("Failed to start Prometheus server", zap.Error(err))
		}
	}()

	// Graceful shutdown handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	zapLogger.Info("Received shutdown signal, shutting down...")

	// Shutdown HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		zapLogger.Error("Failed to shutdown HTTP server", zap.Error(err))
	} else {
		zapLogger.Info("HTTP server stopped")
	}

	// Shutdown Prometheus server
	if err := promServer.Shutdown(ctx); err != nil {
		zapLogger.Error("Failed to shutdown Prometheus server", zap.Error(err))
	} else {
		zapLogger.Info("Prometheus server stopped")
	}

	zapLogger.Info("Service stopped gracefully")
}
