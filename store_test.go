package hermes

import (
	"context"
	"fmt"
	"github.com/themedef/go-hermes/internal/types"
	"sync"
	"testing"
	"time"

	"github.com/themedef/go-hermes/internal/contracts"
)

func withTestStore(t *testing.T) contracts.StoreHandler {
	db := newTestStore()
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing store: %v", err)
		}
	})
	return db
}

func newTestStore() contracts.StoreHandler {
	return NewStore(Config{})
}

// TestStoreSet checks the behavior of the Set method.
func TestStoreSet(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()

	// Scenario 1: Set a regular value and verify with Get
	err := db.Set(ctx, "hello", "world", 0)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	val, err := db.Get(ctx, "hello")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != "world" {
		t.Errorf("Expected 'world', got %v", val)
	}

	// Scenario 2: Set with a negative TTL
	err = db.Set(ctx, "invalidTTL", "val", -5)
	if !IsInvalidTTL(err) {
		t.Errorf("Expected ErrInvalidTTL, got %v", err)
	}

	// Scenario 3: Empty key
	err = db.Set(ctx, "", "val", 0)
	if !IsInvalidKey(err) {
		t.Errorf("Expected ErrInvalidKey, got %v", err)
	}

	// Scenario 4: Setting a key with a large TTL
	ttl := 365 * 24 * 3600
	err = db.Set(ctx, "hugeTTL", "data", ttl)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	val, err = db.Get(ctx, "hugeTTL")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != "data" {
		t.Errorf("Key with large TTL mismatch, got %v", val)
	}

	// Scenario 5: Set nil
	err = db.Set(ctx, "nilValue", nil, 0)
	if err != nil {
		t.Fatalf("Set failed (nil): %v", err)
	}
	val, err = db.Get(ctx, "nilValue")
	if err != nil {
		t.Fatalf("Get failed (nil): %v", err)
	}
	if val != nil {
		t.Errorf("Expected nil, got %v", val)
	}

	// Scenario 6: Set with TTL and check expiration
	err = db.Set(ctx, "tempKey", "tempVal", 1)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	time.Sleep(2 * time.Second)
	_, err = db.Get(ctx, "tempKey")
	if !IsKeyExpired(err) && !IsKeyNotFound(err) {
		t.Errorf("Expected key expiration, got: %v", err)
	}
}

// TestStoreSetNX checks the behavior of the SetNX method.
func TestStoreSetNX(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()

	// Scenario 1: Successful SetNX for a missing key
	ok, err := db.SetNX(ctx, "nxKey", "initial", 0)
	if err != nil {
		t.Fatalf("SetNX failed (first): %v", err)
	}
	if !ok {
		t.Error("SetNX should succeed on missing key")
	}

	// Scenario 2: Attempt SetNX on an existing key
	ok, err = db.SetNX(ctx, "nxKey", "newVal", 0)
	if !IsKeyExists(err) {
		t.Errorf("Expected ErrKeyExists, got %v", err)
	}
	if ok {
		t.Error("SetNX returned ok=true, but key already exists")
	}
}

// TestStoreSetXX checks the behavior of the SetXX method.
func TestStoreSetXX(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()

	// Scenario 1: Attempt to SetXX on a non-existent key
	ok, err := db.SetXX(ctx, "xxKey", "val", 0)
	if !IsKeyNotFound(err) {
		t.Errorf("Expected ErrKeyNotFound, got %v", err)
	}
	if ok {
		t.Error("SetXX returned ok=true, but key doesn't exist")
	}

	// Scenario 2: Successful SetXX for an existing key
	err = db.Set(ctx, "xxKeyReal", "oldVal", 0)
	if err != nil {
		t.Fatalf("Initial Set failed: %v", err)
	}
	ok, err = db.SetXX(ctx, "xxKeyReal", "newVal", 0)
	if err != nil {
		t.Fatalf("SetXX failed: %v", err)
	}
	if !ok {
		t.Error("SetXX should succeed on existing key")
	}
	val, err := db.Get(ctx, "xxKeyReal")
	if err != nil {
		t.Fatalf("Get after SetXX failed: %v", err)
	}
	if val != "newVal" {
		t.Errorf("Expected 'newVal', got %v", val)
	}
}

