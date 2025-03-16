package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RequestCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_gateway_requests_total",
			Help: "Total number of requests",
		},
		[]string{"method", "endpoint", "status"},
	)
	RequestLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "api_gateway_request_duration_seconds",
			Help: "Request latency in seconds",
		},
		[]string{"method", "endpoint"},
	)
	UsersCreated = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "users_created_total",
			Help: "Total number of users created",
		},
	)
)