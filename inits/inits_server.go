package inits

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/fvbock/endless"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/spf13/cast"
	"github.com/tiantianjianbao/craftsman/log"
	"github.com/tiantianjianbao/craftsman/log/zap"
	"github.com/tiantianjianbao/craftsman/transport/gin"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"syscall"
	"time"
)

type server struct {
	C           Config
	Logger      log.Logger
	Clients     *CommonClients
	OptionsHttp []httptransport.ServerOption
	OptionsGrpc []grpctransport.ServerOption
	initTime    time.Time
	startTime   time.Time
}

var (
	ConfigParseErr       = "config parse err:"
	LoggerInitErr        = "logger init err:"
	CommonClientsInitErr = "common client init err:"
	ServerHandleErr      = "server handler err"
)

func NewServer(customConf interface{}) *server {
	initStartTime := time.Now()
	conf := flag.String("config", "./config/config.toml", "rpc config file")

	flag.Parse()
	c, err := NewConfigToml(*conf, customConf)
	if err != nil {
		panic(ConfigParseErr + err.Error())
	}

	separator := "/"
	loggerPath := c.LogPath()
	if strings.LastIndex(loggerPath, separator) != (strings.Count(loggerPath, "") - 1) {
		loggerPath += separator
	}
	loggerPath += c.GetServiceName() + separator + c.GetServiceName()
	logger, err := zap.NewJsonLogger(loggerPath, c.LogRotate(), c.LogLevel())
	if err != nil {
		panic(LoggerInitErr + err.Error())
	}

	clients, err := NewCommonClients(c, logger)
	if err != nil {
		panic(CommonClientsInitErr + err.Error())
	}

	tracer, _ := clients.TraceClients.GetTracer()
	timeDuration := time.Now().Sub(initStartTime).String()
	fmt.Println("server init duration:" + timeDuration)
	logger.Info("server init duration", "time duration", timeDuration)

	return &server{
		C:           c,
		Logger:      logger,
		Clients:     clients,
		OptionsGrpc: DefaultGrpcOptions(logger, tracer),
		OptionsHttp: DefaultHttpOptions(logger, tracer),
		initTime:    initStartTime,
	}
}

func (s *server) Start(handler interface{}) {
	s.startTime = time.Now()
	var err error
	if handler != nil && reflect.TypeOf(handler).AssignableTo(reflect.TypeOf(&grpc.Server{})) {
		err = s.grpcStart(handler.(*grpc.Server))
	} else if reflect.TypeOf(handler).Implements(reflect.TypeOf((*http.Handler)(nil)).Elem()) {
		err = s.httpStart(handler.(http.Handler))
	} else {
		panic(ServerHandleErr)
	}

	fmt.Println(err)
	s.Logger.Info("server stop", "time duration", time.Now().Sub(s.startTime).String(), "start time", s.startTime.String(), "end time", time.Now().String(), "err", err)
	time.Sleep(1 * time.Second)
}

func (s *server) httpStart(h http.Handler) error {
	errs := make(chan error)

	go func() {
		srv := endless.NewServer(fmt.Sprintf(":%d", s.C.GetServicePort()), h)
		srv.SignalHooks[endless.PRE_SIGNAL][syscall.SIGUSR1] = append(
			srv.SignalHooks[endless.PRE_SIGNAL][syscall.SIGUSR1])
		srv.SignalHooks[endless.POST_SIGNAL][syscall.SIGUSR1] = append(
			srv.SignalHooks[endless.POST_SIGNAL][syscall.SIGUSR1])

		errs <- srv.ListenAndServe()
		if errs != nil {
			_ = srv.Close()
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		if err := s.shutdown(); err != nil {
			panic(err)
		}
	}()

	return <-errs
}

func (s *server) grpcStart(server *grpc.Server) error {
	defer server.GracefulStop()
	errs := make(chan error)

	go func() {
		port := fmt.Sprintf(":%d", s.C.GetServicePort())
		sc, err := net.Listen("tcp", port)
		if err != nil {
			errs <- err
		} else {
			errs <- server.Serve(sc)
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		if err := s.shutdown(); err != nil {
			panic(err)
		}
		server.Stop()
	}()

	return <-errs
}

func (s *server) shutdown() (err error) {

	//for _, skyConn := range s.Clients.SkyTracesConn {
	//	skyConn.Close()
	//}

	// 解除服务发现注册
	deRet, err := s.Clients.NacosClient.NacosClient.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          LocalIPString(s.Logger),
		Port:        uint64(s.C.GetServicePort()),
		Cluster:     s.C.GetServiceClusterName(),
		ServiceName: s.C.GetServiceName(),
		GroupName:   s.C.GetServiceGroupName(),
	})
	s.Logger.Info("deregister instance", "deRet", deRet, "err", err)

	//释放rocketmq
	s.Clients.RocketMQClients.Close()

	return
}

func (s *server) NewGrpcHandleSever() *grpc.Server {
	if s.C.GetServiceRunModel() == DebugRunMode {
		return grpc.NewServer()
	}

	opts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandlerContext(func(ctx context.Context, p interface{}) (err error) {
			s.Logger.ErrorT(ctx, "server panic", "err", p)
			return errors.New(cast.ToString(p))
		}),
	}

	return grpc.NewServer(
		grpc_middleware.WithUnaryServerChain(grpc_recovery.UnaryServerInterceptor(opts...)),
		grpc_middleware.WithStreamServerChain(grpc_recovery.StreamServerInterceptor(opts...)),
	)
}

func (s *server) NewHttpHandlerServer() *gin.Client {
	tracer, _ := s.Clients.TraceClients.GetTracer()
	return gin.NewClient(s.C.GetServiceName(), s.C.GetServiceRunModel(), s.Logger, tracer)
}

func (s *server) NewGrpcSvsClient() GrpcClient {
	return NewGrpcClient(s.Logger, s.Clients.TraceClients, s.C)
}
