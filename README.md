# rate
Golang rate limiter for distributed system

## Implementation
According to [Stripe's rate-limiters practice](https://stripe.com/blog/rate-limiters), use `Redis Server` & `Lua Script` to implement a rate limiter based on [token bucket algorithm](https://en.wikipedia.org/wiki/Token_bucket).

## Install
```shell
go get "github.com/wallstreetcn/rate"
```

## Usage
```go
// initialize redis.
SetClient(NewRedisClient(ConfigRedis{
    Host: "127.0.0.1",
    Port: 6379,
    Auth: "",
}))

// setup a 1 ops/s rate limiter.
limiter := NewLimiter(Every(time.Second), 1, "a-sample-operation")
if limiter.Allow() {
    // serve the user request
} else {
    // reject the user request
}
```
