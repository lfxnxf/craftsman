package inits

import (
	"sync"

	"github.com/tiantianjianbao/craftsman/log"
	"github.com/tiantianjianbao/craftsman/transport"
)

const (
	CONNECT_TIMEOUT int = 300
	READ_TIMEOUT    int = 300
	WRITE_TIMEOUT   int = 300
	SLOW_TIME       int = 500
)

var (
	scMap = map[string]transport.ServerClient{}
)

type ServiceClients struct {
	ServerClient map[string]transport.ServerClient
	mut          sync.RWMutex
	logger       log.Logger
}

func NewServiceClients(log log.Logger, serverClients []transport.ServerClient) (*ServiceClients, error) {
	scMap = make(map[string]transport.ServerClient, len(serverClients))

	scs := &ServiceClients{
		logger:       log,
		ServerClient: make(map[string]transport.ServerClient, len(serverClients)),
	}

	for _, sc := range serverClients {

		if _, ok := scs.ServerClient[sc.ServiceName]; ok {
			continue
		}
		scs.ServerClient[sc.ServiceName] = makeDefaultConfig(sc)

		scMap[sc.ServiceName] = makeDefaultConfig(sc)
	}

	return scs, nil
}

func makeDefaultConfig(sc transport.ServerClient) transport.ServerClient {
	connectTime := sc.ConnectTimeout
	if connectTime == 0 {
		sc.ConnectTimeout = CONNECT_TIMEOUT
	}
	readTimeout := sc.ReadTimeout
	if readTimeout == 0 {
		sc.ReadTimeout = READ_TIMEOUT
	}
	writeTimeout := sc.WriteTimeout
	if writeTimeout == 0 {
		sc.WriteTimeout = WRITE_TIMEOUT
	}
	slowTime := sc.SlowTime
	if slowTime == 0 {
		sc.SlowTime = SLOW_TIME
	}
	return sc
}
