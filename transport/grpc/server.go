package grpc

import (
	"context"
	"fmt"
	"github.com/SkyAPM/go2sky"
	"github.com/lfxnxf/craftsman/log/zap"
	"github.com/lfxnxf/craftsman/tracing/sky"
	"google.golang.org/grpc/metadata"

	"github.com/go-kit/kit/endpoint"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	"github.com/lfxnxf/craftsman/tracing"
)

func NewGRPCHandler(endpoint endpoint.Endpoint) *grpctransport.Server {
	var options []grpctransport.ServerOption

	tracer, err := tracing.GetTracer()
	if err != nil {
		log.Error("tracer", "err", err)
	} else {
		if tracer != nil {
			fmt.Println("9999999999999")
			options = append(options, sky.GRPCServerTrace(tracer))
		}
	}

	beforeFunc := func(ctx context.Context, md metadata.MD) context.Context {
		ctx = context.WithValue(ctx, "trace_id", go2sky.TraceID(ctx))

		fmt.Println("mmmmm",(md)["grpc-method"])
		if len((md)["grpc-service"]) >0 {
			ctx = context.WithValue(ctx, "grpc-service", (md)["grpc-service"][0])
		}

		if len((md)["grpc-method"]) >0 {
			fmt.Println("mmmmm",3333)
			ctx = context.WithValue(ctx, "grpc-method", (md)["grpc-method"][0])
		}
		return ctx
	}

	afterFunc := func(ctx context.Context, _ *metadata.MD, _ *metadata.MD) context.Context {
		return ctx
	}

	return grpctransport.NewServer(
		endpoint,
		decodeRequest,
		encodeResponse,
		append(options,
			grpctransport.ServerBefore(beforeFunc),
			grpctransport.ServerAfter(afterFunc), )...,
	)
}

func decodeRequest(ctx context.Context, req interface{}) (interface{}, error) {
	return req, nil
}

func encodeResponse(ctx context.Context, resp interface{}) (interface{}, error) {
	return resp, nil
}
