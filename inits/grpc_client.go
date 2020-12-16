package inits

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SkyAPM/go2sky/propagation"
	"github.com/SkyAPM/go2sky/reporter/grpc/common"
	"github.com/lfxnxf/craftsman/tracing/sky"
	"google.golang.org/grpc/status"
	"strconv"
	"sync"
	"time"

	"github.com/SkyAPM/go2sky"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	"github.com/lfxnxf/craftsman/cache/memory"
	"github.com/lfxnxf/craftsman/log"
	"github.com/lfxnxf/craftsman/tracing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
)

type GrpcClient interface {
	Call(ctx context.Context, service, alisService, method string, request, reply interface{}) (interface{}, error)
	CallAndConvertResp(ctx context.Context, service, alisService, method string, request, reply interface{}, resp ...interface{}) error
}

type grpcClient struct {
	logger   log.Logger
	tracers  *TraceClients
	config   Config
	connsMap memory.ConcurrentMap
}

func NewGrpcClient(logger log.Logger, tracers *TraceClients, config Config) GrpcClient {
	return &grpcClient{
		logger:   logger,
		tracers:  tracers,
		config:   config,
		connsMap: memory.NewConcurrentMap(),
	}
}

func (c *grpcClient) Call(ctx context.Context, service, alisService, method string, request, reply interface{}) (interface{}, error) {
	tracer, _ := c.tracers.GetTracer()
	target := ""
	//兼容旧service，没有服务发现
	serviceClients := c.config.GetServiceClients()
	for _, serviceClient := range serviceClients {
		if serviceClient.ServiceName == service {
			target = serviceClient.Ipport
			break
		}
	}

	if target == "" {
		instance, err := GetOneHealthyInstance(service)()
		if err != nil {
			c.logger.InfoT(ctx, "nacos get instance["+service+"] error", "err", err.Error())
			return nil, err
		}

		target = fmt.Sprintf("%s:%d", instance.Ip, instance.Port)
	}

	key := fmt.Sprintf("%s#%s", service, target)
	c.logger.InfoT(ctx, "get grpc client", "service", service, "target", target)

	var conn *grpc.ClientConn
	_conn, ok := c.connsMap.Get(key)
	if !ok {
		conn, err := GetGrpcConn(target)
		if err != nil {
			c.logger.ErrorT(ctx, "get grpc client", "service", service, "err", err.Error())
			return nil, err
		}
		c.connsMap.Set(key, conn)
		_conn, _ = c.connsMap.Get(key)
	}

	conn, ok = _conn.(*grpc.ClientConn)
	if !ok {
		c.connsMap.Remove(target)
		c.logger.ErrorT(ctx, "get grpc client", "service", service, "err", "conn type error")
		return nil, errors.New("conn type error")
	}

	connStatus := conn.GetState()
	if connStatus != connectivity.Ready && connStatus != connectivity.Idle && connStatus != connectivity.Connecting && connStatus != connectivity.TransientFailure {
		c.connsMap.Remove(target)
		c.logger.ErrorT(ctx, "get grpc client", "service", service, "err", "conn no activity "+strconv.Itoa(int(connStatus)))
		return nil, errors.New("conn no activity")
	}

	client, err := newGRPCClient(ctx, conn, target, c.logger, tracer, service, alisService, method, reply)
	if err != nil {
		c.logger.ErrorT(ctx, "new grpc client", "err", err)
		return nil, err
	}

	res, err := client.Endpoint()(ctx, request)
	if err != nil {
		c.logger.ErrorT(ctx, "client endpoint invoke", "service", service, "method", method, "err", err)
		return nil, err
	}

	return res, err
}

func (c *grpcClient) CallAndConvertResp(ctx context.Context, service, alisService, method string, request, reply interface{}, resp ...interface{}) error {
	res, err := c.Call(ctx, service, alisService, method, request, reply)
	if err != nil {
		return err
	}

	pbStr, _ := json.Marshal(res)
	err = json.Unmarshal(pbStr, &resp)
	if err != nil {
		return err
	}

	return nil
}

