package hermes

import (
	"context"
	"fmt"
	"hash/fnv"
	"log"
	"sync"
	"time"

	"github.com/themedef/go-hermes/internal/contracts"
	"github.com/themedef/go-hermes/internal/logger"
	"github.com/themedef/go-hermes/internal/pubsub"
)

type DataType int

const (
	String DataType = iota
	List
	Hash
)

type Config struct {
	ShardCount       int
	CleanupInterval  time.Duration
	EnableLogging    bool
	LogFile          string
	LogBufferSize    int
	MinLevel         logger.LogLevel
	PubSubBufferSize int
}

type Entry struct {
	Value      interface{}
	Type       DataType
	Expiration time.Time
}

type shard struct {
	mu   sync.RWMutex
	data map[string]Entry
}

type DB struct {
	shards        []*shard
	logger        contracts.LoggerHandler
	pubsub        contracts.PubSubHandler
	config        Config
	transaction   *Transaction
	commands      contracts.CommandsHandler
	cleanupCtx    context.Context
	cleanupCancel context.CancelFunc
}

func NewStore(config Config) contracts.StoreHandler {
	if config.CleanupInterval == 0 {
		config.CleanupInterval = time.Second
	}

	if config.ShardCount < 1 {
		config.ShardCount = 1
	}

	dbLogger, err := logger.NewLogger(logger.Config{
		LogFile:    config.LogFile,
		Enabled:    config.EnableLogging,
		BufferSize: config.LogBufferSize,
		MinLevel:   config.MinLevel,
	})
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	cleanupCtx, cleanupCancel := context.WithCancel(context.Background())

	shards := make([]*shard, config.ShardCount)
	for i := 0; i < config.ShardCount; i++ {
		shards[i] = &shard{
			data: make(map[string]Entry),
		}
	}

	ps := pubsub.NewPubSub(pubsub.Config{BufferSize: config.PubSubBufferSize})

	db := &DB{
		shards:        shards,
		logger:        dbLogger,
		pubsub:        ps,
		config:        config,
		cleanupCtx:    cleanupCtx,
		cleanupCancel: cleanupCancel,
	}
	db.commands = NewCommandAPI(db)

	go db.cleanupExpiredKeys(config.CleanupInterval)

	return db
}

func (db *DB) getShardIndex(key string) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))

	sum := h.Sum32()
	shardIndex := sum % uint32(len(db.shards))

	return int(shardIndex)
}
func ttlSecondsToTime(ttl int) (time.Time, error) {
	if ttl < 0 {
		return time.Time{}, ErrInvalidTTL
	}
	if ttl == 0 {
		return time.Time{}, nil
	}
	return time.Now().Add(time.Duration(ttl) * time.Second), nil
}

func isExpired(e Entry) bool {
	if e.Expiration.IsZero() {
		return false
	}
	return time.Now().After(e.Expiration)
}

func (db *DB) setInternal(ctx context.Context, key string, value interface{}, ttl int, ifExists, ifNotExists bool) (bool, error) {
	select {
	case <-ctx.Done():
		db.logger.Warn("setInternal operation canceled", "key", key)
		return false, ErrContextCanceled
	default:
	}

	if key == "" {
		db.logger.Error("empty key provided")
		return false, ErrInvalidKey
	}

	expiration, err := ttlSecondsToTime(ttl)
	if err != nil {
		db.logger.Error("invalid TTL value in setInternal",
			"key", key,
			"ttl", ttl,
			"error", err)
		return false, err
	}

	sh := db.shards[db.getShardIndex(key)]
	sh.mu.Lock()
	defer sh.mu.Unlock()

	_, exists := sh.data[key]

	if ifExists && !exists {
		db.logger.Warn("key does not exist for XX operation", "key", key)
		return false, ErrKeyNotFound
	}
	if ifNotExists && exists {
		db.logger.Warn("key already exists for NX operation", "key", key)
		return false, ErrKeyExists
	}

	newEntry := Entry{
		Value:      value,
		Expiration: expiration,
		Type:       String,
	}
	sh.data[key] = newEntry

	db.logger.Info("key set successfully",
		"key", key,
		"value", value,
		"ttl", ttl)
	db.pubsub.Publish(key, fmt.Sprintf("SET: %v", value))

	return true, nil
}

func (db *DB) Set(ctx context.Context, key string, value interface{}, ttl int) error {
	_, err := db.setInternal(ctx, key, value, ttl, false, false)
	return err
}

func (db *DB) SetNX(ctx context.Context, key string, value interface{}, ttl int) (bool, error) {
	return db.setInternal(ctx, key, value, ttl, false, true)
}

func (db *DB) SetXX(ctx context.Context, key string, value interface{}, ttl int) (bool, error) {
	return db.setInternal(ctx, key, value, ttl, true, false)
}

