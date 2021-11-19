package logger

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"testing"
)

func newTestLogger() *MyLogger {
	logConf := LogrusConfig{
		ProgramName: "LoggerTest",
		LogFileName: "LoggerTest",
		LogFilePath: "/tmp/",
		Suffix:      "log",
		LogLevel:    DebugLevel,
		IsFormat:    true,
		PrLogMaxLen: 10240,
		MaxSize:     1048576000,
	}
	testLogger := NewLogrus(logConf)
	return testLogger
}

func closeTestLogger(l *MyLogger) {
	l.Close()
}

func Test_LoggerPrint(t *testing.T) {
	testLogger := newTestLogger()
	if testLogger == nil {
		t.Error("newTestLogger() Err.")
	}
	defer closeTestLogger(testLogger)

	testLogger.Debug("PrintDebug")
	testLogger.Info("PrintInfo")
	testLogger.Error("PrintError")
}

func Benchmark_Prinf(b *testing.B) {
	testLogger := newTestLogger()
	if testLogger == nil {
		b.Error("newTestLogger() Err.")
	}
	defer closeTestLogger(testLogger)

	for i := 0; i < b.N; i++ {
		funcName, fileName, lineNo := testLogger.GetContextInfo(1)
		testLogger.WithFields(logrus.Fields{
			"funcName": funcName,
			"fileName": fileName,
			"lineNo":   lineNo,
		}).Debug(fmt.Sprintf("PrintDebug: %d", i))

		/*
		testLogger.WithFields(logrus.Fields{
			"funcName": funcName,
			"fileName": fileName,
			"lineNo":   lineNo,
		}).Info(fmt.Sprintf("PrintDebug: %d", i))

		testLogger.WithFields(logrus.Fields{
			"funcName": funcName,
			"fileName": fileName,
			"lineNo":   lineNo,
		}).Error(fmt.Sprintf("PrintDebug: %d", i))
		*/
	}
}
