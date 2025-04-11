package hermes

import (
	"context"
	"errors"
	"github.com/themedef/go-hermes/internal/contracts"
	"github.com/themedef/go-hermes/internal/types"
	"reflect"
	"sync"
	"testing"
	"time"
)

func setupTestDB() contracts.StoreHandler {
	config := Config{}
	return NewStore(config)
}

// TestTransactionSetCommit checks that a Set inside a transaction is committed correctly.
func TestTransactionSetCommit(t *testing.T) {
	db := setupTestDB()
	tx := db.Transaction()
	ctx := context.Background()

	if err := tx.Set(ctx, "test_key", "value", 60); err != nil {
		t.Fatalf("Set in transaction failed: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	val, err := db.Get(ctx, "test_key")
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if val != "value" {
		t.Fatalf("Expected 'value', got %v", val)
	}
}

// TestTransactionSetRollback checks that a Set is rolled back properly.
func TestTransactionSetRollback(t *testing.T) {
	db := setupTestDB()
	tx := db.Transaction()
	ctx := context.Background()

	if err := tx.Set(ctx, "test_key", "value", 60); err != nil {
		t.Fatalf("Set in transaction failed: %v", err)
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	_, err := db.Get(ctx, "test_key")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("Expected ErrKeyNotFound after rollback, got %v", err)
	}
}

// TestTransactionDeleteCommit checks that Delete is committed correctly.
func TestTransactionDeleteCommit(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	if err := db.Set(ctx, "test_key", "value", 60); err != nil {
		t.Fatalf("Setup Set failed: %v", err)
	}

	tx := db.Transaction()

	if err := tx.Delete(ctx, "test_key"); err != nil {
		t.Fatalf("Delete in transaction failed: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	_, err := db.Get(ctx, "test_key")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("Expected key to be deleted, got err=%v", err)
	}
}

// TestTransactionDeleteRollback checks that Delete is rolled back correctly.
func TestTransactionDeleteRollback(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	if err := db.Set(ctx, "test_key", "value", 60); err != nil {
		t.Fatalf("Setup Set failed: %v", err)
	}

	tx := db.Transaction()

	if err := tx.Delete(ctx, "test_key"); err != nil {
		t.Fatalf("Delete in transaction failed: %v", err)
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	val, err := db.Get(ctx, "test_key")
	if errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("Expected key restored after rollback, but not found")
	} else if err != nil {
		t.Fatalf("Unexpected error after rollback: %v", err)
	}
	if val != "value" {
		t.Fatalf("Expected 'value', got %v", val)
	}
}

// TestTransactionIncrCommit checks that Incr is committed properly.
func TestTransactionIncrCommit(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	if err := db.Set(ctx, "counter", int64(1), 0); err != nil {
		t.Fatalf("Setup Set failed: %v", err)
	}

	tx := db.Transaction()

	if err := tx.Incr(ctx, "counter"); err != nil {
		t.Fatalf("Incr in transaction failed: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	val, err := db.Get(ctx, "counter")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val.(int64) != 2 {
		t.Fatalf("Expected 2, got %v", val)
	}
}

// TestTransactionDecrCommit checks that Decr is committed properly.
func TestTransactionDecrCommit(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	if err := db.Set(ctx, "counter", int64(1), 0); err != nil {
		t.Fatalf("Setup Set failed: %v", err)
	}

	tx := db.Transaction()

	if err := tx.Decr(ctx, "counter"); err != nil {
		t.Fatalf("Decr in transaction failed: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	val, err := db.Get(ctx, "counter")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val.(int64) != 0 {
		t.Fatalf("Expected 0, got %v", val)
	}
}

// TestTransactionIncrRollback checks that an Incr is rolled back properly.
func TestTransactionIncrRollback(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	if err := db.Set(ctx, "counter", int64(1), 0); err != nil {
		t.Fatalf("Setup Set failed: %v", err)
	}

	tx := db.Transaction()

	if err := tx.Incr(ctx, "counter"); err != nil {
		t.Fatalf("Incr failed: %v", err)
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	val, err := db.Get(ctx, "counter")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val.(int64) != 1 {
		t.Fatalf("Expected 1 after rollback, got %v", val)
	}
}

// TestTransactionDecrRollback checks that a Decr is rolled back properly.
func TestTransactionDecrRollback(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	if err := db.Set(ctx, "counter", int64(1), 0); err != nil {
		t.Fatalf("Setup Set failed: %v", err)
	}

	tx := db.Transaction()

	if err := tx.Decr(ctx, "counter"); err != nil {
		t.Fatalf("Decr failed: %v", err)
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	val, err := db.Get(ctx, "counter")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val.(int64) != 1 {
		t.Fatalf("Expected 1 after rollback, got %v", val)
	}
}

// TestTransactionSetNXExists checks that SetNX on an existing key
// causes transaction failure on commit.
func TestTransactionSetNXExists(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	if err := db.Set(ctx, "nxKey", "alreadyHere", 0); err != nil {
		t.Fatalf("Setup Set failed: %v", err)
	}

	tx := db.Transaction()

	if err := tx.SetNX(ctx, "nxKey", "newVal", 0); err != nil {
		t.Fatalf("SetNX returned error: %v", err)
	}

	err := tx.Commit()
	if !errors.Is(err, ErrTransactionFailed) {
		t.Fatalf("Expected ErrTransactionFailed, got: %v", err)
	}

	val, err := db.Get(ctx, "nxKey")
	if err != nil {
		t.Fatalf("Get after failed commit: %v", err)
	}
	if val != "alreadyHere" {
		t.Fatalf("Expected 'alreadyHere', got: %v", val)
	}
}

// TestTransactionSetXXNotExists checks that SetXX on a missing key
// causes transaction failure on commit.
func TestTransactionSetXXNotExists(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	tx := db.Transaction()

	if err := tx.SetXX(ctx, "xxKey", "val", 0); err != nil {
		t.Fatalf("SetXX returned error: %v", err)
	}

	err := tx.Commit()
	if !errors.Is(err, ErrTransactionFailed) {
		t.Fatalf("Expected ErrTransactionFailed, got: %v", err)
	}

	_, err = db.Get(ctx, "xxKey")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("Expected key not to exist, got: %v", err)
	}
}

// TestTransactionSetCASWrongOld checks that a wrong oldValue for CAS leads to transaction failure on commit.
func TestTransactionSetCASWrongOld(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	if err := db.Set(ctx, "casKey", "initial", 0); err != nil {
		t.Fatalf("Setup Set failed: %v", err)
	}

	tx := db.Transaction()

	if err := tx.SetCAS(ctx, "casKey", "wrong", "new", 0); err != nil {
		t.Fatalf("SetCAS returned error: %v", err)
	}

	err := tx.Commit()
	if !errors.Is(err, ErrTransactionFailed) {
		t.Fatalf("Expected ErrTransactionFailed, got: %v", err)
	}

	val, err := db.Get(ctx, "casKey")
	if err != nil {
		t.Fatalf("Get after failed commit: %v", err)
	}
	if val != "initial" {
		t.Fatalf("Expected 'initial', got: %v", val)
	}
}

// TestTransactionExpireCommit checks that Expire inside a transaction extends TTL on commit.
func TestTransactionExpireCommit(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	if err := db.Set(ctx, "ttlKey", "hasTTL", 1); err != nil {
		t.Fatalf("Setup Set failed: %v", err)
	}

	tx := db.Transaction()

	if err := tx.Expire(ctx, "ttlKey", 3); err != nil {
		t.Fatalf("Expire in transaction failed: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	time.Sleep(2 * time.Second)
	val, err := db.Get(ctx, "ttlKey")
	if errors.Is(err, ErrKeyExpired) || errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("Key should still exist, got err=%v", err)
	} else if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if val != "hasTTL" {
		t.Fatalf("Expected 'hasTTL', got %v", val)
	}
}

// TestTransactionExpireRollback checks that an Expire call is rolled back if the transaction is rolled back.
func TestTransactionExpireRollback(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	if err := db.Set(ctx, "ttlKey", "hello", 1); err != nil {
		t.Fatalf("Setup Set failed: %v", err)
	}

	tx := db.Transaction()

	if err := tx.Expire(ctx, "ttlKey", 5); err != nil {
		t.Fatalf("Expire in transaction failed: %v", err)
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	time.Sleep(2 * time.Second)
	_, err := db.Get(ctx, "ttlKey")
	if !errors.Is(err, ErrKeyExpired) && !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("Expected key to be expired/not found after old TTL=1, got err=%v", err)
	}
}

// TestTransactionConcurrentIncrements checks concurrent commits of small Incr transactions.
func TestTransactionConcurrentIncrements(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	if err := db.Set(ctx, "sharedCounter", int64(0), 0); err != nil {
		t.Fatalf("Setup Set failed: %v", err)
	}

	const workers = 2
	const increments = 5
	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < increments; j++ {
				tx := db.Transaction()
				if err := tx.Incr(ctx, "sharedCounter"); err != nil {
					t.Fatalf("Worker %d: Incr failed: %v", id, err)
				}
				if err := tx.Commit(); err != nil {
					t.Fatalf("Worker %d: Commit failed: %v", id, err)
				}
			}
		}(i)
	}

	wg.Wait()

	val, err := db.Get(ctx, "sharedCounter")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	final := val.(int64)
	expected := int64(workers * increments)
	if final != expected {
		t.Fatalf("Expected final=%d, got %d", expected, final)
	}
}

// TestTransactionLPushCommit checks that LPush is committed properly.
func TestTransactionLPushCommit(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	tx := db.Transaction()
	if err := tx.LPush(ctx, "listKey", "val1"); err != nil {
		t.Fatalf("LPush in transaction failed: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	val, err := db.LPop(ctx, "listKey")
	if err != nil {
		t.Fatalf("LPop failed: %v", err)
	}
	if val != "val1" {
		t.Fatalf("Expected 'val1', got %v", val)
	}
}

// TestTransactionLPushRollback checks that LPush is rolled back properly.
func TestTransactionLPushRollback(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	tx := db.Transaction()
	if err := tx.LPush(ctx, "listKey", "val1"); err != nil {
		t.Fatalf("LPush in transaction failed: %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	_, err := db.LPop(ctx, "listKey")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("Expected ErrKeyNotFound (list empty or does not exist), got: %v", err)
	}
}

// TestTransactionRPushCommit checks that RPush is committed properly.
func TestTransactionRPushCommit(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	tx := db.Transaction()
	if err := tx.RPush(ctx, "listKey", "val1"); err != nil {
		t.Fatalf("RPush in transaction failed: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	val, err := db.RPop(ctx, "listKey")
	if err != nil {
		t.Fatalf("RPop failed: %v", err)
	}
	if val != "val1" {
		t.Fatalf("Expected 'val1', got %v", val)
	}
}

// TestTransactionRPushRollback checks that RPush is rolled back properly.
func TestTransactionRPushRollback(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	tx := db.Transaction()
	if err := tx.RPush(ctx, "listKey", "val1"); err != nil {
		t.Fatalf("RPush in transaction failed: %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	_, err := db.RPop(ctx, "listKey")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("Expected ErrKeyNotFound after rollback, got: %v", err)
	}
}

// TestTransactionLPopCommit checks that LPop is committed properly.
func TestTransactionLPopCommit(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	if err := db.LPush(ctx, "listKey", "val1"); err != nil {
		t.Fatalf("Setup LPush failed: %v", err)
	}

	tx := db.Transaction()
	popped, err := tx.LPop(ctx, "listKey")
	if err != nil {
		t.Fatalf("LPop in transaction failed: %v", err)
	}
	if popped != "val1" {
		t.Fatalf("Expected 'val1', got %v", popped)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	_, err = db.LPop(ctx, "listKey")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("Expected empty or nonexistent list after pop, got: %v", err)
	}
}

// TestTransactionLPopRollback checks that LPop is rolled back properly.
func TestTransactionLPopRollback(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	if err := db.LPush(ctx, "listKey", "val2", "val1"); err != nil {
		t.Fatalf("Setup LPush failed: %v", err)
	}

	tx := db.Transaction()
	popped, err := tx.LPop(ctx, "listKey")
	if err != nil {
		t.Fatalf("LPop in transaction failed: %v", err)
	}
	if popped != "val1" {
		t.Fatalf("Expected 'val1', got %v", popped)
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	list, err := db.LRange(ctx, "listKey", 0, -1)
	if err != nil {
		t.Fatalf("LRange failed: %v", err)
	}

	expected := []interface{}{"val1", "val2"}
	if !reflect.DeepEqual(list, expected) {
		t.Fatalf("Expected %v after rollback, got %v", expected, list)
	}
}

// TestTransactionRPopCommit checks that RPop is committed properly.
func TestTransactionRPopCommit(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	if err := db.RPush(ctx, "listKey", "val1"); err != nil {
		t.Fatalf("Setup RPush failed: %v", err)
	}

	tx := db.Transaction()
	popped, err := tx.RPop(ctx, "listKey")
	if err != nil {
		t.Fatalf("RPop in transaction failed: %v", err)
	}
	if popped != "val1" {
		t.Fatalf("Expected 'val1', got %v", popped)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	exists, err := db.Exists(ctx, "listKey")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Fatalf("Expected listKey to not exist after final pop, but it still exists")
	}
}

// TestTransactionRPopRollback checks that RPop is rolled back properly.
func TestTransactionRPopRollback(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	if err := db.LPush(ctx, "listKey", "val2", "val1"); err != nil {
		t.Fatalf("Setup LPush failed: %v", err)
	}

	tx := db.Transaction()
	popped, err := tx.RPop(ctx, "listKey")
	if err != nil {
		t.Fatalf("RPop in transaction failed: %v", err)
	}
	if popped != "val2" {
		t.Fatalf("Expected 'val2', got %v", popped)
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	list, err := db.LRange(ctx, "listKey", 0, -1)
	if err != nil {
		t.Fatalf("LRange failed: %v", err)
	}

	expected := []interface{}{"val1", "val2"}
	if !reflect.DeepEqual(list, expected) {
		t.Fatalf("Expected %v after rollback, got %v", expected, list)
	}
}

// TestTransactionLLenAndLRangeCommit checks that LLen and LRange work after committing changes.
func TestTransactionLLenAndLRangeCommit(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	if err := db.RPush(ctx, "listKey", "one"); err != nil {
		t.Fatalf("Setup RPush failed: %v", err)
	}
	if err := db.RPush(ctx, "listKey", "two"); err != nil {
		t.Fatalf("Setup RPush failed: %v", err)
	}

	tx := db.Transaction()
	if err := tx.RPush(ctx, "listKey", "three"); err != nil {
		t.Fatalf("RPush in transaction failed: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	length, err := db.LLen(ctx, "listKey")
	if err != nil {
		t.Fatalf("LLen error: %v", err)
	}
	if length != 3 {
		t.Fatalf("Expected length=3, got=%d", length)
	}

	values, err := db.LRange(ctx, "listKey", 0, -1)
	if err != nil {
		t.Fatalf("LRange error: %v", err)
	}
	if len(values) != 3 || values[0] != "one" || values[1] != "two" || values[2] != "three" {
		t.Fatalf("LRange mismatch, got=%v", values)
	}
}

// TestTransactionHSetCommit checks that HSet is committed properly.
func TestTransactionHSetCommit(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	tx := db.Transaction()
	if err := tx.HSet(ctx, "hashKey", "field1", "value1", 0); err != nil {
		t.Fatalf("HSet in transaction failed: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	val, err := db.HGet(ctx, "hashKey", "field1")
	if err != nil {
		t.Fatalf("HGet failed: %v", err)
	}
	if val != "value1" {
		t.Fatalf("Expected 'value1', got %v", val)
	}
}

// TestTransactionHSetRollback checks that HSet is rolled back properly.
func TestTransactionHSetRollback(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	tx := db.Transaction()
	if err := tx.HSet(ctx, "hashKey", "field1", "value1", 0); err != nil {
		t.Fatalf("HSet in transaction failed: %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	_, err := db.HGet(ctx, "hashKey", "field1")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("Expected ErrKeyNotFound after rollback, got %v", err)
	}
}

// TestTransactionHDelCommit checks that HDel is committed properly.
func TestTransactionHDelCommit(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	if err := db.HSet(ctx, "hashKey", "field1", "value1", 0); err != nil {
		t.Fatalf("Setup HSet failed: %v", err)
	}

	tx := db.Transaction()
	if err := tx.HDel(ctx, "hashKey", "field1"); err != nil {
		t.Fatalf("HDel in transaction failed: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	_, err := db.HGet(ctx, "hashKey", "field1")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("Expected ErrKeyNotFound, got %v", err)
	}
}

// TestTransactionHDelRollback checks that HDel is rolled back properly.
func TestTransactionHDelRollback(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	if err := db.HSet(ctx, "hashKey", "field1", "value1", 0); err != nil {
		t.Fatalf("Setup HSet failed: %v", err)
	}

	tx := db.Transaction()
	if err := tx.HDel(ctx, "hashKey", "field1"); err != nil {
		t.Fatalf("HDel in transaction failed: %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	val, err := db.HGet(ctx, "hashKey", "field1")
	if err != nil {
		t.Fatalf("HGet after rollback failed: %v", err)
	}
	if val != "value1" {
		t.Fatalf("Expected 'value1', got %v", val)
	}
}

// TestTransactionHGetAllCommit checks that HGetAll sees changes after commit.
func TestTransactionHGetAllCommit(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	if err := db.HSet(ctx, "hashKey", "field1", "val1", 0); err != nil {
		t.Fatalf("Setup HSet failed: %v", err)
	}

	tx := db.Transaction()
	if err := tx.HSet(ctx, "hashKey", "field2", "val2", 0); err != nil {
		t.Fatalf("HSet in transaction failed: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	allFields, err := db.HGetAll(ctx, "hashKey")
	if err != nil {
		t.Fatalf("HGetAll failed: %v", err)
	}
	if len(allFields) != 2 || allFields["field1"] != "val1" || allFields["field2"] != "val2" {
		t.Fatalf("Expected 2 fields {field1=val1, field2=val2}, got %v", allFields)
	}
}

// TestTransactionHExistsAndHLenCommit checks that HExists and HLen reflect changes after commit.
func TestTransactionHExistsAndHLenCommit(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	if err := db.HSet(ctx, "hashKey", "field1", "val1", 0); err != nil {
		t.Fatalf("Setup HSet failed: %v", err)
	}

	tx := db.Transaction()
	if err := tx.HSet(ctx, "hashKey", "field2", "val2", 0); err != nil {
		t.Fatalf("HSet in transaction failed: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	exists, err := db.HExists(ctx, "hashKey", "field1")
	if err != nil {
		t.Fatalf("HExists error: %v", err)
	}
	if !exists {
		t.Fatalf("Expected field1 to exist")
	}

	length, err := db.HLen(ctx, "hashKey")
	if err != nil {
		t.Fatalf("HLen error: %v", err)
	}
	if length != 2 {
		t.Fatalf("Expected length=2, got %d", length)
	}
}

// TestTransactionExistsCommit checks that Exists reflects transaction changes after commit.
func TestTransactionExistsCommit(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	if err := db.Set(ctx, "someKey", "someValue", 0); err != nil {
		t.Fatalf("Setup Set failed: %v", err)
	}

	tx := db.Transaction()
	exists, err := tx.Exists(ctx, "someKey")
	if err != nil {
		t.Fatalf("Exists in transaction error: %v", err)
	}
	if !exists {
		t.Fatalf("Expected someKey to exist")
	}

	if err := tx.Delete(ctx, "someKey"); err != nil {
		t.Fatalf("Delete in transaction failed: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	ok, err := db.Exists(ctx, "someKey")
	if err != nil {
		t.Fatalf("Exists after commit error: %v", err)
	}
	if ok {
		t.Fatalf("Expected someKey to be gone")
	}
}

// TestTransactionTypeCommit checks that Type returns the correct type within a transaction.
func TestTransactionTypeCommit(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	if err := db.Set(ctx, "strKey", "val", 0); err != nil {
		t.Fatalf("Setup Set failed: %v", err)
	}

	tx := db.Transaction()
	typ, err := tx.Type(ctx, "strKey")
	if err != nil {
		t.Fatalf("Type in transaction error: %v", err)
	}
	if typ != types.String {
		t.Fatalf("Expected type=String, got %v", typ)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}
}

// TestTransactionGetWithDetailsCommit checks that GetWithDetails returns correct TTL inside a transaction.
func TestTransactionGetWithDetailsCommit(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	if err := db.Set(ctx, "detailKey", "detailedVal", 5); err != nil {
		t.Fatalf("Setup Set failed: %v", err)
	}

	tx := db.Transaction()
	val, ttl, err := tx.GetWithDetails(ctx, "detailKey")
	if err != nil {
		t.Fatalf("GetWithDetails in transaction failed: %v", err)
	}
	if val != "detailedVal" {
		t.Fatalf("Expected 'detailedVal', got %v", val)
	}
	if ttl <= 0 {
		t.Fatalf("Expected TTL > 0, got %d", ttl)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}
}

// TestTransactionRenameCommit checks that Rename is committed properly.
func TestTransactionRenameCommit(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	if err := db.Set(ctx, "oldKey", "oldVal", 60); err != nil {
		t.Fatalf("Setup Set failed: %v", err)
	}

	val, err := db.Get(ctx, "oldKey")
	if err != nil {
		t.Fatalf("Get(oldKey) failed: %v", err)
	}
	if val != "oldVal" {
		t.Fatalf("Expected oldVal, got %v", val)
	}

	tx := db.Transaction()
	if err := tx.Rename(ctx, "oldKey", "newKey"); err != nil {
		t.Fatalf("Rename in transaction failed: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}
	_, err = db.Get(ctx, "oldKey")
	if !IsKeyNotFound(err) {
		t.Fatalf("Expected oldKey to not exist, got %v", err)
	}

	val, err = db.Get(ctx, "newKey")
	if err != nil {
		t.Fatalf("Get(newKey) failed: %v", err)
	}
	if val != "oldVal" {
		t.Fatalf("Expected 'oldVal' in newKey, got %v", val)
	}
}

// TestTransactionRenameRollback checks that Rename is rolled back properly.
func TestTransactionRenameRollback(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	if err := db.Set(ctx, "oldKey", "oldVal", 0); err != nil {
		t.Fatalf("Setup oldKey failed: %v", err)
	}
	if err := db.Set(ctx, "anotherKey", "someVal", 0); err != nil {
		t.Fatalf("Setup anotherKey failed: %v", err)
	}

	tx := db.Transaction()
	if err := tx.Rename(ctx, "oldKey", "anotherKey"); err != nil {
		t.Fatalf("Rename in transaction failed: %v", err)
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	val, err := db.Get(ctx, "oldKey")
	if err != nil {
		t.Fatalf("Get(oldKey) failed: %v", err)
	}
	if val != "oldVal" {
		t.Fatalf("Expected 'oldVal', got %v", val)
	}

	val2, err := db.Get(ctx, "anotherKey")
	if err != nil {
		t.Fatalf("Get(anotherKey) failed: %v", err)
	}
	if val2 != "someVal" {
		t.Fatalf("Expected 'someVal', got %v", val2)
	}
}

// TestTransactionFindByValueCommit checks that FindByValue sees newly created keys after commit.
func TestTransactionFindByValueCommit(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	if err := db.Set(ctx, "k1", "look", 0); err != nil {
		t.Fatalf("Setup Set(k1) failed: %v", err)
	}

	tx := db.Transaction()
	if err := tx.Set(ctx, "k2", "look", 0); err != nil {
		t.Fatalf("Set(k2) in transaction failed: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	keys, err := db.FindByValue(ctx, "look")
	if err != nil {
		t.Fatalf("FindByValue error: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("Expected 2 keys, got %d (%v)", len(keys), keys)
	}
}

// TestTransactionSAddCommit checks that SAdd is committed properly.
func TestTransactionSAddCommit(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	tx := db.Transaction()
	if err := tx.SAdd(ctx, "colors", "red", "green", "blue"); err != nil {
		t.Fatalf("SAdd in transaction failed: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	members, err := db.SMembers(ctx, "colors")
	if err != nil {
		t.Fatalf("SMembers failed: %v", err)
	}
	if len(members) != 3 {
		t.Fatalf("Expected 3 members, got %d", len(members))
	}
}

// TestTransactionSAddRollback checks that SAdd is rolled back properly.
func TestTransactionSAddRollback(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	tx := db.Transaction()
	if err := tx.SAdd(ctx, "colors", "red", "green"); err != nil {
		t.Fatalf("SAdd in transaction failed: %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	_, err := db.SMembers(ctx, "colors")
	if !IsKeyNotFound(err) {
		t.Fatalf("Expected ErrKeyNotFound after rollback, got %v", err)
	}
}

// TestTransactionSRemCommit checks that SRem is committed properly.
func TestTransactionSRemCommit(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	if err := db.SAdd(ctx, "numbers", 1, 2, 3, 4, 5); err != nil {
		t.Fatalf("SAdd setup failed: %v", err)
	}

	tx := db.Transaction()
	if err := tx.SRem(ctx, "numbers", 2, 4); err != nil {
		t.Fatalf("SRem in transaction failed: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	members, err := db.SMembers(ctx, "numbers")
	if err != nil {
		t.Fatalf("SMembers error: %v", err)
	}
	for _, m := range members {
		if m == 2 || m == 4 {
			t.Fatalf("Expected members 2 and 4 to be removed, but found %v", m)
		}
	}
}

// TestTransactionSRemRollback checks that SRem is rolled back properly.
func TestTransactionSRemRollback(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	if err := db.SAdd(ctx, "numbers", 10, 20, 30); err != nil {
		t.Fatalf("SAdd setup failed: %v", err)
	}

	tx := db.Transaction()
	if err := tx.SRem(ctx, "numbers", 10); err != nil {
		t.Fatalf("SRem in transaction failed: %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	members, err := db.SMembers(ctx, "numbers")
	if err != nil {
		t.Fatalf("SMembers failed: %v", err)
	}
	found := false
	for _, m := range members {
		if m == 10 {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("Expected member 10 to be present after rollback, got %v", members)
	}
}

// TestTransactionSMembersCommit checks that SMembers sees newly added members after commit.
func TestTransactionSMembersCommit(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	tx := db.Transaction()
	if err := tx.SAdd(ctx, "foods", "apple", "banana"); err != nil {
		t.Fatalf("SAdd in transaction failed: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	members, err := db.SMembers(ctx, "foods")
	if err != nil {
		t.Fatalf("SMembers error: %v", err)
	}
	if len(members) != 2 {
		t.Fatalf("Expected 2 members, got %d", len(members))
	}
}

// TestTransactionSIsMemberCommit checks that SIsMember reflects committed changes.
func TestTransactionSIsMemberCommit(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	tx := db.Transaction()
	if err := tx.SAdd(ctx, "planets", "Earth", "Mars"); err != nil {
		t.Fatalf("SAdd in transaction failed: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	isMem, err := db.SIsMember(ctx, "planets", "Earth")
	if err != nil {
		t.Fatalf("SIsMember error: %v", err)
	}
	if !isMem {
		t.Fatalf("Expected Earth to be a member of the set")
	}
}

// TestTransactionSCardCommit checks that SCard returns the correct count after commit.
func TestTransactionSCardCommit(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	// Setup
	if err := db.SAdd(ctx, "letters", "a", "b"); err != nil {
		t.Fatalf("SAdd setup failed: %v", err)
	}

	tx := db.Transaction()
	if err := tx.SAdd(ctx, "letters", "c"); err != nil {
		t.Fatalf("SAdd in transaction failed: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	card, err := db.SCard(ctx, "letters")
	if err != nil {
		t.Fatalf("SCard error: %v", err)
	}
	if card != 3 {
		t.Fatalf("Expected cardinality=3, got=%d", card)
	}
}

// TestTransactionLTrimCommit checks that LTrim changes are committed properly.
func TestTransactionLTrimCommit(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	if err := db.RPush(ctx, "numbersList", 1, 2, 3, 4, 5); err != nil {
		t.Fatalf("RPush setup failed: %v", err)
	}

	tx := db.Transaction()
	if err := tx.LTrim(ctx, "numbersList", 1, 2); err != nil {
		t.Fatalf("LTrim in transaction failed: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	result, err := db.LRange(ctx, "numbersList", 0, -1)
	if err != nil {
		t.Fatalf("LRange after commit failed: %v", err)
	}
	if len(result) != 2 || result[0] != 2 || result[1] != 3 {
		t.Fatalf("Expected [2 3], got %v", result)
	}
}

// TestTransactionLTrimRollback checks that LTrim changes are rolled back if the transaction is rolled back.
func TestTransactionLTrimRollback(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	if err := db.RPush(ctx, "numbersList", 10, 20, 30, 40); err != nil {
		t.Fatalf("RPush setup failed: %v", err)
	}

	tx := db.Transaction()
	if err := tx.LTrim(ctx, "numbersList", 1, 2); err != nil {
		t.Fatalf("LTrim in transaction failed: %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	result, err := db.LRange(ctx, "numbersList", 0, -1)
	if err != nil {
		t.Fatalf("LRange failed: %v", err)
	}
	if len(result) != 4 || result[0] != 10 || result[1] != 20 || result[2] != 30 || result[3] != 40 {
		t.Fatalf("Expected [10 20 30 40], got %v", result)
	}
}