func newGRPCClient(ctx context.Context, conn *grpc.ClientConn, endpoint string, log log.Logger, tracer *go2sky.Tracer, serviceName, alisService, method string, reply interface{}) (*grpctransport.Client, error) {

	var options []grpctransport.ClientOption

	beforeFunc := func(ctx context.Context, md *metadata.MD) context.Context {
		if tracer != nil {
			span, err := tracer.CreateExitSpan(ctx, method, endpoint, func(header string) error {
				(*md)[propagation.Header] = []string{header}
				(*md)[tracing.TagGRPCServer] = []string{serviceName}
				(*md)[tracing.TagGRPCMethod] = []string{method}
				return nil
			})
			if err != nil {
				log.ErrorT(ctx, "create exit span", "err", err)
				return ctx
			}
			span.Tag(tracing.TagGRPCClientStart, "before func")
			span.SetSpanLayer(common.SpanLayer_RPCFramework)
			span.SetComponent(0)
			//span.End()
			return sky.NewContext(ctx, span)
		}

		return ctx
	}

	afterFunc := func(ctx context.Context, _ metadata.MD, _ metadata.MD) context.Context {
		return ctx
	}

	finalizerFunc := func(ctx context.Context, err error) {
		span := sky.SpanFromContext(ctx)
		if span != nil {
			span.Tag(tracing.TagGRPCClientEnd, "finalizer func")
			if status2, ok := status.FromError(err); ok {
				statusCode := strconv.FormatUint(uint64(status2.Code()), 10)
				span.Tag(go2sky.TagStatusCode, statusCode)
			} else {
				span.Tag(go2sky.TagStatusCode, err.Error())
			}
			span.End()
		}

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
			grpctransport.ClientAfter(afterFunc),
			grpctransport.ClientFinalizer(finalizerFunc))...,
	)

	return client, nil
}

func decodeRequest(ctx context.Context, req interface{}) (interface{}, error) {
	return req, nil
}

func encodeResponse(ctx context.Context, resp interface{}) (interface{}, error) {
	return resp, nil
}

var (
	ErrNotFoundClient = errors.New("not found grpc conn")
	ErrConnShutdown   = errors.New("grpc conn shutdown")

	defaultClientPoolCap    = 3
	defaultDialTimeout      = 100 * time.Millisecond
	defaultKeepAlive        = 30 * time.Second
	defaultKeepAliveTimeout = 10 * time.Second
)

func GetGrpcConn(target string) (*grpc.ClientConn, error) {
	grpcConn := NewGrpcConn(
		target,
		NewDefaultClientOption(),
	)
	conn, err := grpcConn.connect()
	if err != nil {
		return conn, err
	}
	err = grpcConn.checkState(conn)
	return conn, err
}

type ClientOption struct {
	DialTimeout      time.Duration
	KeepAlive        time.Duration
	KeepAliveTimeout time.Duration
	ClientPoolSize   int
}

type GrpcConn struct {
	option *ClientOption
	target string
	mtx    sync.RWMutex
}

func NewGrpcConn(target string, option *ClientOption) *GrpcConn {
	return &GrpcConn{
		target: target,
		option: option,
	}
}

func NewDefaultClientOption() *ClientOption {
	return &ClientOption{
		DialTimeout:      defaultDialTimeout,
		KeepAlive:        defaultKeepAlive,
		KeepAliveTimeout: defaultKeepAliveTimeout,
	}
}

//func (c *GrpcConn) getConn() (*grpc.ClientConn, error) {
//	err := c.checkState(c.conn)
//	if err != nil {
//		return c.conn, err
//	}
//	return c.conn, err
//}

func (c *GrpcConn) connect() (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(c.target,
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(c.option.DialTimeout),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    c.option.KeepAlive,
			Timeout: c.option.KeepAliveTimeout},
		),
	)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (c *GrpcConn) checkState(conn *grpc.ClientConn) error {
	state := conn.GetState()
	switch state {
	case connectivity.TransientFailure, connectivity.Shutdown:
		return ErrConnShutdown
	}
	return nil
}

func (c *GrpcConn) close() {
	c.mtx.Lock()
	defer c.mtx.Unlock()
}
