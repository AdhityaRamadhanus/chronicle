package redis

import (
	"time"

	"github.com/go-redis/redis"
)

//CacheService implements chronicle.CacheService interface using redis
type CacheService struct {
	redisClient *redis.Client
}

//NewCacheService construct a new CacheService from redis client
func NewCacheService(redisClient *redis.Client) *CacheService {
	return &CacheService{
		redisClient: redisClient,
	}
}

//Get a cache in bytes from a key
func (c CacheService) Get(key string) (result []byte, err error) {
	return c.redisClient.Get(key).Bytes()
}

//Set cache in bytes with key without expiration
func (c CacheService) Set(key string, value []byte) (err error) {
	return c.redisClient.Set(key, string(value), 0).Err()
}

//SetEx cache in bytes with key with expiration
func (c CacheService) SetEx(key string, value []byte, expiration time.Duration) (err error) {
	return c.redisClient.Set(key, value, expiration).Err()
}
