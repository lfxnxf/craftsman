package http

import (
	"context"
	"fmt"
	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/propagation"
	"net/http"

	"github.com/lfxnxf/craftsman/log/zap"
	"github.com/lfxnxf/craftsman/tracing"
	"github.com/lfxnxf/craftsman/tracing/sky"

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
)

func NewHTTPHandler(urlsHandle map[string]endpoint.Endpoint) http.Handler {
	var options []httptransport.ServerOption

	tracer, err := tracing.GetTracer()
	if err != nil {
		log.Error("tracer", "err", err)
	} else {
		fmt.Println("tracer", tracer)
		if tracer != nil {
			options = append(options, sky.HTTPServerTrace(tracer))
		}
	}

	beforeFunc := func(ctx context.Context, r *http.Request) context.Context {
		if tracer != nil {
			span, err := tracer.CreateExitSpan(ctx, r.URL.Path, r.Host, func(header string) error {
				r.Header.Set(propagation.Header, header)
				ctx = context.WithValue(ctx, "propagation_header", header)
				ctx = context.WithValue(ctx, "trace_id", go2sky.TraceID(ctx))
				return nil
			})
			if err != nil {
				log.ErrorJ(ctx, "tracer", "err", err)
				return ctx
			}
			span.SetComponent(0)
			span.End()
		}
		return ctx
	}

	afterFunc := func(ctx context.Context, w http.ResponseWriter) context.Context {

		return ctx
	}

	m := http.NewServeMux()

	for pattern, e := range urlsHandle {
		fmt.Println("pattern")
		m.Handle(pattern, httptransport.NewServer(
			e,
			decodeRequest,
			encodeResponse,
			append(options,
				httptransport.ServerBefore(beforeFunc),
				httptransport.ServerAfter(afterFunc))...,
		))
	}

	return m
}
