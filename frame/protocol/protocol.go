package protocol

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/mutou1225/go-frame/config"
	"github.com/mutou1225/go-frame/frame/errcode"
	apphttp "github.com/mutou1225/go-frame/implements/http"
	"github.com/mutou1225/go-frame/implements/opentracing"
	"github.com/mutou1225/go-frame/implements/toolkit"
	"github.com/mutou1225/go-frame/logger"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

type ProtocolType int

const (
	ProtocolV1        ProtocolType = 1  // _body _ret _retcode _retinfo
	ProtocolV15       ProtocolType = 15 // _data _ret _retcode _retinfo
	ProtocolV2        ProtocolType = 2  // _data _ret _errCode _errStr
	BasePriceAgentUrl              = "http://baseprice.huishoubao.com"
)

// 参数体 ProtocolV2
type SubsysCommonRsp struct {
	RetMsg  string      `json:"_errStr"`
	Data    interface{} `json:"_data"` //上层应用定义
	RetCode string      `json:"_errCode"`
	Ret     string      `json:"_ret"`
}

// ProtocolV1
type SubsysCommonRspV1 struct {
	Data    interface{} `json:"_data"` //上层应用定义
	RetCode string      `json:"_retcode"`
	Ret     string      `json:"_ret"`
	RetMsg  string      `json:"_retinfo"`
}

// ProtocolV15
type SubsysCommonRspV15 struct {
	Data    interface{} `json:"_data"` //上层应用定义
	RetCode string      `json:"_retcode"`
	Ret     string      `json:"_ret"`
	RetMsg  string      `json:"_retinfo"`
}

// 头部
type SubsysHeader struct {
	CallServiceId string `json:"_callerServiceId" validate:"required"`
	GroupNo       string `json:"_groupNo" validate:"required"`
	Interface     string `json:"_interface" validate:"required"`
	InvokeId      string `json:"_invokeId" validate:"required"`
	MsgType       string `json:"_msgType" validate:"required"`
	Remark        string `json:"_remark"`
	Timestamp     string `json:"_timestamps" validate:"required"`
	Version       string `json:"_version" validate:"required"`
}

//基本请求体定义
type SubsysReqBody struct {
	Head  SubsysHeader `json:"_head" validate:"required"`
	Param interface{}  `json:"_param" validate:"required"` //上层应用定义
}

//基本响应体定义
type SubsysRspBody struct {
	Head *SubsysHeader    `json:"_head"`
	Rsp  *SubsysCommonRsp `json:"_data"`
}

type SubsysRspBodyV1 struct {
	Head *SubsysHeader      `json:"_head"`
	Rsp  *SubsysCommonRspV1 `json:"_body"`
}

type SubsysRspBodyV15 struct {
	Head *SubsysHeader       `json:"_head"`
	Rsp  *SubsysCommonRspV15 `json:"_data"`
}

type SubsysTotal struct {
	PageIndex string `json:"pageIndex"`
	PageSize  string `json:"pageSize"`
	Total     string `json:"total"`
}

// 空返回
func SubsysGetBadHeader() SubsysHeader {
	return SubsysHeader{
		CallServiceId: "unknown",
		GroupNo:       "-1",
		Interface:     "unknown",
		InvokeId:      "unknown",
		MsgType:       "response",
		Remark:        "unknown",
		Timestamp:     toolkit.ConvertToString(toolkit.GetTimeStamp()),
		Version:       "unknown",
	}
}

