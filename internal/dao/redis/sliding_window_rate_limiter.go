package redis

import (
	"context"
	// Uncomment for Challenge #7
	// "fmt"
	// "math/rand"
	// "time"

	// Uncomment for Challenge #7
	// "redisolar-go/internal/dao"
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

	// START Challenge #7
	// Implement a sliding window rate limiter using a sorted set.
	//
	// Steps:
	// 1. Calculate the current time in milliseconds:
	//    now := time.Now().UTC()
	//    nowMs := float64(now.UnixNano()) / 1e6
	//    windowStart := nowMs - rl.windowSizeMs
	//
	// 2. Create a unique member string:
	//    member := fmt.Sprintf("%f-%f", nowMs, rand.Float64())
	//
	// 3. Use a pipeline with three commands:
	//    a. ZADD: Add the member with score = nowMs
	//    b. ZREMRANGEBYSCORE: Remove entries from 0 to windowStart
	//    c. ZCARD: Count the remaining entries
	//
	// 4. Execute the pipeline and check the ZCARD result.
	//    If count > rl.maxHits, return dao.ErrRateLimitExceeded.
	//
	// Hint: Use rl.client.Pipeline() to create a pipeline
	// Hint: Use goredis.Z{Score: nowMs, Member: member} for ZADD
	// Hint: Use fmt.Sprintf("%f", windowStart) for the max score in ZREMRANGEBYSCORE
	_ = key // TODO: remove after implementing
	return nil
	// END Challenge #7
}
