package sky

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/SkyAPM/go2sky"
	"google.golang.org/grpc/metadata"
)

type Tag string

const (
	TagURL             Tag = "url"
	TagStatusCode      Tag = "status_code"
	TagHTTPMethod      Tag = "http.method"
	TagHTTPClient      Tag = "http.client"
	TagHTTPServer      Tag = "http.server"
	TagGRPCClient      Tag = "grpc.client"
	TagGRPCServer      Tag = "grpc.server"
	TagDBType          Tag = "db.type"
	TagDBInstance      Tag = "db.instance"
	TagDBStatement     Tag = "db.statement"
	TagDBBindVariables Tag = "db.bind_vars"
	TagMQQueue         Tag = "mq.queue"
	TagMQBroker        Tag = "mq.broker"
	TagMQTopic         Tag = "mq.topic"
)

const (
	TraceID      = "x-b3-traceid"
	SpanID       = "x-b3-spanid"
	ParentSpanID = "x-b3-parentspanid"
	Sampled      = "x-b3-sampled"
	Flags        = "x-b3-flags"
	Context      = "b3"
)

// Extractor function signature
type Extractor func() (*go2sky.SegmentContext, error)

// Injector function signature
type Injector func(go2sky.SegmentContext) error

// ExtractGRPC will extract a span.Context from the gRPC Request metadata if
// found in B3 header format.
func ExtractGRPC(md *metadata.MD) Extractor {
	return func() (*go2sky.SegmentContext, error) {
		var (
			traceIDHeader      = GetGRPCHeader(md, TraceID)
			spanIDHeader       = GetGRPCHeader(md, SpanID)
			parentSpanIDHeader = GetGRPCHeader(md, ParentSpanID)
		)
		_spanIDHeader, _ := strconv.Atoi(spanIDHeader)
		_parentSpanIDHeader, _ := strconv.Atoi(parentSpanIDHeader)

		return &go2sky.SegmentContext{
			TraceID:      traceIDHeader,
			SpanID:       int32(_spanIDHeader),
			ParentSpanID: int32(_parentSpanIDHeader),
		}, nil
	}
}

// InjectGRPC will inject a span.Context into gRPC metadata.
func InjectGRPC(md *metadata.MD) Injector {
	return func(sc go2sky.SegmentContext) error {
		setGRPCHeader(md, TraceID, sc.TraceID)
		setGRPCHeader(md, SpanID, strconv.Itoa(int(sc.SpanID)))
		setGRPCHeader(md, ParentSpanID, strconv.Itoa(int(sc.ParentSpanID)))
		return nil
	}
}

// GetGRPCHeader retrieves the last value found for a particular key. If key is
// not found it returns an empty string.
func GetGRPCHeader(md *metadata.MD, key string) string {
	v := (*md)[key]
	if len(v) < 1 {
		return ""
	}
	return v[len(v)-1]
}

func setGRPCHeader(md *metadata.MD, key, value string) {
	(*md)[key] = append((*md)[key], value)
}

func traceIDString(traceId []int64) string {
	ii := make([]string, len(traceId))
	for i, v := range traceId {
		ii[i] = fmt.Sprint(v)
	}
	return strings.Join(ii, ".")
}

func traceIDInt64(traceId string) []int64 {
	var traceIds []int64
	_traceIds := strings.Split(traceId, ".")
	for _, traceId := range _traceIds {
		id, _ := strconv.ParseInt(traceId, 10, 64)
		traceIds = append(traceIds, id)
	}
	return traceIds
}

//func traceIDStringByString(traceId string) string {
//	var traceIds []int64
//	_traceIds := strings.Split(traceId, ".")
//	for _, traceId := range _traceIds {
//		id, _ := strconv.ParseInt(traceId, 10, 64)
//		traceIds = append(traceIds, id)
//	}
//	return traceIds
//}