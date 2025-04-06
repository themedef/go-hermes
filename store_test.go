package hermes

import (
	"context"
	"fmt"
	"github.com/themedef/go-hermes/internal/contracts"
	"sync"
	"testing"
	"time"
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

func TestStoreSetAndGet(t *testing.T) {
	t.Run("Set and Get string value", func(t *testing.T) {
		db := withTestStore(t)
		ctx := context.Background()
		if err := db.Set(ctx, "hello", "world", 0); err != nil {
			t.Fatalf("Set failed: %v", err)
		}
		val, err := db.Get(ctx, "hello")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if val != "world" {
			t.Errorf("Expected 'world', got %v", val)
		}
	})
	t.Run("Get non-existent key", func(t *testing.T) {
		db := withTestStore(t)
		_, err := db.Get(context.Background(), "missingKey")
		if !IsKeyNotFound(err) {
			t.Errorf("Expected ErrKeyNotFound, got: %v", err)
		}
	})
	t.Run("Set with TTL and expiration", func(t *testing.T) {
		db := withTestStore(t)
		ctx := context.Background()
		if err := db.Set(ctx, "tempKey", "tempVal", 1); err != nil {
			t.Fatalf("Set failed: %v", err)
		}
		time.Sleep(2 * time.Second)
		_, err := db.Get(ctx, "tempKey")
		if !IsKeyExpired(err) && !IsKeyNotFound(err) {
			t.Errorf("Expected key expiration, got: %v", err)
		}
	})
	t.Run("Set with negative TTL", func(t *testing.T) {
		db := withTestStore(t)
		err := db.Set(context.Background(), "invalidTTL", "val", -5)
		if !IsInvalidTTL(err) {
			t.Errorf("Expected ErrInvalidTTL, got %v", err)
		}
	})
	t.Run("Set with empty key", func(t *testing.T) {
		db := withTestStore(t)
		err := db.Set(context.Background(), "", "val", 0)
		if !IsInvalidKey(err) {
			t.Errorf("Expected ErrInvalidKey, got %v", err)
		}
	})
}

func TestStoreEdgeCases(t *testing.T) {
	t.Run("Set nil value", func(t *testing.T) {
		db := withTestStore(t)
		ctx := context.Background()
		if err := db.Set(ctx, "nilValue", nil, 0); err != nil {
			t.Fatalf("Set failed: %v", err)
		}
		val, err := db.Get(ctx, "nilValue")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if val != nil {
			t.Errorf("Expected nil, got %v", val)
		}
	})
	t.Run("Very large TTL", func(t *testing.T) {
		db := withTestStore(t)
		ttl := 365 * 24 * 3600
		if err := db.Set(context.Background(), "hugeTTL", "data", ttl); err != nil {
			t.Fatalf("Set failed: %v", err)
		}
		val, err := db.Get(context.Background(), "hugeTTL")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if val != "data" {
			t.Errorf("Key with large TTL mismatch")
		}
	})
}

func TestStoreSetNXAndSetXX(t *testing.T) {
	t.Run("SetNX on existing key", func(t *testing.T) {
		db := withTestStore(t)
		ctx := context.Background()
		if err := db.Set(ctx, "nxKey", "val", 0); err != nil {
			t.Fatalf("Set failed: %v", err)
		}
		ok, err := db.SetNX(ctx, "nxKey", "newVal", 0)
		if !IsKeyExists(err) {
			t.Fatalf("Expected ErrKeyExists, got err=%v", err)
		}
		if ok {
			t.Error("SetNX returned ok=true, but key already exists")
		}
	})
	t.Run("SetXX on non-existent key", func(t *testing.T) {
		db := withTestStore(t)
		ok, err := db.SetXX(context.Background(), "xxKey", "val", 0)
		if !IsKeyNotFound(err) {
			t.Fatalf("Expected ErrKeyNotFound, got err=%v", err)
		}
		if ok {
			t.Error("SetXX returned ok=true, but key doesn't exist")
		}
	})
}

func TestStoreSetCAS(t *testing.T) {
	t.Run("SetCAS successful", func(t *testing.T) {
		db := withTestStore(t)
		ctx := context.Background()
		if err := db.Set(ctx, "casKey", "old", 0); err != nil {
			t.Fatalf("Set failed: %v", err)
		}
		if err := db.SetCAS(ctx, "casKey", "old", "new", 0); err != nil {
			t.Fatalf("SetCAS failed: %v", err)
		}
		val, err := db.Get(ctx, "casKey")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if val != "new" {
			t.Error("SetCAS did not update value")
		}
	})
	t.Run("SetCAS mismatch", func(t *testing.T) {
		db := withTestStore(t)
		ctx := context.Background()
		if err := db.Set(ctx, "casKey2", "init", 0); err != nil {
			t.Fatalf("Set failed: %v", err)
		}
		err := db.SetCAS(ctx, "casKey2", "wrongOld", "new", 0)
		if !IsValueMismatch(err) {
			t.Errorf("Expected ErrValueMismatch, got %v", err)
		}
	})
}

func TestStoreGetSet(t *testing.T) {
	t.Run("GetSet basic usage", func(t *testing.T) {
		db := withTestStore(t)
		ctx := context.Background()
		oldVal, err := db.GetSet(ctx, "newKey", "newVal", 0)
		if err != nil {
			t.Fatalf("GetSet failed: %v", err)
		}
		if oldVal != nil {
			t.Errorf("Expected nil for new key, got %v", oldVal)
		}
		if _, err = db.GetSet(ctx, "newKey", "updatedVal", 0); err != nil {
			t.Fatalf("Second GetSet failed: %v", err)
		}
		val, err := db.Get(ctx, "newKey")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if val != "updatedVal" {
			t.Error("GetSet didn't update value")
		}
	})
}

func TestStoreIncrDecr(t *testing.T) {
	t.Run("Incr on new key", func(t *testing.T) {
		db := withTestStore(t)
		val, err := db.Incr(context.Background(), "counter")
		if err != nil {
			t.Fatalf("Incr failed: %v", err)
		}
		if val != 1 {
			t.Errorf("Expected 1, got %d", val)
		}
	})
	t.Run("Incr on existing integer", func(t *testing.T) {
		db := withTestStore(t)
		ctx := context.Background()
		_, _ = db.Incr(ctx, "counter2")
		_, _ = db.Incr(ctx, "counter2")
		val, err := db.Incr(ctx, "counter2")
		if err != nil {
			t.Fatalf("Incr failed: %v", err)
		}
		if val != 3 {
			t.Errorf("Expected 3, got %d", val)
		}
	})
	t.Run("Incr on non-integer", func(t *testing.T) {
		db := withTestStore(t)
		ctx := context.Background()
		if err := db.Set(ctx, "str", "hello", 0); err != nil {
			t.Fatalf("Set failed: %v", err)
		}
		_, err := db.Incr(ctx, "str")
		if !IsInvalidValueType(err) {
			t.Errorf("Expected ErrInvalidValueType, got %v", err)
		}
	})
	t.Run("Decr new key", func(t *testing.T) {
		db := withTestStore(t)
		val, err := db.Decr(context.Background(), "decKey")
		if err != nil {
			t.Fatalf("Decr failed: %v", err)
		}
		if val != -1 {
			t.Errorf("Expected -1, got %d", val)
		}
	})
}

func TestStoreListOperations(t *testing.T) {
	t.Run("LPush / LPop basic", func(t *testing.T) {
		db := withTestStore(t)
		ctx := context.Background()
		if err := db.LPush(ctx, "myList", 1); err != nil {
			t.Fatalf("LPush failed: %v", err)
		}
		if err := db.LPush(ctx, "myList", 2); err != nil {
			t.Fatalf("LPush failed: %v", err)
		}
		val, err := db.LPop(ctx, "myList")
		if err != nil {
			t.Fatalf("LPop failed: %v", err)
		}
		if val != 2 {
			t.Errorf("Expected 2, got %v", val)
		}
	})
	t.Run("RPush / RPop basic", func(t *testing.T) {
		db := withTestStore(t)
		ctx := context.Background()
		if err := db.RPush(ctx, "myList", "a"); err != nil {
			t.Fatalf("RPush failed: %v", err)
		}
		if err := db.RPush(ctx, "myList", "b"); err != nil {
			t.Fatalf("RPush failed: %v", err)
		}
		val, err := db.RPop(ctx, "myList")
		if err != nil {
			t.Fatalf("RPop failed: %v", err)
		}
		if val != "b" {
			t.Errorf("Expected 'b', got %v", val)
		}
	})
	t.Run("Pop from empty list or not found key", func(t *testing.T) {
		db := withTestStore(t)
		_, err := db.LPop(context.Background(), "noList")
		if !IsKeyNotFound(err) {
			t.Errorf("Expected ErrKeyNotFound, got %v", err)
		}
	})
}

func TestStoreHashOperations(t *testing.T) {
	t.Run("HSet / HGet success", func(t *testing.T) {
		db := withTestStore(t)
		ctx := context.Background()
		if err := db.HSet(ctx, "user:1", "name", "Test", 0); err != nil {
			t.Fatalf("HSet failed: %v", err)
		}
		val, err := db.HGet(ctx, "user:1", "name")
		if err != nil {
			t.Fatalf("HGet failed: %v", err)
		}
		if val != "Test" {
			t.Errorf("Expected 'Test', got %v", val)
		}
	})
	t.Run("HGet from nonexistent key", func(t *testing.T) {
		db := withTestStore(t)
		_, err := db.HGet(context.Background(), "nobody", "field")
		if !IsKeyNotFound(err) {
			t.Errorf("Expected ErrKeyNotFound, got %v", err)
		}
	})
	t.Run("HDel field", func(t *testing.T) {
		db := withTestStore(t)
		ctx := context.Background()
		if err := db.HSet(ctx, "session", "token", "abc123", 0); err != nil {
			t.Fatalf("HSet failed: %v", err)
		}
		if err := db.HDel(ctx, "session", "token"); err != nil {
			t.Fatalf("HDel failed: %v", err)
		}
		_, err := db.HGet(ctx, "session", "token")
		if !IsKeyNotFound(err) {
			t.Errorf("Field should be gone, expected ErrKeyNotFound")
		}
	})
	t.Run("HGetAll", func(t *testing.T) {
		db := withTestStore(t)
		ctx := context.Background()
		if err := db.HSet(ctx, "product:1", "name", "Laptop", 0); err != nil {
			t.Fatalf("HSet failed: %v", err)
		}
		if err := db.HSet(ctx, "product:1", "price", 1000, 0); err != nil {
			t.Fatalf("HSet failed: %v", err)
		}
		fields, err := db.HGetAll(ctx, "product:1")
		if err != nil {
			t.Fatalf("HGetAll failed: %v", err)
		}
		if len(fields) != 2 || fields["name"] != "Laptop" || fields["price"] != 1000 {
			t.Error("HGetAll returned incorrect data")
		}
	})
}

func TestStoreHExistsHLen(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()
	if err := db.HSet(ctx, "hash", "field1", "val1", 0); err != nil {
		t.Fatalf("HSet failed: %v", err)
	}
	if err := db.HSet(ctx, "hash", "field2", "val2", 0); err != nil {
		t.Fatalf("HSet failed: %v", err)
	}
	t.Run("HExists", func(t *testing.T) {
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
	})
	t.Run("HLen", func(t *testing.T) {
		count, err := db.HLen(ctx, "hash")
		if err != nil {
			t.Fatalf("HLen error: %v", err)
		}
		if count != 2 {
			t.Errorf("HLen expected 2, got %d", count)
		}
	})
}

func TestStoreExists(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()
	if err := db.Set(ctx, "check", "val", 0); err != nil {
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

func TestStoreExpire(t *testing.T) {
	t.Run("Expire on existing key", func(t *testing.T) {
		db := withTestStore(t)
		ctx := context.Background()
		if err := db.Set(ctx, "upKey", "val", 1); err != nil {
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
	})
	t.Run("Expire on non-existent key", func(t *testing.T) {
		db := withTestStore(t)
		success, err := db.Expire(context.Background(), "missing", 10)
		if err == nil || !IsKeyNotFound(err) {
			t.Errorf("Expected ErrKeyNotFound for missing key, got err=%v, success=%v", err, success)
		}
	})
}

func TestStoreType(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()
	if err := db.Set(ctx, "string", "val", 0); err != nil {
		t.Fatalf("Setup Set failed: %v", err)
	}
	if err := db.LPush(ctx, "list", "elem"); err != nil {
		t.Fatalf("Setup LPush failed: %v", err)
	}
	if err := db.HSet(ctx, "hash", "field", "val", 0); err != nil {
		t.Fatalf("Setup HSet failed: %v", err)
	}
	tests := []struct {
		key      string
		expected DataType
	}{
		{"string", String},
		{"list", List},
		{"hash", Hash},
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

func TestStoreGetWithDetails(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()
	if err := db.Set(ctx, "detailed", "value", 10); err != nil {
		t.Fatalf("Setup Set failed: %v", err)
	}
	val, ttl, err := db.GetWithDetails(ctx, "detailed")
	if err != nil {
		t.Fatalf("GetWithDetails failed: %v", err)
	}
	if val != "value" || ttl <= 0 {
		t.Errorf("Unexpected values: val=%v, ttl=%d", val, ttl)
	}
	if err := db.Set(ctx, "temp", "val", 1); err != nil {
		t.Fatalf("Setup Set failed: %v", err)
	}
	time.Sleep(1100 * time.Millisecond)
	_, _, err = db.GetWithDetails(ctx, "temp")
	if !IsKeyExpired(err) && !IsKeyNotFound(err) {
		t.Errorf("Expected ErrKeyExpired (or ErrKeyNotFound), got %v", err)
	}
}

func TestStoreRename(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()
	if err := db.Set(ctx, "oldKey", "value", 0); err != nil {
		t.Fatalf("Setup Set failed: %v", err)
	}
	t.Run("Successful rename", func(t *testing.T) {
		if err := db.Rename(ctx, "oldKey", "newKey"); err != nil {
			t.Fatalf("Rename failed: %v", err)
		}
		_, err := db.Get(ctx, "newKey")
		if err != nil {
			t.Fatalf("Get after rename failed: %v", err)
		}
	})
	t.Run("Rename to existing key", func(t *testing.T) {
		if err := db.Set(ctx, "targetKey", "existing", 0); err != nil {
			t.Fatalf("Setup Set failed: %v", err)
		}
		err := db.Rename(ctx, "newKey", "targetKey")
		if !IsKeyExists(err) {
			t.Errorf("Expected ErrKeyExists, got %v", err)
		}
	})
}

func TestStoreFindByValue(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()
	if err := db.Set(ctx, "k1", "look", 0); err != nil {
		t.Fatalf("db.Set failed: %v", err)
	}
	if err := db.Set(ctx, "k2", "look", 0); err != nil {
		t.Fatalf("db.Set failed: %v", err)
	}
	if err := db.Set(ctx, "k3", "other", 0); err != nil {
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

func TestStoreDelete(t *testing.T) {
	t.Run("Delete existing key", func(t *testing.T) {
		db := withTestStore(t)
		ctx := context.Background()
		if err := db.Set(ctx, "delKey", "someValue", 0); err != nil {
			t.Fatalf("Set failed: %v", err)
		}
		if err := db.Delete(ctx, "delKey"); err != nil {
			t.Fatalf("Delete failed: %v", err)
		}
		_, err := db.Get(ctx, "delKey")
		if !IsKeyNotFound(err) {
			t.Errorf("Key should be deleted, expected ErrKeyNotFound")
		}
	})
	t.Run("Delete non-existent key", func(t *testing.T) {
		db := withTestStore(t)
		err := db.Delete(context.Background(), "notHere")
		if !IsKeyNotFound(err) {
			t.Errorf("Expected ErrKeyNotFound, got %v", err)
		}
	})
}

func TestStoreDropAll(t *testing.T) {
	t.Run("DropAll clears the DB", func(t *testing.T) {
		db := withTestStore(t)
		ctx := context.Background()
		for i := 0; i < 10; i++ {
			if err := db.Set(ctx, fmt.Sprintf("key%d", i), i, 0); err != nil {
				t.Fatalf("Set failed for key%d: %v", i, err)
			}
		}
		if err := db.DropAll(ctx); err != nil {
			t.Fatalf("DropAll failed: %v", err)
		}
		time.Sleep(50 * time.Millisecond)
		for i := 0; i < 10; i++ {
			_, err := db.Get(ctx, fmt.Sprintf("key%d", i))
			if !IsKeyNotFound(err) {
				t.Errorf("Expected ErrKeyNotFound after DropAll for key%d, got %v", i, err)
			}
		}
	})
}

func TestStoreConcurrency(t *testing.T) {
	t.Run("Concurrent writes to same key", func(t *testing.T) {
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
					if _, err := db.Incr(ctx, "counter"); err != nil {
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
	})
}

func TestPubSubManagement(t *testing.T) {
	db := withTestStore(t)
	key := "testChannel"
	ch1 := db.Subscribe(key)
	ch2 := db.Subscribe(key)
	t.Run("ListSubscriptions", func(t *testing.T) {
		subs := db.ListSubscriptions()
		if len(subs) != 1 || subs[0] != key {
			t.Errorf("ListSubscriptions failed: %v", subs)
		}
	})
	t.Run("CloseAllSubscriptions", func(t *testing.T) {
		db.CloseAllSubscriptionsForKey(key)
		select {
		case _, ok := <-ch1:
			if ok {
				t.Error("Channel 1 not closed")
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("Timeout checking channel 1")
		}
		select {
		case _, ok := <-ch2:
			if ok {
				t.Error("Channel 2 not closed")
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("Timeout checking channel 2")
		}
	})
}

func TestTypeValidation(t *testing.T) {
	db := withTestStore(t)
	ctx := context.Background()
	if err := db.LPush(ctx, "listKey", "item"); err != nil {
		t.Fatalf("LPush failed: %v", err)
	}
	err := db.HSet(ctx, "listKey", "field", "val", 0)
	if !IsInvalidType(err) {
		t.Errorf("Expected ErrInvalidType, got %v", err)
	}
	if err := db.Set(ctx, "strKey", "val", 0); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	err = db.LPush(ctx, "strKey", "item")
	if !IsInvalidType(err) {
		t.Errorf("Expected ErrInvalidType, got %v", err)
	}
}
