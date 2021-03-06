package model

import (
	"github.com/mutou1225/go-frame/example/testapp/appinterface"
	"github.com/mutou1225/go-frame/example/testapp/service/dao"
	"github.com/mutou1225/go-frame/frame/appengine/app"
	"strconv"
)

func TestEsModel(params *appinterface.TestEs) ([]appinterface.TestInfo, int64, error) {
	pageIndex := app.FormatPageIndex(&params.PageIndex)
	pageSize := app.FormatPageSize(&params.PageSize)

	esResq, err := dao.GetInfoFromES(pageIndex, pageSize)
	if err != nil {
		return nil, 0, err
	}

	retList := []appinterface.TestInfo{}
	for _, info := range esResq.Hits.Hits {
		retList = append(retList, appinterface.TestInfo{
			ClassId:     strconv.Itoa(info.Source.ClassId),
			ClassName:   info.Source.ClassName,
			ProductId:   strconv.Itoa(info.Source.ProductId),
			ProductName: info.Source.ProductName,
			BrandId:     strconv.Itoa(info.Source.BrandId),
			BrandName:   info.Source.BrandName,
			PicId:       info.Source.PicId,
		})
	}

	return retList, int64(esResq.Hits.Total), nil
}
