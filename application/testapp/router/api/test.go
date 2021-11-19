package api

import (
	"eva_services_go/application/testapp/apperrors"
	"eva_services_go/application/testapp/appinterface"
	"eva_services_go/application/testapp/service/model"
	"eva_services_go/frame/appengine/app"
	"eva_services_go/frame/protocol"
	"eva_services_go/logger"
	"github.com/gin-gonic/gin"
	"strconv"
)

func TestEsApi(c *gin.Context) {
	var form protocol.SubsysReqBody
	var formParam appinterface.TestEs
	err := app.BindAndValid(c, &form, &formParam)
	if err != nil {
		app.JsonResponse(c, apperrors.INVALID_PARAMS, form.Head, nil, err.Error())
		return
	}

	logger.PrintInfo("formParam: %+v", formParam)

	result, total, err := model.TestEsModel(&formParam)
	if err != nil {
		app.JsonResponse(c, apperrors.GET_TEST_LIST_ERROR, form.Head, nil, err.Error())
		return
	}

	reqData := appinterface.RespTest{}
	reqData.TestList = result
	reqData.Total.PageIndex = formParam.PageIndex
	reqData.Total.PageSize  = formParam.PageSize
	reqData.Total.Total = strconv.FormatInt(total, 10)

	app.JsonResponse(c, apperrors.SUCCESS, form.Head, nil)
}

func TestRedisGetApi(c *gin.Context) {
	var form protocol.SubsysReqBody
	var formParam appinterface.TestSet
	err := app.BindAndValid(c, &form, &formParam)
	if err != nil {
		app.JsonResponse(c, apperrors.INVALID_PARAMS, form.Head, nil, err.Error())
		return
	}

	logger.PrintInfo("formParam: %+v", formParam)

	result, err := model.TestRedisGetModel(&formParam)
	if err != nil {
		app.JsonResponse(c, apperrors.GET_TEST_LIST_ERROR, form.Head, nil, err.Error())
		return
	}

	app.JsonResponse(c, apperrors.SUCCESS, form.Head, result)
}

func TestRedisSetApi(c *gin.Context) {
	var form protocol.SubsysReqBody
	var formParam appinterface.TestSet
	err := app.BindAndValid(c, &form, &formParam)
	if err != nil {
		app.JsonResponse(c, apperrors.INVALID_PARAMS, form.Head, nil, err.Error())
		return
	}

	logger.PrintInfo("formParam: %+v", formParam)

	if err := model.TestRedisSetModel(&formParam); err != nil {
		app.JsonResponse(c, apperrors.GET_TEST_LIST_ERROR, form.Head, nil, err.Error())
		return
	}

	app.JsonResponse(c, apperrors.SUCCESS, form.Head, nil)
}

func TestMysqlGetApi(c *gin.Context) {
	var form protocol.SubsysReqBody
	var formParam appinterface.Test
	err := app.BindAndValid(c, &form, &formParam)
	if err != nil {
		app.JsonResponse(c, apperrors.INVALID_PARAMS, form.Head, nil, err.Error())
		return
	}

	logger.PrintInfo("formParam: %+v", formParam)

	result, total, err := model.TestMysqlGetModel(&formParam)
	if err != nil {
		app.JsonResponse(c, apperrors.GET_TEST_LIST_ERROR, form.Head, nil, err.Error())
		return
	}

	retData := appinterface.RespTest{}
	retData.Total.Total = strconv.Itoa(total)
	retData.Total.PageIndex = formParam.PageIndex
	retData.Total.PageSize = formParam.PageSize
	retData.TestList = result

	app.JsonResponse(c, apperrors.SUCCESS, form.Head, retData)
}

func TestMysqlSetApi(c *gin.Context) {
	var form protocol.SubsysReqBody
	var formParam appinterface.TestInfo
	err := app.BindAndValid(c, &form, &formParam)
	if err != nil {
		app.JsonResponse(c, apperrors.INVALID_PARAMS, form.Head, nil, err.Error())
		return
	}

	logger.PrintInfo("formParam: %+v", formParam)

	if err := model.TestMysqlSetModel(&formParam); err != nil {
		app.JsonResponse(c, apperrors.GET_TEST_LIST_ERROR, form.Head, nil, err.Error())
		return
	}

	app.JsonResponse(c, apperrors.SUCCESS, form.Head, nil)
}



