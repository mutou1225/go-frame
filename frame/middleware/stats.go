package middleware

import (
	"fmt"
	"go-frame/implements/toolkit"
	"time"

	"github.com/gin-gonic/gin"
	metrics "github.com/rcrowley/go-metrics"
)

const (
	ginLatencyMetric = "gin.latency"
	ginStatusMetric  = "gin.status"
	ginRequestMetric = "gin.request"
)

var (
	DefaultRegistry = metrics.NewRegistry()
)

//Report from default metric registry
func StatsReport() metrics.Registry {
	return DefaultRegistry
}

//RequestStats middleware
func RequestStats() gin.HandlerFunc {
	go func() {
		t1 := time.NewTimer(toolkit.GetNextDaySecond())
		for {
			select {
			case <-t1.C:
				DefaultRegistry = metrics.NewRegistry()
				t1.Reset(toolkit.GetNextDaySecond())
			}
		}
	}()

	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		latency := metrics.GetOrRegisterTimer(fmt.Sprintf("%s.%s", ginRequestMetric, c.Request.URL.Path), DefaultRegistry)
		latency.UpdateSince(start)

		status := metrics.GetOrRegisterMeter(fmt.Sprintf("%s.%d", ginStatusMetric, c.Writer.Status()), DefaultRegistry)
		status.Mark(1)
	}
}
