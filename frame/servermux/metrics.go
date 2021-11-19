package servermux

import (
	"expvar"
	"fmt"
	"gitee.com/cristiane/go-common/ptool"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"strings"
)

var appStats = expvar.NewMap("appstats")

/*
func NewServerMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux = GetElasticMux(mux)
	//mux = metrics_mux.GetPProfMux(mux)
	mux = GetPrometheusMux(mux)
	return mux
}

func GetPrometheusMux(mux *http.ServeMux) *http.ServeMux {
	mux.Handle("/metrics", promhttp.Handler())
	return mux
}

func GetElasticMux(mux *http.ServeMux) *http.ServeMux {
	mux.HandleFunc("/debug/vars", metricsHandler)
	return mux
}

// metricsHandler print expvar data.
func metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	appStats.Set("Goroutine", expvar.Func(ptool.GetGoroutineCount))
	appStats.Set("Threadcreate", expvar.Func(ptool.GetThreadCreateCount))
	appStats.Set("Block", expvar.Func(ptool.GetBlockCount))
	appStats.Set("Mutex", expvar.Func(ptool.GetMutexCount))
	appStats.Set("Heap", expvar.Func(ptool.GetHeapCount))

	first := true
	report := func(key string, value interface{}) {
		if !first {
			fmt.Fprintf(w, ",\n")
		}
		first = false
		if str, ok := value.(string); ok {
			fmt.Fprintf(w, "%q: %q", key, str)
		} else {
			fmt.Fprintf(w, "%q: %v", key, value)
		}
	}

	fmt.Fprintf(w, "{\n")
	expvar.Do(func(kv expvar.KeyValue) {
		report(kv.Key, kv.Value)
	})
	fmt.Fprintf(w, "\n}\n")
}
 */

func MetricsHandler(c *gin.Context) {
	promhttp.Handler().ServeHTTP(c.Writer, c.Request)
}

func ExpvarHandler(c *gin.Context) {
	appStats.Set("Goroutine", expvar.Func(ptool.GetGoroutineCount))
	appStats.Set("Threadcreate", expvar.Func(ptool.GetThreadCreateCount))
	appStats.Set("Block", expvar.Func(ptool.GetBlockCount))
	appStats.Set("Mutex", expvar.Func(ptool.GetMutexCount))
	appStats.Set("Heap", expvar.Func(ptool.GetHeapCount))

	first := true
	var msg strings.Builder
	report := func(key string, value interface{}) {
		if !first {
			msg.WriteString(",")
		}
		first = false
		if str, ok := value.(string); ok {
			msg.WriteString(fmt.Sprintf("%q: %q", key, str))
		} else {
			msg.WriteString(fmt.Sprintf("%q: %v", key, value))
		}
	}

	msg.WriteString("{")
	expvar.Do(func(kv expvar.KeyValue) {
		report(kv.Key, kv.Value)
	})
	msg.WriteString("}")

	c.String(http.StatusOK, msg.String())
}

