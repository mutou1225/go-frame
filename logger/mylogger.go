package logger

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"go-frame/config"
	"go-frame/implements/http"
	"log"
	"reflect"
	"time"
)

var (
	myLogger       *MyLogger
	myLoggerReport *MyLogger
	myLoggerZipkin *MyLogger
	hostname       string
)

const (
	knownFrames = 3
)

// 初始化日志系统
func InitLogger(appName, fileName string) {
	hostname, _ = http.ExternalIP()
	logLevel, err := logrus.ParseLevel(config.GLogConfig.LogConfig.LogLevel)
	if err != nil {
		logLevel = InfoLevel
	}

	logConf := LogrusConfig{
		ProgramName: appName,
		LogFileName: fileName,
		LogFilePath: config.GLogConfig.LogConfig.LogFilePath,
		Suffix:      config.GLogConfig.LogConfig.Suffix,
		LogLevel:    logLevel,
		IsFormat:    true,
		PrLogMaxLen: config.GLogConfig.LogConfig.LineSize,
		MaxSize:     config.GLogConfig.LogConfig.MaxSize,
	}
	myLogger = NewLogrus(logConf)
	myLogger.InitColor()

	// 接口上报
	logConf.Suffix = config.GLogConfig.ReportConfig.Suffix
	logConf.LogLevel = DebugLevel
	logConf.IsFormat = false
	log.Printf("logConf: %+v", logConf)
	myLoggerReport = NewLogrus(logConf)

	// Zipkin上报
	logConf.LogFilePath = config.GLogConfig.ZipkinConfig.LogFilePath
	logConf.Suffix = config.GLogConfig.ZipkinConfig.Suffix
	logConf.LogLevel = DebugLevel
	logConf.IsFormat = false
	myLoggerZipkin = NewLogrus(logConf)
}

func ExitLogger() {
	myLoggerReport.Close()
	myLoggerZipkin.Close()
	myLogger.Close()
}

// 打印 Debug 日志
func PrintDebug(format string, v ...interface{}) {
	PrintDebugCalldepth(knownFrames, format, v...)
}

func PrintDebugCalldepth(calldepth int, format string, v ...interface{}) {
	var strLog string
	if len(v) > 0 {
		strLog = fmt.Sprintf(format, v...)
	} else {
		strLog = format
	}

	funcName, fileName, lineNo := myLogger.GetContextInfo(calldepth)
	myLogger.WithFields(logrus.Fields{
		"funcName": funcName,
		"fileName": fileName,
		"lineNo":   lineNo,
	}).Debug(strLog)
}

// 打印 Info 日志
func PrintInfo(format string, v ...interface{}) {
	PrintInfoCalldepth(knownFrames, format, v...)
}

func PrintInfoCalldepth(calldepth int, format string, v ...interface{}) {
	var strLog string
	if len(v) > 0 {
		strLog = fmt.Sprintf(format, v...)
	} else {
		strLog = format
	}

	if myLogger == nil {
		log.Println(strLog)
		return
	}

	funcName, fileName, lineNo := myLogger.GetContextInfo(calldepth)
	myLogger.WithFields(logrus.Fields{
		"funcName": funcName,
		"fileName": fileName,
		"lineNo":   lineNo,
	}).Info(strLog)
}

// 打印 Error 日志
func PrintError(format string, v ...interface{}) {
	PrintErrorCalldepth(knownFrames, format, v...)
}

func PrintErrorCalldepth(calldepth int, format string, v ...interface{}) {
	var strLog string
	if len(v) > 0 {
		strLog = fmt.Sprintf(format, v...)
	} else {
		strLog = format
	}

	if myLogger == nil {
		log.Println(strLog)
		return
	}

	funcName, fileName, lineNo := myLogger.GetContextInfo(calldepth)
	myLogger.WithFields(logrus.Fields{
		"funcName": funcName,
		"fileName": fileName,
		"lineNo":   lineNo,
	}).Error(strLog)
}

// 打印 Panic 日志，并抛出 Panic
func PrintPanic(format string, v ...interface{}) {
	var strLog string
	if len(v) > 0 {
		strLog = fmt.Sprintf(format, v...)
	} else {
		strLog = format
	}

	if myLogger == nil {
		log.Panic(strLog)
		return
	}

	funcName, fileName, lineNo := myLogger.GetContextInfo(knownFrames)
	myLogger.WithFields(logrus.Fields{
		"funcName": funcName,
		"fileName": fileName,
		"lineNo":   lineNo,
	}).Panic(strLog)
}

// 打印 Report 日志
func PrintReport(callerName, calleeName, calleeNode, methods string, errCode int, timeConsume float64) {
	if myLoggerReport == nil {
		log.Println("PrintReport() <nil>")
		return
	}

	myLoggerReport.Info(fmt.Sprintf("1|%d|%s|%s|%s|%s|%s|%d|%0.3f",
		time.Now().UTC().Unix(), callerName, hostname,
		calleeName, calleeNode, methods, errCode, timeConsume))
}

// 打印 Report 日志
func PrintReportByTime(calleeName, calleeNode, methods string, errCode int, startTimeNano int64) {
	nowTime := time.Now().UTC().UnixNano()
	diffTime := (float64(nowTime) - float64(startTimeNano)) / float64(time.Millisecond)
	PrintReport(config.GetServerName(), calleeName, calleeNode, methods, errCode, diffTime)
}

// 打印 Zipkin 日志
func PrintZipkin(s string) {
	if myLoggerZipkin == nil {
		log.Println("PrintZipkin() <nil>")
		return
	}

	myLoggerZipkin.Info(s)
}

// 打印各种类型数据
func PrinfInterface(v interface{}) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				PrintInfo("PrinfInterface() 捕获到了panic产生的异常: %v", err)
			}
		}()

		val := reflect.ValueOf(v) //获取reflect.Type类型
		switch val.Kind() {
		case reflect.Struct:
			printStruct(v)
		default:
			PrintInfo("%+v", v)
		}
	}()
}

/*
私有函数
*/
func printValue(val reflect.Value, typ reflect.StructField) {
	switch val.Kind() {
	case reflect.Struct:
		for i := 0; i < val.NumField(); i++ {
			printValue(val.Field(i), typ.Type.Field(i))
		}
	case reflect.Bool:
		fallthrough
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fallthrough
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		fallthrough
	case reflect.Float32, reflect.Float64:
		fallthrough
	case reflect.String:
		fallthrough
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice, reflect.UnsafePointer:
		fallthrough
	case reflect.Array:
		fallthrough
	default:
		PrintInfo("%s\t%+v", typ.Name, val.Interface())
	}
}

func printStruct(a interface{}) {
	typ := reflect.TypeOf(a)
	val := reflect.ValueOf(a) //获取reflect.Type类型

	kd := val.Kind() //获取到a对应的类别
	if kd != reflect.Struct {
		return
	}

	//遍历结构体的所有字段
	for i := 0; i < val.NumField(); i++ {
		printValue(val.Field(i), typ.Field(i))
	}

	//获取到该结构体有多少个方法
	PrintInfo("struct has %d methods", val.NumMethod())
}
