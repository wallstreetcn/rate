Golang rate limiter for distributed system
======
[![Build Status](https://travis-ci.org/wallstreetcn/rate.svg?branch=master)](https://travis-ci.org/wallstreetcn/rate)
[![Coverage Status](https://coveralls.io/repos/github/wallstreetcn/rate/badge.svg?branch=master)](https://coveralls.io/github/wallstreetcn/rate?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/wallstreetcn/rate)](https://goreportcard.com/report/github.com/wallstreetcn/rate)
[![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/wallstreetcn/rate/master/LICENSE)


## Implementation
According to [Stripe's rate-limiters practice](https://stripe.com/blog/rate-limiters), use `Redis Server` & `Lua Script` to implement a rate limiter based on [token bucket algorithm](https://en.wikipedia.org/wiki/Token_bucket).
在github.com/wallstreetcn/rate的基础上适配了codis和最新版本的redis客户端。

## Install
~~go get "github.com/wallstreetcn/rate"~~
```shell
go get "github.com/bigo-sg/rate"
```

## Usage
```go
import (
    rate "github.com/bigo-sg/rate/redis"
)

// initialize redis.
rate.SetRedis(&rate.ConfigRedis{
    Host: "127.0.0.1",
    Port: 6379,
    Auth: "",
}

// setup a 1 ops/s rate limiter.
limiter := rate.NewLimiter(Every(time.Second), 1, "a-sample-operation")
if limiter.Allow() {
    // serve the user request
} else {
    // reject the user request
}

// setup a 1000 ops/s rate limiter.
limiter := rate.NewLimiter(Every(time.Second/time.Duration(1000)), 1000, "a-sample-operation")
if limiter.Allow() {
    // serve the user request
} else {
    // reject the user request
}
```
