package redis

import (
	"github.com/redis/go-redis/v9"

	"redisolar-go/internal/keyschema"
)

type RedisDao struct {
	Client    *redis.Client
	KeySchema *keyschema.KeySchema
}

func NewRedisDao(client *redis.Client, ks *keyschema.KeySchema) RedisDao {
	return RedisDao{Client: client, KeySchema: ks}
}
