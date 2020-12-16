package sky

import (
	"context"
	"github.com/SkyAPM/go2sky/propagation"
	"strconv"

	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/reporter/grpc/common"
	"github.com/go-kit/kit/log"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/tiantianjianbao/craftsman/tracing"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func GRPCClientTrace(tracer *go2sky.Tracer, options ...TracerOption) kitgrpc.ClientOption {
	config := tracerOptions{
		tags:      make(map[string]string),
		name:      tracing.TagGRPCClient,
		logger:    log.NewNopLogger(),
		propagate: true,
	}

	for _, option := range options {
		option(&config)
	}

	clientBefore := kitgrpc.ClientBefore(
		func(ctx context.Context, md *metadata.MD) context.Context {
			method, _ := ctx.Value(kitgrpc.ContextKeyRequestMethod).(string)
			span, ctx, err := tracer.CreateEntrySpan(ctx, method, func() (string, error) {
				return "", nil
			})
			if err != nil {
				_ = config.logger.Log("err", err)
				return ctx
			}
			defer span.End()

			span.SetComponent(0)
			span.SetSpanLayer(common.SpanLayer_RPCFramework)
			span.Tag(tracing.TagGRPCClient, "client before")

			//fmt.Println("grpc-client before go2sky.TraceID", go2sky.TraceID(ctx))
			return NewContext(ctx, span)
		},
	)

	clientAfter := kitgrpc.ClientAfter(
		func(ctx context.Context, _ metadata.MD, _ metadata.MD) context.Context {
			return ctx
		},
	)

	clientFinalizer := kitgrpc.ClientFinalizer(
		func(ctx context.Context, err error) {

			method, _ := ctx.Value(kitgrpc.ContextKeyRequestMethod).(string)

			span, ctx, err := tracer.CreateEntrySpan(ctx, method, func() (string, error) {
				return "", nil
			})
			if err != nil {
				_ = config.logger.Log("err", err)
				return
			}
			defer span.End()

			span.SetComponent(0)
			span.SetSpanLayer(common.SpanLayer_RPCFramework)
			span.Tag(tracing.TagGRPCClient, "client finalizer")

			if status2, ok := status.FromError(err); ok {
				statusCode := strconv.FormatUint(uint64(status2.Code()), 10)
				span.Tag(go2sky.TagStatusCode, statusCode)
			} else {
				span.Tag(go2sky.TagStatusCode, err.Error())
			}
		},
	)

	return func(c *kitgrpc.Client) {
		clientBefore(c)
		clientAfter(c)
		clientFinalizer(c)
	}

}

func GRPCServerTrace(tracer *go2sky.Tracer, options ...TracerOption) kitgrpc.ServerOption {
	config := tracerOptions{
		tags:      make(map[string]string),
		name:      tracing.TagGRPCServer,
		logger:    log.NewNopLogger(),
		propagate: true,
	}

	for _, option := range options {
		option(&config)
	}

	serverBefore := kitgrpc.ServerBefore(
		func(ctx context.Context, md metadata.MD) context.Context {

			var (
				method string
			)

			if len(md[tracing.TagGRPCMethod]) > 0 {
				method = md[tracing.TagGRPCMethod][0]
			}

			span, ctx, err := tracer.CreateEntrySpan(ctx, method, func() (string, error) {
				if h, ok := md[propagation.Header]; ok && h != nil && len(h) > 0 {
					return h[0], nil
				}

				return "", nil
			})
			if err != nil {
				_ = config.logger.Log("err", err)
				return ctx
			}
			//defer span.End()

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

			span.SetComponent(0)
			span.Tag(tracing.TagGRPCServerStart, "server before")
			span.SetSpanLayer(common.SpanLayer_RPCFramework)
			return NewContext(ctx, span)
		},
	)

	serverAfter := kitgrpc.ServerAfter(
		func(ctx context.Context, _ *metadata.MD, _ *metadata.MD) context.Context {
			return ctx
		},
	)

	serverFinalizer := kitgrpc.ServerFinalizer(
		func(ctx context.Context, err error) {
			span := SpanFromContext(ctx)

			if span != nil {
				span.Tag(tracing.TagGRPCServerEnd, "server finalizer")
				if status2, ok := status.FromError(err); ok {
					statusCode := strconv.FormatUint(uint64(status2.Code()), 10)
					span.Tag(go2sky.TagStatusCode, statusCode)
				} else {
					span.Tag(go2sky.TagStatusCode, err.Error())
				}
				span.End()
			}

			//fmt.Println("grpc-server serverFinalizer go2sky.TraceID", go2sky.TraceID(ctx))
		},
	)

	return func(s *kitgrpc.Server) {
		serverBefore(s)
		serverAfter(s)
		serverFinalizer(s)
	}
}
