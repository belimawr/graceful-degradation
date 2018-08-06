package cache

import (
	"context"

	"github.com/go-redis/redis"
	"github.com/rs/zerolog"
)

type redisCache struct {
	client redis.UniversalClient
}

// New returns a Redis implementation of Cache
func New(addr string) Cache {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	return redisCache{
		client: client,
	}

}
func (r redisCache) Get(ctx context.Context, key string) ([]byte, error) {
	logger := zerolog.Ctx(ctx)

	value, err := r.client.Get(key).Bytes()
	if err == redis.Nil {
		return []byte{}, ErrNotFound
	}

	if err != nil {
		logger.Error().Err(err).Msg("reding from cache")
		return []byte{}, err
	}

	return value, nil
}

func (r redisCache) Set(ctx context.Context, key string, value []byte) error {
	logger := zerolog.Ctx(ctx)

	err := r.client.Set(key, value, 0).Err()
	if err != nil {
		logger.Error().Err(err).Msgf("could not set key %q from cache", key)
		return err
	}

	return nil
}