func (db *DB) Get(ctx context.Context, key string) (interface{}, error) {
	select {
	case <-ctx.Done():
		db.logger.Warn("Get operation canceled", "key", key)
		return nil, ErrContextCanceled
	default:
	}

	sh := db.shards[db.getShardIndex(key)]
	sh.mu.RLock()
	entry, exists := sh.data[key]
	sh.mu.RUnlock()

	if !exists {
		db.logger.Warn("attempt to Get a non-existent key", "key", key)
		return nil, ErrKeyNotFound
	}

	if isExpired(entry) {
		db.logger.Info("key expired in Get", "key", key)

		sh.mu.Lock()
		if latestEntry, ok := sh.data[key]; ok && isExpired(latestEntry) {
			delete(sh.data, key)
		}
		sh.mu.Unlock()

		return nil, ErrKeyExpired
	}

	db.logger.Info("Get operation successful", "key", key)
	return entry.Value, nil
}

func (db *DB) SetCAS(ctx context.Context, key string, oldValue, newValue interface{}, ttl int) error {
	select {
	case <-ctx.Done():
		db.logger.Warn("SetCAS operation canceled", "key", key)
		return ErrContextCanceled
	default:
	}

	expiration, err := ttlSecondsToTime(ttl)
	if err != nil {
		db.logger.Error("invalid TTL value in SetCAS",
			"key", key,
			"ttl", ttl,
			"error", err)
		return err
	}

	sh := db.shards[db.getShardIndex(key)]
	sh.mu.Lock()
	defer sh.mu.Unlock()

	entry, exists := sh.data[key]
	if !exists || isExpired(entry) {
		if exists {
			delete(sh.data, key)
			db.logger.Info("auto-removed expired key in SetCAS", "key", key)
		}
		db.logger.Warn("key not found or expired in SetCAS", "key", key)
		return fmt.Errorf("%w: SetCAS", ErrKeyNotFound)
	}

	if entry.Value != oldValue {
		db.logger.Warn("value mismatch in SetCAS",
			"key", key,
			"expected", oldValue,
			"actual", entry.Value)
		return fmt.Errorf("%w: SetCAS", ErrValueMismatch)
	}

	newEntry := Entry{
		Value:      newValue,
		Expiration: expiration,
		Type:       entry.Type,
	}
	sh.data[key] = newEntry

	db.logger.Info("CAS update successful",
		"key", key,
		"old_value", oldValue,
		"new_value", newValue,
		"ttl", ttl)
	db.pubsub.Publish(key, fmt.Sprintf("CAS: %v -> %v", oldValue, newValue))

	return nil
}

func (db *DB) GetSet(ctx context.Context, key string, newValue interface{}, ttl int) (interface{}, error) {
	select {
	case <-ctx.Done():
		db.logger.Warn("GetSet operation canceled", "key", key)
		return nil, ErrContextCanceled
	default:
	}

	expiration, err := ttlSecondsToTime(ttl)
	if err != nil {
		db.logger.Error("invalid TTL value in GetSet", "key", key, "ttl", ttl, "error", err)
		return nil, err
	}

	sh := db.shards[db.getShardIndex(key)]
	sh.mu.Lock()
	defer sh.mu.Unlock()

	entry, exists := sh.data[key]

	if exists && isExpired(entry) {
		delete(sh.data, key)
		exists = false
		db.logger.Info("GetSet removed expired key", "key", key)
	}

	var oldValue interface{}
	if exists {
		oldValue = entry.Value
	} else {
		oldValue = nil
	}

	newEntry := Entry{
		Value:      newValue,
		Type:       String,
		Expiration: expiration,
	}

	sh.data[key] = newEntry
	db.logger.Info("GetSet operation successful", "key", key, "oldValue", oldValue, "newValue", newValue, "ttl", ttl)
	db.pubsub.Publish(key, fmt.Sprintf("GETSET: %v -> %v", oldValue, newValue))

	return oldValue, nil
}

func (db *DB) Incr(ctx context.Context, key string) (int64, error) {
	select {
	case <-ctx.Done():
		db.logger.Warn("Incr operation canceled", "key", key)
		return 0, ErrContextCanceled
	default:
	}

	sh := db.shards[db.getShardIndex(key)]
	sh.mu.Lock()
	defer sh.mu.Unlock()

	entry, exists := sh.data[key]
	if !exists || isExpired(entry) {
		sh.data[key] = Entry{Value: int64(1), Type: String}
		db.logger.Info("Incr created new key with value=1", "key", key)
		return 1, nil
	}

	val, ok := entry.Value.(int64)
	if !ok {
		db.logger.Error("Incr failed: value is not an int64", "key", key)
		return 0, ErrInvalidValueType
	}

	val++
	entry.Value = val
	sh.data[key] = entry

	db.logger.Info("Incr operation successful", "key", key, "newVal", val)
	return val, nil
}

