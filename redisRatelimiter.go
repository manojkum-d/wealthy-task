package main

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisTokenBucketLimiter struct {
	client     *redis.Client
	key        string
	maxTokens  int
	interval   time.Duration
	leakRate   float64
	luaScript  *redis.Script
}

func NewRedisTokenBucketLimiter(client *redis.Client, key string, maxTokens int, interval time.Duration) *RedisTokenBucketLimiter {
	// Lua token bucket script
	lua := redis.NewScript(`
local key = KEYS[1]
local max_tokens = tonumber(ARGV[1])
local interval = tonumber(ARGV[2])
local now = tonumber(ARGV[3])

local data = redis.call("HMGET", key, "tokens", "lastRefill")
local tokens = tonumber(data[1]) or max_tokens
local lastRefill = tonumber(data[2]) or now

-- refill
if now - lastRefill >= interval then
	tokens = max_tokens
	lastRefill = now
end

if tokens > 0 then
	tokens = tokens - 1
	redis.call("HMSET", key, "tokens", tokens, "lastRefill", lastRefill)
	return 1
else
	return 0
end
`)

	return &RedisTokenBucketLimiter{
		client:    client,
		key:       key,
		maxTokens: maxTokens,
		interval:  interval,
		luaScript: lua,
	}
}

func (r *RedisTokenBucketLimiter) RedisCheck(ctx context.Context) error {
	now := time.Now().Unix()
	res, err := r.luaScript.Run(ctx, r.client, []string{r.key},
		r.maxTokens,
		int64(r.interval.Seconds()),
		now,
	).Result()
	if err != nil {
		return err
	}

	if res.(int64) == 1 {
		// Token available, allow immediately
		return nil
	}

	// No token available, log once and wait until interval
	log.Printf("Rate limit reached, waiting %v before retrying...", r.interval)
	select {
	case <-time.After(r.interval):
	case <-ctx.Done():
		return ctx.Err()
	}

	// Retry after wait
	return r.RedisCheck(ctx)
}

