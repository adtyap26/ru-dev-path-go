package redis

import (
	"context"
	"errors"
	"testing"
	"time"

	"redisolar-go/internal/dao"
)

const tenSecondsMs = 10 * 1000

// Challenge #7
func TestSlidingWindow_WithinLimitInsideWindow(t *testing.T) {
	t.Skip("Remove for Challenge #7")

	client := testRedisClient(t)
	cleanupTestKeys(t, client)
	ks := testKeySchema()
	limiter := NewSlidingWindowRateLimiter(client, ks, tenSecondsMs, 10)
	ctx := context.Background()

	exceptionCount := 0
	for i := 0; i < 10; i++ {
		err := limiter.Hit(ctx, "foo")
		if errors.Is(err, dao.ErrRateLimitExceeded) {
			exceptionCount++
		} else if err != nil {
			t.Fatalf("Hit failed: %v", err)
		}
	}

	if exceptionCount != 0 {
		t.Fatalf("expected 0 rate limit exceptions, got %d", exceptionCount)
	}
}

func TestSlidingWindow_ExceedsLimitInsideWindow(t *testing.T) {
	t.Skip("Remove for Challenge #7")

	client := testRedisClient(t)
	cleanupTestKeys(t, client)
	ks := testKeySchema()
	limiter := NewSlidingWindowRateLimiter(client, ks, tenSecondsMs, 10)
	ctx := context.Background()

	exceptionCount := 0
	for i := 0; i < 12; i++ {
		err := limiter.Hit(ctx, "foo")
		if errors.Is(err, dao.ErrRateLimitExceeded) {
			exceptionCount++
		} else if err != nil {
			t.Fatalf("Hit failed: %v", err)
		}
	}

	if exceptionCount != 2 {
		t.Fatalf("expected 2 rate limit exceptions, got %d", exceptionCount)
	}
}

func TestSlidingWindow_ExceedsLimitOutsideWindow(t *testing.T) {
	t.Skip("Remove for Challenge #7")

	client := testRedisClient(t)
	cleanupTestKeys(t, client)
	ks := testKeySchema()
	// 100ms window
	limiter := NewSlidingWindowRateLimiter(client, ks, 100, 10)
	ctx := context.Background()

	for i := 0; i < 10; i++ {
		if err := limiter.Hit(ctx, "foo"); err != nil {
			t.Fatalf("Hit failed: %v", err)
		}
	}

	// Sleep to let the window move and thus allow an 11th request.
	time.Sleep(1 * time.Second)

	raised := false
	err := limiter.Hit(ctx, "foo")
	if errors.Is(err, dao.ErrRateLimitExceeded) {
		raised = true
	} else if err != nil {
		t.Fatalf("Hit failed: %v", err)
	}

	if raised {
		t.Fatal("expected no rate limit exception after window passed")
	}
}
