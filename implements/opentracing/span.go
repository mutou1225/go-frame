package opentracing

import (
	"fmt"
	"github.com/mutou1225/go-frame/implements/toolkit"
	"time"
)

type Span struct {
	spanId            string
	spanName          string
	duration          int64
	startUpTime       int64
	startCalendarTime int64
	sampled           uint8
	tags              map[string]string
}

func NewSpan(name string) *Span {
	return &Span{
		toolkit.RandomHexadecimal(),
		name,
		0,
		time.Now().UTC().UnixNano()/1000,
		time.Now().UTC().UnixNano()/1000,
		1,
		make(map[string]string),
	}
}

func (s *Span) SetSpanName(name string) {
	s.spanName = name
}

func (s *Span) SetTag(k, v string) {
	s.tags[k] = v
}

func (s *Span) SetSampled(sampled uint8) {
	s.sampled = sampled
}

func (s *Span) GetSpanId() string {
	return s.spanId
}

func (s *Span) GetSampled() uint8 {
	return s.sampled
}

func (s *Span) GetDuration() int64 {
	return time.Now().UTC().UnixNano()/1000 - s.startUpTime
}

func (s *Span) Finish() {
	s.duration = time.Now().UTC().UnixNano()/1000 - s.startUpTime
}

func (s *Span) FinishByDuration(duration int64) {
	s.duration = duration
}

func (s *Span) GetContextStr() string {
	return fmt.Sprintf("x-b3-spanid:%s,x-b3-sampled:%d,x-b3-flags:0", s.spanId, s.sampled)
}
