package httprequest

import (
	"context"
	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/reporter/grpc/common"
	"github.com/mozillazg/request"
	"github.com/lfxnxf/craftsman/log"
	"net/http"
	"time"
)

// Client is alias Args
type Client struct {
	timeOut time.Duration
	logger  log.Logger
	tracer  *go2sky.Tracer
}

type HttpRequest struct {
	*request.Request
	logger log.Logger
	tracer *go2sky.Tracer
}

// NewClient return a *Client
func NewClient(timeOut time.Duration, logger log.Logger, tracer *go2sky.Tracer) *Client {
	return &Client{timeOut: timeOut, logger: logger, tracer: tracer}
}

func (c *Client) GetRequest() *HttpRequest {
	return &HttpRequest{
		Request: request.NewRequest(&http.Client{Timeout: c.timeOut}),
		logger:  c.logger,
		tracer:  c.tracer,
	}
}

func (req *HttpRequest) goTracerAndExec(ctx context.Context, f func() (resp *request.Response, err error)) (resp *request.Response, err error) {
	if req.tracer != nil {
		span, err := req.tracer.CreateExitSpan(ctx, "nil", "nil", func(header string) error {
			return nil
		})
		if err != nil {
			req.logger.ErrorT(ctx, "request tracer error", "err", err.Error())
			return f()
		}
		resp, err = f()

		if err != nil {
			span.Tag(go2sky.TagStatusCode, err.Error())
			req.logger.ErrorT(ctx, "http request err", "err", err.Error())
		} else {
			span.Tag(go2sky.TagStatusCode, resp.Status)
		}

		if resp != nil && resp.Response != nil {
			span.SetOperationName(resp.Request.Method)
			span.SetPeer(resp.Request.Host)
			span.Tag(go2sky.TagURL, resp.Request.URL.String())
		}

		span.SetSpanLayer(common.SpanLayer_Http)
		span.End()
	}

	return f()
}

// Get issues a GET to the specified URL.
//
// url can be string or *url.URL or ur.URL
func (req *HttpRequest) Get(ctx context.Context, url interface{}) (resp *request.Response, err error) {

	return req.goTracerAndExec(ctx, func() (resp *request.Response, err error) {
		return req.Request.Get(url)
	})
}

// Head issues a HEAD to the specified URL.
//
// url can be string or *url.URL or ur.URL
func (req *HttpRequest) Head(ctx context.Context, url interface{}) (resp *request.Response, err error) {
	return req.goTracerAndExec(ctx, func() (resp *request.Response, err error) {
		return req.Request.Head(url)
	})
}

// Post issues a POST to the specified URL.
//
// url can be string or *url.URL or ur.URL
func (req *HttpRequest) Post(ctx context.Context, url interface{}) (resp *request.Response, err error) {
	return req.goTracerAndExec(ctx, func() (resp *request.Response, err error) {
		return req.Request.Post(url)
	})
}

// PostForm send post form request.
//
// url can be string or *url.URL or ur.URL
//
// data can be map[string]string or map[string][]string or string or io.Reader
//
// 	data := map[string]string{
// 		"a": "1",
// 		"b": "2",
// 	}
//
// 	data := map[string][]string{
// 		"a": []string{"1", "2"},
// 		"b": []string{"2", "3"},
// 	}
//
// 	data : = "a=1&b=2"
//
// 	data : = strings.NewReader("a=1&b=2")
//
func (req *HttpRequest) PostForm(ctx context.Context, url interface{}, data interface{}) (resp *request.Response, err error) {
	return req.goTracerAndExec(ctx, func() (resp *request.Response, err error) {
		return req.Request.PostForm(url, data)
	})

}

// Put issues a PUT to the specified URL.
//
// url can be string or *url.URL or ur.URL
func (req *HttpRequest) Put(ctx context.Context, url interface{}) (resp *request.Response, err error) {
	return req.goTracerAndExec(ctx, func() (resp *request.Response, err error) {
		return req.Request.Put(url)
	})
}

// Patch issues a PATCH to the specified URL.
//
// url can be string or *url.URL or ur.URL
func (req *HttpRequest) Patch(ctx context.Context, url interface{}) (resp *request.Response, err error) {
	return req.goTracerAndExec(ctx, func() (resp *request.Response, err error) {
		return req.Request.Patch(url)
	})
}

// Delete issues a DELETE to the specified URL.
//
// url can be string or *url.URL or ur.URL
func (req *HttpRequest) Delete(ctx context.Context, url interface{}) (resp *request.Response, err error) {
	return req.goTracerAndExec(ctx, func() (resp *request.Response, err error) {
		return req.Request.Delete(url)
	})
}

// Options issues a OPTIONS to the specified URL.
//
// url can be string or *url.URL or ur.URL
func (req *HttpRequest) Options(ctx context.Context, url interface{}) (resp *request.Response, err error) {
	return req.goTracerAndExec(ctx, func() (resp *request.Response, err error) {
		return req.Request.Options(url)
	})
}
