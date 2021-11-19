package app

import (
	"eva_services_go/frame/errcode"
	"github.com/gin-gonic/gin"
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
