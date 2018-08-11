package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis"
	"github.com/rs/zerolog"
)

type redisCache struct {
	client redis.UniversalClient
}

// New returns a Redis implementation of Cache
func New(addr string, timeout time.Duration) Cache {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
		DialTimeout:  timeout / 2,
		PoolTimeout:  timeout / 2,
	})

	return redisCache{
		client: client,
	}

}
func (r redisCache) Get(ctx context.Context, key string) ([]byte, error) {
	logger := zerolog.Ctx(ctx).
		With().
		Str("_function", "redisCache.Get").
		Logger()

	value, err := r.client.Get(key).Bytes()
	if err == redis.Nil {
		return []byte{}, ErrNotFound
	}

	if err != nil {
		logger.Debug().Err(err).Msgf("reading from cache: %q", key)
		return []byte{}, err
	}

	return value, nil
}

func (r redisCache) Set(ctx context.Context, key string, value []byte) error {
	logger := zerolog.Ctx(ctx).
		With().
		Str("_function", "redisCache.Set").
		Logger()

	err := r.client.Set(key, value, 0).Err()
	if err != nil {
		logger.Debug().Err(err).Msgf("could not set key %q from cache", key)
		return err
	}

	return nil
}
