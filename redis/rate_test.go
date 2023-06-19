// Package redis provides a rate limiter based on redis.
package redis

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func setup() {
	if err := SetRedis(redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
	})); err != nil {
		panic(fmt.Sprintf("fail to initialize redis client: %v", err))
	}
}

func teardown() {
}

func TestMain(m *testing.M) {
	setup()
	retCode := m.Run()
	teardown()
	os.Exit(retCode)
}

func Test1QPS(t *testing.T) {
	assert.NotNil(t, Client(), "redis client should not be empty")

	limiter := NewLimiter(Every(time.Second), 1, "Test1QPS")
	assert.NotNil(t, limiter)

	assert.True(t, limiter.Allow(), "first access should be allowed")
	assert.False(t, limiter.Allow(), "second access should be rejected")
}

func Test1QP2S(t *testing.T) {
	assert.NotNil(t, Client(), "redis client should not be empty")

	limiter := NewLimiter(Every(2*time.Second), 1, "Test1QP2S")
	assert.NotNil(t, limiter)

	assert.True(t, limiter.Allow(), "first access should be allowed")
	assert.False(t, limiter.Allow(), "second access should be rejected")
	<-time.After(2 * time.Second)
	assert.True(t, limiter.Allow(), "third access should be allowed")
}

func Test10QPS(t *testing.T) {
	assert.NotNil(t, Client(), "redis client should not be empty")

	limiter := NewLimiter(Every(100*time.Millisecond), 10, "Test10QPS")
	assert.NotNil(t, limiter)

	for i := 0; i < 10; i++ {
		assert.True(t, limiter.Allow(), "access should be allowed")
	}
	assert.False(t, limiter.Allow(), "access should be rejected")
}

func TestConcurrent10QPS(t *testing.T) {
	assert.NotNil(t, Client(), "redis client should not be empty")

	var count = 5
	var limiters []*Limiter

	for i := 0; i < count; i++ {
		limiters = append(limiters, NewLimiter(Every(100*time.Millisecond), 10, "TestConcurrent10QPS"))
		assert.NotNil(t, limiters[i])
	}

	var wg sync.WaitGroup
	wg.Add(count)

	var l sync.Mutex
	totalAllows := 0

	for i := 0; i < count; i++ {
		go func(index int) {
			for j := 0; j < 10; j++ {
				if limiters[index].Allow() {
					l.Lock()
					totalAllows++
					l.Unlock()
				}
			}
			wg.Done()
		}(i)
	}

	wg.Wait()
	assert.Equal(t, 10, totalAllows)
}