// TestStoreGet checks the behavior of the Get method.
func TestStoreGet(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()

	// Scenario 1: Get a missing key
	_, err := db.Get(ctx, "missingKey")
	if !IsKeyNotFound(err) {
		t.Errorf("Expected ErrKeyNotFound, got: %v", err)
	}

	// Scenario 2: Basic Set + Get
	err = db.Set(ctx, "basicKey", "basicVal", 0)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	val, err := db.Get(ctx, "basicKey")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != "basicVal" {
		t.Errorf("Expected 'basicVal', got %v", val)
	}

	// Scenario 3: Checking TTL expiration
	err = db.Set(ctx, "expireKey", "willExpire", 1)
	if err != nil {
		t.Fatalf("Set with TTL failed: %v", err)
	}
	time.Sleep(2 * time.Second)
	_, err = db.Get(ctx, "expireKey")
	if !IsKeyExpired(err) && !IsKeyNotFound(err) {
		t.Errorf("Expected expiration, got %v", err)
	}
}

// TestStoreSetCAS checks the behavior of the SetCAS method.
func TestStoreSetCAS(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()

	// Scenario 1: Successful CAS
	err := db.Set(ctx, "casKey", "old", 0)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	err = db.SetCAS(ctx, "casKey", "old", "new", 0)
	if err != nil {
		t.Fatalf("SetCAS failed: %v", err)
	}
	val, err := db.Get(ctx, "casKey")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != "new" {
		t.Errorf("SetCAS did not update value; got %v", val)
	}

	// Scenario 2: CAS mismatch
	err = db.Set(ctx, "casKey2", "init", 0)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	err = db.SetCAS(ctx, "casKey2", "wrongOld", "new", 0)
	if !IsValueMismatch(err) {
		t.Errorf("Expected ErrValueMismatch, got %v", err)
	}
}

// TestStoreGetSet checks the behavior of the GetSet method.
func TestStoreGetSet(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()

	// Scenario 1: First GetSet call on a new key
	oldVal, err := db.GetSet(ctx, "gsKey", "newVal", 0)
	if err != nil {
		t.Fatalf("GetSet failed: %v", err)
	}
	if oldVal != nil {
		t.Errorf("Expected nil for new key, got %v", oldVal)
	}

	// Scenario 2: Second GetSet call
	oldVal, err = db.GetSet(ctx, "gsKey", "updatedVal", 0)
	if err != nil {
		t.Fatalf("Second GetSet failed: %v", err)
	}
	if oldVal != "newVal" {
		t.Errorf("Expected 'newVal' as oldVal, got %v", oldVal)
	}
	val, err := db.Get(ctx, "gsKey")
	if err != nil {
		t.Fatalf("Get after GetSet failed: %v", err)
	}
	if val != "updatedVal" {
		t.Errorf("GetSet didn't update value, got %v", val)
	}
}

// TestStoreIncr checks the behavior of the Incr method.
func TestStoreIncr(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()

	// Scenario 1: Incr on a new key
	val, err := db.Incr(ctx, "counter")
	if err != nil && !IsKeyNotFound(err) {
		t.Fatalf("Incr returned unexpected error: %v", err)
	}
	if val != 1 {
		t.Errorf("Expected 1, got %d", val)
	}

	// Scenario 2: Multiple Incr calls on the same key
	for i := 0; i < 2; i++ {
		_, err = db.Incr(ctx, "counter")
		if err != nil {
			t.Fatalf("Incr iteration %d failed: %v", i, err)
		}
	}
	val, err = db.Incr(ctx, "counter")
	if err != nil {
		t.Fatalf("Final Incr failed: %v", err)
	}
	if val != 4 {
		t.Errorf("Expected 4, got %d", val)
	}

	// Scenario 3: Incr on a non-integer value
	err = db.Set(ctx, "str", "hello", 0)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	_, err = db.Incr(ctx, "str")
	if !IsInvalidValueType(err) {
		t.Errorf("Expected ErrInvalidValueType, got %v", err)
	}
}

