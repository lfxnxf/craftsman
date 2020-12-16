package middleware

import (
	"context"
	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/propagation"
	"github.com/SkyAPM/go2sky/reporter/grpc/common"
	"github.com/gin-gonic/gin"
	"github.com/lfxnxf/craftsman/log"
	"strconv"
	"time"
)

const (
	TraceIdKey        = "trace_id"
	PropagationHeader = "propagation_header"
)

func trace(tracer *go2sky.Tracer, logger log.Logger) gin.HandlerFunc {

	return func(c *gin.Context) {
		operationName := c.Request.Method + ":" + c.Request.URL.Path
		span, ctx, err := tracer.CreateEntrySpan(c.Request.Context(), operationName, func() (string, error) {
			traceId := c.Request.Header.Get(propagation.Header)

			return traceId, nil
		})

		if err != nil {
			logger.ErrorT(c.Request.Context(), "tracer init err", "err", err.Error())
			c.Next()
			return
		}

		//span.SetComponent(int32(inits.LocalIpInt))
		span.Tag(go2sky.TagHTTPMethod, c.Request.Method)
		span.Tag(go2sky.TagURL, c.Request.Host+c.Request.URL.Path)
		span.SetSpanLayer(common.SpanLayer_Http)

		traceId := go2sky.TraceID(ctx)
		ctx = context.WithValue(ctx, TraceIdKey, traceId)
		ctx = context.WithValue(ctx, PropagationHeader, c.Request.Header.Get(propagation.Header))
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		if len(c.Errors) > 0 {
			span.Error(time.Now(), c.Errors.String())
			//logger.ErrorT(c.Request.Context(), c.Errors.String())
		}

		span.Tag(go2sky.TagStatusCode, strconv.Itoa(c.Writer.Status()))
		span.End()
	}
}
