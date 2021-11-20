package model

import (
	"go-frame/application/testapp/appinterface"
	"go-frame/implements/storage"
	"go-frame/logger"
	"time"
)

func TestRedisSetModel(params *appinterface.TestSet) error {
	redisOpt, err := storage.GetRedisCon()
	if err != nil {
		logger.PrintError("GetRedisCon Err: %s", err.Error())
	}

	if err := redisOpt.Set("redis_set_test", params, time.Hour); err != nil {
		logger.PrintError("redis Set Err: %s", err.Error())
	}

	return nil
}

func TestRedisGetModel(params *appinterface.TestSet) (*appinterface.TestSet, error) {
	redisOpt, err := storage.GetRedisCon()
	if err != nil {
		logger.PrintError("GetRedisCon Err: %s", err.Error())
	}

	retData := appinterface.TestSet{}
	if err := redisOpt.Get("redis_set_test", &retData); err != nil {
		logger.PrintError("redis Set Err: %s", err.Error())
	}

	return &retData, nil
}
