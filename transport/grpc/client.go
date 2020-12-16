package grpc

import (
	"context"
	"time"

	"github.com/SkyAPM/go2sky"
	"github.com/lfxnxf/craftsman/log"
	"github.com/lfxnxf/craftsman/tracing/sky"

	"github.com/SkyAPM/go2sky/propagation"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func NewGRPCClient(ctx context.Context, log log.Logger, tracer *go2sky.Tracer, endpoint,
	serviceName, alisService, method string, reply interface{}) (*grpctransport.Client, *grpc.ClientConn, error) {

	var options []grpctransport.ClientOption

	if tracer != nil {
		options = append(options, sky.GRPCClientTrace(tracer))
	}

	conn, err := grpc.Dial(endpoint, grpc.WithInsecure(), grpc.WithTimeout(time.Second))
	if err != nil {
		return nil, nil, err
	}
	//defer conn.Close()

	beforeFunc := func(ctx context.Context, md *metadata.MD) context.Context {
		if tracer != nil {
			span, err := tracer.CreateExitSpan(ctx, method, endpoint, func(header string) error {
				proHeader := ctx.Value("propagation_header")
				proHeaders, ok := proHeader.(string)
				if ok {
					(*md)[propagation.Header] = []string{proHeaders}
				}
				(*md)["grpc-service"] = []string{serviceName}
				(*md)["grpc-method"] = []string{method}
				return nil
			})
			if err != nil {
				log.ErrorT(ctx, "create exit span", "err", err)
				return ctx
			}
			span.SetComponent(0)
			span.End()
		}
		return ctx
	}

	afterFunc := func(ctx context.Context, _ metadata.MD, _ metadata.MD) context.Context {
		return ctx
	}
	client := grpctransport.NewClient(
		conn,
		alisService,
		method,
		decodeRequest,
		encodeResponse,
		reply,
		append(options,
			grpctransport.ClientBefore(beforeFunc),
			grpctransport.ClientAfter(afterFunc))...,
	)

	return client, conn, nil
}
