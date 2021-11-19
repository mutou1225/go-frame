package logger

import (
	"bytes"
	"crypto/md5"
	"eva_services_go/implements/http"
	"fmt"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	PanicLevel = logrus.PanicLevel
	FatalLevel = logrus.FatalLevel
	ErrorLevel = logrus.ErrorLevel
	WarnLevel  = logrus.WarnLevel
	InfoLevel  = logrus.InfoLevel
	DebugLevel = logrus.DebugLevel

	ltLayout    = "20060102-150405"
	ltLayoutDay = "2006-01-02"
)

type LogrusConfig struct {
	ProgramName string       // 应用名，用于日志目录
	LogFileName string       // 日志文件名
	LogFilePath string       // 日志文件路径
	Suffix      string       // 日志后缀
	LogLevel    logrus.Level // 日志打印级别
	IsFormat    bool         // 是否格式化输出
	PrLogMaxLen int          // 打印日志最大长度
	MaxSize     int64        // 日志文件最大长度
	ipAddr      string
	logIndex    int
}

type MyLogger struct {
	*logrus.Logger
	fileHandle *os.File
}

func NewLogrus(logConf LogrusConfig) *MyLogger {
	lConfig := logConf
	if lConfig.ProgramName == "" {
		lConfig.ProgramName = "unknown"
	}

	if lConfig.LogFileName == "" {
		lConfig.LogFileName = lConfig.ProgramName
	}

	if lConfig.LogFilePath == "" {
		lConfig.LogFilePath = "/data/log/cgi_server"
	} else {
		lConfig.LogFilePath = path.Clean(lConfig.LogFilePath)
	}

	if lConfig.LogLevel == 0 {
		lConfig.LogLevel = logrus.DebugLevel
	}

	ip, err := http.ExternalIP()
	if err != nil {
		ip = "127.0.0.1"
	}
	lConfig.ipAddr = ip

	newLogrus := logrus.New()
	newLogrus.SetFormatter(&lConfig)
	newLogrus.SetLevel(lConfig.LogLevel)
	newLogrus.ExitFunc = func(i int) {}
	newLogger := &MyLogger{newLogrus, nil}

	if err := os.MkdirAll(fmt.Sprintf("%s/%s", lConfig.LogFilePath, lConfig.ProgramName), os.ModePerm); err != nil {
		log.Printf("~~~~~~ os.MkdirAll() Err: %s", err.Error())
		newLogrus.SetOutput(os.Stdout)
		return newLogger
	}

	filePath := lConfig.getLogFilePath()
	fileHandle, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		log.Printf("~~~~~~ os.OpenFile(%s) Err: %s", filePath, err.Error())
		newLogrus.SetOutput(os.Stdout)
		return newLogger
	} else {
		newLogrus.SetOutput(fileHandle)
		newLogger.fileHandle = fileHandle
		log.Printf("~~~~~~ logger: %s", filePath)
	}

	// 启动日志回滚
	go lConfig.rotateLogs(newLogger, filePath)

	return newLogger
}

func (l *MyLogger) Close() {
	//l.Close()

	if l.fileHandle != nil {
		_ = l.fileHandle.Close()
		l.fileHandle = nil
	}
}

// 获取日志打印内容的文件信息
func (l *MyLogger) GetContextInfo(calldepth int) (string, string, int) {
	pc, file, line, ok := runtime.Caller(calldepth)
	if !ok {
		// 尝试
		for calldepth = calldepth - 1; calldepth >= 0; calldepth-- {
			pc, file, line, ok = runtime.Caller(calldepth)
			if ok {
				break
			}
		}

		if !ok {
			return "unknown", "unknown", 0
		}
	}
	_, filename := path.Split(file)
	defunc := runtime.FuncForPC(pc).Name()

	// 需要全路径则修改这里
	funcList := strings.Split(defunc, "/")
	index := len(funcList) - 3 // 返回后3层的路径
	if index < 0 {
		index = 0
	}
	defunc = path.Join(funcList[index:]...)

	return defunc, filename, line
}

// 获取日志打印内容前缀
func (f *LogrusConfig) getLogDatePrefix(curTime time.Time) string {
	milliSecond := curTime.UTC().UnixNano() / int64(time.Millisecond)
	leftMs := milliSecond % 1000
	dateStr := curTime.Format(ltLayout)
	return fmt.Sprintf("%s-%d", dateStr, leftMs)
}

func (f *LogrusConfig) getGoroutineId() uint64 {
	b := make([]byte, 64)
	runtime.Stack(b, false)
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

func (f *LogrusConfig) getSessionId(goid uint64) string {
	digestBytes := md5.Sum([]byte(fmt.Sprintf("%s%d", f.ipAddr, goid)))
	md5Str := fmt.Sprintf("%x", digestBytes)
	return md5Str[0:16]
}

func (f *LogrusConfig) Format(entry *logrus.Entry) ([]byte, error) {
	if f.PrLogMaxLen != 0 && len(entry.Message) > f.PrLogMaxLen {
		entry.Message = fmt.Sprintf("%s......\n", entry.Message[0:f.PrLogMaxLen])
	}

	if !f.IsFormat {
		return []byte(fmt.Sprintf("%s\n", entry.Message)), nil
	}

	goid := f.getGoroutineId()
	logData := fmt.Sprintf("%s|%s|%d|%s|%d|%s|%s|%v|%v|%v|%s\n",
		f.getLogDatePrefix(entry.Time),
		strings.ToUpper(entry.Level.String()),
		goid,
		f.getSessionId(goid),
		os.Getpid(),
		f.ipAddr,
		f.ProgramName,
		entry.Data["fileName"],
		entry.Data["funcName"],
		entry.Data["lineNo"],
		entry.Message)

	return []byte(logData), nil
}

func (f *LogrusConfig) getFileSize(path string) int64 {
	fileInfo, err := os.Stat(path)
	if err == nil {
		return fileInfo.Size()
	}
	if os.IsNotExist(err) {
		return 0
	}
	return 0
}

func (f *LogrusConfig) getLogFilePath() string {
	logSuffix := time.Now().Format(ltLayoutDay)
	strLogPath := os.TempDir() + "/" + f.LogFileName
	for f.logIndex = 0; f.logIndex < 1000; f.logIndex++ {
		strSuffix := ""
		if f.logIndex > 0 {
			strSuffix = fmt.Sprintf("-%02d", f.logIndex)
		}
		strLogPath = fmt.Sprintf("%s/%s/%s.%s.%s%s",
			f.LogFilePath,
			f.ProgramName,
			f.LogFileName,
			f.Suffix,
			logSuffix,
			strSuffix)

		logLen := f.getFileSize(strLogPath)
		if logLen == 0 || f.MaxSize == 0 || logLen < f.MaxSize {
			break
		}
	}
	return strLogPath
}

func (f *LogrusConfig) rotateLogs(newLogger *MyLogger, logFilePath string) {
	if newLogger.fileHandle == nil {
		return
	}

	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		strLogPathNew := f.getLogFilePath()
		if strLogPathNew != logFilePath {
			newHandle, err := os.OpenFile(strLogPathNew, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
			if err == nil {
				logFilePath = strLogPathNew

				newLogger.SetOutput(newHandle)
				_ = newLogger.fileHandle.Close()
				newLogger.fileHandle = newHandle
			}
		}
	}
}
