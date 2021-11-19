package middleware

import (
	"github.com/gin-gonic/gin"
)

func InitContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		//logID := toolkit.RandomLowerLetterString(16)
		//ctx := opentracing.SetLogIdToContext(c.Request.Context(), logID)
		//c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
