package appstorage

import (
	"errors"
	"github.com/mutou1225/go-frame/config"
	es "github.com/mutou1225/go-frame/implements/elasticsearch"
	"github.com/mutou1225/go-frame/implements/rabbitmq"
	"github.com/mutou1225/go-frame/implements/storage"
	"github.com/mutou1225/go-frame/logger"
)

var (
	EvaluateToolsRMQ *rabbitmq.RabbitMQ
)

func InitStorage() {
	// 初始化Mysql
	if err := storage.NewMysqlDB(storage.PriceMysql, config.GetDBHost(),
		config.GetDBUser(), config.GetDBPassword(),
		config.GetDBName(), config.GetDBPort(),
		config.GetMysqlPoolMin(), config.GetMysqlPoolMax(),
		config.GetMysqlIdleTime(), config.GetMysqlConnTime(), true); err != nil {
		logger.PrintError("InitDB() error:%s", err.Error())
	}

	// 连接Mgo
	if err := storage.InitMongo(config.GetServerName(), config.GetMongoDBAddStr(),
		config.GetMongoDBName(), config.GetMongoDBReplica(),
		config.GetMongoDBUsert(), config.GetMongoDBPasswd(),
		uint64(config.GetMgodbPoolMax()), uint64(config.GetMgodbPoolMin()),
		int64(config.GetMgodbIdleTime()), config.GetMgodbConnTime()); err != nil {
		logger.PrintError("InitMgoEx() error:%s", err.Error())
	}

	// es
	if err := es.InitEsClient(config.GetESHost()); err != nil {
		logger.PrintPanic("InitEsClient() error:%s", err.Error())
	}

	if err := storage.NewRedisCon(); err != nil {
		logger.PrintPanic("NewRedisCon() error:%s", err.Error())
	}

	if clientMQ, err := rabbitmq.NewRabbitMQ(
		config.GetRabbitMQUser(),
		config.GetRabbitMQPassword(),
		config.GetRabbitMQHost(),
		config.GetRabbitMQPort(),
		config.GetRabbitMQEvaVhost()); err != nil {
		logger.PrintError("rabbitmq.NewRabbitMQ() Err: %s", err.Error())
	} else {
		EvaluateToolsRMQ = clientMQ
	}
}

func GetRabbitMQClient() (*rabbitmq.RabbitMQ, error) {
	if EvaluateToolsRMQ == nil {
		return nil, errors.New("MQ Client <nil>")
	}
	return EvaluateToolsRMQ, nil
}

func ExitStorage() {
	storage.ExitDB()
	storage.CloseMgo()
	storage.CloseRedisCon()
}
