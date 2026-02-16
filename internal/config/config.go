package config

import "os"

type Config struct {
	RedisHost      string
	RedisPort      string
	RedisKeyPrefix string
	RedisUsername  string
	RedisPassword  string
	UseGeoSiteAPI  bool
	ServerPort     string
}

func Load() Config {
	c := Config{
		RedisHost:      getEnv("REDIS_HOST", "redis-19256.redis.alldataint.com"),
		RedisPort:      getEnv("REDIS_PORT", "19256"),
		RedisKeyPrefix: getEnv("REDIS_KEY_PREFIX", "ru102py-app"),
		RedisUsername:  os.Getenv("REDISOLAR_REDIS_USERNAME"),
		RedisPassword:  os.Getenv("REDISOLAR_REDIS_PASSWORD"),
		UseGeoSiteAPI:  getEnv("USE_GEO_SITE_API", "true") == "true",
		ServerPort:     getEnv("SERVER_PORT", "8081"),
	}
	return c
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
