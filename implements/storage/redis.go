package storage

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-redis/redis/v8"
	jsoniter "github.com/json-iterator/go"
	"go-frame/config"
	"go-frame/implements/opentracing"
	"go-frame/implements/toolkit"
	"go-frame/logger"
	"log"
	"strconv"
	"strings"
	"time"
)

type RedisOpt struct {
	client    *redis.Client
	startTime int64
}

var (
	pRedisOpt *RedisOpt
	//pRedisOptConf *redis.Options
)

type RediGoInterface interface {
	IsConnect() error               // 判断redis是否连接
	IsRedisValueNil(err error) bool // redis返回数据是否是nil
	Del(key ...string) error
	Expire(key string, expiration time.Duration) (bool, error)
	Get(key string, value interface{}) error
	Set(key string, value interface{}, expiration time.Duration) error
	HGet(key string, fields interface{}, value interface{}) error
	HSet(key string, fields interface{}, value interface{}) error
	HDel(key string, fields ...string) error
	Incr(key string) (int64, error)
	MGet(key []string, value map[string]interface{}, retNil bool) error
	MSet(value map[string]interface{}) error
	HIncr(key, field string, incr int) (int64, error)
	HMGet(key string, fields []string, value map[string]interface{}, retNil bool) error
	HMSet(key string, value map[string]interface{}) error // 官方不推荐使用
	ZRange(key string, start, stop int64) ([]string, error)
}

func redisOptions() *redis.Options {
	return &redis.Options{
		DialTimeout:  3 * time.Second,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,

		MaxRetries: -1,
		MaxConnAge: 5 * time.Minute,

		PoolSize:           3,
		MinIdleConns:       1,
		PoolTimeout:        time.Minute, // 连接池忙，等待时间
		IdleTimeout:        3 * time.Minute,
		IdleCheckFrequency: time.Minute,
	}
}

func updatePoolStates() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		stats := pRedisOpt.client.PoolStats()
		log.Printf("Redis PoolStats: %+v", stats)
	}
}

// 正常redis应该只会使用一个实例，所以暂时不做多实例应用
func NewRedisCon() error {
	opt := redisOptions()
	opt.Addr = config.GetRedisHost() + ":" + strconv.Itoa(config.GetRedisPort())
	opt.Password = config.GetRedisAuth()
	opt.DB = 0

	if config.GetRedisPoolMax() != 0 {
		opt.PoolSize = config.GetRedisPoolMax()
	}

	// 不配0，新连接建立会比较慢
	if config.GetRedisPoolMin() != 0 {
		opt.MinIdleConns = config.GetRedisPoolMin()
	}

	client := redis.NewClient(opt)
	if client == nil {
		return errors.New("NewRedisCon Fail!")
	}

	//pRedisOptConf = opt // 用于更新
	pRedisOpt = &RedisOpt{client: client}

	go updatePoolStates()

	return nil
}

func NewRedisConByInfo(host, passwd string, port, dbIndex, poolMax, minIdle int) error {
	opt := redisOptions()
	opt.Addr = host + ":" + strconv.Itoa(port)
	opt.Password = passwd
	opt.DB = dbIndex
	opt.PoolSize = poolMax
	opt.MinIdleConns = minIdle

	if poolMax == 0 {
		opt.PoolSize = 1
		opt.MinIdleConns = 1
	}

	client := redis.NewClient(opt)
	if client == nil {
		return errors.New("NewRedisCon Fail!")
	}

	//pRedisOptConf = opt // 用于更新
	pRedisOpt = &RedisOpt{client: client}

	go updatePoolStates()

	return nil
}

func GetRedisCon() (*RedisOpt, error) {
	if pRedisOpt != nil && pRedisOpt.client != nil {
		return pRedisOpt, nil
	}
	return nil, errors.New("Redis Con <nil>")
}

// 关闭redis
func CloseRedisCon() {
	if pRedisOpt != nil {
		if pRedisOpt.client != nil {
			_ = pRedisOpt.client.Close()
		}
	}
}

