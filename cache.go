package chronicle

import (
	"time"
)

type CacheKeyBuilder interface {
	BuildKey(obj interface{}) string
}

type CacheService interface {
	Set(key string, value []byte) error
	SetEx(key string, value []byte, expirationInSeconds time.Duration) error
	Get(key string) ([]byte, error)
}