// TestStoreDecr checks the behavior of the Decr method.
func TestStoreDecr(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()

	// Scenario 1: Decr on a new key
	val, err := db.Decr(ctx, "decKey")
	if err != nil && !IsKeyNotFound(err) {
		t.Fatalf("Decr returned unexpected error: %v", err)
	}
	if val != -1 {
		t.Errorf("Expected -1, got %d", val)
	}

	// Scenario 2: Multiple Decr calls
	for i := 0; i < 2; i++ {
		_, err = db.Decr(ctx, "decKey")
		if err != nil {
			t.Fatalf("Decr iteration %d failed: %v", i, err)
		}
	}
	val, err = db.Decr(ctx, "decKey")
	if err != nil {
		t.Fatalf("Final Decr failed: %v", err)
	}
	if val != -4 {
		t.Errorf("Expected -4, got %d", val)
	}

	// Scenario 3: Decr on a non-integer value
	err = db.Set(ctx, "strD", "hello", 0)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	_, err = db.Decr(ctx, "strD")
	if !IsInvalidValueType(err) {
		t.Errorf("Expected ErrInvalidValueType, got %v", err)
	}
}

// TestStoreIncrBy checks the behavior of the IncrBy method.
func TestStoreIncrBy(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()

	// Scenario: IncrBy on a new key
	val, err := db.IncrBy(ctx, "incrByKey", 10)
	if err != nil && !IsKeyNotFound(err) {
		t.Fatalf("IncrBy returned unexpected error: %v", err)
	}
	if val != 10 {
		t.Errorf("Expected 10, got %d", val)
	}

	// Scenario: IncrBy on an existing int64
	val, err = db.IncrBy(ctx, "incrByKey", 5)
	if err != nil {
		t.Fatalf("Second IncrBy failed: %v", err)
	}
	if val != 15 {
		t.Errorf("Expected 15, got %d", val)
	}
}

// TestStoreDecrBy checks the behavior of the DecrBy method.
func TestStoreDecrBy(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()

	// Check DecrBy as negative IncrBy
	_, err := db.DecrBy(ctx, "decrByKey", 5)
	if err != nil && !IsKeyNotFound(err) {
		t.Fatalf("DecrBy returned unexpected error: %v", err)
	}
	val, err := db.Get(ctx, "decrByKey")
	if err != nil {
		t.Fatalf("Get after DecrBy failed: %v", err)
	}
	if val.(int64) != -5 {
		t.Errorf("Expected -5, got %v", val)
	}

	// Second decrement
	_, err = db.DecrBy(ctx, "decrByKey", 10)
	if err != nil {
		t.Fatalf("Second DecrBy failed: %v", err)
	}
	val, err = db.Get(ctx, "decrByKey")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val.(int64) != -15 {
		t.Errorf("Expected -15, got %v", val)
	}
}

// TestStoreLPush checks the behavior of the LPush method.
func TestStoreLPush(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()

	// Simple LPush check
	err := db.LPush(ctx, "myList", 1)
	if err != nil {
		t.Fatalf("LPush failed: %v", err)
	}
	err = db.LPush(ctx, "myList", 2)
	if err != nil {
		t.Fatalf("LPush failed: %v", err)
	}
	val, err := db.LPop(ctx, "myList")
	if err != nil {
		t.Fatalf("LPop failed: %v", err)
	}
	if val != 2 {
		t.Errorf("Expected 2, got %v", val)
	}

	// Attempt LPush on a non-list key
	err = db.Set(ctx, "strKey", "val", 0)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	err = db.LPush(ctx, "strKey", "item")
	if !IsInvalidType(err) {
		t.Errorf("Expected ErrInvalidType, got %v", err)
	}
}