func (r *RedisOpt) OpenTracing(cmd string) {
	// opentracing
	duration := (time.Now().UTC().UnixNano() - r.startTime) / int64(time.Microsecond)
	cmdList := strings.Split(cmd, ":")
	ot := opentracing.GetOpenTracing()
	spanId, _ := ot.StartChildSpan("Redis")
	ot.SetChildTag(spanId, "db.type", "redis")
	ot.SetChildTag(spanId, "db.statement", cmdList[0])
	ot.EndChildSpanByDuration(spanId, duration)
}

// 检测redis是否连接
func (r *RedisOpt) IsConnect() error {
	if r == nil || r.client == nil {
		return errors.New("RedisOpt client Is nil!")
	}
	r.startTime = time.Now().UTC().UnixNano()
	return nil
}

// 判断redis返回的err结果是否是redis无数据
func (r *RedisOpt) IsRedisValueNil(err error) bool {
	if err == nil || err.Error() == "redis: nil" {
		return true
	}
	return false
}

// Del操作
func (r *RedisOpt) Del(key ...string) error {
	if err := r.IsConnect(); err != nil {
		return err
	}

	rComd := r.client.Del(context.Background(), key...)
	r.OpenTracing(rComd.String())
	logger.PrintInfoCalldepth(3, "rrdisCmd: %s", rComd.String())

	if rComd.Err() != nil {
		logger.PrintErrorCalldepth(3, "RedisError: %s", rComd.Err().Error())
		return rComd.Err()
	}

	return nil
}

// Expire操作
func (r *RedisOpt) Expire(key string, expiration time.Duration) (bool, error) {
	if err := r.IsConnect(); err != nil {
		return false, err
	}

	rComd := r.client.Expire(context.Background(), key, expiration)
	r.OpenTracing(rComd.String())
	logger.PrintInfoCalldepth(3, "rrdisCmd: %s", rComd.String())

	vaule, err := rComd.Result()
	if err != nil {
		logger.PrintErrorCalldepth(3, "RedisError: %s", rComd.Err().Error())
		return false, err
	}
	return vaule, nil
}

// Get操作
func (r *RedisOpt) Get(key string, value interface{}) error {
	if err := r.IsConnect(); err != nil {
		return err
	}

	rComd := r.client.Get(context.Background(), key)
	r.OpenTracing(rComd.String())
	logger.PrintInfoCalldepth(3, "rrdisCmd: %s", rComd.String())

	if tmp, err := rComd.Bytes(); err != nil {
		if !r.IsRedisValueNil(err) {
			logger.PrintErrorCalldepth(3, "Redis Get Error: %s", err.Error())
		}
		return err
	} else {
		var json = jsoniter.ConfigCompatibleWithStandardLibrary
		if err = json.Unmarshal(tmp[:], value); err != nil {
			logger.PrintErrorCalldepth(3, "json.Unmarshal() %s", err.Error())
			return err
		}
	}

	return nil
}

// Set操作
// expiration: 0 表示没有过期时间
func (r *RedisOpt) Set(key string, value interface{}, expiration time.Duration) error {
	if err := r.IsConnect(); err != nil {
		return err
	}

	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		logger.PrintErrorCalldepth(3, "RedisOpt::Set() json.Marshal() Err: %s", err.Error())
		return err
	}

	rComd := r.client.Set(context.Background(), key, jsonBytes, expiration)
	r.OpenTracing(rComd.String())
	logger.PrintInfoCalldepth(3, "rrdisCmd: %s", rComd.String())

	if rComd.Err() != nil {
		logger.PrintErrorCalldepth(3, "RedisError: %s", rComd.Err().Error())
		return rComd.Err()
	}

	return nil
}

// HGet操作
func (r *RedisOpt) HGet(key string, fields interface{}, value interface{}) error {
	if err := r.IsConnect(); err != nil {
		return err
	}

	rComd := r.client.HGet(context.Background(), key, toolkit.ConvertToString(fields))
	r.OpenTracing(rComd.String())
	logger.PrintInfoCalldepth(3, "rrdisCmd: %s", rComd.String())

	if tmp, err := rComd.Bytes(); err != nil {
		if !r.IsRedisValueNil(err) {
			logger.PrintErrorCalldepth(3, "Redis HGet Error: %s", err.Error())
		}
		return err
	} else {
		var json = jsoniter.ConfigCompatibleWithStandardLibrary
		if err = json.Unmarshal(tmp[:], value); err != nil {
			logger.PrintErrorCalldepth(3, "json.Unmarshal() %s", err.Error())
			return err
		}
	}

	return nil
}

