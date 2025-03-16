package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RequestCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_service_requests_total",
			Help: "Total number of requests handled by the User Service",
		},
		[]string{"method", "endpoint", "status"},
	)
	RequestLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "user_service_request_duration_seconds",
			Help: "Latency of requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
)