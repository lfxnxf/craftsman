package zap

import (
	"golang.org/x/net/context"
)

const (
	traceIDKey = "trace_id"
)

func traceID(ctx context.Context) string {
	traceId := ctx.Value(traceIDKey)
	if traceId, ok := traceId.(string); ok {
		return traceId
	}
	return ""
}