// Set操作
func (r *RedisOpt) HSet(key string, fields interface{}, value interface{}) error {
	if err := r.IsConnect(); err != nil {
		return err
	}

	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		logger.PrintErrorCalldepth(3, "RedisOpt::Set() json.Marshal() Err: %s", err.Error())
		return err
	}

	rComd := r.client.HSet(context.Background(), key, toolkit.ConvertToString(fields), jsonBytes)
	r.OpenTracing(rComd.String())
	logger.PrintInfoCalldepth(3, "rrdisCmd: %s", rComd.String())

	if rComd.Err() != nil {
		logger.PrintErrorCalldepth(3, "RedisError: %s", rComd.Err().Error())
		return rComd.Err()
	}

	return nil
}

// HDel操作
func (r *RedisOpt) HDel(key string, fields ...string) error {
	if err := r.IsConnect(); err != nil {
		return err
	}

	rComd := r.client.HDel(context.Background(), key, fields...)
	r.OpenTracing(rComd.String())
	logger.PrintInfoCalldepth(3, "rrdisCmd: %s", rComd.String())

	if rComd.Err() != nil {
		logger.PrintErrorCalldepth(3, "RedisError: %s", rComd.Err().Error())
		return rComd.Err()
	}

	return nil
}

// Incr操作
func (r *RedisOpt) Incr(key string) (int64, error) {
	if err := r.IsConnect(); err != nil {
		return -1, err
	}

	rComd := r.client.Incr(context.Background(), key)
	r.OpenTracing(rComd.String())
	logger.PrintInfoCalldepth(3, "rrdisCmd: %s", rComd.String())

	vaule, err := rComd.Result()
	if err != nil {
		logger.PrintErrorCalldepth(3, "RedisError: %s", rComd.Err().Error())
		return -1, err
	}
	return vaule, nil
}

// MGet操作
func (r *RedisOpt) MGet(key []string, value map[string]interface{}, retNil bool) error {
	if err := r.IsConnect(); err != nil {
		return err
	}

	sliceCmd := r.client.MGet(context.Background(), key[:]...)
	r.OpenTracing(sliceCmd.String())
	logger.PrintInfoCalldepth(3, "rrdisCmd: %s", sliceCmd.String())

	vaule, err := sliceCmd.Result()
	if err != nil {
		if !r.IsRedisValueNil(err) {
			logger.PrintErrorCalldepth(3, "Redis MGet Error: %s", err.Error())
		}
		return err
	}

	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	for i, k := range vaule {
		if k == nil {
			if retNil {
				value[key[i]] = nil
			}
			continue
		}

		if err = json.Unmarshal([]byte(k.(string)), value[key[i]]); err != nil {
			logger.PrintErrorCalldepth(3, "json.Unmarshal(%v) %s", k, err.Error())
			return err
		}
	}

	return nil
}

// MSet操作
func (r *RedisOpt) MSet(value map[string]interface{}) error {
	if err := r.IsConnect(); err != nil {
		return err
	}

	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	tmpMap := make(map[string]interface{})
	for k, v := range value {
		if jsonBytes, err := json.Marshal(v); err != nil {
			logger.PrintErrorCalldepth(3, "RedisOpt::MSet() json.Marshal() Err: %s", err.Error())
			return err
		} else {
			tmpMap[k] = jsonBytes
		}
	}

	statusCmd := r.client.MSet(context.Background(), tmpMap)
	r.OpenTracing(statusCmd.String())
	logger.PrintInfoCalldepth(3, "rrdisCmd: %s", statusCmd.String())

	if statusCmd.Err() != nil {
		logger.PrintErrorCalldepth(3, "RedisError: %s", statusCmd.Err().Error())
		return statusCmd.Err()
	}

	return nil
}

