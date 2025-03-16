package config

import (
	"os"
)

// RabbitMQ configuration structure
type RabbitMQ struct {
	URL      string
	Queue    string
	Exchange string
}

// Config structure containing application configurations
type Config struct {
	GRPCUserServiceAddr string
	HTTPPort            string
	JWTSecret           string
	JaegerEndpoint      string
	RabbitMQ            RabbitMQ
}

// Option type for functional options pattern
type Option func(*Config)

// NewConfig creates a new Config instance with optional functional parameters
func NewConfig(options ...Option) *Config {
	// Default Config
	config := &Config{
		GRPCUserServiceAddr: getEnvOrDefault("GRPC_USER_SERVICE_ADDR", "localhost:50051"),
		HTTPPort:            getEnvOrDefault("HTTP_PORT", ":8080"),
		JWTSecret:           getEnvOrDefault("JWT_SECRET", "your-secret-key"),
		JaegerEndpoint:      getEnvOrDefault("JAEGER_ENDPOINT", "http://localhost:14268/api/traces"),
		RabbitMQ: RabbitMQ{
			URL:      getEnvOrDefault("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
			Queue:    getEnvOrDefault("RABBITMQ_QUEUE", "user_tasks"),
			Exchange: getEnvOrDefault("RABBITMQ_EXCHANGE", "user_exchange"),
		},
	}

	// Apply provided options
	for _, option := range options {
		option(config)
	}

	return config
}

// Option function to override gRPC User Service Address
func WithGRPCUserServiceAddr(addr string) Option {
	return func(c *Config) {
		c.GRPCUserServiceAddr = addr
	}
}

// Option function to override HTTP Port
func WithHTTPPort(port string) Option {
	return func(c *Config) {
		c.HTTPPort = port
	}
}

// Option function to override JWT Secret
func WithJWTSecret(secret string) Option {
	return func(c *Config) {
		c.JWTSecret = secret
	}
}

// Option function to override Jaeger Endpoint
func WithJaegerEndpoint(endpoint string) Option {
	return func(c *Config) {
		c.JaegerEndpoint = endpoint
	}
}

// Option function to override RabbitMQ settings
func WithRabbitMQ(url, queue, exchange string) Option {
	return func(c *Config) {
		c.RabbitMQ = RabbitMQ{
			URL:      url,
			Queue:    queue,
			Exchange: exchange,
		}
	}
}

// Helper function to get an environment variable or return a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
