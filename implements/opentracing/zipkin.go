package opentracing

import (
	"errors"
	"fmt"
	"github.com/patrickmn/go-cache"
	"go-frame/implements/http"
	"go-frame/implements/toolkit"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	defZipkin      Zipkin
	logZipkinCache = cache.New(10*time.Minute, time.Minute)
)

type Zipkin struct {
	name         string
	serverName   string
	traceId      string
	parentSpanId string
	ipv4         string
	span         *Span
	childSpans   []*Span
	mutex        sync.Mutex
}

func ZipkinInit(name string) {
	ip, err := http.ExternalIP()
	if err != nil {
		ip = "127.0.0.1"
	}
	defZipkin.ipv4 = ip

	defZipkin.serverName = name
	if defZipkin.serverName == "" {
		defZipkin.serverName = "unknown"
	}
}

func NewZipkin(logid string) *Zipkin {
	if z, ok := logZipkinCache.Get(logid); ok {
		return z.(*Zipkin)
	}

	zipkin := defZipkin
	logZipkinCache.Set(logid, &zipkin, time.Minute*3)
	return &zipkin
}

func DelZipkin(logid string) {
	logZipkinCache.Delete(logid)
}

func (z *Zipkin) Clear() {
	if z.span != nil {
		z.span = nil
	}

	if len(z.childSpans) > 0 {
		z.childSpans = make([]*Span, 0, 0)
	}

	z.traceId = ""
	z.parentSpanId = ""
}

func (z *Zipkin) SetName(name string) {
	z.name = name
	if z.span != nil {
		z.span.spanName = name
	}
}

func (z *Zipkin) SetServerName(name string) {
	z.serverName = name
}

func (z *Zipkin) SetTag(k, v string) error {
	if z.span == nil {
		return errors.New("span nil")
	}

	z.mutex.Lock()
	defer z.mutex.Unlock()

	z.span.SetTag(k, v)
	return nil
}

func (z *Zipkin) SetChildTag(k, v string, index int) error {
	if index >= len(z.childSpans) {
		return errors.New("index out of range")
	}

	if index < 0 {
		if index = len(z.childSpans) - 1; index < 0 {
			return errors.New("child index out of range")
		}
	}

	z.mutex.Lock()
	defer z.mutex.Unlock()

	z.childSpans[index].SetTag(k, v)
	return nil
}

func (z *Zipkin) SetFromContext(context string) {
	m := make(map[string]string)
	infoSplit := strings.Split(context, ",")
	for _, split := range infoSplit {
		nodeSplit := strings.Split(split, ":")
		if len(nodeSplit) == 2 {
			m[nodeSplit[0]] = nodeSplit[1]
		}
	}

	z.Clear()
	z.span = NewSpan(z.name)

	if strTraceId, ok := m[TraceIdKeyName]; ok {
		z.traceId = strTraceId
	} else {
		z.traceId = toolkit.RandomHexadecimal()
	}

	if strSpanId, ok := m[SpanIdKeyName]; ok {
		z.parentSpanId = strSpanId
	}

	iSampled := -1
	if strSampled, ok := m[SampledKeyName]; ok {
		if sampled, err := strconv.Atoi(strSampled); err == nil {
			iSampled = sampled
		}
	}
	z.span.sampled = uint8(iSampled)
}

func (z *Zipkin) GetChildSpanContextString(index int) (string, error) {
	if index >= len(z.childSpans) {
		return "", errors.New("index out of range")
	}

	if index < 0 {
		if index = len(z.childSpans) - 1; index < 0 {
			return "", errors.New("child index out of range")
		}
	}

	span := z.childSpans[index]
	return fmt.Sprintf("%s,x-b3-traceid:%s,x-b3-parentspanid:%s", span.GetContextStr(), z.traceId, z.parentSpanId), nil
}

func (z *Zipkin) StartChildSpan(name string) (int, error) {
	if z.span == nil {
		return -1, errors.New("span nil")
	}

	z.mutex.Lock()
	defer z.mutex.Unlock()

	z.childSpans = append(z.childSpans, NewSpan(name))
	index := len(z.childSpans) - 1
	return index, nil
}

func (z *Zipkin) EndChildSpan(index int) error {
	if index >= len(z.childSpans) {
		return errors.New("index out of range")
	}

	if index < 0 {
		if index = len(z.childSpans) - 1; index < 0 {
			return errors.New("child index out of range")
		}
	}

	z.childSpans[index].Finish()
	return nil
}

func (z *Zipkin) EndChildSpanByDuration(index int, duration int64) error {
	if index >= len(z.childSpans) {
		return errors.New("index out of range")
	}

	if index < 0 {
		if index = len(z.childSpans) - 1; index < 0 {
			return errors.New("child index out of range")
		}
	}

	z.childSpans[index].FinishByDuration(duration)
	return nil
}

func (z *Zipkin) Dump() ([]map[string]interface{}, error) {
	if z.span == nil {
		return nil, errors.New("span nil")
	}

	z.span.Finish()

	var mapList []map[string]interface{}
	mapList = append(mapList, map[string]interface{}{
		"traceId":  z.traceId,
		"parentId": z.parentSpanId,
		"localEndpoint": map[string]string{
			"serviceName": z.serverName,
			"ipv4":        z.ipv4,
		},
		"kind": "SERVER",

		"id":        z.span.spanId,
		"name":      z.span.spanName,
		"timestamp": z.span.startUpTime,
		"duration":  z.span.duration,
		"tags":      z.span.tags,
	})

	for _, cp := range z.childSpans {
		mapList = append(mapList, map[string]interface{}{
			"traceId":  z.traceId,
			"parentId": z.span.spanId,
			"localEndpoint": map[string]string{
				"serviceName": z.serverName,
				"ipv4":        z.ipv4,
			},
			"kind": "CLIENT",

			"id":        cp.spanId,
			"name":      cp.spanName,
			"timestamp": cp.startUpTime,
			"duration":  cp.duration,
			"tags":      cp.tags,
		})
	}

	return mapList, nil
}
