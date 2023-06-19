package redis

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

// ConfigRedis sets the Redis.
type ConfigRedis struct {
	Addr        string `yaml:"addr"` // host:port
	Password    string `yaml:"password"`
	IdleTimeout int    `yaml:"idle_timeout"`
}

// newRedisClient constructs a redis client.
func newRedisClient(config ConfigRedis) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		PoolSize: 200,
	})

	err := checkRedisClient(client)
	if err != nil {
		return nil
	}
	return client
}

// check redis client connection.
func checkRedisClient(client *redis.Client) error {
	if _, err := client.Ping(context.Background()).Result(); err != nil {
		log.Println("fail to ping redis client: ", err)
		return err
	}
	return nil
}
