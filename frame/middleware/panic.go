package middleware

import (
	"eva_services_go/frame/errcode"
	"eva_services_go/frame/protocol"
	"eva_services_go/logger"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

func ThrowPanic() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func(c *gin.Context) {
			if err := recover(); err != nil {
				logger.PrintError("gin Panic: %s", errcode.GetSystemPanic(err))

				reqMsg := protocol.SubsysReqBody{}
				body, _ := ioutil.ReadAll(c.Request.Body)

				var json = jsoniter.ConfigCompatibleWithStandardLibrary
				_ = json.Unmarshal(body, &reqMsg)
				reqMsg.Head.MsgType = "response"
				reqMsg.Head.Timestamp = strconv.FormatInt(time.Now().UTC().Unix(), 10)
				jsonResponsev2(c, http.StatusInternalServerError, errcode.ERROR_SERVER_ERROR, reqMsg.Head)
				c.Abort()
			}
		}(c)
		c.Next()
	}
}
