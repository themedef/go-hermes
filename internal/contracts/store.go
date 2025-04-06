package contracts

import (
	"context"
)

type StoreHandler interface {
	Set(ctx context.Context, key string, value interface{}, ttl int) error
	SetNX(ctx context.Context, key string, value interface{}, ttl int) (bool, error)
	SetXX(ctx context.Context, key string, value interface{}, ttl int) (bool, error)
	Get(ctx context.Context, key string) (interface{}, error)
	SetCAS(ctx context.Context, key string, oldVal, newVal interface{}, ttl int) error
	GetSet(ctx context.Context, key string, newValue interface{}, ttl int) (interface{}, error)
	Incr(ctx context.Context, key string) (int64, error)
	Decr(ctx context.Context, key string) (int64, error)
	IncrBy(ctx context.Context, key string, increment int64) (int64, error)
	DecrBy(ctx context.Context, key string, decrement int64) (int64, error)
	LPush(ctx context.Context, key string, values ...interface{}) error
	RPush(ctx context.Context, key string, values ...interface{}) error
	LPop(ctx context.Context, key string) (interface{}, error)
	RPop(ctx context.Context, key string) (interface{}, error)
	LLen(ctx context.Context, key string) (int, error)
	LRange(ctx context.Context, key string, start, end int) ([]interface{}, error)
	HSet(ctx context.Context, key string, field string, value interface{}, ttl int) error
	HGet(ctx context.Context, key string, field string) (interface{}, error)
	HDel(ctx context.Context, key string, field string) error
	HGetAll(ctx context.Context, key string) (map[string]interface{}, error)
	HExists(ctx context.Context, key string, field string) (bool, error)
	HLen(ctx context.Context, key string) (int, error)
	Exists(ctx context.Context, key string) (bool, error)
	Expire(ctx context.Context, key string, ttl int) (bool, error)
	Persist(ctx context.Context, key string) (bool, error)
	Type(ctx context.Context, key string) (interface{}, error)
	GetWithDetails(ctx context.Context, key string) (interface{}, int, error)
	Rename(ctx context.Context, oldKey, newKey string) error
	FindByValue(ctx context.Context, value interface{}) ([]string, error)
	Delete(ctx context.Context, key string) error
	DropAll(ctx context.Context) error
	Subscribe(key string) chan string
	Unsubscribe(key string, ch chan string)
	ListSubscriptions() []string
	CloseAllSubscriptionsForKey(key string)
	Logger() LoggerHandler
	Commands() CommandsHandler
	Transaction() TransactionHandler
	Close() error
}
