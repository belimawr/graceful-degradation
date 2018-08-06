package cache

import (
	"context"
	"errors"
)

// ErrNotFound - this error is returned if the requested key
// is not present on cache
var ErrNotFound = errors.New("key not found")

// Cache - Key/Value cache interface
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte) error
}
