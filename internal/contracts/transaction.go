package contracts

import "context"

type TransactionHandler interface {
	Commit() error
	Rollback() error
	Set(ctx context.Context, key string, value interface{}, ttl int) error
	SetNX(ctx context.Context, key string, value interface{}, ttl int) error
	SetXX(ctx context.Context, key string, value interface{}, ttl int) error
	Get(ctx context.Context, key string) (interface{}, error)
	SetCAS(ctx context.Context, key string, oldValue, newValue interface{}, ttl int) error
	GetSet(ctx context.Context, key string, newValue interface{}, ttl int) (interface{}, error)
	Incr(ctx context.Context, key string) error
	Decr(ctx context.Context, key string) error
	IncrBy(ctx context.Context, key string, increment int64) error
	DecrBy(ctx context.Context, key string, decrement int64) error
	LPush(ctx context.Context, key string, values ...interface{}) error
	RPush(ctx context.Context, key string, values ...interface{}) error
	LPop(ctx context.Context, key string) (interface{}, error)
	RPop(ctx context.Context, key string) (interface{}, error)
	LLen(ctx context.Context, key string) (int, error)
	LRange(ctx context.Context, key string, start, end int) ([]interface{}, error)
	HSet(ctx context.Context, key, field string, value interface{}, ttl int) error
	HGet(ctx context.Context, key, field string) (interface{}, error)
	HDel(ctx context.Context, key, field string) error
	HGetAll(ctx context.Context, key string) (map[string]interface{}, error)
	HExists(ctx context.Context, key, field string) (bool, error)
	HLen(ctx context.Context, key string) (int, error)
	Exists(ctx context.Context, key string) (bool, error)
	Expire(ctx context.Context, key string, ttl int) error
	Persist(ctx context.Context, key string) error
	Type(ctx context.Context, key string) (interface{}, error)
	GetWithDetails(ctx context.Context, key string) (interface{}, int, error)
	Rename(ctx context.Context, oldKey, newKey string) error
	FindByValue(ctx context.Context, value interface{}) ([]string, error)
	Delete(ctx context.Context, key string) error
}
