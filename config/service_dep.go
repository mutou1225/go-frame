package config

import (
	"encoding/xml"
	"eva_services_go/implements/watcher"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

var (
	gServerconfig     = &xmlServerConfig{}
	gServerConfigFile = ""
	serverCaller      *map[string]string //Caller <id key>
	serverCallee      *map[string]CalleeConfig
	serverOther       *map[string]string //Caller <k, v>
)

type xmlServerConfig struct {
	XMLName      xml.Name       `xml:"xml"`
	ServerConfig ServerConfig   `xml:"Server"`
	MysqlPool    DBPoolConfig   `xml:"MysqlPool"`
	MgodbPool    DBPoolConfig   `xml:"MgodbPool"`
	RedisPool    DBPoolConfig   `xml:"RedisPool"`
	Caller       []callerConfig `xml:"Caller"`
	Callee       []CalleeConfig `xml:"Callee"`
	Other        []OtherConfig  `xml:"Other"`
}

type ServerConfig struct {
	ServerId    int    `xml:"ServerId"`
	ServerName  string `xml:"ServerName"`
	ServerModel string `xml:"ServerModel"`
	ServerPort  int    `xml:"ServerPort"`
	MonitorPort int    `xml:"MonitorPort"`
	LogFileName string `xml:"LogFileName"`
}

type DBPoolConfig struct {
	PoolMin  int `xml:"PoolMin"`
	PoolMax  int `xml:"PoolMax"`
	IdleTime int `xml:"IdleTime"`
	ConnTime int `xml:"ConnTimeout"`
}

type callerConfig struct {
	Id  int    `xml:"id"`
	Key string `xml:"key"`
}

type CalleeConfig struct {
	ServerId   int    `xml:"ServerId"`
	ServerName string `xml:"ServerName"`
	ServerUrl  string `xml:"ServerUrl"`
	ServerKey  string `xml:"ServerKey"`
}

type OtherConfig struct {
	Key   string `xml:"Key"`
	Value string `xml:"Value"`
}

// 初始化app的依赖配置
func InitServerConfig(configFile string) {
	gServerConfigFile = configFile
	if configFile == "" {
		log.Println("InitServerConfig() Error! configFile Empty!")
		return
	}

	file, err := os.Open(configFile)
	if err != nil {
		panic(fmt.Sprintf("open config error:%s", err.Error()))
	}
	defer file.Close() // 关闭文件

	data, err := ioutil.ReadAll(file)
	if err != nil {
		panic(fmt.Sprintf("load config error:%s", err.Error()))
	}

	err = xml.Unmarshal(data, &gServerconfig)
	if err != nil {
		panic(fmt.Sprintf("load config error:%s", err.Error()))
	}

	initCallerInfo()
	initCalleeInfo()
	initOtherInfo()

	log.Printf("gServerconfig: %+v", gServerconfig)
	log.Println(strings.Repeat("~", 37))

	go updateServerConfig()
}

func updateServerConfig() {
	w, err := watcher.FileWatcher(gServerConfigFile, configUpdateTime)
	if err != nil {
		log.Printf("updateServerConfig() Err: %s", err.Error())
		return
	}

	for {
		select {
		case event := <-w.Event:
			log.Println(event)
			if file, err := os.Open(gServerConfigFile); err == nil {
				if data, err := ioutil.ReadAll(file); err == nil {
					tmpConfig := &xmlServerConfig{}
					if err := xml.Unmarshal(data, tmpConfig); err == nil {
						gServerconfig = tmpConfig

						initCallerInfo()
						initCalleeInfo()
						initOtherInfo()

						log.Printf("gServerconfig: %+v", gServerconfig)
						log.Println(strings.Repeat("~", 37))
					}
				}
				file.Close() // 关闭文件
			}
		case err := <-w.Error:
			log.Printf("updateServerConfig() Err: %s", err.Error())
		case <-w.Closed:
			log.Printf("updateServerConfig() Closed")
			return
		}
	}
}

// 初始化可请求的server id
func initCallerInfo() {
	// 初始化
	serverCallerTmp := make(map[string]string)
	for _, caller := range gServerconfig.Caller {
		serverCallerTmp[strconv.Itoa(caller.Id)] = caller.Key
	}
	serverCaller = &serverCallerTmp

	log.Println("Caller:", serverCaller)
	log.Println(strings.Repeat("~", 37))
}

// 初始化被调方信息
func initCalleeInfo() {
	// 初始化
	serverCalleeTmp := make(map[string]CalleeConfig)
	for _, callee := range gServerconfig.Callee {
		serverCalleeTmp[strconv.Itoa(callee.ServerId)] = callee
		serverCalleeTmp[callee.ServerName] = callee
		log.Printf("Callee: %+v", callee)
	}
	serverCallee = &serverCalleeTmp
	log.Println(strings.Repeat("~", 37))
}

func initOtherInfo() {
	serverOtherTmp := make(map[string]string)
	for _, other := range gServerconfig.Other {
		serverOtherTmp[other.Key] = other.Value
		log.Printf("Other: %+v", other)
	}
	serverOther = &serverOtherTmp
	log.Println(strings.Repeat("~", 37))
}

// 获取Caller的key
func GetCallerKey(serId string) (key string, ok bool) {
	key, ok = (*serverCaller)[serId]
	return
}

func GetOtherValue(key string) (value string, ok bool) {
	value, ok = (*serverOther)[key]
	return
}

// 获取全部配置信息
func GetAllSerConfig() *xmlServerConfig {
	return gServerconfig
}

// 获取本应用的 ServerId
func GetServerId() int {
	return gServerconfig.ServerConfig.ServerId
}

func GetServerIdStr() string {
	return strconv.Itoa(gServerconfig.ServerConfig.ServerId)
}

// 获取本应用的 ServerName
func GetServerName() string {
	return gServerconfig.ServerConfig.ServerName
}

// 获取本应用的 ServerPort
func GetServerPort() int {
	return gServerconfig.ServerConfig.ServerPort
}

// 获取本应用的 MonitorPort
func GetSerMonitorPort() int {
	return gServerconfig.ServerConfig.MonitorPort
}

// 获取本应用的 LogFileName
func GetSerLogFileName() string {
	return gServerconfig.ServerConfig.LogFileName
}

// 获取本应用的 Mysql PoolMin
func GetMysqlPoolMin() int {
	return gServerconfig.MysqlPool.PoolMin
}

// 获取本应用的 Mysql PoolMax
func GetMysqlPoolMax() int {
	return gServerconfig.MysqlPool.PoolMax
}

// 获取本应用的 Mysql IdleTime
func GetMysqlIdleTime() int {
	return gServerconfig.MysqlPool.IdleTime
}

// 获取本应用的 Mysql ConnTime
func GetMysqlConnTime() int {
	return gServerconfig.MysqlPool.ConnTime
}

// 获取本应用的 Mgodb PoolMin
func GetMgodbPoolMin() int {
	return gServerconfig.MgodbPool.PoolMin
}

// 获取本应用的 Mgodb PoolMax
func GetMgodbPoolMax() int {
	return gServerconfig.MgodbPool.PoolMax
}

// 获取本应用的 Mgodb IdleTime
func GetMgodbIdleTime() int {
	return gServerconfig.MgodbPool.IdleTime
}

// 获取本应用的 Mgodb PoolMin
func GetRedisPoolMin() int {
	return gServerconfig.RedisPool.PoolMin
}

// 获取本应用的 Mgodb PoolMax
func GetRedisPoolMax() int {
	return gServerconfig.RedisPool.PoolMax
}

// 获取本应用的 Mgodb IdleTime
func GetRedisIdleTime() int {
	return gServerconfig.RedisPool.IdleTime
}

// 获取本应用的 Mgodb ConnTime
func GetMgodbConnTime() int {
	return gServerconfig.MgodbPool.ConnTime
}

// 获取被调方信息
func GetCalleeByServerId(serId string) (callss CalleeConfig, ok bool) {
	callss, ok = (*serverCallee)[serId]
	return
}

// 获取被调方信息
func GetCalleeByServerName(serName string) (callss CalleeConfig, ok bool) {
	callss, ok = (*serverCallee)[serName]
	return
}
