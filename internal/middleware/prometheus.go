package middleware

import (
	"artifactor/internal/metrics"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}

		metrics.HttpRequestsTotal.WithLabelValues(
			c.Request.Method,
			path,
			strconv.Itoa(c.Writer.Status()),
		).Inc()

		metrics.HttpRequestDuration.WithLabelValues(
			c.Request.Method,
			path,
		).Observe(time.Since(start).Seconds())
	}
}
