package app

import (
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/mutou1225/go-frame/frame/protocol"
	"github.com/mutou1225/go-frame/implements/toolkit"
	"github.com/mutou1225/go-frame/logger"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	translations "github.com/go-playground/validator/v10/translations/zh"
)

const (
	EXPORT_IMPORT_FILE_PATH = "/cdn/root/base_price/download/"
	DOWNLOAD_URL_PATH       = "http://baseprice.huishoubao.com.cn/download/"
)

//
func BindAndValid(c *gin.Context, form *protocol.SubsysReqBody, formParam interface{}) error {
	form.Param = formParam
	if err := c.ShouldBindJSON(form); err != nil {
		return err
	}

	// 设置时间戳，用于计算监控的耗时
	form.Head.Timestamp = strconv.FormatInt(toolkit.GetNanoTimeStamp(), 10)

	validate := validator.New()
	err := validate.Struct(form)
	if err != nil {
		//验证器注册翻译器
		uni := ut.New(zh.New())
		trans, _ := uni.GetTranslator("zh")
		translations.RegisterDefaultTranslations(validate, trans)

		for _, verr := range err.(validator.ValidationErrors) {
			return errors.New(verr.Translate(trans))
		}
	}

	return nil
}

func ValidatorStruct(s interface{}) error {
	validate := validator.New()
	err := validate.Struct(s)
	if err != nil {
		//验证器注册翻译器
		uni := ut.New(zh.New())
		trans, _ := uni.GetTranslator("zh")
		translations.RegisterDefaultTranslations(validate, trans)

		for _, verr := range err.(validator.ValidationErrors) {
			return errors.New(verr.Translate(trans))
		}
	}

	return nil
}

func FormatPageIndex(pageIndex *string) int {
	if *pageIndex == "" {
		logger.PrintInfo("pageIndex empty")
		*pageIndex = "0"
		return 0
	}

	i, e := strconv.Atoi(*pageIndex)
	if e != nil {
		logger.PrintInfo("FormatPageIndex() Error! index: %s, err: %s", pageIndex, e.Error())
		*pageIndex = "0"
		return 0
	}
	return i
}

func FormatPageSize(pageSize *string) int {
	if *pageSize == "" {
		logger.PrintInfo("pageSize empty")
		*pageSize = "10"
		return 10
	}

	i, e := strconv.Atoi(*pageSize)
	if e != nil {
		logger.PrintInfo("FormatPageIndex() Error! index: %s, err: %s", pageSize, e.Error())
		*pageSize = "10"
		return 10
	}

	if i > 500 {
		*pageSize = "500"
		return 500
	}

	return i
}
