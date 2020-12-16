package inits

import (
	"github.com/lfxnxf/craftsman/cache/redis"
	"github.com/lfxnxf/craftsman/log"
)

type RedisClients struct {
	RedisClient map[string]*redis.Redis
	logger      log.Logger
}

func NewRedisClient(log log.Logger, tracerClients *TraceClients, redisConfigs []redis.RedisConfig) (*RedisClients, error) {
	rdsClients := &RedisClients{
		logger:      log,
		RedisClient: make(map[string]*redis.Redis, len(redisConfigs)),
	}

	tracer, _ := tracerClients.GetTracer()
	for _, config := range redisConfigs {
		client, err := redis.NewRedis(&config, log, tracer)

		if err != nil {
			log.Error("new redis", "config", config, "err", err)
			continue
		}
		if client == nil {
			log.Error("new redis", "config", config, "client init failed")
			continue
		}
		rdsClients.RedisClient[config.ServerName] = client
	}
	return rdsClients, nil
}
