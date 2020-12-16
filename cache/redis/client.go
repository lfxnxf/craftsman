package redis

import (
	"errors"
	"github.com/SkyAPM/go2sky"

	"github.com/tiantianjianbao/craftsman/log"
)

var (
	ClientNotInit = errors.New("redis client not init ")
)

var redisClient map[string]*Redis

func InitRedisClient(logger log.Logger, tracer *go2sky.Tracer, redisConfigs []RedisConfig) error {
	if redisClient == nil {
		redisClient = make(map[string]*Redis)
	}

	for _, config := range redisConfigs {
		client, err := NewRedis(&config, logger, tracer)
		if err != nil {
			continue
		}
		if client == nil {
			continue
		}
		redisClient[config.ServerName] = client
	}
	return nil
}

func GetRedis(service string) (*Redis, error) {
	if client, ok := redisClient[service]; ok {
		return client, nil
	}
	return nil, ClientNotInit
}