// HMGet操作
// 参数value的interface{}需要是指针类型
func (r *RedisOpt) HMGet(key string, fields []string, value map[string]interface{}, retNil bool) error {
	if err := r.IsConnect(); err != nil {
		return err
	}

	sliceCmd := r.client.HMGet(context.Background(), key, fields[:]...)
	r.OpenTracing(sliceCmd.String())
	logger.PrintInfoCalldepth(3, "rrdisCmd: %s", sliceCmd.String())

	result, err := sliceCmd.Result()
	if err != nil {
		logger.PrintErrorCalldepth(3, "RedisError: %s", sliceCmd.Err().Error())
		return err
	}

	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	for i, k := range result {
		if k == nil {
			if retNil {
				value[fields[i]] = nil
			}
			continue
		}

		if err = json.Unmarshal([]byte(k.(string)), value[fields[i]]); err != nil {
			logger.PrintErrorCalldepth(3, "json.Unmarshal(%v) %s", k, err.Error())
			return err
		}

		/*
		val := reflect.ValueOf(value[fields[i]]) //获取reflect.Type类型
		switch val.Kind() {
		case reflect.Int,reflect.Int8,reflect.Int16,reflect.Int32,reflect.Int64,reflect.Uint,reflect.Uint8,reflect.Uint16,reflect.Uint32,reflect.Uint64:
			var tem int
			if err = json.Unmarshal([]byte(k.(string)), &tem); err != nil {
				logger.PrintErrorCalldepth(3, "json.Unmarshal(%v) %s", k, err.Error())
				return err
			}
			value[fields[i]] = tem
		default:
			tem := value[fields[i]]
			if err = json.Unmarshal([]byte(k.(string)), &tem); err != nil {
				logger.PrintErrorCalldepth(3, "json.Unmarshal(%v) %s", k, err.Error())
				return err
			}
			value[fields[i]] = tem
		}
		*/
	}

	return nil
}

// HMSet操作
// HMSet is a deprecated version of HSet left for compatibility with Redis 3.
func (r *RedisOpt) HMSet(key string, value map[string]interface{}) error {
	if err := r.IsConnect(); err != nil {
		return err
	}

	tmpMap := make(map[string]interface{})
	for k, v := range value {
		if jsonBytes, err := json.Marshal(v); err != nil {
			logger.PrintErrorCalldepth(3, "RedisOpt::MSet() json.Marshal() Err: %s", err.Error())
			return err
		} else {
			tmpMap[k] = jsonBytes
		}
	}

	statusCmd := r.client.HMSet(context.Background(), key, tmpMap)
	r.OpenTracing(statusCmd.String())
	logger.PrintInfoCalldepth(3, "rrdisCmd: %s", statusCmd.String())

	if statusCmd.Err() != nil {
		logger.PrintErrorCalldepth(3, "RedisError: %s", statusCmd.Err().Error())
		return statusCmd.Err()
	}

	return nil
}

// Incr操作
func (r *RedisOpt) HIncr(key, field string, incr int) (int64, error) {
	if err := r.IsConnect(); err != nil {
		return -1, err
	}

	intCmd := r.client.HIncrBy(context.Background(), key, field, int64(incr))
	r.OpenTracing(intCmd.String())
	logger.PrintInfoCalldepth(3, "rrdisCmd: %s", intCmd.String())

	vaule, err := intCmd.Result()
	if err != nil {
		logger.PrintErrorCalldepth(3, "RedisError: %s", intCmd.Err().Error())
		return -1, err
	}
	return vaule, nil
}

func (r *RedisOpt) ZRange(key string, start, stop int64) ([]string, error) {
	if err := r.IsConnect(); err != nil {
		return nil, err
	}

	sliceCmd := r.client.ZRange(context.Background(), key, start, stop)
	r.OpenTracing(sliceCmd.String())
	logger.PrintInfoCalldepth(3, "rrdisCmd: %s", sliceCmd.String())

	if sliceCmd.Err() != nil {
		logger.PrintErrorCalldepth(3, "RedisError: %s", sliceCmd.Err().Error())
		return nil, sliceCmd.Err()
	}

	return sliceCmd.Result()
}
