package app

import (
	"github.com/gin-gonic/gin"
	"github.com/mutou1225/go-frame/frame/errcode"
	"net/http"
	"time"
)

// 打印/请求
func IndexApi(c *gin.Context) {
	JsonResponseOther(c, http.StatusOK, errcode.SUCCESS, "Welcome")
	return
}

// ping
func PingApi(c *gin.Context) {
	JsonResponseOther(c, http.StatusOK, errcode.SUCCESS, time.Now())
}
