package contracts

import "context"

type TransactionHandler interface {
	Commit() error
	Rollback() error

	Get(ctx context.Context, key string) (interface{}, error)
	Exists(ctx context.Context, key string) (bool, error)
	HGet(ctx context.Context, key, field string) (interface{}, error)
	HGetAll(ctx context.Context, key string) (map[string]interface{}, error)
	GetWithDetails(ctx context.Context, key string) (interface{}, int, error)
	FindByValue(ctx context.Context, value interface{}) ([]string, error)

	Set(ctx context.Context, key string, value interface{}, ttl int) error
	SetNX(ctx context.Context, key string, value interface{}, ttl int) error
	SetXX(ctx context.Context, key string, value interface{}, ttl int) error
	SetCAS(ctx context.Context, key string, oldValue, newValue interface{}, ttl int) error
	Delete(ctx context.Context, key string) error
	Incr(ctx context.Context, key string) error
	Decr(ctx context.Context, key string) error

	LPush(ctx context.Context, key string, values ...interface{}) error
	RPush(ctx context.Context, key string, values ...interface{}) error
	LPop(ctx context.Context, key string) (interface{}, error)
	RPop(ctx context.Context, key string) (interface{}, error)

	UpdateTTL(ctx context.Context, key string, ttl int) error
	HSet(ctx context.Context, key, field string, value interface{}, ttl int) error
	HDel(ctx context.Context, key, field string) error

	LLen(ctx context.Context, key string) (int, error)
	LRange(ctx context.Context, key string, start, end int) ([]interface{}, error)
	HExists(ctx context.Context, key, field string) (bool, error)
	HLen(ctx context.Context, key string) (int, error)
	Type(ctx context.Context, key string) (interface{}, error)
	Rename(ctx context.Context, oldKey, newKey string) error
}
