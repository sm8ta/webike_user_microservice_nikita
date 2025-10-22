package ports

import (
	"time"

	"github.com/gin-gonic/gin"
)

type MetricsPort interface {
	IncrementCounter(name string, labels map[string]string)
	RecordDuration(name string, duration time.Duration, labels map[string]string)
	RecordMetrics(c *gin.Context, start time.Time)
}
