package storage

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"testing"
	"time"
)

func setupMysql() {
	err := NewMysqlDB(TestMysql, "EVA_BD_HOST", "eva", "123456", "test",
		3307, 2, 1, 30, 30, true)
	if err != nil {
		log.Panicf("InitMongo() Err: %s", err.Error())
	}
}

func teardownMysql() {
	ExitDB()
}

type tTest struct {
	Fid          int       `gorm:"primary_key;column:Fid"`
	Fname        string    `gorm:"column:Fname"`
	Fupdate_time time.Time `gorm:"column:Fupdate_time"`
}

func TestMysqlApi(t *testing.T) {
	mysqlDB := GetDBHandle(TestMysql)
	if mysqlDB == nil {
		t.Error("GetDBHandle() <nil>")
	}

	data := tTest{
		1,
		strconv.Itoa(int(time.Now().UTC().Unix())),
		time.Now(),
	}

	db := mysqlDB.Table("t_test")
	db = db.Save(data)
	if err := GetDBError(db); err != nil {
		t.Errorf("db.Save() Err: %s", err.Error())
	}

	db = mysqlDB.Table("t_test").Select("Fid, Fname, Fupdate_time")
	db = db.Where("Fname LIKE ?", fmt.Sprintf("%%%d%%", 16))

	var total int64
	db = db.Count(&total)
	if err := GetDBError(db); err != nil {
		t.Errorf("db.Count() Err: %s", err.Error())
	}
	t.Logf("total: %d", total)

	if total == 0 {
		t.Errorf("db.Count() total = 0")
	}

	var dest []tTest
	db = db.Limit(10).Offset(0).Scan(&dest)
	if err := GetDBError(db); err != nil {
		t.Errorf("db.Count() Err: %s", err.Error())
	}
}

func Benchmark_Find(b *testing.B) {
	mysqlDB := GetDBHandle(TestMysql)
	if mysqlDB == nil {
		b.Error("GetDBHandle() <nil>")
	}

	for i := 1; i < b.N; i++ {
		var dest []tTest
		db := mysqlDB.Table("t_test").Select("Fid, Fname, Fupdate_time").Find(&dest)
		db = db.Limit(50).Offset(0)
		if err := GetDBError(db); err != nil {
			b.Errorf("db.Count() Err: %s", err.Error())
		}
	}
}

func Benchmark_Create(b *testing.B) {
	mysqlDB := GetDBHandle(TestMysql)
	if mysqlDB == nil {
		b.Error("GetDBHandle() <nil>")
	}

	for i := 1; i < b.N; i++ {
		data2 := tTest{
			Fid: 0,
			Fname:        strconv.Itoa(int(time.Now().UTC().Unix())),
			Fupdate_time: time.Now(),
		}

		db := mysqlDB.Table("t_test").Create(&data2)
		if err := GetDBError(db); err != nil {
			b.Errorf("db.Save() Err: %s", err.Error())
		}
	}
}

func TestMain(m *testing.M) {
	setupMysql()
	code := m.Run()
	teardownMysql()
	os.Exit(code)
}
