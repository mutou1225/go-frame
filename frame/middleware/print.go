package middleware

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/mutou1225/go-frame/implements/opentracing"
	"github.com/mutou1225/go-frame/logger"
	"io/ioutil"
	"net/http"
)

func PrintPostData() gin.HandlerFunc {
	return func(c *gin.Context) {

		body, _ := ioutil.ReadAll(c.Request.Body)

		logger.PrintInfo("%sRequest Interface Statr ......%s", logger.Purple, logger.Reset)
		logger.PrintInfo("%sPath:%s %s", logger.Purple, logger.Reset, c.Request.URL.Path)
		//logger.PrintInfo("Url Host: %s", c.Request.URL.Host)
		strTracing := ""
		for k, v := range c.Request.Header {
			logger.PrintInfo("%sHeader:%s %s, value: %s", logger.Purple, logger.Reset, k, v)

			if k == opentracing.TraceHeader {
				strTracing = v[0]
			}
		}
		logger.PrintInfo("%spacket:%s %s", logger.Purple, logger.Reset, body)

		// 开启OpenTracing
		opentracing.GetOpenTracing().FromContextSetName(strTracing, c.Request.URL.Path)

		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		c.Next()
	}
}

// GetIP gets a requests IP address by reading off the forwarded-for
// header (for proxies) and falls back to use the remote address.
func GetRequestsIp(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}
