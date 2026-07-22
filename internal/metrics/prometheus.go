// Package metrics provides Prometheus metrics endpoint for Gin.
package metrics

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// PrometheusHandler returns a Gin handler that serves Prometheus metrics.
func PrometheusHandler() gin.HandlerFunc {
	return gin.WrapH(promhttp.Handler())
}
