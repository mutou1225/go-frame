package middleware

import (
	"context"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"go-frame/frame/errcode"
	"go-frame/frame/protocol"
	"golang.org/x/time/rate"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

// maxBurstSize 每秒允许最大请求数
func Limiter(maxBurstSize int) gin.HandlerFunc {
	limiter := rate.NewLimiter(rate.Every(time.Second*1), maxBurstSize)
	return func(c *gin.Context) {
		ctx, _ := context.WithTimeout(context.Background(), time.Millisecond * 200)
		if err := limiter.Wait(ctx); err == nil {
			c.Next()
			return
		}

		reqMsg := protocol.SubsysReqBody{}
		body, _ := ioutil.ReadAll(c.Request.Body)

		var json = jsoniter.ConfigCompatibleWithStandardLibrary
		_ = json.Unmarshal(body, &reqMsg)
		reqMsg.Head.MsgType = "response"
		reqMsg.Head.Timestamp = strconv.FormatInt(time.Now().UTC().Unix(), 10)
		jsonResponsev2(c, http.StatusInternalServerError, errcode.ERROR_LIMINT, reqMsg.Head)
		c.Abort()
		return
	}
}