// 请求价格接入层
func RequestAccessModel(url, strInterface string, params interface{}, response interface{}) error {
	errCode := 1
	startTime := toolkit.GetNanoTimeStamp()
	defer func(errCode *int) {
		logger.PrintReportByTime("BasePriceAgent", url, strInterface, *errCode, startTime)
	}(&errCode)

	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	jsonStr, err := json.Marshal(params)
	if err != nil {
		logger.PrintError("json.Marshal() Err: %s", err.Error())
		return err
	}

	logger.PrintInfo("curl -d'%s' %s", string(jsonStr), url)

	// opentracing
	ot := opentracing.GetOpenTracing()
	spanId, _ := ot.StartChildSpan("BasePriceAgent")
	ot.SetChildTag(spanId, url, strInterface)
	spanContext, _ := ot.GetChildSpanContext(spanId)
	defer ot.EndChildSpan(spanId)

	// 3 分钟超时
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Minute)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonStr))
	if err != nil {
		logger.PrintError("http.NewRequestWithContext() Err: %s", err.Error())
		return err
	}

	req.Header["OPENTRACER-INFO"] = []string{spanContext}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("Connection", "keep-alive")

	logger.PrintInfo("Header: %v", req.Header)

	client := apphttp.CreateHTTPClient()
	if client == nil {
		logger.PrintError("http.Client <nil>")
		return errors.New("http.Client <nil>")
	}
	rsp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		io.Copy(ioutil.Discard, rsp.Body)
		logger.PrintError("RequestCgiModel[%s] Response Status Code: %d", url, rsp.StatusCode)
		return errors.New(fmt.Sprintf("Response Status Code: %d", rsp.StatusCode))
	}

	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		logger.PrintError("ioutil.ReadAll() Err: %s", err.Error())
		return err
	}

	logger.PrintInfo("response: %s", string(body))
	err = json.Unmarshal(body, response)
	if err != nil {
		logger.PrintError("json.Unmarshal() Err: %s", err.Error())
		return err
	}

	// 设置成功错误码
	errCode = 0

	return nil
}

// 请求外部服务（二层应用）
// calleeName 即为上报的服务名
// 通过 config.GetServerIdStr() 得到 callerServiceId 值
// 通过 config.GetCalleeByServerName(calleeName) 得到 signKey 值
// 统一规则：调用信息只放在xml配置文件中，防止写死在代码中，以及配置到其他地方，如数据库等。
type RequestCgiHandle struct {
	PolType    ProtocolType  //协议类型
	CalleeName string        //被调方名称
	Url        string        //被调方Url
	Interface  string        //被调方Interface
	MsgBody    interface{}   //请求参数 (Param部分)
	Timeout    time.Duration //请求超时时间
}

// 返回空的Handle
func NewCgiHandleDef(pType ProtocolType, timeOut time.Duration) *RequestCgiHandle {
	return &RequestCgiHandle{
		PolType: pType,
		Timeout: timeOut,
	}
}

// 返回Handle
func NewCgiHandle(pType ProtocolType, calleeName, strUrl, strInterface string,
	msgBody interface{}, timeOut time.Duration) *RequestCgiHandle {
	return &RequestCgiHandle{
		pType,
		calleeName,
		strUrl,
		strInterface,
		msgBody,
		timeOut,
	}
}

// 设置被调方的信息
func (h *RequestCgiHandle) SetCalleeInfo(calleeName, strUrl, strInterface string) {
	h.CalleeName = calleeName
	h.Url = strUrl
	h.Interface = strInterface
}

// 设置请求参数 (Param部分)
func (h *RequestCgiHandle) SetRequestBody(msgBody interface{}) {
	h.MsgBody = msgBody
}

