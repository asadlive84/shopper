package config

import (
	"log"
	"os"
	"strconv"
)

// Config structure for application configuration
type Config struct {
	GRPCPort       string
	JaegerEndpoint string
	RabbitMQ       RabbitMQ
	Postgres       Postgres
	MongoDB        MongoDB
}

// RabbitMQ configuration structure
type RabbitMQ struct {
	URL      string
	Queue    string
	Exchange string
}

// PostgreSQL configuration structure
type Postgres struct {
	DSN string
}

// MongoDB configuration structure
type MongoDB struct {
	URI      string
	Database string
}

// Option type for functional options pattern
type Option func(*Config)

// NewConfig creates a new Config instance with optional functional parameters
func NewConfig(options ...Option) *Config {
	cfg := &Config{
		GRPCPort:       getEnv("GRPC_PORT", ":50052"),
		JaegerEndpoint: getEnv("JAEGER_ENDPOINT", "http://localhost:14268/api/traces"),
		RabbitMQ: RabbitMQ{
			URL:      getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
			Queue:    getEnv("RABBITMQ_QUEUE", "user_tasks"),
			Exchange: getEnv("RABBITMQ_EXCHANGE", "user_exchange"),
		},
		Postgres: Postgres{
			DSN: getEnv("POSTGRES_DSN", "host=localhost user=postgres password=postgres dbname=userdb port=5438 sslmode=disable"),
		},
		MongoDB: MongoDB{
			URI:      getEnv("MONGODB_URI", "mongodb://localhost:27017"),
			Database: getEnv("MONGODB_DATABASE", "userdb"),
		},
	}

	// Apply functional options
	for _, option := range options {
		option(cfg)
	}

	return cfg
}

// Helper function to get environment variable with default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// Helper function to get an environment variable as an integer
func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		log.Println("Invalid integer value for", key, ":", value)
	}
	return defaultValue
}

// Option function to override gRPC Port
func WithGRPCPort(port string) Option {
	return func(c *Config) {
		c.GRPCPort = port
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

// Option function to override PostgreSQL DSN
func WithPostgresDSN(dsn string) Option {
	return func(c *Config) {
		c.Postgres = Postgres{DSN: dsn}
	}
}

// Option function to override MongoDB settings
func WithMongoDB(uri, database string) Option {
	return func(c *Config) {
		c.MongoDB = MongoDB{
			URI:      uri,
			Database: database,
		}
	}
}
