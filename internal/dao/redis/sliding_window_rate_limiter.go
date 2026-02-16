package redis

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"redisolar-go/internal/dao"
	"redisolar-go/internal/keyschema"

	goredis "github.com/redis/go-redis/v9"
)

type SlidingWindowRateLimiter struct {
	client       *goredis.Client
	keySchema    *keyschema.KeySchema
	windowSizeMs float64
	maxHits      int
}

func NewSlidingWindowRateLimiter(client *goredis.Client, ks *keyschema.KeySchema, windowSizeMs float64, maxHits int) *SlidingWindowRateLimiter {
	return &SlidingWindowRateLimiter{
		client:       client,
		keySchema:    ks,
		windowSizeMs: windowSizeMs,
		maxHits:      maxHits,
	}
}

func (rl *SlidingWindowRateLimiter) Hit(ctx context.Context, name string) error {
	key := rl.keySchema.SlidingWindowRateLimiterKey(name, int(rl.windowSizeMs), rl.maxHits)

	now := time.Now().UTC()
	nowMs := float64(now.UnixNano()) / 1e6
	windowStart := nowMs - rl.windowSizeMs
	member := fmt.Sprintf("%f-%f", nowMs, rand.Float64())

	pipe := rl.client.Pipeline()
	pipe.ZAdd(ctx, key, goredis.Z{Score: nowMs, Member: member})
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%f", windowStart))
	cardCmd := pipe.ZCard(ctx, key)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}

	count, _ := cardCmd.Result()
	if count > int64(rl.maxHits) {
		return dao.ErrRateLimitExceeded
	}
	return nil
}
