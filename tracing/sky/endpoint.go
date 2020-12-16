package sky

import (
	"context"
	"github.com/SkyAPM/go2sky"
	"github.com/go-kit/kit/endpoint"
)

func TraceServer(tracer *go2sky.Tracer, operationName string) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			// TODO
			return next(ctx, request)
		}
	}
}

func TraceClient(tracer *go2sky.Tracer, operationName string) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			// TODO
			return next(ctx, request)
		}
	}
}
