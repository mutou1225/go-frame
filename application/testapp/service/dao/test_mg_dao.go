package dao

import (
	"context"
	"errors"
	"eva_services_go/implements/storage"
	"eva_services_go/implements/toolkit"
	"eva_services_go/logger"
	"go.mongodb.org/mongo-driver/bson"
)

type TestSearch struct {
	Id                  int
	Keyword             string
	PageIndex, PageSize int
}

type StPriceAdjPlan struct {
	Id   int    `bson:"Fid"`
	Name string `bson:"Fname"`
}

func GetTestList(params TestSearch) ([]StPriceAdjPlan, int64, error) {
	mgoColl := storage.MgoCollection{"base_price", "t_test"}
	collection := mgoColl.GetMgoCollection()
	if collection == nil {
		logger.PrintError("get mongo session failed")
		return nil, 0, errors.New("get mongo session failed")
	}

	query := bson.M{}
	if params.Id > 0 {
		query["Fid"] = params.Id
	}

	if params.Keyword != "" {
		orFilter := []bson.M{}
		orFilter = append(orFilter, bson.M{"Fname": bson.M{"$regex": params.Keyword, "$options": "$i"}})

		if toolkit.StrIsAllNum(params.Keyword) {
			orFilter = append(orFilter, bson.M{"Fid": toolkit.StrAtoi(params.Keyword)})
		}

		query["$or"] = orFilter
	}

	field := &bson.D{
		{"Fid", 1},
		{"Fname", 1},
	}
	logger.PrintInfo("query: %+v", query)
	logger.PrintInfo("field: %+v", field)

	queryHandle := collection.Find(context.Background(), &query)

	// 获取总数
	total, err := queryHandle.Count()
	if err != nil {
		logger.PrintError("find monogo db failed:%s", err.Error())
		return nil, 0, err
	}
	logger.PrintInfo("Mgo count: %d", total)

	if total == 0 {
		return []StPriceAdjPlan{}, 0, nil
	}

	// 查询信息
	queryHandle.Select(field).Sort("-Fid")
	if params.PageSize > 0 {
		queryHandle.Skip(int64(params.PageIndex * params.PageSize)).Limit(int64(params.PageSize))
	}

	var result []StPriceAdjPlan
	err = queryHandle.All(&result)
	if err != nil {
		logger.PrintError("find monogo db failed:%s", err.Error())
		return nil, 0, err
	}

	logger.PrintInfo("Mgo result: %+v", result)

	return result, total, nil
}