// TestStoreRPush checks the behavior of the RPush method.
func TestStoreRPush(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()

	// Simple RPush check
	err := db.RPush(ctx, "myListRP", "a")
	if err != nil {
		t.Fatalf("RPush failed: %v", err)
	}
	err = db.RPush(ctx, "myListRP", "b")
	if err != nil {
		t.Fatalf("RPush failed: %v", err)
	}
	val, err := db.RPop(ctx, "myListRP")
	if err != nil {
		t.Fatalf("RPop failed: %v", err)
	}
	if val != "b" {
		t.Errorf("Expected 'b', got %v", val)
	}

	// Attempt RPush on a non-list key
	err = db.Set(ctx, "strKeyRP", "val", 0)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	err = db.RPush(ctx, "strKeyRP", "x")
	if !IsInvalidType(err) {
		t.Errorf("Expected ErrInvalidType, got %v", err)
	}
}

// TestStoreLPop checks the behavior of the LPop method.
func TestStoreLPop(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()

	// LPop on a missing key
	_, err := db.LPop(ctx, "noList")
	if !IsKeyNotFound(err) {
		t.Errorf("Expected ErrKeyNotFound, got %v", err)
	}

	// Normal scenario
	err = db.LPush(ctx, "popList", 1, 2)
	if err != nil {
		t.Fatalf("LPush for popList failed: %v", err)
	}
	val, err := db.LPop(ctx, "popList")
	if err != nil {
		t.Fatalf("LPop failed: %v", err)
	}
	if val != 2 {
		t.Errorf("Expected 2, got %v", val)
	}
}

// TestStoreRPop checks the behavior of the RPop method.
func TestStoreRPop(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()

	// RPop on a missing key
	_, err := db.RPop(ctx, "noList2")
	if !IsKeyNotFound(err) {
		t.Errorf("Expected ErrKeyNotFound, got %v", err)
	}

	// Normal scenario
	err = db.RPush(ctx, "popList2", "a", "b")
	if err != nil {
		t.Fatalf("RPush failed: %v", err)
	}
	val, err := db.RPop(ctx, "popList2")
	if err != nil {
		t.Fatalf("RPop failed: %v", err)
	}
	if val != "b" {
		t.Errorf("Expected 'b', got %v", val)
	}
}

// TestStoreLLen checks the behavior of the LLen method.
func TestStoreLLen(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()

	err := db.LPush(ctx, "lenList", "x", "y", "z")
	if err != nil {
		t.Fatalf("LPush failed: %v", err)
	}
	length, err := db.LLen(ctx, "lenList")
	if err != nil {
		t.Fatalf("LLen failed: %v", err)
	}
	if length != 3 {
		t.Errorf("Expected length=3, got %d", length)
	}
}

// TestStoreLRange checks the behavior of the LRange method.
func TestStoreLRange(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()

	err := db.RPush(ctx, "rangeList", 1, 2, 3, 4, 5)
	if err != nil {
		t.Fatalf("RPush failed: %v", err)
	}
	// Take elements from index 1 to 3
	vals, err := db.LRange(ctx, "rangeList", 1, 3)
	if err != nil {
		t.Fatalf("LRange failed: %v", err)
	}
	if len(vals) != 3 || vals[0] != 2 || vals[1] != 3 || vals[2] != 4 {
		t.Errorf("LRange returned unexpected slice: %v", vals)
	}
}

// TestStoreLTrim checks the behavior of the LTrim method.
func TestStoreLTrim(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()

	err := db.RPush(ctx, "trimList", 1, 2, 3, 4, 5)
	if err != nil {
		t.Fatalf("RPush failed: %v", err)
	}
	// Keep elements from index 1..2
	err = db.LTrim(ctx, "trimList", 1, 2)
	if err != nil {
		t.Fatalf("LTrim failed: %v", err)
	}
	vals, err := db.LRange(ctx, "trimList", 0, -1)
	if err != nil {
		t.Fatalf("LRange after trim failed: %v", err)
	}
	if len(vals) != 2 || vals[0] != 2 || vals[1] != 3 {
		t.Errorf("Unexpected trimmed list: %v", vals)
	}
}