func (db *DB) Decr(ctx context.Context, key string) (int64, error) {
	select {
	case <-ctx.Done():
		db.logger.Warn("Decr operation canceled", "key", key)
		return 0, ErrContextCanceled
	default:
	}

	sh := db.shards[db.getShardIndex(key)]
	sh.mu.Lock()
	defer sh.mu.Unlock()

	entry, exists := sh.data[key]
	if !exists || isExpired(entry) {
		sh.data[key] = Entry{Value: int64(-1), Type: String}
		db.logger.Info("Decr created new key with value=-1", "key", key)
		return -1, nil
	}

	val, ok := entry.Value.(int64)
	if !ok {
		db.logger.Error("Decr failed: value is not an int64", "key", key)
		return 0, ErrInvalidValueType
	}

	val--
	entry.Value = val
	sh.data[key] = entry

	db.logger.Info("Decr operation successful", "key", key, "newVal", val)
	return val, nil
}

func (db *DB) IncrBy(ctx context.Context, key string, increment int64) (int64, error) {
	select {
	case <-ctx.Done():
		db.logger.Warn("IncrBy operation canceled", "key", key)
		return 0, ErrContextCanceled
	default:
	}

	sh := db.shards[db.getShardIndex(key)]
	sh.mu.Lock()
	defer sh.mu.Unlock()

	entry, exists := sh.data[key]
	if !exists || isExpired(entry) {
		newVal := increment
		sh.data[key] = Entry{Value: newVal, Type: String}
		db.logger.Info("IncrBy created key", "key", key, "value", newVal)
		return newVal, nil
	}

	if entry.Type != String {
		db.logger.Error("IncrBy type mismatch", "key", key, "type", entry.Type)
		return 0, ErrInvalidType
	}

	current, ok := entry.Value.(int64)
	if !ok {
		db.logger.Error("IncrBy value not int64", "key", key)
		return 0, ErrInvalidValueType
	}

	current += increment
	entry.Value = current
	sh.data[key] = entry
	db.logger.Info("IncrBy success", "key", key, "newValue", current)
	return current, nil
}

func (db *DB) DecrBy(ctx context.Context, key string, decrement int64) (int64, error) {
	return db.IncrBy(ctx, key, -decrement)
}

func (db *DB) LPush(ctx context.Context, key string, values ...interface{}) error {
	select {
	case <-ctx.Done():
		db.logger.Warn("LPush operation canceled", "key", key)
		return ErrContextCanceled
	default:
	}

	if key == "" {
		db.logger.Error("LPush failed: empty key")
		return ErrInvalidKey
	}

	if len(values) == 0 {
		db.logger.Warn("LPush called with no values", "key", key)
		return ErrEmptyValues
	}

	sh := db.shards[db.getShardIndex(key)]
	sh.mu.Lock()
	defer sh.mu.Unlock()

	entry, exists := sh.data[key]

	if exists && isExpired(entry) {
		delete(sh.data, key)
		exists = false
		db.logger.Info("LPush removed expired key before pushing", "key", key)
	}

	reversed := make([]interface{}, len(values))
	for i, j := 0, len(values)-1; i < len(values); i, j = i+1, j-1 {
		reversed[i] = values[j]
	}

	if exists {
		if entry.Type != List {
			db.logger.Error("LPush failed: existing key is not a list", "key", key)
			return ErrInvalidType
		}

		list, ok := entry.Value.([]interface{})
		if !ok {
			db.logger.Error("LPush failed: stored value is not []interface{}", "key", key)
			return ErrInvalidType
		}

		newList := make([]interface{}, 0, len(reversed)+len(list))
		newList = append(newList, reversed...)
		newList = append(newList, list...)
		entry.Value = newList
	} else {
		entry = Entry{
			Value:      reversed,
			Type:       List,
			Expiration: time.Time{},
		}
	}

	sh.data[key] = entry
	db.pubsub.Publish(key, fmt.Sprintf("LPush: %v", values))
	db.logger.Info("LPush operation successful",
		"key", key,
		"values", values,
		"count", len(values))
	return nil
}
func (db *DB) RPush(ctx context.Context, key string, values ...interface{}) error {
	select {
	case <-ctx.Done():
		db.logger.Warn("RPush operation canceled", "key", key)
		return ErrContextCanceled
	default:
	}

	if len(values) == 0 {
		db.logger.Warn("RPush called with no values", "key", key)
		return ErrEmptyValues
	}

	sh := db.shards[db.getShardIndex(key)]
	sh.mu.Lock()
	defer sh.mu.Unlock()

	entry, exists := sh.data[key]

	if exists && isExpired(entry) {
		delete(sh.data, key)
		exists = false
		db.logger.Info("RPush removed expired key before pushing", "key", key)
	}

	if exists {
		if entry.Type != List {
			db.logger.Error("RPush failed: existing key is not a list", "key", key)
			return ErrInvalidType
		}

		list, ok := entry.Value.([]interface{})
		if !ok {
			db.logger.Error("RPush failed: stored value is not []interface{}", "key", key)
			return ErrInvalidType
		}

		entry.Value = append(list, values...)
	} else {
		entry = Entry{
			Value:      values,
			Type:       List,
			Expiration: time.Time{},
		}
	}

	sh.data[key] = entry
	db.pubsub.Publish(key, fmt.Sprintf("RPush: %v", values))
	db.logger.Info("RPush operation successful",
		"key", key,
		"values", values,
		"count", len(values))
	return nil
}

