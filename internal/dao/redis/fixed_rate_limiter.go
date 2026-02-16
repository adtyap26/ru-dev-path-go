package redis

import (
	"context"
	"time"

	"redisolar-go/internal/dao"
	"redisolar-go/internal/keyschema"

	goredis "github.com/redis/go-redis/v9"
)

type FixedRateLimiter struct {
	client     *goredis.Client
	keySchema  *keyschema.KeySchema
	interval   int // minute interval
	expiration int // seconds
	maxHits    int
}

func NewFixedRateLimiter(client *goredis.Client, ks *keyschema.KeySchema, intervalMinutes int, maxHits int) *FixedRateLimiter {
	return &FixedRateLimiter{
		client:     client,
		keySchema:  ks,
		interval:   intervalMinutes,
		expiration: intervalMinutes * 60,
		maxHits:    maxHits,
	}
}

func (rl *FixedRateLimiter) Hit(ctx context.Context, name string) error {
	now := time.Now()
	minuteOfDay := now.Hour()*60 + now.Minute()
	minuteBlock := minuteOfDay / rl.interval
	key := rl.keySchema.FixedRateLimiterKey(name, minuteBlock, rl.maxHits)

	pipe := rl.client.Pipeline()
	incrCmd := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, time.Duration(rl.expiration)*time.Second)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}

	hits, _ := incrCmd.Result()
	if hits > int64(rl.maxHits) {
		return dao.ErrRateLimitExceeded
	}
	return nil
}
