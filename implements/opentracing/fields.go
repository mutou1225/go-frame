package opentracing

const (
	TraceIdKeyName      = "x-b3-traceid"
	SpanIdKeyName       = "x-b3-spanid"
	ParentSpanIdKeyName = "x-b3-parentspanid"
	SampledKeyName      = "x-b3-sampled"
	FlagsKeyName        = "x-b3-flags"
)

type contextKey uint

const (
	TraceHeader = "Opentracer-Info"

	contextLogID contextKey = iota
)


