package storage

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"os"
	"testing"
	"time"
)

func setupMongo() {
	err := InitMongo("Test", "dev_db01", "base_price", "vpcdev",
		"admin", "vpc_dev", 2, 1, 30, 30)
	if err != nil {
		log.Panicf("InitMongo() Err: %s", err.Error())
	}
}

func teardownMongo() {
	CloseMgo()
}

type coll_test struct {
	Id   int64     `bson:"id"`
	Desc string    `bson:"desc"`
	time time.Time `bson:"create_time"`
}

func TestMongoApi(t *testing.T) {
	mgoColl := MgoCollection{"test_db", "test"}
	collection := mgoColl.GetMgoCollection()
	if collection == nil {
		t.Error("GetMgoCollection() Error")
	}

	data := coll_test{
		time.Now().UTC().Unix(),
		"123",
		time.Now().UTC().Add(time.Duration(8) * time.Hour),
	}

	_, err := collection.InsertOne(context.Background(), data)
	if err != nil {
		t.Errorf("InsertOne() Err: %s", err.Error())
	}

	filter := bson.M{
		"id": data.Id,
	}
	total, err := collection.Find(context.Background(), filter).Count()
	if err != nil {
		t.Errorf("InsertOne() Err: %s", err.Error())
	}

	if total != 1 {
		t.Errorf("total Error")
	}

	var result coll_test
	err = collection.Find(context.Background(), filter).One(&result)
	if err != nil {
		t.Errorf("InsertOne() Err: %s", err.Error())
	}

	if result.Id != data.Id {
		t.Errorf("FindOne() result Error: %d != %d", result.Id, data.Id)
	}
}

func Benchmark_MongoInsert(b *testing.B) {
	data := coll_test{
		time.Now().UTC().Unix(),
		"123",
		time.Now().UTC().Add(time.Duration(8) * time.Hour),
	}

	for i := 1; i < b.N; i++ {
		mgoColl := MgoCollection{"test_db", "test"}
		collection := mgoColl.GetMgoCollection()
		if collection == nil {
			b.Error("GetMgoCollection() Error")
		}

		_, err := collection.InsertOne(context.Background(), data)
		if err != nil {
			b.Errorf("InsertOne() Err: %s", err.Error())
		}
	}
}

func Benchmark_MongoFind(b *testing.B) {
	for i := 1; i < b.N; i++ {
		mgoColl := MgoCollection{"test_db", "test"}
		collection := mgoColl.GetMgoCollection()
		if collection == nil {
			b.Error("GetMgoCollection() Error")
		}

		filter := bson.M{}
		var result []coll_test
		err := collection.Find(context.Background(), filter).Skip(0).Limit(50).All(&result)
		if err != nil {
			b.Errorf("InsertOne() Err: %s", err.Error())
		}
	}
}

func TestMain(m *testing.M) {
	setupMongo()
	code := m.Run()
	teardownMongo()
	os.Exit(code)
}
