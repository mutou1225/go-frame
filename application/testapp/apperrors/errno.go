package apperrors

import (
	"go-frame/frame/errcode"
)

const (
	RetSUCCESS  = 0
	RetCodeBase = 80017000
)

var (
	SUCCESS 				= errcode.AppError{ErrorCode: 0, ErrorInfo: "SUCCESS"}

	// 公共错误码
	INVALID_PARAMS 			= errcode.AppError{ErrorCode: RetCodeBase + 100, ErrorInfo:"请求参数错误"}
	DB_OPERATION_FAILED 	= errcode.AppError{ErrorCode: RetCodeBase + 101, ErrorInfo:"数据库操作失败"}
	GET_INCREMENT_ID_FAIL 	= errcode.AppError{ErrorCode: RetCodeBase + 102, ErrorInfo:"获取自增id失败"}
	NAME_EXISTS				= errcode.AppError{ErrorCode: RetCodeBase + 103, ErrorInfo:"名称已存在"}

	// 业务
	GET_TEST_LIST_ERROR     = errcode.AppError{ErrorCode: RetCodeBase + 300, ErrorInfo:"获取信息失败"}

)