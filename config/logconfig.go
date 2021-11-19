package config

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var GLogConfig = xmlLogConfig{}

type xmlLogConfig struct {
	XMLName      xml.Name `xml:"xml"`
	LogColor     bool     `xml:"LogColor"`
	LogConfig    xmlLog   `xml:"Log"`
	ReportConfig xmlLog   `xml:"report"`
	ZipkinConfig xmlLog   `xml:"Zipkin"`
}

type xmlLog struct {
	LogFilePath string `xml:"FilePath"`
	AdapterName string `xml:"AdapterName"`
	LogLevel    string `xml:"LogLevel"`
	MaxSize     int64  `xml:"MaxSize"`
	LineSize    int    `xml:"LineSize"`
	Maxdays     int    `xml:"Maxdays"`
	Suffix      string `xml:"Suffix"`
}

// 初始化系统配置文件
func InitLogConfig() {
	file, err := os.Open("/huishoubao/config/GoAppLogConfig.xml")
	if err != nil {
		fmt.Sprintf("open config error:%s", err.Error())
		return
	}
	defer file.Close() // 关闭文件

	data, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Sprintf("load config error:%s", err.Error())
		return
	}

	err = xml.Unmarshal(data, &GLogConfig)
	if err != nil {
		fmt.Sprintf("load config error:%s", err.Error())
		return
	}

	log.Printf("GLogConfig: %+v", GLogConfig)
	log.Println(strings.Repeat("~", 37))
}
