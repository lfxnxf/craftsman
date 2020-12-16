package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/tiantianjianbao/craftsman/tracing"
	"github.com/tiantianjianbao/craftsman/tracing/sky"

	"github.com/SkyAPM/go2sky/propagation"
	httptransport "github.com/go-kit/kit/transport/http"
)

type Response struct {
	Body   io.ReadCloser
	String string
}

func NewHTTPClient(instance, method, rawurl string) (*httptransport.Client, error) {
	var options []httptransport.ClientOption

	tracer, err := tracing.GetTracer()
	if err != nil {
		fmt.Println(err)
	} else {
		options = append(options, sky.HTTPClientTrace(tracer))
	}

	var (
		encode = func(context.Context, *http.Request, interface{}) error { return nil }
		decode = func(_ context.Context, r *http.Response) (interface{}, error) {
			return Response{r.Body, ""}, nil
		}
	)

	beforeFunc := func(ctx context.Context, r *http.Request) context.Context {
		span, err := tracer.CreateExitSpan(ctx, r.URL.Path, r.Host, func(header string) error {
			r.Header.Set(propagation.Header, header)
			return nil
		})
		if err != nil {
			return ctx
		}
		span.SetComponent(0)
		span.End()
		return ctx
	}

	afterFunc := func(ctx context.Context, r *http.Response) context.Context {
		return ctx
	}

	u, err := url.Parse(rawurl)
	if err != nil {
		fmt.Println(err)
	}

	c := httptransport.NewClient(
		method,
		u,
		encode,
		decode,
		append(options,
			httptransport.ClientBefore(beforeFunc),
			httptransport.ClientAfter(afterFunc))...,
	)
	return c, nil
}

func decodeRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	return r, nil
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(response)
	return nil
}
