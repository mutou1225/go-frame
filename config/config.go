package config

import (
	"encoding/xml"
	"fmt"
	"go-frame/implements/watcher"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	gConfig     = &xmlConfig{}
	gConfigFile = ""
)

const (
	configUpdateTime = 2 * time.Minute
)

type xmlConfig struct {
	XmlName       xml.Name     `xml:"xml"`
	IsTestEnv     bool         `xml:"IsTest"`
	Environment   string       `xml:"Environment"`
	PrintLen      int          `xml:"PrintLen"`
	EnableZipkin  bool         `xml:"EnableZipkin"`
	MysqlDB       xmlDb        `xml:"DB"`
	SlaveDB       xmlDb        `xml:"SlaveDB"`
	HsbDB         xmlDb        `xml:"HsbDB"`
	BiDB          xmlDb        `xml:"BiDB"`
	Redis         xmlRedis     `xml:"REDIS"`
	RabbitMQ      xmlRabbitMQ  `xml:"RABBITMQ"`
	ElasticSearch xmlES        `xml:"ElasticSearch"`
	Kafka         xmlKafka     `xml:"KAFKA"`
	MongoDB       []xmlMongoDB `xml:"MongoDB"`
	MongoDBCfg    xmlMgDBCfg   `xml:"MongoDBCfg"`
	DingTalk      xmlDingTalk  `xml:"DingTalk"`
}

type xmlDb struct {
	HostName    string `xml:"HostName"`
	Port        int    `xml:"Port"`
	UserName    string `xml:"UserName"`
	Password    string `xml:"Password"`
	DBName      string `xml:"DBName"`
	PageDBName  string `xml:"PageDBName"`
	ReportName  string `xml:"ReportName"`
	IdleTimeout int    `xml:"IdleTimeout"`
}

type xmlRedis struct {
	Host       string `xml:"Host"`
	Port       int    `xml:"Port"`
	Auth       string `xml:"Auth"`
	ReportName string `xml:"ReportName"`
}

type xmlRabbitMQ struct {
	Host       string `xml:"Host"`
	Port       int    `xml:"Port"`
	UserName   string `xml:"UserName"`
	Password   string `xml:"Password"`
	RootVhost  string `xml:"Vhost"`
	EvaVhost   string `xml:"EvaVhost"`
	ChanVhost  string `xml:"ChanVhost"`
	ReportName string `xml:"ReportName"`
}

type xmlES struct {
	Host          string `xml:"Host"`
	Index         string `xml:"Index"`
	IndexProduct  string `xml:"IndexProduct"`
	IndexPlatform string `xml:"IndexPlatform"`
	IndexStandard string `xml:"IndexStandard"`
	ReportName    string `xml:"ReportName"`
}

type xmlKafka struct {
	Host       string `xml:"Host"`
	ReportName string `xml:"ReportName"`
}

type xmlMongoDB struct {
	Host     string `xml:"Host"`
	Port     int    `xml:"Port"`
	User     string `xml:"User"`
	Password string `xml:"Password"`
}

type xmlMgDBCfg struct {
	DBName     string `xml:"DBName"`
	Replica    string `xml:"Replica"`
	ReportName string `xml:"ReportName"`
}

type xmlDingTalk struct {
	Operate     string `xml:"Operate"`
	OperateConf string `xml:"OperateConf"`
	Develop     string `xml:"Develop"`
}

// 初始化系统配置文件
func InitConfig(configFile string) {
	if configFile == "" {
		configFile = "/huishoubao/config/tinyxml2/eva_pro_config.xml"
	}
	gConfigFile = configFile

	file, err := os.Open(configFile)
	if err != nil {
		panic(fmt.Sprintf("open config error:%s", err.Error()))
	}
	defer file.Close() // 关闭文件

	data, err := ioutil.ReadAll(file)
	if err != nil {
		panic(fmt.Sprintf("load config error:%s", err.Error()))
	}

	err = xml.Unmarshal(data, &gConfig)
	if err != nil {
		panic(fmt.Sprintf("load config error:%s", err.Error()))
	}

	log.Printf("Gconfig: %+v", gConfig)
	log.Println(strings.Repeat("~", 37))

	go updateConfig()
}

func updateConfig() {
	w, err := watcher.FileWatcher(gConfigFile, configUpdateTime)
	if err != nil {
		log.Printf("updateConfig() Err: %s", err.Error())
		return
	}

	for {
		select {
		case event := <-w.Event:
			log.Println(event)
			if file, err := os.Open(gConfigFile); err == nil {
				if data, err := ioutil.ReadAll(file); err == nil {
					tmpConfig := &xmlConfig{}
					if err := xml.Unmarshal(data, tmpConfig); err == nil {
						gConfig = tmpConfig
						log.Printf("updateConfig: %+v", gConfig)
					}
				}
				file.Close() // 关闭文件
			}
		case err := <-w.Error:
			log.Printf("updateConfig() Err: %s", err.Error())
		case <-w.Closed:
			log.Printf("updateConfig() Closed")
			return
		}
	}
}

