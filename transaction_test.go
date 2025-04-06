package hermes

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"
)

func setupTestDB() *DB {
	config := Config{}
	return NewStore(config)
}

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
	if errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("Expected key='value', but got not found")
	} else if err != nil {
		t.Fatalf("Unexpected Get error: %v", err)
	}
	if val != "value" {
		t.Fatalf("Expected 'value', got %v", val)
	}
}

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

func TestTransactionDeleteCommit(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()
	_ = db.Set(ctx, "test_key", "value", 60)

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

func TestTransactionDeleteRollback(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()
	_ = db.Set(ctx, "test_key", "value", 60)

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

func TestTransactionIncrCommit(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()
	_ = db.Set(ctx, "counter", int64(1), 0)

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

func TestTransactionDecrCommit(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()
	_ = db.Set(ctx, "counter", int64(1), 0)

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

func TestTransactionIncrRollback(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()
	_ = db.Set(ctx, "counter", int64(1), 0)

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

func TestTransactionDecrRollback(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()
	_ = db.Set(ctx, "counter", int64(1), 0)

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

func TestTransactionSetNXExists(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()
	_ = db.Set(ctx, "nxKey", "alreadyHere", 0)

	tx := db.Transaction()

	err := tx.SetNX(ctx, "nxKey", "newVal", 0)
	if err != nil {
		t.Fatalf("SetNX returned error: %v", err)
	}

	err = tx.Commit()
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

func TestTransactionSetXXNotExists(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	tx := db.Transaction()

	err := tx.SetXX(ctx, "xxKey", "val", 0)
	if err != nil {
		t.Fatalf("SetXX returned error: %v", err)
	}

	err = tx.Commit()
	if !errors.Is(err, ErrTransactionFailed) {
		t.Fatalf("Expected ErrTransactionFailed, got: %v", err)
	}

	_, err = db.Get(ctx, "xxKey")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("Expected key not to exist, got: %v", err)
	}
}

func TestTransactionSetCASWrongOld(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()
	_ = db.Set(ctx, "casKey", "initial", 0)

	tx := db.Transaction()

	err := tx.SetCAS(ctx, "casKey", "wrong", "new", 0)
	if err != nil {
		t.Fatalf("SetCAS returned error: %v", err)
	}

	err = tx.Commit()
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

func TestTransactionExpireCommit(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()
	_ = db.Set(ctx, "ttlKey", "hasTTL", 1)

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

func TestTransactionExpireRollback(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()
	_ = db.Set(ctx, "ttlKey", "hello", 1)

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

func TestTransactionConcurrentIncrements(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()
	_ = db.Set(ctx, "sharedCounter", int64(0), 0)

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

func TestTransactionLPopCommit(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	_ = db.LPush(ctx, "listKey", "val1")

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

func TestTransactionLPopRollback(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	err := db.LPush(ctx, "listKey", "val2", "val1")
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
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

func TestTransactionRPopCommit(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	_ = db.RPush(ctx, "listKey", "val1")

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

func TestTransactionRPopRollback(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	if err := db.LPush(ctx, "listKey", "val2", "val1"); err != nil {
		t.Fatalf("Setup failed: %v", err)
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

func TestTransactionLLenAndLRangeCommit(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	_ = db.RPush(ctx, "listKey", "one")
	_ = db.RPush(ctx, "listKey", "two")

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

func TestTransactionHDelCommit(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	_ = db.HSet(ctx, "hashKey", "field1", "value1", 0)

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

func TestTransactionHDelRollback(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	_ = db.HSet(ctx, "hashKey", "field1", "value1", 0)

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

func TestTransactionHGetAllCommit(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	_ = db.HSet(ctx, "hashKey", "field1", "val1", 0)

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

func TestTransactionHExistsAndHLenCommit(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	_ = db.HSet(ctx, "hashKey", "field1", "val1", 0)

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

func TestTransactionExistsCommit(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	_ = db.Set(ctx, "someKey", "someValue", 0)

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

	ok, _ := db.Exists(ctx, "someKey")
	if ok {
		t.Fatalf("Expected someKey to be gone")
	}
}

func TestTransactionTypeCommit(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	_ = db.Set(ctx, "strKey", "val", 0)

	tx := db.Transaction()
	typ, err := tx.Type(ctx, "strKey")
	if err != nil {
		t.Fatalf("Type in transaction error: %v", err)
	}
	if typ != String {
		t.Fatalf("Expected type=String, got %v", typ)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}
}

func TestTransactionGetWithDetailsCommit(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	_ = db.Set(ctx, "detailKey", "detailedVal", 5)

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

func TestTransactionRenameCommit(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	_ = db.Set(ctx, "oldKey", "oldVal", 0)

	tx := db.Transaction()
	if err := tx.Rename(ctx, "oldKey", "newKey"); err != nil {
		t.Fatalf("Rename in transaction failed: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	_, err := db.Get(ctx, "oldKey")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("Expected oldKey to not exist, got %v", err)
	}

	val, err := db.Get(ctx, "newKey")
	if err != nil {
		t.Fatalf("Get(newKey) failed: %v", err)
	}
	if val != "oldVal" {
		t.Fatalf("Expected 'oldVal', got %v", val)
	}
}

func TestTransactionRenameRollback(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	_ = db.Set(ctx, "oldKey", "oldVal", 0)
	_ = db.Set(ctx, "anotherKey", "someVal", 0)

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

func TestTransactionFindByValueCommit(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()

	_ = db.Set(ctx, "k1", "look", 0)

	tx := db.Transaction()
	if err := tx.Set(ctx, "k2", "look", 0); err != nil {
		t.Fatalf("Set in transaction failed: %v", err)
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
