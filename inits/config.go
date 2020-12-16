package inits

import (
	"github.com/tiantianjianbao/craftsman/mq/rocketmq"
	"github.com/tiantianjianbao/craftsman/opensearch"
	"time"

	"github.com/tiantianjianbao/craftsman/cache/redis"
	"github.com/tiantianjianbao/craftsman/db/sql"
	"github.com/tiantianjianbao/craftsman/log"
	"github.com/tiantianjianbao/craftsman/sd"
	"github.com/tiantianjianbao/craftsman/tracing"
	"github.com/tiantianjianbao/craftsman/transport"
)

type DefaultConfig struct {
	Server transport.Server `toml:"server"`

	Log log.Log `toml:"log"`

	Trace tracing.Trace `toml:"trace"`

	ServiceDiscovery sd.ServiceDiscovery `toml:"service_discovery"`

	ServerClient []transport.ServerClient `toml:"server_client"`

	Redis []redis.ConfigToml `toml:"redis"`

	Database []sql.SQLGroupConfig `toml:"database"`

	OpenSearch []opensearch.OpenSearchConfig `toml:"opensearch"`

	RocketMQ []rocketmq.RocketMQConfig `toml:"rocketmq"`
}

type Config interface {
	// Service
	GetServiceName() string
	GetServicePort() int
	GetServiceProto() string
	GetServiceClusterName() string
	GetServiceGroupName() string
	GetServiceRunModel() string
	IdleTimeout() time.Duration
	KeepAliveInterval() time.Duration

	// Logger
	LogLevel() string
	LogRotate() string
	LogPath() string

	// Trace
	GetTraceConfig() tracing.Trace

	// ServiceDiscovery
	GetServiceDiscoveryConfig() sd.ServiceDiscovery

	// Redis
	GetRedisConfig() []redis.RedisConfig

	// MySQL
	GetSQLConfig() []sql.SQLGroupConfig

	// OpenSearch
	GetOpenSearchConfig() []opensearch.OpenSearchConfig

	//RocketMQ
	GetRocketMQConfig() []rocketmq.RocketMQConfig

	// Service client
	GetServiceClients() []transport.ServerClient
}
