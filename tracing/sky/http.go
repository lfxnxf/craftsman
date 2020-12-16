package sky

import (
	"context"
	"github.com/tiantianjianbao/craftsman/tracing"
	"net/http"
	"strconv"

	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/propagation"
	"github.com/SkyAPM/go2sky/reporter/grpc/common"
	"github.com/go-kit/kit/log"
	kithttp "github.com/go-kit/kit/transport/http"
)

func HTTPClientTrace(tracer *go2sky.Tracer, options ...TracerOption) kithttp.ClientOption {
	config := tracerOptions{
		tags:      make(map[string]string),
		name:      tracing.TagHTTPClient,
		logger:    log.NewNopLogger(),
		propagate: true,
	}

	for _, option := range options {
		option(&config)
	}

	clientBefore := kithttp.ClientBefore(
		func(ctx context.Context, req *http.Request) context.Context {

			config.name = req.URL.Path

			span, ctx, err := tracer.CreateEntrySpan(req.Context(), req.URL.Path, func() (s string, e error) {
				return req.Header.Get(propagation.Header), nil
			})
			if err != nil {
				_ = config.logger.Log("err", err)
				return ctx
			}

			span.SetComponent(0)
			span.Tag(go2sky.TagHTTPMethod, req.Method)
			span.Tag(go2sky.TagURL, req.Host+req.URL.Path)
			span.SetSpanLayer(common.SpanLayer_Http)
			span.Tag(tracing.TagHTTPClient, "client before")
			span.End()
			//fmt.Println("http.client clientBefore go2sky.TraceID", go2sky.TraceID(ctx))
			return NewContext(ctx, span)
		},
	)

	clientAfter := kithttp.ClientAfter(
		func(ctx context.Context, res *http.Response) context.Context {
			span, err := tracer.CreateExitSpan(ctx, res.Request.URL.Path, res.Request.Host, func(header string) error {
				return nil
			})
			if err != nil {
				_ = config.logger.Log("err", err)
				return ctx
			}

			span.SetComponent(0)
			span.Tag(tracing.TagHTTPClient, "client after")
			span.Tag(go2sky.TagStatusCode, strconv.Itoa(res.StatusCode))
			span.End()
			//fmt.Println("http.client clientAfter go2sky.TraceID", go2sky.TraceID(ctx))
			return NewContext(ctx, span)
		},
	)

	clientFinalizer := kithttp.ClientFinalizer(
		func(ctx context.Context, err error) {

		},
	)

	return func(c *kithttp.Client) {
		clientBefore(c)
		clientAfter(c)
		clientFinalizer(c)
	}
}

func HTTPServerTrace(tracer *go2sky.Tracer, options ...TracerOption) kithttp.ServerOption {
	config := tracerOptions{
		tags:      make(map[string]string),
		name:      tracing.TagHTTPServer,
		logger:    log.NewNopLogger(),
		propagate: true,
	}

	for _, option := range options {
		option(&config)
	}

	serverBefore := kithttp.ServerBefore(
		func(ctx context.Context, req *http.Request) context.Context {
			config.name = req.URL.Path
			span, ctx, err := tracer.CreateEntrySpan(req.Context(), config.name, func() (string, error) {
				return req.Header.Get(propagation.Header), nil
			})
			if err != nil {
				_ = config.logger.Log("err", err)
				return ctx
			}

			ctx = context.WithValue(ctx, tracing.PropagationHeader, req.Header.Get(propagation.Header))
			//ctx = context.WithValue(ctx, tracing.TraceId, go2sky.TraceID(ctx))
			ctx = context.WithValue(ctx, tracing.TraceId, go2sky.TraceID(ctx))

			span.SetComponent(0)
			span.Tag(go2sky.TagHTTPMethod, req.Method)
			span.Tag(go2sky.TagURL, req.Host+req.URL.Path)
			span.SetSpanLayer(common.SpanLayer_Http)
			span.Tag(tracing.TagHTTPServer, "server before")
			//span.End()
			//fmt.Println("http.service serverBefore go2sky.TraceID", go2sky.TraceID(ctx))

			return NewContext(ctx, span)
		},
	)

	serverAfter := kithttp.ServerAfter(
		func(ctx context.Context, res http.ResponseWriter) context.Context {
			return ctx
		},
	)

	serverFinalizer := kithttp.ServerFinalizer(
		func(ctx context.Context, code int, r *http.Request) {
			/*span, err := tracer.CreateExitSpan(ctx, r.URL.Path, r.Host, func(header string) error {
				return nil
			})
			if err != nil {
				_ = config.logger.Log("err", err)
				return
			}
			span.SetComponent(0)*/
			span := SpanFromContext(ctx)
			if span != nil {
				span.Tag(go2sky.TagStatusCode, strconv.Itoa(code))
				span.Tag(tracing.TagHTTPServer, "server finalizer")
				span.End()
				//fmt.Println("http.service serverFinalizer go2sky.TraceID", go2sky.TraceID(ctx))
			}
			return
		},
	)

	return func(s *kithttp.Server) {
		serverBefore(s)
		serverAfter(s)
		serverFinalizer(s)
	}
}
