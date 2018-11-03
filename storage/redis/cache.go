package redis

import (
	"time"

	"github.com/go-redis/redis"
)

type CacheService struct {
	redisClient *redis.Client
}

func NewCacheService(redisClient *redis.Client) *CacheService {
	return &CacheService{
		redisClient: redisClient,
	}
}

func (c CacheService) Get(key string) (result []byte, err error) {
	return c.redisClient.Get(key).Bytes()
}

func (c CacheService) Set(key string, value []byte) (err error) {
	return c.redisClient.Set(key, string(value), 0).Err()
}

func (c CacheService) SetEx(key string, value []byte, expiration time.Duration) (err error) {
	return c.redisClient.Set(key, value, expiration).Err()
}
