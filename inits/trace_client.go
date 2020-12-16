package inits

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/propagation"
	"github.com/SkyAPM/go2sky/reporter"
	"github.com/lfxnxf/craftsman/log"
	"github.com/lfxnxf/craftsman/tracing"
	"google.golang.org/grpc/metadata"
)

type TraceClients struct {
	SkyTraces     []*go2sky.Tracer
	SkyTracesConn map[string]go2sky.Reporter
	logger        log.Logger
	BalanceType   string
}

const (
	Random = "random"
)

func (t *TraceClients) GetTracer() (*go2sky.Tracer, error) {
	if len(t.SkyTraces) > 0 {
		var balancer tracing.Balancer
		endpointer := tracing.FixedEndpointer(t.SkyTraces)
		switch t.BalanceType {
		case Random:
			balancer = tracing.NewRandom(endpointer, 100)
		default:
			balancer = tracing.NewRandom(endpointer, 100)
		}

		endpoint, err := balancer.Endpoint()
		if err != nil {
			return nil, err
		}
		return endpoint, nil
	} else {
		return nil, errors.New("tracer no register")
	}
}

func NewSkyTraceClient(log log.Logger, serviceName string, traceConfig tracing.Trace) (*TraceClients, error) {
	endpoints := strings.Split(traceConfig.Ipport, ",")

	traceClients := &TraceClients{
		SkyTracesConn: make(map[string]go2sky.Reporter, len(endpoints)),
		logger:        log,
		BalanceType:   traceConfig.Balancetype,
	}

	for _, endpoint := range endpoints {
		r, err := reporter.NewGRPCReporter(endpoint)
		if err != nil {
			log.Error("new reporter", "endpoint", endpoint, "err", err)
			continue
		}
		traceClients.SkyTracesConn[endpoint] = r
		tracer, err := go2sky.NewTracer(serviceName, go2sky.WithReporter(r))
		if err != nil {
			log.Error("new tracer", "endpoint", endpoint, "err", err)
			continue
		}
		// TODO 待优化
		tracer.WaitUntilRegister()
		traceClients.SkyTraces = append(traceClients.SkyTraces, tracer)
	}

	return traceClients, nil
}

const (
	PropagationHeader = "propagation_header"
	TraceId           = "trace_id"
)

// TODO 待优化
func HttpBeforeFuncCreateExitSpan(ctx context.Context, tracer *go2sky.Tracer, r *http.Request) context.Context {
	if tracer != nil {
		span, err := tracer.CreateExitSpan(ctx, r.URL.Path, r.Host, func(header string) error {
			r.Header.Set(propagation.Header, header)
			ctx = context.WithValue(ctx, tracing.PropagationHeader, header)
			//ctx = context.WithValue(ctx, tracing.TraceId, go2sky.TraceID(ctx))
			ctx = context.WithValue(ctx, tracing.TraceId, r.Header.Get(tracing.TraceId))
			return nil
		})
		if err != nil {
			return ctx
		}
		span.SetComponent(0)
		span.Tag(tracing.TagHTTPClient, "client before func")
		span.End()
	}
	return ctx
}

// TODO 待优化
func GrpcBeforeFuncCreateExitSpan(ctx context.Context, md metadata.MD, tracer *go2sky.Tracer) context.Context {
	if tracer != nil {
		traceId := go2sky.TraceID(ctx)
		if traceId != "" {
			ctx = context.WithValue(ctx, tracing.TraceId, go2sky.TraceID(ctx))

			if len((md)[tracing.TagGRPCServer]) > 0 {
				ctx = context.WithValue(ctx, tracing.TagGRPCServer, (md)[tracing.TagGRPCServer][0])
			}

			if len((md)[tracing.TagGRPCMethod]) > 0 {
				ctx = context.WithValue(ctx, tracing.TagGRPCMethod, (md)[tracing.TagGRPCMethod][0])
			}
		}
	}
	return ctx
}

func healthEndpoints(endpoints []string) {
	for {
		time.Sleep(time.Second * 1)
		for _, endpoint := range endpoints {
			fmt.Println(endpoint)
		}
	}
}