// TestStoreHSet checks the behavior of the HSet method.
func TestStoreHSet(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()

	err := db.HSet(ctx, "user:1", "name", "Test", 0)
	if err != nil {
		t.Fatalf("HSet failed: %v", err)
	}
	// Attempt HSet on a non-hash key
	err = db.Set(ctx, "strType", "something", 0)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	err = db.HSet(ctx, "strType", "field", "val", 0)
	if !IsInvalidType(err) {
		t.Errorf("Expected ErrInvalidType, got %v", err)
	}
}

// TestStoreHGet checks the behavior of the HGet method.
func TestStoreHGet(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()

	_, err := db.HGet(ctx, "nobody", "field")
	if !IsKeyNotFound(err) {
		t.Errorf("Expected ErrKeyNotFound, got %v", err)
	}

	err = db.HSet(ctx, "user:2", "name", "Bob", 0)
	if err != nil {
		t.Fatalf("HSet failed: %v", err)
	}
	val, err := db.HGet(ctx, "user:2", "name")
	if err != nil {
		t.Fatalf("HGet failed: %v", err)
	}
	if val != "Bob" {
		t.Errorf("Expected 'Bob', got %v", val)
	}
}

// TestStoreHDel checks the behavior of the HDel method.
func TestStoreHDel(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()

	err := db.HSet(ctx, "session", "token", "abc123", 0)
	if err != nil {
		t.Fatalf("HSet failed: %v", err)
	}
	err = db.HDel(ctx, "session", "token")
	if err != nil {
		t.Fatalf("HDel failed: %v", err)
	}
	_, err = db.HGet(ctx, "session", "token")
	if !IsKeyNotFound(err) {
		t.Errorf("Expected ErrKeyNotFound after HDel, got %v", err)
	}
}

// TestStoreHGetAll checks the behavior of the HGetAll method.
func TestStoreHGetAll(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()

	err := db.HSet(ctx, "product:1", "name", "Laptop", 0)
	if err != nil {
		t.Fatalf("HSet failed: %v", err)
	}
	err = db.HSet(ctx, "product:1", "price", 1000, 0)
	if err != nil {
		t.Fatalf("HSet failed: %v", err)
	}
	fields, err := db.HGetAll(ctx, "product:1")
	if err != nil {
		t.Fatalf("HGetAll failed: %v", err)
	}
	if len(fields) != 2 || fields["name"] != "Laptop" || fields["price"] != 1000 {
		t.Errorf("HGetAll returned incorrect data: %v", fields)
	}
}

// TestStoreHExists checks the behavior of the HExists method.
func TestStoreHExists(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()

	err := db.HSet(ctx, "hash", "field1", "val1", 0)
	if err != nil {
		t.Fatalf("HSet failed: %v", err)
	}
	exists, err := db.HExists(ctx, "hash", "field1")
	if err != nil {
		t.Fatalf("HExists error: %v", err)
	}
	if !exists {
		t.Error("HExists failed to find existing field")
	}
	exists, err = db.HExists(ctx, "hash", "missing")
	if err != nil {
		t.Fatalf("HExists error: %v", err)
	}
	if exists {
		t.Error("HExists found non-existing field")
	}
}

// TestStoreHLen checks the behavior of the HLen method.
func TestStoreHLen(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()

	err := db.HSet(ctx, "hashLen", "field1", "val1", 0)
	if err != nil {
		t.Fatalf("HSet failed: %v", err)
	}
	err = db.HSet(ctx, "hashLen", "field2", "val2", 0)
	if err != nil {
		t.Fatalf("HSet failed: %v", err)
	}
	count, err := db.HLen(ctx, "hashLen")
	if err != nil {
		t.Fatalf("HLen error: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2, got %d", count)
	}
}

