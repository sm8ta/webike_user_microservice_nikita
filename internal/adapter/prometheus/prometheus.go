package prometheus

import (
	"fmt"
	"time"
	"webike_services/webike_User-microservice_Nikita/internal/core/ports"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

type PrometheusAdapter struct {
	httpRequestsTotal   *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec
}

func NewPrometheusAdapter() ports.MetricsPort {
	adapter := &PrometheusAdapter{
		httpRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"path", "method", "status", "app_name"},
		),
		httpRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "api_request_duration_seconds",
				Help:    "Duration API requests",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"path", "method", "status", "app_name"},
		),
	}

	prometheus.MustRegister(adapter.httpRequestsTotal)
	prometheus.MustRegister(adapter.httpRequestDuration)

	// ебаная строчка
	adapter.httpRequestsTotal.WithLabelValues("/health", "GET", "200", "user_microservice").Add(0)
	return adapter
}

func (p *PrometheusAdapter) IncrementCounter(name string, labels map[string]string) {
	p.httpRequestsTotal.WithLabelValues(
		labels["path"],
		labels["method"],
		labels["status"],
		"user_microservice",
	).Inc()
}

func (p *PrometheusAdapter) RecordDuration(name string, duration time.Duration, labels map[string]string) {
	p.httpRequestDuration.WithLabelValues(
		labels["path"],
		labels["method"],
		labels["status"],
		"user_microservice",
	).Observe(duration.Seconds())
}

func (p *PrometheusAdapter) RecordMetrics(c *gin.Context, start time.Time) {
	status := fmt.Sprintf("%d", c.Writer.Status())
	labels := map[string]string{
		"path":   c.Request.URL.Path,
		"method": c.Request.Method,
		"status": status,
	}

	p.IncrementCounter("http_requests_total", labels)
	p.RecordDuration("api_request_duration_seconds", time.Since(start), labels)
}
