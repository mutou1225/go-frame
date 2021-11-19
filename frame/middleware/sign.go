package middleware

import (
	"bytes"
	"crypto/md5"
	"eva_services_go/config"
	"eva_services_go/frame/errcode"
	"eva_services_go/frame/protocol"
	"eva_services_go/logger"
	"fmt"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

func apiSign(req string, key string) string {
	signStr := req + "_" + key
	digestBytes := md5.Sum([]byte(signStr))
	md5Str := fmt.Sprintf("%x", digestBytes)
	return md5Str
}

// 签名校验
func CheckCallSign() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqMsg := protocol.SubsysReqBody{}
		body, _ := ioutil.ReadAll(c.Request.Body)

		var json = jsoniter.ConfigCompatibleWithStandardLibrary
		err := json.Unmarshal(body, &reqMsg)
		if err != nil {
			logger.PrintInfo("请求参数解析失败")
			jsonResponsev2(c, http.StatusBadRequest, errcode.INVALID_PARAMS, protocol.SubsysGetBadHeader())
			c.Abort()
			return
		}

		errCode := errcode.SUCCESS
		for i := 0; i < 1; i++ {
			if reqMsg.Head.MsgType != "request" {
				logger.PrintInfo("_msgType 参数错误")
				errCode = errcode.INVALID_PARAMS
				break
			}

			callerId := c.Request.Header.Get("HSB-OPENAPI-CALLERSERVICEID")
			if callerId == "" {
				logger.PrintInfo("http协议头部HTTP_HSB_OPENAPI_CALLERSERVICEID值为空或不存在!")
				errCode = errcode.ERRO_SERVICE_ID_FIELD_NO_EXIST
				break
			}

			sign := c.Request.Header.Get("HSB-OPENAPI-SIGNATURE")
			if sign == "" {
				logger.PrintInfo("http协议头部HTTP_HSB_OPENAPI_SIGNATURE为空或者不存在!")
				errCode = errcode.ERROR_SIGN_FIELD_NO_EXIST
				break
			}

			callerKey, ok := config.GetCallerKey(callerId)
			if !ok {
				logger.PrintInfo("非法的ServerId!")
				errCode = errcode.ERROR_DENY_SERVICE_ID
				break
			}

			reqData := string(body)
			localSign := apiSign(reqData, callerKey)
			if localSign != sign {
				logger.PrintInfo("签名检验失败 ")
				logger.PrintInfo("Request:%s", sign)
				logger.PrintInfo("Local:%s", localSign)
				errCode = errcode.ERROR_SIGN
				break
			}
		}

		if errCode.ErrorCode != 0 {
			reqMsg.Head.MsgType = "response"
			reqMsg.Head.Timestamp = strconv.FormatInt(time.Now().UTC().Unix(), 10)
			jsonResponsev2(c, http.StatusUnauthorized, errCode, reqMsg.Head)
			c.Abort()
		} else {
			c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
			c.Next()
		}
	}
}
