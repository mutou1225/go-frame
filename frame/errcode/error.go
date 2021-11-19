package errcode

import (
	"bytes"
	"fmt"
	"runtime"
)

const (
	RetCodeSuccess           = 0
	RetCodePriceAdjConf      = 70017000
	RetCodeEvaluateConf      = 70018000
	RetCodeBasePriceEva      = 70019000
	RetCodeEvaluateProduct   = 70020000
	RetCodeEvaluatePrice     = 70021000
	RetCodeCheckTemplateConf = 70022000
	RetCodeEvaluateCheckV3   = 70023000
	RetCodeBaseProductV3     = 70024000
)

type AppError struct {
	ErrorCode int
	ErrorInfo string
}

func (a AppError) Error() string {
	return a.ErrorInfo
}

var (
	SUCCESS                        = AppError{ErrorCode: 0, ErrorInfo: "SUCCESS"}
	INVALID_PARAMS                 = AppError{ErrorCode: 400, ErrorInfo: "请求参数错误"}
	ID_NOT_EMPTY                   = AppError{ErrorCode: 4001, ErrorInfo: "ID为空"}
	ERROR_TOKEN_EMPTY              = AppError{ErrorCode: 4002, ErrorInfo: "token为空"}
	ERROR_TOKEN_INVALID            = AppError{ErrorCode: 4003, ErrorInfo: "token无效"}
	ERROR_TOKEN_EXPIRE             = AppError{ErrorCode: 4004, ErrorInfo: "token过期"}
	ERROR_USER_NOT_EXIST           = AppError{ErrorCode: 4005, ErrorInfo: "用户不存在"}
	ERROR_SERVER_ERROR             = AppError{ErrorCode: 500, ErrorInfo: "服务内部错误"}
	ERROR_DATA_NOT_EXIST           = AppError{ErrorCode: 5001, ErrorInfo: "记录不存在"}
	ERROR_CONFIG_PARSE             = AppError{ErrorCode: 5002, ErrorInfo: "解析配置出错"}
	ERROR_SIGN_FIELD_NO_EXIST      = AppError{ErrorCode: 5003, ErrorInfo: "没有签名字段"}
	ERROR_SIGN                     = AppError{ErrorCode: 5004, ErrorInfo: "签名错误"}
	ERRO_SERVICE_ID_FIELD_NO_EXIST = AppError{ErrorCode: 5005, ErrorInfo: "没有服务ID"}
	ERROR_DENY_SERVICE_ID          = AppError{ErrorCode: 5006, ErrorInfo: "服务未授权"}
	ERROR_LOST_SIGN_DATA           = AppError{ErrorCode: 5007, ErrorInfo: "没有签名数据"}
	RetCode_ERR_CACHE_INIT         = AppError{ErrorCode: 5008, ErrorInfo: "redis初始化失败"}
	ERROR_LIMINT                   = AppError{ErrorCode: 5009, ErrorInfo: "请求过快"}
)

// 自定义失败：错误码不变，在原错误信息的基础上，增加自定义错误信息
func CustomError(appError AppError, errInfo string) AppError {
	return AppError{
		ErrorCode: appError.ErrorCode,
		ErrorInfo: appError.ErrorInfo + ": " + errInfo,
	}
}

func NewError(errCode int, errInfo string) AppError {
	return AppError{
		ErrorCode: errCode,
		ErrorInfo: errInfo,
	}
}

func NewErrorByError(appError AppError) *AppError {
	return &AppError{
		ErrorCode: appError.ErrorCode,
		ErrorInfo: appError.ErrorInfo,
	}
}

func GetSystemPanic(err interface{}) string {
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "%v\n", err)
	for i := 1; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
	}
	return buf.String()
}

/*
func GetMsg(code int) string {
	msg, ok := MsgFlags[code]
	if ok {
		return msg
	}
	return MsgFlags[ERROR]
}
*/
