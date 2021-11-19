package storage

import (
	"log"
	"os"
	"testing"
	"time"
)

func setupRedis() {
	err := NewRedisConByInfo("EVA_REDIS_HOST", "hsb_redis_123", 6379,
		0, 2, 1)
	if err != nil {
		log.Panicf("InitMongo() Err: %s", err.Error())
	}
}

func teardownRedis() {
	CloseRedisCon()
}

func BenchmarkRedisOpt_Get(b *testing.B) {
	redisPool, err := GetRedisCon()
	if err != nil {
		b.Errorf("GetRedisCon() Err: %s", err.Error())
	}

	value := "testValue"
	if err := redisPool.Set("testKey", value, 10*time.Second); err != nil {
		b.Errorf("redis.Set() Err: %s", err.Error())
	}

	for i := 0; i < b.N; i++ {
		var data string
		if err := redisPool.Get("testKey", &data); err != nil {
			b.Errorf("redis.Get() Err: %s", err.Error())
		}
	}
}

func BenchmarkRedisOpt_Set(b *testing.B) {
	redisPool, err := GetRedisCon()
	if err != nil {
		b.Errorf("GetRedisCon() Err: %s", err.Error())
	}

	value := "testValue"
	for i := 0; i < b.N; i++ {
		if err := redisPool.Set("testKey", value, 10*time.Second); err != nil {
			b.Errorf("redis.Set() Err: %s", err.Error())
		}
	}
}

func BenchmarkRedisOpt_HSet(b *testing.B) {
	redisPool, err := GetRedisCon()
	if err != nil {
		b.Errorf("GetRedisCon() Err: %s", err.Error())
	}

	fields := "testField"
	for i := 0; i < b.N; i++ {
		if err := redisPool.HSet("testHKey", fields, i); err != nil {
			b.Errorf("redis.HSet() Err: %s", err.Error())
		}
	}
}

func BenchmarkRedisOpt_HGet(b *testing.B) {
	redisPool, err := GetRedisCon()
	if err != nil {
		b.Errorf("GetRedisCon() Err: %s", err.Error())
	}

	fields := "testField"
	var value int
	for i := 0; i < b.N; i++ {
		if err := redisPool.HGet("testHKey", fields, &value); err != nil {
			b.Errorf("redis.HGet() Err: %s", err.Error())
		}
	}
}

func TestRedisApi(t *testing.T) {
	redisPool, err := GetRedisCon()
	if err != nil {
		t.Errorf("GetRedisCon() Err: %s", err.Error())
	}

	value := "testValue"
	if err := redisPool.Set("testKey", value, 10*time.Hour); err != nil {
		t.Errorf("redis.Set() Err: %s", err.Error())
	}

	if ok, err := redisPool.Expire("testKey", 10*time.Hour); err != nil {
		t.Errorf("redis.Set() Err: %s", err.Error())
	} else if !ok {
		t.Errorf("redis.Set() ok = false")
	}

	var data string
	if err := redisPool.Get("testKey", &data); err != nil {
		t.Errorf("redis.Get() Err: %s", err.Error())
	} else if data != value {
		t.Errorf("redis.Get() Error data[%s] != value[%s]", data, value)
	}

	if err := redisPool.Del("testKey"); err != nil {
		t.Errorf("redis.Del() Err: %s", err.Error())
	}

	value2 := 100
	if err := redisPool.HSet("testHKey", "testF", value2); err != nil {
		t.Errorf("redis.HSet() Err: %s", err.Error())
	}

	var value3 int
	if err := redisPool.HGet("testHKey", "testF", &value3); err != nil {
		t.Errorf("redis.HGet() Err: %s", err.Error())
	} else if value2 != value3 {
		t.Errorf("redis.HGet() Error value2[%d] != value3[%d]", value2, value3)
	}

	if err := redisPool.HDel("testHKey", "testF"); err != nil {
		t.Errorf("redis.HDel() Err: %s", err.Error())
	}

	hmsetMap := map[string]interface{}{
		"field1": int(1),
		"field2": int(2),
		"field3": int(3),
		"field4": int(4),
		"field5": int(5),
		"field6": int(6),
		"field7": int(7),
		"field8": int(8),
		"field9": int(9),
		"field0": int(0),
	}
	if err := redisPool.HMSet("testHKey", hmsetMap); err != nil {
		t.Errorf("redis.HMSet() Err: %s", err.Error())
	}

	fields := []string{
		"field1",
		"field2",
		"field3",
		"field4",
		"field5",
		"field6",
		"field7",
		"field8",
		"field9",
		"field0",
	}

	temp := make([]int, 10)
	hmsetMap2 := map[string]interface{}{
		"field1": &temp[0],
		"field2": &temp[1],
		"field3": &temp[2],
		"field4": &temp[3],
		"field5": &temp[4],
		"field6": &temp[5],
		"field7": &temp[6],
		"field8": &temp[7],
		"field9": &temp[8],
		"field0": &temp[9],
	}
	if err := redisPool.HMGet("testHKey", fields, hmsetMap2, true); err != nil {
		t.Errorf("redis.HMGet() Err: %s", err.Error())
	}

	for k, v := range hmsetMap2 {
		if hmsetMap[k] != *(v.(*int)) {
			t.Errorf("redisPool.HMGet() Err: %d != %d", v, hmsetMap[k])
		}
	}

	if err := redisPool.Del("testHKey"); err != nil {
		t.Errorf("redis.Del() Err: %s", err.Error())
	}
}

func TestMain(m *testing.M) {
	setupRedis()
	code := m.Run()
	teardownRedis()
	os.Exit(code)
}
