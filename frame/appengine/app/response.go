package app

import (
	"eva_services_go/frame/errcode"
	"eva_services_go/frame/protocol"
	"eva_services_go/implements/opentracing"
	"eva_services_go/implements/toolkit"
	"eva_services_go/logger"
	"fmt"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"net/http"
	"strconv"
	"time"
)

// 接口响应数据结构封装
func JsonResponse(ctx *gin.Context, err errcode.AppError, rspHead protocol.SubsysHeader, data interface{}, errInfo ...string) {
	errStr := ""
	if len(errInfo) > 0 {
		errStr = " " + fmt.Sprint(errInfo)
	}

	// 设置Head参数
	reqTimestamp := rspHead.Timestamp
	rspHead.MsgType = "response"
	rspHead.Timestamp = strconv.FormatInt(toolkit.GetTimeStamp(), 10)

	// 组装响应报文和发送
	dataInfo := protocol.SubsysCommonRsp{
		RetMsg:  err.ErrorInfo + errStr,
		Data:    data,
		RetCode: strconv.Itoa(err.ErrorCode),
		Ret:     strconv.Itoa(err.ErrorCode),
	}
	respDate := protocol.SubsysRspBody{Head: &rspHead, Rsp: &dataInfo}
	ctx.JSON(http.StatusOK, respDate)

	// 计算耗时，推送监控统计
	reqTime, _ := strconv.ParseInt(reqTimestamp, 10, 64)
	tconsum := float32(toolkit.GetNanoTimeStamp()-reqTime) / float32(time.Millisecond)

	// 打印响应报文
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	respData, _ := json.Marshal(respDate)
	logger.PrintInfo("%soutPacket %s[%.3fms]%s %s", logger.Red, logger.DarkGreen, tconsum, logger.Reset, respData)

	tracing := opentracing.GetOpenTracing()
	if err.ErrorCode != 0 {
		tracing.SetTag("error", strconv.Itoa(err.ErrorCode))
		tracing.SetTag("error.kind", err.ErrorInfo)
	}
	tracing.Dump()
}

// 接口响应数据结构封装(二层协议V1版本)
func JsonResponseV1(ctx *gin.Context, err errcode.AppError, rspHead protocol.SubsysHeader, data interface{}, errInfo ...string) {
	errStr := ""
	if len(errInfo) > 0 {
		errStr = " " + fmt.Sprint(errInfo)
	}

	// 设置Head参数
	reqTimestamp := rspHead.Timestamp
	rspHead.MsgType = "response"
	rspHead.Timestamp = strconv.FormatInt(toolkit.GetTimeStamp(), 10)

	// 组装响应报文和发送
	dataInfo := protocol.SubsysCommonRspV1{
		RetMsg:  err.ErrorInfo + errStr,
		Data:    data,
		RetCode: strconv.Itoa(err.ErrorCode),
		Ret:     strconv.Itoa(err.ErrorCode),
	}
	respDate := protocol.SubsysRspBodyV1{Head: &rspHead, Rsp: &dataInfo}
	ctx.JSON(http.StatusOK, respDate)

	// 计算耗时，推送监控统计
	reqTime, _ := strconv.ParseInt(reqTimestamp, 10, 64)
	tconsum := float32(toolkit.GetNanoTimeStamp()-reqTime) / float32(time.Millisecond)

	// 打印响应报文
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	respData, _ := json.Marshal(respDate)
	logger.PrintInfo("%soutPacket %s[%.3fms]%s %s", logger.Red, logger.DarkGreen, tconsum, logger.Reset, respData)

	tracing := opentracing.GetOpenTracing()
	if err.ErrorCode != 0 {
		tracing.SetTag("error", strconv.Itoa(err.ErrorCode))
		tracing.SetTag("error.kind", err.ErrorInfo)
	}
	tracing.Dump()
}

// 接口响应数据结构封装
func JsonResponseOther(ctx *gin.Context, httpCode int, err errcode.AppError, data interface{}) {
	ctx.JSON(httpCode, gin.H{
		"code": err.ErrorCode,
		"msg":  err.ErrorInfo,
		"data": data,
	})

	opentracing.GetOpenTracing().Dump()
}

// 接口响应数据结构封装
func ProtoBufResponse(ctx *gin.Context, httpCode int, data interface{}) {
	ctx.ProtoBuf(httpCode, data)
}
