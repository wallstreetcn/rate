// Package redis provides a rate limiter based on redis.
package redis

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"log"
	"math"
	"time"

	redis "github.com/go-redis/redis"
)

var (
	redisClient *redis.Client
	scriptHash  string
)

const script = `
local tokens_key = KEYS[1]
local timestamp_key = KEYS[2]

local rate = tonumber(ARGV[1])
local capacity = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local requested = tonumber(ARGV[4])

local fill_time = capacity/rate
local ttl = math.floor(fill_time*2)

local last_tokens = tonumber(redis.call("get", tokens_key))
if last_tokens == nil then
    last_tokens = capacity
end

local last_refreshed = tonumber(redis.call("get", timestamp_key))
if last_refreshed == nil then
    last_refreshed = 0
end

local delta = math.max(0, now-last_refreshed)
local filled_tokens = math.min(capacity, last_tokens+(delta*rate))
local allowed = filled_tokens >= requested
local new_tokens = filled_tokens
if allowed then
    new_tokens = filled_tokens - requested
end

redis.call("setex", tokens_key, ttl, new_tokens)
redis.call("setex", timestamp_key, ttl, now)

return { allowed, new_tokens }
`

// Client indicates the redis client of the rate limiter.
func Client() *redis.Client {
	return redisClient
}

// SetRedis sets the redis client.
func SetRedis(config *ConfigRedis) error {
	if config == nil {
		return errors.New("redis config is empty")
	}

	redisClient = newRedisClient(*config)
	if redisClient == nil {
		return errors.New("redis client is nil")
	}

	go func() {
		timer := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-timer.C:
				loadScript()
			}
		}
	}()
	return loadScript()
}

func loadScript() error {
	if redisClient == nil {
		return errors.New("redis client is nil")
	}

	scriptHash = fmt.Sprintf("%x", sha1.Sum([]byte(script)))
	exists, err := redisClient.ScriptExists(scriptHash).Result()
	if err != nil {
		return err
	}

	// load script when missing.
	if !exists[0] {
		_, err := redisClient.ScriptLoad(script).Result()
		if err != nil {
			return err
		}
	}
	return nil
}

// Limit defines the maximum frequency of some events.
// Limit is represented as number of events per second.
// A zero Limit allows no events.
type Limit float64

// Inf is the infinite rate limit; it allows all events (even if burst is zero).
const Inf = Limit(math.MaxFloat64)

// A Limiter controls how frequently events are allowed to happen.
type Limiter struct {
	limit Limit
	burst int

	// mu sync.Mutex

	key string
}

// NewLimiter returns a new Limiter that allows events up to rate r and permits
// bursts of at most b tokens.
func NewLimiter(r Limit, b int, key string) *Limiter {
	return &Limiter{
		limit: r,
		burst: b,
		key:   key,
	}
}

// Every converts a minimum time interval between events to a Limit.
func Every(interval time.Duration) Limit {
	if interval <= 0 {
		return Inf
	}
	return 1 / Limit(interval.Seconds())
}

// Allow is shorthand for AllowN(time.Now(), 1).
func (lim *Limiter) Allow() bool {
	return lim.AllowN(time.Now(), 1)
}

// AllowN reports whether n events may happen at time now.
// Use this method if you intend to drop / skip events that exceed the rate limit.
// Otherwise use Reserve or Wait.
func (lim *Limiter) AllowN(now time.Time, n int) bool {
	return lim.reserveN(now, n).ok
}

// A Reservation holds information about events that are permitted by a Limiter to happen after a delay.
// A Reservation may be canceled, which may enable the Limiter to permit additional events.
type Reservation struct {
	ok     bool
	tokens int
}

func (lim *Limiter) reserveN(now time.Time, n int) Reservation {
	if redisClient == nil {
		return Reservation{
			ok:     true,
			tokens: n,
		}
	}

	results, err := redisClient.EvalSha(
		scriptHash,
		[]string{lim.key + ".tokens", lim.key + ".ts"},
		float64(lim.limit),
		lim.burst,
		now.Unix(),
		n,
	).Result()
	if err != nil {
		log.Println("fail to call rate limit: ", err)
		return Reservation{
			ok:     true,
			tokens: n,
		}
	}

	rs, ok := results.([]interface{})
	if ok {
		newTokens, _ := rs[1].(int64)
		return Reservation{
			ok:     rs[0] == int64(1),
			tokens: int(newTokens),
		}
	}

	log.Println("fail to transform results")
	return Reservation{
		ok:     true,
		tokens: n,
	}
}
