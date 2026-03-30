package filter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisSlidingWindow uses a Redis sorted set for a sliding window counter.
type RedisSlidingWindow struct {
	rdb    *redis.Client
	prefix string
	limit  int
	window time.Duration
}

// NewRedisSlidingWindow creates a Redis-backed limiter.
func NewRedisSlidingWindow(rdb *redis.Client, keyPrefix string, limit int, window time.Duration) *RedisSlidingWindow {
	if keyPrefix == "" {
		keyPrefix = "nexus:rl:"
	}
	return &RedisSlidingWindow{rdb: rdb, prefix: keyPrefix, limit: limit, window: window}
}

func (r *RedisSlidingWindow) Allow(key string) bool {
	if r.rdb == nil {
		return true
	}
	k := r.prefix + key
	now := float64(time.Now().UnixNano())
	cutoff := now - float64(r.window.Nanoseconds())
	ctx := context.Background()
	if err := r.rdb.ZRemRangeByScore(ctx, k, "-inf", fmt.Sprintf("%f", cutoff)).Err(); err != nil {
		return true
	}
	n, err := r.rdb.ZCard(ctx, k).Result()
	if err != nil {
		return true
	}
	if int(n) >= r.limit {
		return false
	}
	member := fmt.Sprintf("%f", now)
	if err := r.rdb.ZAdd(ctx, k, redis.Z{Score: now, Member: member}).Err(); err != nil {
		return true
	}
	_ = r.rdb.Expire(ctx, k, r.window+time.Second).Err()
	return true
}
