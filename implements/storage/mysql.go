package storage

import (
	"errors"
	"fmt"
	"go-frame/implements/opentracing"
	"go-frame/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
	"strconv"
	"sync"
	"time"
)

type MysqlType int

const (
	_ MysqlType = iota
	TestMysql
	PriceMysql
	DataWarehouse
	HsbMysql
)

const (
	mysqlConnTimeout = 3
)

var (
	gDbHandle  = make(map[MysqlType]*gorm.DB)
	gDdCfgMap  = make(map[MysqlType]*dbCfgInfo)
	mysqlmutex sync.Mutex
)

type dbCfgInfo struct {
	connStr                    string
	maxOpen, maxIdle, idleTime int
	debug                      bool
}

// dbName: 数据库实例名称，用于获取数据库句柄
func NewMysqlDB(myType MysqlType, host, user, pwd, database string, port, maxOpen, maxIdle, idleTime, connTime int, debug bool) error {
	logger.PrintInfo("NewMysqlDB Start")

	mysqlmutex.Lock()
	defer mysqlmutex.Unlock()

	// 是否存在
	if conn, ok := gDbHandle[myType]; ok && conn != nil {
		return nil
	}

	connTimeout := connTime
	if connTimeout == 0 {
		connTimeout = mysqlConnTimeout
	}
	connStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=%ds",
		user, pwd, host, port, database, connTimeout)

	// 配置
	gDdCfgMap[myType] = &dbCfgInfo{
		connStr:  connStr,
		maxOpen:  maxOpen,
		maxIdle:  maxIdle,
		idleTime: idleTime,
		debug:    debug,
	}

	dbHandle, err := gDdCfgMap[myType].newConn()
	gDbHandle[myType] = dbHandle
	if err != nil {
		gDbHandle[myType] = nil
		return err
	}

	return nil
}

func (c *dbCfgInfo) newConn() (*gorm.DB, error) {
	logger.PrintInfo("Mysql connStr: %v", c.connStr)
	dbHandle, err := gorm.Open(mysql.Open(c.connStr), &gorm.Config{
		Logger: glogger.New(DbLogger{}, glogger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  glogger.Info,
			IgnoreRecordNotFoundError: false,
			Colorful:                  false,
		}),
	})
	if err != nil {
		logger.PrintError("gorm.Open() Err: %s", err.Error())
		return nil, err
	}

	sqlDB, err := dbHandle.DB()
	if err != nil {
		logger.PrintError("gorm.DB() Err: %s", err.Error())
		return nil, err
	}

	sqlDB.SetMaxOpenConns(c.maxOpen)
	sqlDB.SetMaxIdleConns(c.maxIdle)
	sqlDB.SetConnMaxIdleTime(time.Duration(c.idleTime) * time.Second)
	return dbHandle, nil
}

func reconnMysqlDB(mysqlType MysqlType) (*gorm.DB, error) {
	logger.PrintInfo("reconnMysqlDB Start")

	mysqlmutex.Lock()
	defer mysqlmutex.Unlock()

	if conn, ok := gDbHandle[mysqlType]; ok && conn != nil {
		return conn, nil
	}

	var dbCfg *dbCfgInfo
	dbCfg, ok := gDdCfgMap[mysqlType]
	if !ok {
		logger.PrintInfo("reconnection() Err: cfg empty")
		return nil, errors.New("Err: cfg empty")
	}

	dbHandle, err := dbCfg.newConn()
	gDbHandle[mysqlType] = dbHandle
	if err != nil {
		gDbHandle[mysqlType] = nil
		return nil, err
	}
	return dbHandle, nil
}

// 获取一个连接
func GetDBHandle(mysqlType MysqlType) *gorm.DB {
	dbHandle, ok := gDbHandle[mysqlType]
	if ok && dbHandle != nil {
		return dbHandle
	} else {
		conn, _ := reconnMysqlDB(mysqlType)
		return conn
	}
	return nil
}

// 关闭连接
func ExitDB() {
	for _, dbHandle := range gDbHandle {
		if dbHandle != nil {
			if sqlDB, err := dbHandle.DB(); err != nil {
				_ = sqlDB.Close()
			}
			dbHandle = nil
		}
	}
}

// 获取mysql执行的错误信息
func GetDBError(db *gorm.DB) error {
	return db.Error
}

type DbLogger struct{}

func (log DbLogger) Printf(s string, v ...interface{}) {
	vLen := len(v)
	if vLen == 4 {
		logger.PrintInfoCalldepth(5, "%s %s[%.3fms rows:%v]%s", v[3], logger.Green, v[1], v[2], logger.Reset)

		f, _ := strconv.ParseFloat(fmt.Sprintf("%s", v[1]), 32)
		ot := opentracing.GetOpenTracing()
		spanId, _ := ot.StartChildSpan("Mysql")
		ot.SetChildTag(spanId, "db.type", "mysql")
		ot.SetChildTag(spanId, "db.statement", fmt.Sprintf("%s", v[3]))
		ot.EndChildSpanByDuration(spanId, int64(f * 1000))
	} else if vLen == 5 {
		logger.PrintErrorCalldepth(5, "%s %s%s%s %s[%.3fms rows:%v]%s", v[4], logger.Red, v[1], logger.Reset, logger.Yellow, v[2], v[3], logger.Reset)

		f, _ := strconv.ParseFloat(fmt.Sprintf("%s", v[2]), 32)
		ot := opentracing.GetOpenTracing()
		spanId, _ := ot.StartChildSpan("Mysql")
		ot.SetChildTag(spanId, "db.type", "mysql")
		ot.SetChildTag(spanId, "db.statement", fmt.Sprintf("%s", v[4]))
		ot.EndChildSpanByDuration(spanId, int64(f * 1000))
	} else {
		logger.PrintInfoCalldepth(5, s, v...)
	}
}
