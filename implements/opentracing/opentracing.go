package opentracing

import (
	"bytes"
	"errors"
	"eva_services_go/logger"
	jsoniter "github.com/json-iterator/go"
	"runtime"
)

var (
	gOpenTracing *OpenTracing
	//limiterSet   = cache.New(5*time.Minute, time.Minute)
)

type OpenTracing struct {
	ServerName string
}

func NewOpenTracing(name string) *OpenTracing {
	ZipkinInit(name)
	gOpenTracing = &OpenTracing{name}
	return gOpenTracing
}

func GetOpenTracing() *OpenTracing {
	if gOpenTracing == nil {
		NewOpenTracing("unknown")
	}
	return gOpenTracing
}

func (ot *OpenTracing) SetServerName(name string) {
	ZipkinInit(name)
	ot.ServerName = name
}

func (ot *OpenTracing) getGoroutineId() string {
	b := make([]byte, 64)
	runtime.Stack(b, false)
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	return string(b)
}

func (ot *OpenTracing) FromContextSetName(strContext, name string) {
	logId := ot.getGoroutineId()
	if logId == "" {
		return
	}

	zipkin := NewZipkin(logId)
	zipkin.SetFromContext(strContext)
	zipkin.SetName(name)
}

func (ot *OpenTracing) SetTag(k, v string) {
	logId := ot.getGoroutineId()
	if logId == "" {
		return
	}
	NewZipkin(logId).SetTag(k, v)
}

func (ot *OpenTracing) Dump() {
	logId := ot.getGoroutineId()
	if logId == "" {
		return
	}

	zipkin := NewZipkin(logId)
	defer DelZipkin(logId)

	if m, err := zipkin.Dump(); err == nil {
		var json = jsoniter.ConfigCompatibleWithStandardLibrary
		jsonBytes, err := json.Marshal(m)
		if err == nil {
			logger.PrintZipkin(string(jsonBytes))
		}
	}
}

// 返回Child Span的index，错误时index为小于0
func (ot *OpenTracing) StartChildSpan(name string) (int, error) {
	logId := ot.getGoroutineId()
	if logId == "" {
		return -1, errors.New("OpenTracing logId empty")
	}

	return NewZipkin(logId).StartChildSpan(name)
}

func (ot *OpenTracing) SetChildTag(index int, k, v string) {
	logId := ot.getGoroutineId()
	if logId == "" {
		return
	}

	if index < 0 {
		return
	}

	NewZipkin(logId).SetChildTag(k, v, index)
}

func (ot *OpenTracing) GetChildSpanContext(index int) (string, error) {
	logId := ot.getGoroutineId()
	if logId == "" {
		return "", nil
	}

	if index < 0 {
		return "", nil
	}

	return NewZipkin(logId).GetChildSpanContextString(index)
}

func (ot *OpenTracing) EndChildSpan(index int) {
	logId := ot.getGoroutineId()
	if logId == "" {
		return
	}

	if index < 0 {
		return
	}

	NewZipkin(logId).EndChildSpan(index)
}

func (ot *OpenTracing) EndChildSpanByDuration(index int, duration int64) {
	logId := ot.getGoroutineId()
	if logId == "" {
		return
	}

	if index < 0 {
		return
	}

	NewZipkin(logId).EndChildSpanByDuration(index, duration)
}
