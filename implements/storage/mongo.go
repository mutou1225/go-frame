package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/mutou1225/go-frame/implements/opentracing"
	"github.com/mutou1225/go-frame/logger"
	"github.com/qiniu/qmgo"
	"github.com/qiniu/qmgo/options"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/event"
	opts "go.mongodb.org/mongo-driver/mongo/options"
	"sync"
	"time"
)

var (
	mgoClient *qmgo.Client
	mgoCfgMap *mgoCfgInfo
	mgomutex  sync.Mutex
)

const (
	mgoConnectTimeout = 3
)

type mgoCfgInfo struct {
	cfg     *qmgo.Config
	appName string
	replica string
}

// Init Mgo Pool
func InitMongo(appName, addStr, dbName, replicaSet, user, passwd string, maxPoolSize, minPoolSize uint64,
	idleTime int64, connTime int) error {
	logger.PrintInfo("InitMongo Start ...")
	defer logger.PrintInfo("... InitMongo End")

	mgomutex.Lock()
	defer mgomutex.Unlock()

	if mgoClient != nil {
		return nil
	}

	// 设置连接超时时间
	connTimeout := int64(connTime * 1000)
	if connTimeout == 0 {
		connTimeout = int64(mgoConnectTimeout * 1000)
	}

	// 配置信息
	mgoCfgMap = &mgoCfgInfo{
		cfg: &qmgo.Config{
			Uri:              "mongodb://" + addStr,
			Database:         dbName,
			MaxPoolSize:      &maxPoolSize,
			MinPoolSize:      &minPoolSize,
			ConnectTimeoutMS: &connTimeout,
			Auth: &qmgo.Credential{
				AuthSource:  "admin",
				Username:    user,
				Password:    passwd,
				PasswordSet: true,
			},
		},
		appName: appName,
		replica: replicaSet,
	}

	client, err := mgoCfgMap.connMongo()
	if err != nil {
		return err
	}
	mgoClient = client

	return nil
}

func (c *mgoCfgInfo) loggerMongo() *event.CommandMonitor {
	commandMap := make(map[int64]bson.Raw)
	timerMap := make(map[int64]int64)
	mapMutex := sync.Mutex{}
	return &event.CommandMonitor{
		Started: func(_ context.Context, evt *event.CommandStartedEvent) {
			logger.PrintInfoCalldepth(9, "%sMongo Command:%s %v", logger.Blue, logger.Reset, evt.Command)

			mapMutex.Lock()
			commandMap[evt.RequestID] = evt.Command
			timerMap[evt.RequestID] = time.Now().UTC().UnixNano()
			mapMutex.Unlock()
		},
		Succeeded: func(_ context.Context, evt *event.CommandSucceededEvent) {
			logger.PrintInfoCalldepth(9, "%sMongo Reply:%s %v", logger.Blue, logger.Reset, evt.Reply)

			mapMutex.Lock()
			command := commandMap[evt.RequestID]
			duration := (time.Now().UTC().UnixNano() - timerMap[evt.RequestID]) / int64(time.Microsecond)
			delete(commandMap, evt.RequestID)
			delete(timerMap, evt.RequestID)
			mapMutex.Unlock()

			ot := opentracing.GetOpenTracing()
			spanId, _ := ot.StartChildSpan("Mongo")
			ot.SetChildTag(spanId, "db.type", "mongodb")
			ot.SetChildTag(spanId, "db.statement", fmt.Sprintf("%v", command))
			ot.SetChildTag(spanId, "db.result", "0")
			ot.EndChildSpanByDuration(spanId, duration)
		},
		Failed: func(_ context.Context, evt *event.CommandFailedEvent) {
			logger.PrintInfoCalldepth(9, "%sMongo Failure:%s %s", logger.Red, logger.Reset, evt.Failure)

			mapMutex.Lock()
			command := commandMap[evt.RequestID]
			duration := (time.Now().UTC().UnixNano() - timerMap[evt.RequestID]) / int64(time.Microsecond)
			delete(commandMap, evt.RequestID)
			delete(timerMap, evt.RequestID)
			mapMutex.Unlock()

			ot := opentracing.GetOpenTracing()
			spanId, _ := ot.StartChildSpan("Mongo")
			ot.SetChildTag(spanId, "db.type", "mongodb")
			ot.SetChildTag(spanId, "db.statement", fmt.Sprintf("%v", command))
			ot.SetChildTag(spanId, "db.result", fmt.Sprintf("%v", evt.Failure))
			ot.EndChildSpanByDuration(spanId, duration)
		},
	}
}

func (c *mgoCfgInfo) connMongo() (*qmgo.Client, error) {
	opt := opts.Client().SetAppName(c.appName).SetReplicaSet(c.replica).SetMonitor(c.loggerMongo())
	client, err := qmgo.NewClient(context.Background(), c.cfg, options.ClientOptions{ClientOptions: opt})
	if err != nil {
		logger.PrintError("qmgo.NewClient() Err: %s", err.Error())
		return nil, err
	}
	return client, nil
}

func reconnMongo() *qmgo.Client {
	logger.PrintInfo("reconnMongo Start")

	mgomutex.Lock()
	defer mgomutex.Unlock()

	if mgoClient != nil {
		return mgoClient
	}

	if mgoCfgMap == nil {
		logger.PrintInfo("reconnMongo() Err: cfg empty")
		return nil
	}

	client, err := mgoCfgMap.connMongo()
	if err != nil {
		return nil
	}
	mgoClient = client

	return mgoClient
}

// Close Mgo
func CloseMgo() {
	if mgoClient != nil {
		_ = mgoClient.Close(context.Background())
		mgoClient = nil
	}
}

// 获取 MgoSession
func GetMgoSession() (*qmgo.Session, error) {
	if mgoClient == nil {
		conn := reconnMongo()
		if conn != nil {
			return conn.Session()
		}
		return nil, errors.New("Mgo Client is Nil")
	}
	return mgoClient.Session()
}

// 获取 MgoCollection
// db: Database
// coll: Collection
func GetMgoCollection(db, coll string) *qmgo.Collection {
	if mgoClient == nil {
		conn := reconnMongo()
		if conn != nil {
			return conn.Database(db).Collection(coll)
		}
		return nil
	}
	return mgoClient.Database(db).Collection(coll)
}

// mongodb struct
type MgoCollection struct {
	Database   string
	Collection string
}

// 获取 MgoSession
func (mc MgoCollection) GetMgoCollection() *qmgo.Collection {
	if mgoClient == nil {
		conn := reconnMongo()
		if conn != nil {
			return conn.Database(mc.Database).Collection(mc.Collection)
		}
		return nil
	}
	return mgoClient.Database(mc.Database).Collection(mc.Collection)
}

// 获取 MgoCollection
func (mc MgoCollection) GetMgoSession() (*qmgo.Session, error) {
	if mgoClient == nil {
		conn := reconnMongo()
		if conn != nil {
			return conn.Session()
		}
		return nil, errors.New("Mgo Client is Nil")
	}
	return mgoClient.Session()
}

// 获取 QmgoClient，用于使用集合操作
func (mc MgoCollection) GetQmgoClient() (cli *qmgo.QmgoClient, err error) {
	var db *qmgo.Database
	if mgoClient == nil {
		conn := reconnMongo()
		if conn == nil {
			return nil, errors.New("Mgo Client is Nil")
		}
		db = conn.Database(mc.Database)
	} else {
		db = mgoClient.Database(mc.Database)
	}

	coll := db.Collection(mc.Collection)
	cli = &qmgo.QmgoClient{
		Client:     mgoClient,
		Database:   db,
		Collection: coll,
	}
	return
}