func (db *DB) LPop(ctx context.Context, key string) (interface{}, error) {
	select {
	case <-ctx.Done():
		db.logger.Warn("LPop operation canceled", "key", key)
		return nil, ErrContextCanceled
	default:
	}

	sh := db.shards[db.getShardIndex(key)]
	sh.mu.Lock()
	defer sh.mu.Unlock()

	entry, exists := sh.data[key]
	if !exists || isExpired(entry) {
		db.logger.Warn("LPop failed: key not found or expired", "key", key)
		return nil, ErrKeyNotFound
	}

	if entry.Type != List {
		db.logger.Error("LPop failed: existing key is not a list", "key", key)
		return nil, ErrInvalidType
	}

	list, ok := entry.Value.([]interface{})
	if !ok || len(list) == 0 {
		db.logger.Warn("LPop failed: list empty or invalid type", "key", key)
		return nil, ErrEmptyList
	}

	val := list[0]
	list = list[1:]
	if len(list) == 0 {
		delete(sh.data, key)
		db.logger.Info("LPop removed the key as the list is now empty", "key", key)
	} else {
		entry.Value = list
		sh.data[key] = entry
	}

	db.logger.Info("LPop operation successful", "key", key, "poppedValue", val)
	return val, nil
}

func (db *DB) RPop(ctx context.Context, key string) (interface{}, error) {
	select {
	case <-ctx.Done():
		db.logger.Warn("RPop operation canceled", "key", key)
		return nil, ErrContextCanceled
	default:
	}

	sh := db.shards[db.getShardIndex(key)]
	sh.mu.Lock()
	defer sh.mu.Unlock()

	entry, exists := sh.data[key]
	if !exists || isExpired(entry) {
		db.logger.Warn("RPop failed: key not found or expired", "key", key)
		return nil, ErrKeyNotFound
	}

	if entry.Type != List {
		db.logger.Error("RPop failed: existing key is not a list", "key", key)
		return nil, ErrInvalidType
	}

	list, ok := entry.Value.([]interface{})
	if !ok || len(list) == 0 {
		db.logger.Warn("RPop failed: list empty or invalid type", "key", key)
		return nil, ErrEmptyList
	}

	val := list[len(list)-1]
	list = list[:len(list)-1]
	if len(list) == 0 {
		delete(sh.data, key)
		db.logger.Info("RPop removed the key as the list is now empty", "key", key)
	} else {
		entry.Value = list
		sh.data[key] = entry
	}

	db.logger.Info("RPop operation successful", "key", key, "poppedValue", val)
	return val, nil
}

func (db *DB) LLen(ctx context.Context, key string) (int, error) {
	select {
	case <-ctx.Done():
		db.logger.Warn("LLen operation canceled", "key", key)
		return 0, ErrContextCanceled
	default:
	}

	sh := db.shards[db.getShardIndex(key)]
	sh.mu.RLock()
	defer sh.mu.RUnlock()

	entry, exists := sh.data[key]
	if !exists || isExpired(entry) {
		db.logger.Warn("LLen failed: key not found or expired", "key", key)
		return 0, ErrKeyNotFound
	}

	if entry.Type != List {
		db.logger.Error("LLen failed: key is not a list", "key", key)
		return 0, ErrInvalidType
	}

	list, ok := entry.Value.([]interface{})
	if !ok {
		db.logger.Error("LLen failed: value is not a valid list", "key", key)
		return 0, ErrInvalidType
	}

	length := len(list)
	db.logger.Info("LLen operation successful", "key", key, "length", length)
	return length, nil
}