// TestStoreSAdd checks the behavior of the SAdd method.
func TestStoreSAdd(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()

	// Add multiple elements
	err := db.SAdd(ctx, "colors", "red", "green", "blue")
	if err != nil {
		t.Fatalf("SAdd failed: %v", err)
	}

	// Verify that the key was created
	members, err := db.SMembers(ctx, "colors")
	if err != nil {
		t.Fatalf("SMembers failed: %v", err)
	}
	if len(members) != 3 {
		t.Errorf("Expected 3 members, got %d", len(members))
	}

	// Attempt SAdd on a non-set key
	err = db.Set(ctx, "notASet", "value", 0)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	err = db.SAdd(ctx, "notASet", "this", "fail")
	if !IsInvalidType(err) {
		t.Errorf("Expected ErrInvalidType, got %v", err)
	}
}

// TestStoreSRem checks the behavior of the SRem method.
func TestStoreSRem(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()

	err := db.SAdd(ctx, "numbers", 1, 2, 3, 4, 5)
	if err != nil {
		t.Fatalf("SAdd failed: %v", err)
	}
	err = db.SRem(ctx, "numbers", 2, 4)
	if err != nil {
		t.Fatalf("SRem failed: %v", err)
	}
	members, err := db.SMembers(ctx, "numbers")
	if err != nil {
		t.Fatalf("SMembers failed: %v", err)
	}
	for _, m := range members {
		if m == 2 || m == 4 {
			t.Errorf("Expected members 2 and 4 to be removed, but found %v", m)
		}
	}
}

// TestStoreSMembers checks the behavior of the SMembers method.
func TestStoreSMembers(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()

	// Create a set
	err := db.SAdd(ctx, "fruits", "apple", "banana", "cherry")
	if err != nil {
		t.Fatalf("SAdd failed: %v", err)
	}
	members, err := db.SMembers(ctx, "fruits")
	if err != nil {
		t.Fatalf("SMembers failed: %v", err)
	}
	if len(members) != 3 {
		t.Errorf("Expected 3, got %d", len(members))
	}
}

// TestStoreSIsMember checks the behavior of the SIsMember method.
func TestStoreSIsMember(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()

	err := db.SAdd(ctx, "planets", "Earth", "Mars", "Venus")
	if err != nil {
		t.Fatalf("SAdd failed: %v", err)
	}
	isMem, err := db.SIsMember(ctx, "planets", "Earth")
	if err != nil {
		t.Fatalf("SIsMember failed: %v", err)
	}
	if !isMem {
		t.Error("Expected Earth to be a member")
	}
	isMem, err = db.SIsMember(ctx, "planets", "Jupiter")
	if err != nil {
		t.Fatalf("SIsMember failed: %v", err)
	}
	if isMem {
		t.Error("Jupiter should not be a member")
	}
}

// TestStoreSCard checks the behavior of the SCard method.
func TestStoreSCard(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()

	err := db.SAdd(ctx, "letters", "a", "b", "c")
	if err != nil {
		t.Fatalf("SAdd failed: %v", err)
	}
	card, err := db.SCard(ctx, "letters")
	if err != nil {
		t.Fatalf("SCard failed: %v", err)
	}
	if card != 3 {
		t.Errorf("Expected 3, got %d", card)
	}
}

// TestStoreExists checks the behavior of the Exists method.
func TestStoreExists(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()

	err := db.Set(ctx, "check", "val", 0)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	ok, err := db.Exists(ctx, "check")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !ok {
		t.Error("Key should exist")
	}
	ok, _ = db.Exists(ctx, "notexists")
	if ok {
		t.Error("Key should not exist")
	}
}

// TestStoreExpire checks the behavior of the Expire method.
func TestStoreExpire(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()

	err := db.Set(ctx, "upKey", "val", 1)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	success, err := db.Expire(ctx, "upKey", 3)
	if err != nil {
		t.Fatalf("Expire failed: %v", err)
	}
	if !success {
		t.Fatalf("Expire did not succeed on existing key")
	}
	time.Sleep(1500 * time.Millisecond)
	val, err := db.Get(ctx, "upKey")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != "val" {
		t.Errorf("Value mismatch after Expire: expected 'val', got %v", val)
	}

	// Expire on a missing key
	success, err = db.Expire(ctx, "missing", 10)
	if !IsKeyNotFound(err) {
		t.Errorf("Expected ErrKeyNotFound, got err=%v, success=%v", err, success)
	}
}

