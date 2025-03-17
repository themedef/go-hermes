package contracts

import (
	"context"
	"github.com/themedef/go-hermes/internal/logger"
)

type Store interface {
	Set(ctx context.Context, key string, value interface{}, ttl int) error
	SetNX(ctx context.Context, key string, value interface{}, ttl int) (bool, error)
	SetXX(ctx context.Context, key string, value interface{}, ttl int) (bool, error)
	Get(ctx context.Context, key string) (interface{}, bool, error)
	Delete(ctx context.Context, key string) (bool, error)
	SetCAS(ctx context.Context, key string, oldValue, newValue interface{}, ttl int) (bool, error)
	Incr(ctx context.Context, key string) (int64, bool, error)
	Decr(ctx context.Context, key string) (int64, bool, error)
	LPush(ctx context.Context, key string, value interface{}) error
	RPush(ctx context.Context, key string, value interface{}) error
	LPop(ctx context.Context, key string) (interface{}, bool, error)
	RPop(ctx context.Context, key string) (interface{}, bool, error)
	HSet(ctx context.Context, key string, field string, value interface{}, ttl int) error
	HGet(ctx context.Context, key string, field string) (interface{}, bool, error)
	HDel(ctx context.Context, key string, field string) error
	HGetAll(ctx context.Context, key string) (map[string]interface{}, error)
	FindByValue(ctx context.Context, value interface{}) ([]string, error)
	UpdateTTL(ctx context.Context, key string, ttl int) error
	FlushAll(ctx context.Context) error
	Subscribe(key string) chan string
	Unsubscribe(key string, ch chan string)
	Publish(key, message string)
	ClosePubSub()
	Logger() *logger.Logger
}