func (db *DB) LRange(ctx context.Context, key string, start, end int) ([]interface{}, error) {
	select {
	case <-ctx.Done():
		db.logger.Warn("LRange operation canceled", "key", key)
		return nil, ErrContextCanceled
	default:
	}

	sh := db.shards[db.getShardIndex(key)]
	sh.mu.RLock()
	defer sh.mu.RUnlock()

	entry, exists := sh.data[key]
	if !exists || isExpired(entry) {
		db.logger.Warn("LRange failed: key not found or expired", "key", key)
		return nil, ErrKeyNotFound
	}

	if entry.Type != List {
		db.logger.Error("LRange failed: existing key is not a list", "key", key)
		return nil, ErrInvalidType
	}

	list, ok := entry.Value.([]interface{})
	if !ok {
		db.logger.Error("LRange failed: value is not a valid list", "key", key)
		return nil, ErrInvalidType
	}

	length := len(list)

	if start < 0 {
		start = length + start
	}
	if end < 0 {
		end = length + end
	}

	if start < 0 {
		start = 0
	}
	if end >= length {
		end = length - 1
	}

	if start > end || start >= length {
		return []interface{}{}, nil
	}

	result := list[start : end+1]

	db.logger.Info("LRange operation successful", "key", key, "start", start, "end", end, "resultLength", len(result))
	return result, nil
}

func (db *DB) HSet(ctx context.Context, key string, field string, value interface{}, ttl int) error {
	select {
	case <-ctx.Done():
		db.logger.Warn("HSet operation canceled", "key", key)
		return ErrContextCanceled
	default:
	}

	expiration, err := ttlSecondsToTime(ttl)
	if err != nil {
		db.logger.Error("invalid TTL value in HSet", "key", key, "ttl", ttl, "error", err)
		return err
	}

	sh := db.shards[db.getShardIndex(key)]
	sh.mu.Lock()
	defer sh.mu.Unlock()

	entry, exists := sh.data[key]
	if exists && isExpired(entry) {
		delete(sh.data, key)
		exists = false
		db.logger.Info("HSet removed expired key before setting hash field", "key", key)
	}

	if !exists {
		entry = Entry{
			Type:       Hash,
			Value:      make(map[string]interface{}),
			Expiration: expiration,
		}
	} else {
		if entry.Type != Hash {
			db.logger.Error("HSet failed: existing key is not a hash", "key", key)
			return ErrInvalidType
		}
		if ttl == 0 {
			expiration = entry.Expiration
		}
	}

	hash := entry.Value.(map[string]interface{})
	hash[field] = value

	sh.data[key] = Entry{
		Value:      hash,
		Type:       Hash,
		Expiration: expiration,
	}

	db.logger.Info("HSet operation successful", "key", key, "field", field, "value", value, "ttl", ttl)
	return nil
}

func (db *DB) HGet(ctx context.Context, key string, field string) (interface{}, error) {
	select {
	case <-ctx.Done():
		db.logger.Warn("HGet operation canceled", "key", key)
		return nil, ErrContextCanceled
	default:
	}

	sh := db.shards[db.getShardIndex(key)]
	sh.mu.RLock()
	defer sh.mu.RUnlock()

	entry, exists := sh.data[key]
	if !exists || isExpired(entry) {
		db.logger.Warn("HGet failed: key not found or expired", "key", key)
		return nil, ErrKeyNotFound
	}

	if entry.Type != Hash {
		db.logger.Error("HGet failed: existing key is not a hash", "key", key)
		return nil, ErrInvalidType
	}

	hash := entry.Value.(map[string]interface{})
	val, found := hash[field]
	if !found {
		db.logger.Warn("HGet failed: field not found in hash", "key", key, "field", field)
		return nil, ErrKeyNotFound
	}

	db.logger.Info("HGet operation successful", "key", key, "field", field, "value", val)
	return val, nil
}

func (db *DB) HDel(ctx context.Context, key string, field string) error {
	select {
	case <-ctx.Done():
		db.logger.Warn("HDel operation canceled", "key", key)
		return ErrContextCanceled
	default:
	}

	sh := db.shards[db.getShardIndex(key)]
	sh.mu.Lock()
	defer sh.mu.Unlock()

	entry, exists := sh.data[key]
	if !exists || isExpired(entry) {
		db.logger.Warn("HDel failed: key not found or expired", "key", key)
		return ErrKeyNotFound
	}

	if entry.Type != Hash {
		db.logger.Error("HDel failed: existing key is not a hash", "key", key)
		return ErrInvalidType
	}

	hash := entry.Value.(map[string]interface{})
	delete(hash, field)

	if len(hash) == 0 {
		delete(sh.data, key)
		db.logger.Info("HDel removed the entire hash because it became empty", "key", key)
	} else {
		entry.Value = hash
		sh.data[key] = entry
	}
	db.logger.Info("HDel operation successful", "key", key, "field", field)
	return nil
}