// TestStorePersist checks the behavior of the Persist method.
func TestStorePersist(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()

	// Set a key with TTL
	err := db.Set(ctx, "tempKey", "toPersist", 1)
	if err != nil {
		t.Fatalf("Set with TTL failed: %v", err)
	}
	// Immediately call Persist — TTL should be cleared
	ok, err := db.Persist(ctx, "tempKey")
	if err != nil {
		t.Fatalf("Persist failed: %v", err)
	}
	if !ok {
		t.Error("Persist should return true for existing key with TTL")
	}
	time.Sleep(2 * time.Second)
	// Check that it did not expire
	val, err := db.Get(ctx, "tempKey")
	if err != nil {
		t.Fatalf("Get after Persist failed: %v", err)
	}
	if val != "toPersist" {
		t.Errorf("Expected 'toPersist', got %v", val)
	}
}

// TestStoreType checks the behavior of the Type method.
func TestStoreType(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()

	err := db.Set(ctx, "string", "val", 0)
	if err != nil {
		t.Fatalf("Setup Set failed: %v", err)
	}
	err = db.LPush(ctx, "list", "elem")
	if err != nil {
		t.Fatalf("Setup LPush failed: %v", err)
	}
	err = db.HSet(ctx, "hash", "field", "val", 0)
	if err != nil {
		t.Fatalf("Setup HSet failed: %v", err)
	}

	tests := []struct {
		key      string
		expected types.DataType
	}{
		{"string", types.String},
		{"list", types.List},
		{"hash", types.Hash},
	}
	for _, tt := range tests {
		dtype, err := db.Type(ctx, tt.key)
		if err != nil {
			t.Fatalf("Type failed for %s: %v", tt.key, err)
		}
		if dtype != tt.expected {
			t.Errorf("Type mismatch for %s: got %v, expected %v", tt.key, dtype, tt.expected)
		}
	}
}

// TestStoreGetWithDetails checks the behavior of the GetWithDetails method.
func TestStoreGetWithDetails(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()

	err := db.Set(ctx, "detailed", "value", 10)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	val, ttl, err := db.GetWithDetails(ctx, "detailed")
	if err != nil {
		t.Fatalf("GetWithDetails failed: %v", err)
	}
	if val != "value" || ttl <= 0 {
		t.Errorf("Unexpected values: val=%v, ttl=%d", val, ttl)
	}

	err = db.Set(ctx, "temp", "val", 1)
	if err != nil {
		t.Fatalf("Setup Set failed: %v", err)
	}
	time.Sleep(1100 * time.Millisecond)
	_, _, err = db.GetWithDetails(ctx, "temp")
	if !IsKeyExpired(err) && !IsKeyNotFound(err) {
		t.Errorf("Expected key expired or not found, got %v", err)
	}
}

// TestStoreRename checks the behavior of the Rename method.
func TestStoreRename(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()

	err := db.Set(ctx, "oldKey", "value", 0)
	if err != nil {
		t.Fatalf("Setup Set failed: %v", err)
	}
	// Rename to a new key
	err = db.Rename(ctx, "oldKey", "newKey")
	if err != nil {
		t.Fatalf("Rename failed: %v", err)
	}
	_, err = db.Get(ctx, "newKey")
	if err != nil {
		t.Fatalf("Get after rename failed: %v", err)
	}

	// Rename to an existing key => error
	err = db.Set(ctx, "targetKey", "existing", 0)
	if err != nil {
		t.Fatalf("Setup Set failed: %v", err)
	}
	err = db.Rename(ctx, "newKey", "targetKey")
	if !IsKeyExists(err) {
		t.Errorf("Expected ErrKeyExists, got %v", err)
	}
}

