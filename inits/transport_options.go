package inits

import (
	"context"
	"github.com/SkyAPM/go2sky"
	"github.com/go-kit/kit/transport"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/lfxnxf/craftsman/log"
	"github.com/lfxnxf/craftsman/tracing/sky"
	"net/http"
)

func DefaultHttpOptions(logger log.Logger, tracer *go2sky.Tracer) []httptransport.ServerOption {
	var optionsHttp []httptransport.ServerOption
	optionsHttp = append(optionsHttp,
		httptransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		httptransport.ServerBefore(func(ctx context.Context, request *http.Request) context.Context {
			ctx = context.WithValue(ctx, APPUserIdKey, request.Header.Get(APPUserIdKey))
			ctx = context.WithValue(ctx, GatewayIpKey, request.Header.Get(GatewayIpKey))
			ctx = context.WithValue(ctx, APPWidthKey, request.Header.Get(APPWidthKey))
			ctx = context.WithValue(ctx, APPHeightKey, request.Header.Get(APPHeightKey))
			ctx = context.WithValue(ctx, APPVersionKey, request.Header.Get(APPVersionKey))
			ctx = context.WithValue(ctx, APPChannelKey, request.Header.Get(APPChannelKey))
			ctx = context.WithValue(ctx, APPModelKey, request.Header.Get(APPModelKey))
			ctx = context.WithValue(ctx, APPNameKey, request.Header.Get(APPNameKey))
			ctx = context.WithValue(ctx, APPIMEIKey, request.Header.Get(APPIMEIKey))
			return ctx
		}))

	if tracer != nil {
		optionsHttp = append(optionsHttp, sky.HTTPServerTrace(tracer))
	}

	return optionsHttp
}

func DefaultGrpcOptions(logger log.Logger, tracer *go2sky.Tracer) []grpctransport.ServerOption {
	var optionsGrpc []grpctransport.ServerOption
	optionsGrpc = append(optionsGrpc, grpctransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)))
	if tracer != nil {
		optionsGrpc = append(optionsGrpc, sky.GRPCServerTrace(tracer))
	}

	return optionsGrpc
}
