package infrastructure

import (
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	HttpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"method", "path", "status"})

	HttpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Duration of HTTP requests",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})

	DatabaseQueryDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "db_query_duration_seconds",
		Help:    "Duration of Database queries",
		Buckets: prometheus.DefBuckets,
	}, []string{"operation", "status"}) // operation: select, insert, etc.

	RedisCommandDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "redis_command_duration_seconds",
		Help:    "Duration of Redis commands",
		Buckets: prometheus.DefBuckets,
	}, []string{"command", "status"})
)

func SetupMetrics(app *fiber.App) {
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))
}
