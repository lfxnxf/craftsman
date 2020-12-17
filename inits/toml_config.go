package inits

import (
	"github.com/SkyAPM/go2sky"
	"github.com/lfxnxf/craftsman/httprequest"
	//"github.com/lfxnxf/craftsman/mq/rocketmq"

	"github.com/lfxnxf/craftsman/opensearch"
	"time"

	"github.com/lfxnxf/craftsman/cache/redis"
	"github.com/lfxnxf/craftsman/config/toml"
	"github.com/lfxnxf/craftsman/db/sql"
	"github.com/lfxnxf/craftsman/log"
	"github.com/lfxnxf/craftsman/sd"
	"github.com/lfxnxf/craftsman/tracing"
	"github.com/lfxnxf/craftsman/transport"
)

var (
	serviceName string
	config      Config
	//logRotate   string
	//tomlDefault *DefaultConfig
)

type CommonClients struct {
	TraceClients      *TraceClients
	NacosClient       *NacosClient
	ServiceClients    *ServiceClients
	RedisClients      *RedisClients
	SQLClients        *SQLClients
	OpenSearchClients *OpenSearchClients
	HttpClient        *httprequest.Client
	//RocketMQClients   *rocketmq.Client
}

type ConfigToml struct {
	config        *toml.Config
	defaultConfig *DefaultConfig
}

func (c ConfigToml) GetServiceName() string {
	return c.defaultConfig.Server.ServiceName
}

func (c ConfigToml) GetServiceRunModel() string {
	return c.defaultConfig.Server.RunModel
}

func (c ConfigToml) GetServicePort() int {
	return c.defaultConfig.Server.Port
}

func (c ConfigToml) GetServiceProto() string {
	return c.defaultConfig.Server.Proto
}

func (c ConfigToml) GetServiceClusterName() string {
	return c.defaultConfig.Server.ClusterName
}

func (c ConfigToml) GetServiceGroupName() string {
	return c.defaultConfig.Server.GroupName
}

func (c ConfigToml) IdleTimeout() time.Duration {
	return time.Duration(c.defaultConfig.Server.TCP.IdleTimeout) * time.Second
}

func (c ConfigToml) KeepAliveInterval() time.Duration {
	return time.Duration(c.defaultConfig.Server.TCP.KeepliveInterval) * time.Second
}

func (c ConfigToml) LogLevel() string {
	return c.defaultConfig.Log.Level
}

func (c ConfigToml) LogRotate() string {
	return c.defaultConfig.Log.Rotate
}

func (c ConfigToml) LogPath() string {
	return c.defaultConfig.Log.LogPath
}

func (c ConfigToml) GetServiceDiscoveryConfig() sd.ServiceDiscovery {
	return c.defaultConfig.ServiceDiscovery
}

func (c ConfigToml) GetTraceConfig() tracing.Trace {
	return c.defaultConfig.Trace
}

func (c ConfigToml) GetRedisConfig() []redis.RedisConfig {
	var redisConfig []redis.RedisConfig
	if len(c.defaultConfig.Redis) == 0 {
		return redisConfig
	}
	for _, defaultConfig := range c.defaultConfig.Redis {

		var rc redis.RedisConfig
		rc.ServerName = defaultConfig.ServerName
		rc.Addr = defaultConfig.Addr
		rc.Password = defaultConfig.Password
		rc.MaxIdle = defaultConfig.MaxIdle
		rc.MaxActive = defaultConfig.MaxActive
		rc.IdleTimeout = defaultConfig.IdleTimeout
		rc.ConnectTimeout = defaultConfig.ConnectTimeout
		rc.ReadTimeout = defaultConfig.ReadTimeout
		rc.WriteTimeout = defaultConfig.WriteTimeout
		rc.Database = defaultConfig.Database
		rc.Retry = defaultConfig.Retry
		redisConfig = append(redisConfig, rc)
	}
	return redisConfig
}

func (c ConfigToml) GetSQLConfig() []sql.SQLGroupConfig {
	return c.defaultConfig.Database
}

func (c ConfigToml) GetOpenSearchConfig() []opensearch.OpenSearchConfig {
	return c.defaultConfig.OpenSearch
}

func (c ConfigToml) GetServiceClients() []transport.ServerClient {
	return c.defaultConfig.ServerClient
}

//func (c ConfigToml) GetRocketMQConfig() []rocketmq.RocketMQConfig {
//	return c.defaultConfig.RocketMQ
//}

func GetServiceName() string {
	return serviceName
}

func NewConfigToml(path string, v interface{}) (*ConfigToml, error) {
	tomlDefault := &DefaultConfig{}

	tomlConfig, err := NewConfig(path, &tomlDefault)
	if err != nil {
		return nil, err
	}

	err = ParseTomlConfig(path, v)
	if err != nil {
		return nil, err
	}

	configToml := &ConfigToml{
		config:        tomlConfig,
		defaultConfig: tomlDefault,
	}

	serviceName = configToml.GetServiceName()

	config = configToml

	return configToml, nil
}

func NewCommonClients(c Config, logger log.Logger) (*CommonClients, error) {
	comClients := &CommonClients{}

	// new sky tracer client
	var tracer *go2sky.Tracer
	if c.GetTraceConfig().Ipport != "" {
		traceClients, err := NewSkyTraceClient(logger, c.GetServiceName(), c.GetTraceConfig())
		if err != nil {
			return comClients, err
		}
		comClients.TraceClients = traceClients
		tracer, _ = comClients.TraceClients.GetTracer()
	}

	// new nacos client
	if c.GetServiceDiscoveryConfig().Ipport != "" {
		nacosClient, err := NewNacosClient(logger, c.GetServiceDiscoveryConfig())
		if err != nil {
			return comClients, err
		}
		comClients.NacosClient = nacosClient
	}

	// new service client
	if len(c.GetServiceClients()) > 0 {
		serviceClients, err := NewServiceClients(logger, c.GetServiceClients())
		if err != nil {
			return comClients, err
		}
		comClients.ServiceClients = serviceClients
	}

	// new redis
	if len(c.GetRedisConfig()) > 0 {
		rdsClient, err := NewRedisClient(logger, comClients.TraceClients, c.GetRedisConfig())
		if err != nil {
			return comClients, err
		}
		comClients.RedisClients = rdsClient
	}

	// new Mysql
	if len(c.GetSQLConfig()) > 0 {
		sqlClients, err := NewSQLClients(logger, comClients.TraceClients, c.GetSQLConfig())
		if err != nil {
			return comClients, err
		}
		comClients.SQLClients = sqlClients
	}

	// new OpenSearch
	if len(c.GetOpenSearchConfig()) > 0 {
		opensearchClients, err := NewOpenSearchClients(logger, comClients.TraceClients, c.GetOpenSearchConfig())
		if err != nil {
			return comClients, err
		}
		comClients.OpenSearchClients = opensearchClients
	}

	// new RocketMQ
	//if len(c.GetRocketMQConfig()) > 0 {
	//	rocketMQClients, err := rocketmq.NewClient(c.GetServiceName(), c.GetRocketMQConfig(), logger, tracer)
	//	if err != nil {
	//		return comClients, err
	//	}
	//	comClients.RocketMQClients = rocketMQClients
	//}

	//new http client
	comClients.HttpClient = httprequest.NewClient(1*time.Second, logger, tracer)

	//  register instance
	if comClients.NacosClient != nil {
		err := RegisterInstance(logger, c, comClients.NacosClient)
		if err != nil {
			return comClients, err
		}
	}

	return comClients, nil
}
