package middleware

import (
	"github.com/gin-gonic/gin"
	"go-frame/frame/errcode"
	"net/http"
)

var auth = NewJWT()

func CheckUserToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("token")
		if token == "" {
			jsonResponsev2(c, http.StatusUnauthorized, errcode.ERROR_TOKEN_EMPTY, "")
			c.Abort()
			return
		}
		claims, err := auth.ParseToken(token)
		if err != nil {
			jsonResponsev2(c, http.StatusUnauthorized, errcode.ERROR_TOKEN_INVALID, "")
			c.Abort()
			return
		}
		if claims == nil || claims.UID == 0 {
			jsonResponsev2(c, http.StatusUnauthorized, errcode.ERROR_TOKEN_INVALID, "")
			c.Abort()
			return
		}
		//log.Printf("token == %v, \n uid = %v", token, claims.UID)

		c.Set("uid", claims.UID)
		c.Next()
	}
}
