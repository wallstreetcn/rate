package redis

import (
	"fmt"
	"log"
	"strconv"
	"time"

	redis "github.com/go-redis/redis"
)

// ConfigRedis sets the Redis.
type ConfigRedis struct {
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	Auth        string `yaml:"auth"`
	IdleTimeout int    `yaml:"idle_timeout"`
}

// NewRedisRing constructs a redis ring.
func NewRedisRing(config ...ConfigRedis) *redis.Ring {
	servers := map[string]string{}
	var pass string
	var timeoutRead, timeoutWrite, timeoutConnect, timeoutIdle time.Duration

	for i, v := range config {
		servers[strconv.Itoa(i)] = fmt.Sprintf("%s:%d", v.Host, v.Port)
		if i == 0 {
			pass = v.Auth
			timeoutIdle = time.Second * time.Duration(v.IdleTimeout)
		}
	}
	client := redis.NewRing(&redis.RingOptions{
		Addrs:        servers,
		Password:     pass,
		IdleTimeout:  timeoutIdle,
		ReadTimeout:  timeoutRead,
		WriteTimeout: timeoutWrite,
		DialTimeout:  timeoutConnect,
		PoolSize:     200,
	})

	if _, err := client.Ping().Result(); err != nil {
		log.Println("fail to initialize redis ring: ", err)
		client = nil
	}

	return client
}

// NewRedisClient constructs a redis client.
func NewRedisClient(config ConfigRedis) *redis.Client {
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