// 获取全部配置信息
func GetAllConfig() *xmlConfig {
	return gConfig
}

// 是不是开发测试环境
func IsTest() bool {
	return gConfig.IsTestEnv
}

// 环境标识
func GetEnvironment() string {
	return gConfig.Environment
}

// 打印长度
func GetPrintLen() int {
	return gConfig.PrintLen
}

// MysqlDB Host
func GetDBHost() string {
	return gConfig.MysqlDB.HostName
}

// MysqlDB UserName
func GetDBUser() string {
	return gConfig.MysqlDB.UserName
}

// MysqlDB Password
func GetDBPassword() string {
	return gConfig.MysqlDB.Password
}

// MysqlDB Port
func GetDBPort() int {
	return gConfig.MysqlDB.Port
}

// MysqlDB DBName
func GetDBName() string {
	return gConfig.MysqlDB.DBName
}

// MysqlDB IdleTimeout
func GetDBIdleTimeout() int {
	return gConfig.MysqlDB.IdleTimeout
}

// HsbDB Host
func GetHsbDBHost() string {
	return gConfig.HsbDB.HostName
}

// HsbDB UserName
func GetHsbDBUser() string {
	return gConfig.HsbDB.UserName
}

// HsbDB Password
func GetHsbDBPassword() string {
	return gConfig.HsbDB.Password
}

// HsbDB Port
func GetHsbDBPort() int {
	return gConfig.HsbDB.Port
}

// HsbDB DBName
func GetHsbDBName() string {
	return gConfig.HsbDB.DBName
}

// HsbDB IdleTimeout
func GetHsbDBIdleTimeout() int {
	return gConfig.HsbDB.IdleTimeout
}

// Mysql BiDB Host
func GetBiDBHost() string {
	return gConfig.BiDB.HostName
}

// Mysql BiDB UserName
func GetBiDBUser() string {
	return gConfig.BiDB.UserName
}

// Mysql BiDB Password
func GetBiDBPassword() string {
	return gConfig.BiDB.Password
}

// Mysql BiDB Port
func GetBiDBPort() int {
	return gConfig.BiDB.Port
}

// Mysql BiDB DBName
func GetBiDBName() string {
	return gConfig.BiDB.DBName
}

// Mysql BiDB IdleTimeout
func GetBiDBIdleTimeout() int {
	return gConfig.BiDB.IdleTimeout
}

// Redis Host
func GetRedisHost() string {
	return gConfig.Redis.Host
}

// Redis Auth
func GetRedisAuth() string {
	return gConfig.Redis.Auth
}

// Redis Port
func GetRedisPort() int {
	return gConfig.Redis.Port
}

// RabbitMQ Host
func GetRabbitMQHost() string {
	return gConfig.RabbitMQ.Host
}

// RabbitMQ Port
func GetRabbitMQPort() int {
	return gConfig.RabbitMQ.Port
}

// RabbitMQ UserName
func GetRabbitMQUser() string {
	return gConfig.RabbitMQ.UserName
}

func GetRabbitMQPassword() string {
	return gConfig.RabbitMQ.Password
}

func GetRabbitMQEvaVhost() string {
	return gConfig.RabbitMQ.EvaVhost
}

func GetMongoDBAdd() []string {
	addr := make([]string, 0)
	for _, confg := range gConfig.MongoDB {
		addr = append(addr, confg.Host+":"+strconv.Itoa(confg.Port))
	}
	return addr
}

func GetMongoDBAddStr() string {
	var addStr strings.Builder
	for i, confg := range gConfig.MongoDB {
		if i == 0 {
			addStr.WriteString(fmt.Sprintf("%s:%d", confg.Host, confg.Port))
		} else {
			addStr.WriteString(fmt.Sprintf(",%s:%d", confg.Host, confg.Port))
		}
	}
	return addStr.String()
}

func GetMongoDBUsert() string {
	for _, confg := range gConfig.MongoDB {
		if confg.User != "" {
			return confg.User
		}
	}
	return ""
}

func GetMongoDBPasswd() string {
	for _, confg := range gConfig.MongoDB {
		if confg.User != "" {
			return confg.Password
		}
	}
	return ""
}

func GetMongoDBName() string {
	return gConfig.MongoDBCfg.DBName
}

func GetMongoDBReplica() string {
	return gConfig.MongoDBCfg.Replica
}

// MysqlDB Host
func GetESHost() string {
	return gConfig.ElasticSearch.Host
}

// 获取钉钉token
func GetDingTalkOperate() string {
	return gConfig.DingTalk.Operate
}

// 获取钉钉token
func GetDingTalkOperateConf() string {
	return gConfig.DingTalk.OperateConf
}

// 获取钉钉token
func GetDingTalkDevelop() string {
	return gConfig.DingTalk.Develop
}
