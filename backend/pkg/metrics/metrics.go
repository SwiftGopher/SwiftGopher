package metrics

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

var (
	totalHttpRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "api_total_http_requests",
		Help: "Total HTTP requests",
	}, []string{})
	totalOkRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "api_total_ok_requests",
		Help: "Total successful requests",
	}, []string{"path", "status"})

	totalErrorRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "api_total_error_requests",
		Help: "Total error requests",
	}, []string{"path", "status"})

	totalOrders = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "api_total_orders",
		Help: "Total orders count",
	}, []string{})
)

func NewCustomRegistry() *prometheus.Registry {
	customRegistry := prometheus.NewRegistry()

	customRegistry.MustRegister(
		totalHttpRequests,
		totalOkRequests,
		totalErrorRequests,
		totalOrders,
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	return customRegistry
}

func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()
		status := strconv.Itoa(c.Writer.Status())

		totalHttpRequests.WithLabelValues().Inc()

		if c.Writer.Status() < 400 {
			totalOkRequests.WithLabelValues(method+" "+path, status).Inc()
		} else {
			totalErrorRequests.WithLabelValues(method+" "+path, status).Inc()
		}

		if c.Writer.Status() == 201 && path == "/orders" {
			totalOrders.WithLabelValues().Inc()
		}
	}
}