func (db *DB) HGetAll(ctx context.Context, key string) (map[string]interface{}, error) {
	select {
	case <-ctx.Done():
		db.logger.Warn("HGetAll operation canceled", "key", key)
		return nil, ErrContextCanceled
	default:
	}

	sh := db.shards[db.getShardIndex(key)]
	sh.mu.RLock()
	defer sh.mu.RUnlock()

	entry, exists := sh.data[key]
	if !exists || isExpired(entry) {
		db.logger.Warn("HGetAll failed: key not found or expired", "key", key)
		return nil, ErrKeyNotFound
	}

	if entry.Type != Hash {
		db.logger.Error("HGetAll failed: existing key is not a hash", "key", key)
		return nil, ErrInvalidType
	}

	hash := entry.Value.(map[string]interface{})
	result := make(map[string]interface{}, len(hash))
	for k, v := range hash {
		result[k] = v
	}
	db.logger.Info("HGetAll operation successful", "key", key, "fieldsCount", len(result))
	return result, nil
}

func (db *DB) HExists(ctx context.Context, key string, field string) (bool, error) {
	select {
	case <-ctx.Done():
		db.logger.Warn("HExists operation canceled", "key", key)
		return false, ErrContextCanceled
	default:
	}

	sh := db.shards[db.getShardIndex(key)]
	sh.mu.RLock()
	defer sh.mu.RUnlock()

	entry, exists := sh.data[key]
	if !exists || isExpired(entry) {
		db.logger.Warn("HExists failed: key not found or expired", "key", key)
		return false, ErrKeyNotFound
	}

	if entry.Type != Hash {
		db.logger.Error("HExists failed: key is not a hash", "key", key)
		return false, ErrInvalidType
	}

	hash, ok := entry.Value.(map[string]interface{})
	if !ok {
		db.logger.Error("HExists failed: value is not a valid hash", "key", key)
		return false, ErrInvalidType
	}

	_, found := hash[field]
	db.logger.Info("HExists check", "key", key, "field", field, "exists", found)
	return found, nil
}

func (db *DB) HLen(ctx context.Context, key string) (int, error) {
	select {
	case <-ctx.Done():
		db.logger.Warn("HLen operation canceled", "key", key)
		return 0, ErrContextCanceled
	default:
	}

	sh := db.shards[db.getShardIndex(key)]
	sh.mu.RLock()
	defer sh.mu.RUnlock()

	entry, exists := sh.data[key]
	if !exists || isExpired(entry) {
		db.logger.Warn("HLen failed: key not found or expired", "key", key)
		return 0, ErrKeyNotFound
	}

	if entry.Type != Hash {
		db.logger.Error("HLen failed: key is not a hash", "key", key)
		return 0, ErrInvalidType
	}

	hash, ok := entry.Value.(map[string]interface{})
	if !ok {
		db.logger.Error("HLen failed: value is not a valid hash", "key", key)
		return 0, ErrInvalidType
	}

	length := len(hash)
	db.logger.Info("HLen operation successful", "key", key, "fieldCount", length)
	return length, nil
}

func (db *DB) Exists(ctx context.Context, key string) (bool, error) {
	select {
	case <-ctx.Done():
		db.logger.Warn("Exists operation canceled", "key", key)
		return false, ErrContextCanceled
	default:
	}

	sh := db.shards[db.getShardIndex(key)]
	sh.mu.RLock()
	defer sh.mu.RUnlock()

	entry, exists := sh.data[key]
	if !exists {
		db.logger.Info("Exists check: key does not exist", "key", key)
		return false, nil
	}
	if isExpired(entry) {
		db.logger.Info("Exists check: key is expired", "key", key)
		return false, nil
	}

	db.logger.Info("Exists check: key is present", "key", key)
	return true, nil
}

func (db *DB) Expire(ctx context.Context, key string, ttl int) (bool, error) {
	select {
	case <-ctx.Done():
		db.logger.Warn("Expire operation canceled", "key", key)
		return false, ErrContextCanceled
	default:
	}

	expiration, err := ttlSecondsToTime(ttl)
	if err != nil {
		db.logger.Error("invalid TTL in Expire", "key", key, "ttl", ttl, "error", err)
		return false, err
	}

	sh := db.shards[db.getShardIndex(key)]
	sh.mu.Lock()
	defer sh.mu.Unlock()

	entry, exists := sh.data[key]
	if !exists || isExpired(entry) {
		return false, nil
	}

	entry.Expiration = expiration
	sh.data[key] = entry
	db.logger.Info("Expire set", "key", key, "ttl", ttl)
	return true, nil
}

func (db *DB) Persist(ctx context.Context, key string) (bool, error) {
	select {
	case <-ctx.Done():
		db.logger.Warn("Persist operation canceled", "key", key)
		return false, ErrContextCanceled
	default:
	}

	sh := db.shards[db.getShardIndex(key)]
	sh.mu.Lock()
	defer sh.mu.Unlock()

	entry, exists := sh.data[key]
	if !exists || isExpired(entry) {
		return false, nil
	}

	if entry.Expiration.IsZero() {
		return false, nil
	}

	entry.Expiration = time.Time{}
	sh.data[key] = entry
	db.logger.Info("Persist successful", "key", key)
	return true, nil
}

