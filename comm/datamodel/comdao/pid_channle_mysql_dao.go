package comdao

import (
	"errors"
	"go-frame/comm/datamodel/commodel"
	"go-frame/implements/storage"
	"go-frame/implements/toolkit"
	"go-frame/logger"
	"fmt"
)

// 查询渠道信息
func GetChannelByName(name string) ([]commodel.TChannel, error) {
	if name == "" {
		logger.PrintError("GetChannelByName() Params Error")
		return nil, errors.New("GetChannelByName Params Error")
	}

	mysqlDB := storage.GetDBHandle(storage.PriceMysql)
	if mysqlDB == nil {
		logger.PrintError("get mysql handle failed")
		return nil, errors.New("get mysql handle failed")
	}

	db := mysqlDB.Table(commodel.TTChannel).Select("Fchannel_id, Fchannel_name")
	db = db.Where("Fchannel_name LIKE ?", fmt.Sprintf("%%%s%%", name))

	if toolkit.StrIsAllNum(name) {
		db = db.Or("Fchannel_id = ?", toolkit.StrAtoi(name))
	}

	var results []commodel.TChannel
	db = db.Scan(&results)

	return results, storage.GetDBError(db)
}

func GetChannelById(id ... int) ([]commodel.TChannel, error) {
	if len(id) == 0 {
		logger.PrintInfo("GetChannelById() Params Empty")
		return []commodel.TChannel{}, nil
	}

	mysqlDB := storage.GetDBHandle(storage.PriceMysql)
	if mysqlDB == nil {
		logger.PrintError("get mysql handle failed")
		return nil, errors.New("get mysql handle failed")
	}

	var results []commodel.TChannel
	db := mysqlDB.Table(commodel.TTChannel).Select("Fchannel_id, Fchannel_name").Where("Fchannel_id in (?)", id).Scan(&results)

	return results, storage.GetDBError(db)
}

// 查询pid信息
func GetPidByName(name string) ([]commodel.TPid, error) {
	if name == "" {
		logger.PrintError("GetChannelByName() Params Error")
		return nil, errors.New("GetChannelByName Params Error")
	}

	mysqlDB := storage.GetDBHandle(storage.PriceMysql)
	if mysqlDB == nil {
		logger.PrintError("get mysql handle failed")
		return nil, errors.New("get mysql handle failed")
	}

	db := mysqlDB.Table(commodel.TTag).Select("t_tag.Ftag_id, t_tag.Ftag_name, t_maptag.Fp_id, t_tag.Fchannel_id, t_channel.Fchannel_name")
	db = db.Joins("left join t_maptag on t_maptag.Ftag_id = t_tag.Ftag_id")
	db = db.Joins("left join t_channel on t_channel.Fchannel_id = t_tag.Fchannel_id")
	db = db.Where("Ftag_name LIKE ?", fmt.Sprintf("%%%s%%", name))

	if toolkit.StrIsAllNum(name) {
		db = db.Or("t_maptag.Fp_id = ?", toolkit.StrAtoi(name))
	}

	var results []commodel.TPid
	db = db.Scan(&results)

	return results, storage.GetDBError(db)
}

func GetPidById(id ... int) ([]commodel.TPid, error) {
	if len(id) == 0 {
		logger.PrintInfo("GetPidById() Params Empty")
		return []commodel.TPid{}, nil
	}

	mysqlDB := storage.GetDBHandle(storage.PriceMysql)
	if mysqlDB == nil {
		logger.PrintError("get mysql handle failed")
		return nil, errors.New("get mysql handle failed")
	}

	var results []commodel.TPid
	db := mysqlDB.Table(commodel.TTag).Select("t_tag.Ftag_id, t_tag.Ftag_name, t_maptag.Fp_id, t_tag.Fchannel_id, t_channel.Fchannel_name")
	db = db.Joins("left join t_maptag on t_maptag.Ftag_id = t_tag.Ftag_id")
	db = db.Joins("left join t_channel on t_channel.Fchannel_id = t_tag.Fchannel_id")
	db = db.Where("t_maptag.Fp_id in (?)", id).Scan(&results)

	return results, storage.GetDBError(db)
}
