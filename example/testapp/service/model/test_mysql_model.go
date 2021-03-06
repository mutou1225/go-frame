package model

import (
	"github.com/mutou1225/go-frame/example/testapp/appinterface"
	"github.com/mutou1225/go-frame/example/testapp/service/dao"
	"github.com/mutou1225/go-frame/frame/appengine/app"
	"github.com/mutou1225/go-frame/implements/toolkit"
	"strconv"
)

func TestMysqlGetModel(params *appinterface.Test) ([]appinterface.TestInfo, int64, error) {

	search := &dao.ProductSearch{
		Id        : toolkit.StrAtoi(params.Id),
		Status    : toolkit.StrAtoi(params.Status),
		Keyword   : params.Keyword,
		PageIndex : app.FormatPageIndex(&params.PageIndex),
		PageSize  : app.FormatPageIndex(&params.PageSize),
	}

	retData, total, err := dao.GetInfoFromMysql(search)
	if err != nil {
		return nil, 0, err
	}

	testList := []appinterface.TestInfo{}
	for _, info := range retData {
		testList = append(testList, appinterface.TestInfo{
			ClassId     : strconv.Itoa(info.ClassId),
			ClassName   : "",
			ProductId   : strconv.Itoa(info.ProductId),
			ProductName : info.ProductName,
			BrandId     : strconv.Itoa(info.BrandId),
			BrandName   : "",
			PicId       : info.PicId,
		})
	}

	return testList, total, nil
}

func TestMysqlSetModel(params *appinterface.TestInfo) error {
	info := &dao.TProduct{
		ProductId   : toolkit.StrAtoi(params.ProductId),
		ProductName : params.ProductName,
		ClassId     : toolkit.StrAtoi(params.ProductId),
		BrandId     : toolkit.StrAtoi(params.ProductId),
		PicId       : params.PicId,
	}

	err := dao.SetInfoFromMysql(info)
	if err != nil {
		return err
	}

	return nil
}
