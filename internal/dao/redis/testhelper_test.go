package redis

import (
	"context"
	"fmt"
	"os"
	"testing"

	goredis "github.com/redis/go-redis/v9"

	"redisolar-go/internal/keyschema"
)

const testKeyPrefix = "ru102py-test"

func testRedisClient(t *testing.T) *goredis.Client {
	t.Helper()
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "redis-19256.redis.alldataint.com"
	}
	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "19256"
	}
	client := goredis.NewClient(&goredis.Options{
		Addr: fmt.Sprintf("%s:%s", host, port),
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		t.Fatalf("Cannot connect to Redis at %s:%s: %v", host, port, err)
	}
	return client
}

func testKeySchema() *keyschema.KeySchema {
	return keyschema.New(testKeyPrefix)
}

func testBase(t *testing.T) RedisDao {
	t.Helper()
	return RedisDao{
		Client:    testRedisClient(t),
		KeySchema: testKeySchema(),
	}
}

// cleanupTestKeys deletes all keys matching the test prefix after each test.
func cleanupTestKeys(t *testing.T, client *goredis.Client) {
	t.Helper()
	t.Cleanup(func() {
		ctx := context.Background()
		iter := client.Scan(ctx, 0, testKeyPrefix+":*", 0).Iterator()
		for iter.Next(ctx) {
			client.Del(ctx, iter.Val())
		}
	})
}
