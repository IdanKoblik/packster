package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HttpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "packster_http_requests_total",
		Help: "Total HTTP requests by method, path, and status code",
	}, []string{"method", "path", "status"})

	HttpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "packster_http_request_duration_seconds",
		Help:    "HTTP request latency in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})

	AuthCacheHits = promauto.NewCounter(prometheus.CounterOpts{
		Name: "packster_auth_cache_hits_total",
		Help: "Total auth token lookups served from Redis cache",
	})

	AuthCacheMisses = promauto.NewCounter(prometheus.CounterOpts{
		Name: "packster_auth_cache_misses_total",
		Help: "Total auth token lookups that fell through to MySQL",
	})

	AuthFailures = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "packster_auth_failures_total",
		Help: "Total authentication failures by reason",
	}, []string{"reason"})

	ArtifactUploadsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "packster_artifact_uploads_total",
		Help: "Total artifact uploads by product and status",
	}, []string{"product", "status"})

	ArtifactDownloadsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "packster_artifact_downloads_total",
		Help: "Total artifact downloads by product",
	}, []string{"product"})

	ArtifactUploadBytesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "packster_artifact_upload_bytes_total",
		Help: "Total bytes uploaded by product",
	}, []string{"product"})

	MySQLUp = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "packster_mysql_up",
		Help: "MySQL availability (1 = up, 0 = down)",
	})

	RedisUp = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "packster_redis_up",
		Help: "Redis availability (1 = up, 0 = down)",
	})
)
