package middleware

import (
	"eva_services_go/frame/errcode"
	"eva_services_go/implements/opentracing"
	"github.com/gin-gonic/gin"
	"strconv"
)

// 接口响应数据结构封装
func jsonResponsev2(ctx *gin.Context, httpCode int, err errcode.AppError, head interface{}) {
	ctx.JSON(httpCode, gin.H{
		"_data": gin.H{
			"_ret": strconv.Itoa(err.ErrorCode),
			"_errCode": strconv.Itoa(err.ErrorCode),
			"_errStr": err.ErrorInfo,
		},
		"_head": head,
	})

	ot := opentracing.GetOpenTracing()
	if err.ErrorCode != 0 {
		ot.SetTag("error", strconv.Itoa(err.ErrorCode))
		ot.SetTag("error.kind", err.ErrorInfo)
	}
	ot.Dump()
}
