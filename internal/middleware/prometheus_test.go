package middleware

import (
	"artifactor/internal/metrics"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestPrometheusMiddleware_IncrementsCounter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(PrometheusMiddleware())
	router.GET("/api/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	before := testutil.ToFloat64(metrics.HttpRequestsTotal.WithLabelValues("GET", "/api/test", "200"))

	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/test", nil))

	after := testutil.ToFloat64(metrics.HttpRequestsTotal.WithLabelValues("GET", "/api/test", "200"))
	assert.Equal(t, float64(1), after-before)
}

func TestPrometheusMiddleware_RecordsErrorStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(PrometheusMiddleware())
	router.GET("/api/fail", func(c *gin.Context) {
		c.Status(http.StatusInternalServerError)
	})

	before := testutil.ToFloat64(metrics.HttpRequestsTotal.WithLabelValues("GET", "/api/fail", "500"))

	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/fail", nil))

	after := testutil.ToFloat64(metrics.HttpRequestsTotal.WithLabelValues("GET", "/api/fail", "500"))
	assert.Equal(t, float64(1), after-before)
}

func TestPrometheusMiddleware_UnknownPathLabel(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// No routes registered — gin returns 404 and FullPath() is empty.
	router := gin.New()
	router.Use(PrometheusMiddleware())

	before := testutil.ToFloat64(metrics.HttpRequestsTotal.WithLabelValues("GET", "unknown", "404"))

	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/nonexistent", nil))

	after := testutil.ToFloat64(metrics.HttpRequestsTotal.WithLabelValues("GET", "unknown", "404"))
	assert.Equal(t, float64(1), after-before)
}
