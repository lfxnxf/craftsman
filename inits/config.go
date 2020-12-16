package inits

import (
	"github.com/lfxnxf/craftsman/mq/rocketmq"
	"github.com/lfxnxf/craftsman/opensearch"
	"time"

	"github.com/lfxnxf/craftsman/cache/redis"
	"github.com/lfxnxf/craftsman/db/sql"
	"github.com/lfxnxf/craftsman/log"
	"github.com/lfxnxf/craftsman/sd"
	"github.com/lfxnxf/craftsman/tracing"
	"github.com/lfxnxf/craftsman/transport"
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