func (db *DB) Type(ctx context.Context, key string) (interface{}, error) {
	select {
	case <-ctx.Done():
		db.logger.Warn("Type operation canceled", "key", key)
		return -1, ErrContextCanceled
	default:
	}

	sh := db.shards[db.getShardIndex(key)]
	sh.mu.RLock()
	defer sh.mu.RUnlock()

	entry, exists := sh.data[key]
	if !exists || isExpired(entry) {
		db.logger.Warn("Type check failed: key not found or expired", "key", key)
		return -1, ErrKeyNotFound
	}

	db.logger.Info("Type operation successful", "key", key, "type", entry.Type)
	return entry.Type, nil
}

func (db *DB) GetWithDetails(ctx context.Context, key string) (interface{}, int, error) {
	select {
	case <-ctx.Done():
		db.logger.Warn("GetWithDetails operation canceled", "key", key)
		return nil, 0, ErrContextCanceled
	default:
	}

	sh := db.shards[db.getShardIndex(key)]
	sh.mu.RLock()
	defer sh.mu.RUnlock()

	entry, exists := sh.data[key]
	if !exists {
		db.logger.Warn("GetWithDetails failed: key not found", "key", key)
		return nil, 0, ErrKeyNotFound
	}

	if isExpired(entry) {
		db.logger.Info("GetWithDetails: key is expired", "key", key)
		return nil, 0, ErrKeyExpired
	}

	var ttl int
	if entry.Expiration.IsZero() {
		ttl = -1
	} else {
		ttl = int(time.Until(entry.Expiration).Seconds())
	}
	db.logger.Info("GetWithDetails operation successful", "key", key, "ttl", ttl)
	return entry.Value, ttl, nil
}

func (db *DB) Rename(ctx context.Context, oldKey, newKey string) error {
	select {
	case <-ctx.Done():
		db.logger.Warn("Rename operation canceled", "oldKey", oldKey, "newKey", newKey)
		return ErrContextCanceled
	default:
	}

	if oldKey == "" || newKey == "" {
		db.logger.Error("Rename failed: key cannot be empty", "oldKey", oldKey, "newKey", newKey)
		return ErrInvalidKey
	}

	if oldKey == newKey {
		db.logger.Warn("Rename skipped: oldKey and newKey are the same", "key", oldKey)
		return nil
	}

	oldShard := db.shards[db.getShardIndex(oldKey)]
	newShard := db.shards[db.getShardIndex(newKey)]

	oldShard.mu.Lock()
	defer oldShard.mu.Unlock()

	entry, exists := oldShard.data[oldKey]
	if !exists || isExpired(entry) {
		db.logger.Warn("Rename failed: oldKey not found or expired", "oldKey", oldKey)
		return ErrKeyNotFound
	}

	if oldShard != newShard {
		newShard.mu.Lock()
		defer newShard.mu.Unlock()

		if _, conflict := newShard.data[newKey]; conflict {
			db.logger.Warn("Rename failed: newKey already exists", "newKey", newKey)
			return ErrKeyExists
		}

		newShard.data[newKey] = entry
		delete(oldShard.data, oldKey)

	} else {
		if _, conflict := oldShard.data[newKey]; conflict {
			db.logger.Warn("Rename failed: newKey already exists", "newKey", newKey)
			return ErrKeyExists
		}
		oldShard.data[newKey] = entry
		delete(oldShard.data, oldKey)
	}

	db.logger.Info("Rename operation successful", "oldKey", oldKey, "newKey", newKey)
	db.pubsub.Publish(oldKey, "RENAMED")
	db.pubsub.Publish(newKey, "CREATED (via rename)")

	return nil
}

func (db *DB) FindByValue(ctx context.Context, value interface{}) ([]string, error) {
	select {
	case <-ctx.Done():
		db.logger.Warn("FindByValue operation canceled", "value", value)
		return nil, ErrContextCanceled
	default:
	}

	var keys []string

	for _, sh := range db.shards {
		sh.mu.RLock()
		for k, entry := range sh.data {
			if !isExpired(entry) && entry.Value == value {
				keys = append(keys, k)
			}
		}
		sh.mu.RUnlock()
	}

	if len(keys) == 0 {
		db.logger.Warn("no keys found for given value", "value", value)
		return nil, ErrKeyNotFound
	}
	db.logger.Info("FindByValue operation successful", "value", value, "foundKeys", keys)
	return keys, nil
}

func (db *DB) Delete(ctx context.Context, key string) error {
	select {
	case <-ctx.Done():
		db.logger.Warn("Delete operation canceled", "key", key)
		return ErrContextCanceled
	default:
	}

	sh := db.shards[db.getShardIndex(key)]
	sh.mu.Lock()
	defer sh.mu.Unlock()

	_, exists := sh.data[key]
	if !exists {
		db.logger.Warn("attempt to Delete a non-existent key", "key", key)
		return ErrKeyNotFound
	}

	delete(sh.data, key)
	db.pubsub.Publish(key, "DELETE")
	db.logger.Info("Delete operation successful", "key", key)
	return nil
}

