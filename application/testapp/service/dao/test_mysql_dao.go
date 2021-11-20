package dao

import (
	"errors"
	"go-frame/implements/storage"
	"go-frame/logger"
	"fmt"
)

const TTProduct = "t_product"

type TProduct struct {
	ProductId   int    `gorm:"primary_key;column:Fproduct_id"` // 产品ID
	ProductName string `gorm:"column:Fproduct_name"`
	ClassId     int    `gorm:"column:Fclass_id"`
	BrandId     int    `gorm:"column:Fbrand_id"`
	PicId       string `gorm:"column:Fpic_id"`
}

type ProductSearch struct {
	Id        int
	Status    int
	Keyword   string
	PageIndex int
	PageSize  int
}

func GetInfoFromMysql(search *ProductSearch) ([]TProduct, int64, error) {
	mysqlDB := storage.GetDBHandle(storage.PriceMysql)
	if mysqlDB == nil {
		logger.PrintError("get mysql handle failed")
		return nil, 0, errors.New("get mysql handle failed")
	}

	condition := make(map[string]interface{})
	if search.Id > 0 {
		condition["Fproduct_id"] = search.Id
	}

	var retData []TProduct
	db := mysqlDB.Table(TTProduct).Select("Fproduct_id, Fproduct_name, Fclass_id, Fbrand_id, Fpic_id")

	if search.Id > 0 {
		db = db.Where("Fid = ?", search.Id)
	}

	if search.Keyword != "" {
		db = db.Where("Fproduct_name LIKE ?", fmt.Sprintf("%%%s%%", search.Keyword))
	}

	var count int64
	err := storage.GetDBError(db.Count(&count))
	if err != nil {
		return nil, 0, err
	} else if count == 0 {
		return retData, 0, nil
	}

	if search.PageSize > 0 {
		db = db.Offset(search.PageSize*search.PageIndex).Limit(search.PageSize)
	}

	db = db.Scan(&retData)

	return retData, count, storage.GetDBError(db)
}

func SetInfoFromMysql(info *TProduct) error {
	mysqlDB := storage.GetDBHandle(storage.PriceMysql)
	if mysqlDB == nil {
		logger.PrintError("get mysql handle failed")
		return errors.New("get mysql handle failed")
	}

	return storage.GetDBError(mysqlDB.Table(TTProduct).Updates(info))
}