// 请求服务
// response 为返回结果，指针类型
func (h *RequestCgiHandle) RequestCgiModel(response interface{}) error {
	retCode := errcode.NewErrorByError(errcode.ERROR_SERVER_ERROR)
	startTime := toolkit.GetNanoTimeStamp()
	defer func() {
		logger.PrintReportByTime(h.CalleeName, h.Url, h.Interface, retCode.ErrorCode, startTime)
	}()

	callerServiceId := config.GetServerIdStr()
	callee, ok := config.GetCalleeByServerName(h.CalleeName)
	if !ok {
		retCode.ErrorInfo = fmt.Sprintf("GetCalleeByServerName() 未找到配置信息: %s", h.CalleeName)
		logger.PrintError(retCode.ErrorInfo)
		return retCode
	}

	timestamp := fmt.Sprintf("%d", toolkit.GetTimeStamp())
	_, sessionId, _ := toolkit.GetUniqId(h.Interface)
	head := SubsysHeader{
		CallServiceId: callerServiceId,
		GroupNo:       "1",
		Interface:     h.Interface,
		InvokeId:      sessionId,
		MsgType:       "request",
		Remark:        "",
		Timestamp:     toolkit.ConvertToString(timestamp),
		Version:       "0.01",
	}
	request := SubsysReqBody{
		Head:  head,
		Param: h.MsgBody,
	}

	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	jsonStr, err := json.Marshal(request)
	if err != nil {
		logger.PrintError("json.Marshal() Err: %s", err.Error())
		retCode.ErrorInfo = err.Error()
		return retCode
	}

	signStr := toolkit.ApiSign(string(jsonStr), callee.ServerKey)

	// opentracing
	ot := opentracing.GetOpenTracing()
	spanId, _ := ot.StartChildSpan(h.CalleeName)
	ot.SetChildTag(spanId, h.Url, h.Interface)
	spanContext, _ := ot.GetChildSpanContext(spanId)
	defer func() {
		if retCode.ErrorCode != errcode.RetCodeSuccess {
			ot.SetChildTag(spanId, "error", strconv.Itoa(retCode.ErrorCode))
			ot.SetChildTag(spanId, "error.kind", retCode.ErrorInfo)
		}
		ot.EndChildSpan(spanId)
	}()

	// 设置超时
	ctx, _ := context.WithTimeout(context.Background(), h.Timeout)
	req, err := http.NewRequestWithContext(ctx, "POST", h.Url, bytes.NewReader(jsonStr))
	if err != nil {
		logger.PrintError("http.NewRequestWithContext() Err: %s", err.Error())
		retCode.ErrorInfo = err.Error()
		return retCode
	}

	req.Header["OPENTRACER-INFO"] = []string{spanContext}
	req.Header["HSB-OPENAPI-CALLERSERVICEID"] = []string{callerServiceId}
	req.Header["HSB-OPENAPI-SIGNATURE"] = []string{signStr}
	req.Header.Set("content-type", "application/json")

	logger.PrintInfo("curl -H'HSB-OPENAPI-CALLERSERVICEID:%s' -H'HSB-OPENAPI-SIGNATURE:%s' -H'OPENTRACER-INFO:%s' -d'%s' %s",
		callerServiceId, signStr, spanContext, string(jsonStr), h.Url)

	client := apphttp.CreateHTTPClient()
	if client == nil {
		retCode.ErrorInfo = "http.Client <nil>"
		return retCode
	}
	rsp, err := client.Do(req)
	if err != nil {
		logger.PrintError("client.Do() Err: %s", err.Error())
		retCode.ErrorInfo = err.Error()
		return retCode
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		io.Copy(ioutil.Discard, rsp.Body)
		logger.PrintError("RequestCgiModel[%s] Response Status Code: %d", h.Url, rsp.StatusCode)
		retCode.ErrorCode = rsp.StatusCode
		retCode.ErrorInfo = fmt.Sprintf("Response Status Code: %d", rsp.StatusCode)
		return retCode
	}

	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		logger.PrintError("ioutil.ReadAll() Err: %s", err.Error())
		retCode.ErrorInfo = err.Error()
		return retCode
	}

	logger.PrintInfo("Response: %s", string(body))

	if h.PolType == ProtocolV2 {
		retData := SubsysRspBody{
			Rsp: &SubsysCommonRsp{
				Data: response,
			},
		}

		err = json.Unmarshal(body, &retData)
		if err != nil {
			retCode.ErrorInfo = err.Error()
			return retCode
		} else if retData.Rsp.Ret != "0" {
			logger.PrintError("retStr: %s", retData.Rsp.RetMsg)
			retCode.ErrorCode = toolkit.StrAtoi(retData.Rsp.Ret)
			retCode.ErrorInfo = retData.Rsp.RetMsg
			return retCode
		}
	} else if h.PolType == ProtocolV1 {
		retData := SubsysRspBodyV1{
			Rsp: &SubsysCommonRspV1{
				Data: response,
			},
		}

		err = json.Unmarshal(body, &retData)
		if err != nil {
			retCode.ErrorInfo = err.Error()
			return retCode
		} else if retData.Rsp.Ret != "0" {
			logger.PrintError("retStr: %s", retData.Rsp.RetMsg)
			retCode.ErrorCode = toolkit.StrAtoi(retData.Rsp.Ret)
			retCode.ErrorInfo = retData.Rsp.RetMsg
			return retCode
		}
	} else {
		retData := SubsysRspBodyV15{
			Rsp: &SubsysCommonRspV15{
				Data: response,
			},
		}

		err = json.Unmarshal(body, &retData)
		if err != nil {
			retCode.ErrorInfo = err.Error()
			return retCode
		} else if retData.Rsp.Ret != "0" {
			logger.PrintError("retStr: %s", retData.Rsp.RetMsg)
			retCode.ErrorCode = toolkit.StrAtoi(retData.Rsp.Ret)
			retCode.ErrorInfo = retData.Rsp.RetMsg
			return retCode
		}
	}

	// 设置成功错误码
	retCode.ErrorCode = 0
	retCode.ErrorInfo = "SUCCESS"

	return nil
}