func (db *DB) DropAll(ctx context.Context) error {
	select {
	case <-ctx.Done():
		db.logger.Warn("DropAll operation canceled")
		return ErrContextCanceled
	default:
	}

	for _, sh := range db.shards {
		sh.mu.Lock()
		for key := range sh.data {
			db.pubsub.Publish(key, "FLUSH_ALL")
			delete(sh.data, key)
		}
		sh.mu.Unlock()
	}
	db.logger.Info("DropAll operation completed: all keys removed")
	return nil
}

func (db *DB) cleanupExpiredKeys(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	const (
		basePercentage    = 0.25
		aggressionFactor  = 1.2
		cooldownThreshold = 0.1
	)

	for {
		select {
		case <-ticker.C:
			var totalProcessed int
			var totalExpired int

			for _, sh := range db.shards {
				processed, expired := db.processShard(sh, basePercentage, aggressionFactor, cooldownThreshold)
				totalProcessed += processed
				totalExpired += expired
			}

			db.logger.Info("Global cleanup stats",
				"total_processed", totalProcessed,
				"total_expired", totalExpired,
				"efficiency", safeDivide(totalExpired, totalProcessed),
			)

		case <-db.cleanupCtx.Done():
			db.logger.Info("cleanupExpiredKeys: shutting down")
			return
		}
	}
}

func (db *DB) processShard(sh *shard, basePct, aggression float64, cooldownThr float64) (int, int) {
	sh.mu.RLock()
	totalKeys := len(sh.data)
	checkLimit := int(float64(totalKeys) * basePct)
	sh.mu.RUnlock()

	if checkLimit < 1 {
		return 0, 0
	}

	sh.mu.Lock()
	defer sh.mu.Unlock()

	var expiredKeys []string
	checked := 0
	aggressiveMode := false

	for key, entry := range sh.data {
		if checked >= checkLimit {
			break
		}

		checked++
		if isExpired(entry) {
			expiredKeys = append(expiredKeys, key)

			if len(expiredKeys) > int(float64(checked)*cooldownThr) {
				checkLimit = int(float64(checkLimit) * aggression)
				aggressiveMode = true
			}
		}
	}

	deleted := 0
	for _, key := range expiredKeys {
		if entry, exists := sh.data[key]; exists && isExpired(entry) {
			delete(sh.data, key)
			db.pubsub.Publish(key, "EXPIRED")
			deleted++
		}
	}

	if deleted > 0 {
		efficiency := safeDivide(deleted, checked)

		if aggressiveMode {
			db.logger.Warn("Aggressive cleanup activated",
				"checked", checked,
				"deleted", deleted,
				"efficiency", efficiency,
				"aggressive", aggressiveMode,
			)
		} else {
			db.logger.Debug("Shard cleanup stats",
				"checked", checked,
				"deleted", deleted,
				"efficiency", efficiency,
				"aggressive", aggressiveMode,
			)
		}
	}

	return checked, deleted
}

func safeDivide(a, b int) float64 {
	if b == 0 {
		return 0.0
	}
	return float64(a) / float64(b)
}

func (db *DB) Subscribe(key string) chan string {
	db.logger.Debug("New subscription", "key", key)
	return db.pubsub.Subscribe(key)
}

func (db *DB) Unsubscribe(key string, ch chan string) {
	db.logger.Debug("Removing subscription", "key", key)
	db.pubsub.Unsubscribe(key, ch)
}

func (db *DB) ListSubscriptions() []string {
	return db.pubsub.ListSubscribers()
}

func (db *DB) CloseAllSubscriptionsForKey(key string) {
	db.logger.Warn("Closing all subscriptions for key", "key", key)
	db.pubsub.UnsubscribeAllForKey(key)
}

func (db *DB) Logger() contracts.LoggerHandler {
	return db.logger
}

func (db *DB) Commands() contracts.CommandsHandler {
	return db.commands
}

func (db *DB) Transaction() contracts.TransactionHandler {
	return NewTransaction(db)
}

func (db *DB) Close() error {
	if db == nil {
		return nil
	}

	db.logger.Info("Shutting down store...")

	if db.cleanupCancel != nil {
		db.cleanupCancel()
	}

	if db.pubsub != nil {
		db.pubsub.Close()
		db.logger.Info("PubSub closed successfully")
	}

	if db.logger != nil {
		err := db.logger.Close()
		if err != nil {
			db.logger.Error("Logger close error:", err)
			return err
		}
	}

	db.logger.Info("Logger closed successfully")
	return nil
}
