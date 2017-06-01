package redis

import (
	"fmt"
	"log"

	redis "github.com/go-redis/redis"
)

// ConfigRedis sets the Redis.
type ConfigRedis struct {
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	Auth        string `yaml:"auth"`
	IdleTimeout int    `yaml:"idle_timeout"`
}

// newRedisClient constructs a redis client.
func newRedisClient(config ConfigRedis) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Auth,
		PoolSize: 200,
	})

	if _, err := client.Ping().Result(); err != nil {
		log.Println("fail to initialize redis client: ", err)
		client = nil
	}

	return client
}