// TestStoreFindByValue checks the behavior of the FindByValue method.
func TestStoreFindByValue(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()

	err := db.Set(ctx, "k1", "look", 0)
	if err != nil {
		t.Fatalf("db.Set failed: %v", err)
	}
	err = db.Set(ctx, "k2", "look", 0)
	if err != nil {
		t.Fatalf("db.Set failed: %v", err)
	}
	err = db.Set(ctx, "k3", "other", 0)
	if err != nil {
		t.Fatalf("db.Set failed: %v", err)
	}
	keys, err := db.FindByValue(ctx, "look")
	if err != nil {
		t.Fatalf("FindByValue failed: %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("Expected 2, got %d", len(keys))
	}
}

// TestStoreDelete checks the behavior of the Delete method.
func TestStoreDelete(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()

	// Delete an existing key
	err := db.Set(ctx, "delKey", "someValue", 0)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	err = db.Delete(ctx, "delKey")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	_, err = db.Get(ctx, "delKey")
	if !IsKeyNotFound(err) {
		t.Errorf("Expected ErrKeyNotFound after delete, got %v", err)
	}

	// Delete a missing key
	err = db.Delete(ctx, "notHere")
	if !IsKeyNotFound(err) {
		t.Errorf("Expected ErrKeyNotFound, got %v", err)
	}
}

// TestStoreDropAll checks the behavior of the DropAll method.
func TestStoreDropAll(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()

	for i := 0; i < 10; i++ {
		err := db.Set(ctx, fmt.Sprintf("key%d", i), i, 0)
		if err != nil {
			t.Fatalf("Set failed for key%d: %v", i, err)
		}
	}
	err := db.DropAll(ctx)
	if err != nil {
		t.Fatalf("DropAll failed: %v", err)
	}
	time.Sleep(50 * time.Millisecond)
	for i := 0; i < 10; i++ {
		_, err := db.Get(ctx, fmt.Sprintf("key%d", i))
		if !IsKeyNotFound(err) {
			t.Errorf("Expected ErrKeyNotFound after DropAll for key%d, got %v", i, err)
		}
	}
}

// TestStoreConcurrency checks parallel operations on the same key.
func TestStoreConcurrency(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()

	const workers = 5
	const increments = 100
	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < increments; j++ {
				_, err := db.Incr(ctx, "counter")
				if err != nil && !IsKeyNotFound(err) {
					t.Errorf("Incr error: %v", err)
				}
			}
		}()
	}
	wg.Wait()

	val, err := db.Get(ctx, "counter")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	expected := int64(workers * increments)
	if val.(int64) != expected {
		t.Errorf("Expected %d, got %v", expected, val)
	}
}

// TestStorePubSub checks methods related to PubSub.
func TestStorePubSub(t *testing.T) {
	db := withTestStore(t)
	key := "testChannel"

	ch1 := db.Subscribe(key)
	ch2 := db.Subscribe(key)

	// ListSubscriptions
	subs := db.ListSubscriptions()
	if len(subs) != 1 || subs[0] != key {
		t.Errorf("ListSubscriptions failed: %v", subs)
	}

	// CloseAllSubscriptionsForKey
	db.CloseAllSubscriptionsForKey(key)

	select {
	case _, ok := <-ch1:
		if ok {
			t.Error("Channel 1 not closed properly")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout checking channel 1")
	}

	select {
	case _, ok := <-ch2:
		if ok {
			t.Error("Channel 2 not closed properly")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout checking channel 2")
	}
}

// TestStoreTypeValidation checks for type errors when using incorrect operations.
func TestStoreTypeValidation(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()

	err := db.LPush(ctx, "listKey", "item")
	if err != nil {
		t.Fatalf("LPush failed: %v", err)
	}
	// Attempt HSet on a list
	err = db.HSet(ctx, "listKey", "field", "val", 0)
	if !IsInvalidType(err) {
		t.Errorf("Expected ErrInvalidType, got %v", err)
	}

	err = db.Set(ctx, "strKey", "val", 0)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	// Attempt LPush on a string
	err = db.LPush(ctx, "strKey", "item")
	if !IsInvalidType(err) {
		t.Errorf("Expected ErrInvalidType, got %v", err)
	}
}
