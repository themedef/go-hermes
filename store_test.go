package hermes

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func newTestStore() *DB {
	return NewStore(Config{})
}

func TestStoreSetAndGet(t *testing.T) {
	t.Run("Set and Get string value", func(t *testing.T) {
		db := newTestStore()
		defer db.Close()

		ctx := context.Background()
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
	})
	t.Run("Get non-existent key", func(t *testing.T) {
		db := newTestStore()
		defer db.Close()

		_, err := db.Get(context.Background(), "missingKey")
		if !errors.Is(err, ErrKeyNotFound) {
			t.Errorf("Expected ErrKeyNotFound, got: %v", err)
		}
	})

	t.Run("Set with TTL and expiration", func(t *testing.T) {
		db := newTestStore()
		defer db.Close()

		ctx := context.Background()
		err := db.Set(ctx, "tempKey", "tempVal", 1)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		time.Sleep(2 * time.Second)
		_, err = db.Get(ctx, "tempKey")
		if !errors.Is(err, ErrKeyExpired) && !errors.Is(err, ErrKeyNotFound) {
			t.Errorf("Expected key expiration, got: %v", err)
		}
	})

	t.Run("Set with negative TTL", func(t *testing.T) {
		db := newTestStore()
		defer db.Close()

		err := db.Set(context.Background(), "invalidTTL", "val", -5)
		if !errors.Is(err, ErrInvalidTTL) {
			t.Errorf("Expected ErrInvalidTTL, got %v", err)
		}
	})

	t.Run("Set with empty key", func(t *testing.T) {
		db := newTestStore()
		defer db.Close()

		err := db.Set(context.Background(), "", "val", 0)
		if !errors.Is(err, ErrInvalidKey) {
			t.Errorf("Expected ErrInvalidKey, got %v", err)
		}
	})
}

func TestStoreEdgeCases(t *testing.T) {
	t.Run("Set nil value", func(t *testing.T) {
		db := newTestStore()
		defer db.Close()

		err := db.Set(context.Background(), "nilValue", nil, 0)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}
		val, err := db.Get(context.Background(), "nilValue")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if val != nil {
			t.Errorf("Expected nil, got %v", val)
		}
	})

	t.Run("Very large TTL", func(t *testing.T) {
		db := newTestStore()
		defer db.Close()

		ttl := 365 * 24 * 3600
		err := db.Set(context.Background(), "hugeTTL", "data", ttl)
		if err != nil {
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
		db := newTestStore()
		defer db.Close()

		ctx := context.Background()
		_ = db.Set(ctx, "nxKey", "val", 0)

		ok, err := db.SetNX(ctx, "nxKey", "newVal", 0)
		if !errors.Is(err, ErrKeyExists) {
			t.Fatalf("Expected ErrKeyExists, got err=%v", err)
		}
		if ok {
			t.Error("SetNX returned ok=true, but key already exists")
		}
	})

	t.Run("SetXX on non-existent key", func(t *testing.T) {
		db := newTestStore()
		defer db.Close()

		ok, err := db.SetXX(context.Background(), "xxKey", "val", 0)
		if !errors.Is(err, ErrKeyNotFound) {
			t.Fatalf("Expected ErrKeyNotFound, got err=%v", err)
		}
		if ok {
			t.Error("SetXX returned ok=true, but key doesn't exist")
		}
	})
}

func TestStoreSetCAS(t *testing.T) {
	t.Run("SetCAS successful", func(t *testing.T) {
		db := newTestStore()
		defer db.Close()

		ctx := context.Background()
		_ = db.Set(ctx, "casKey", "old", 0)
		err := db.SetCAS(ctx, "casKey", "old", "new", 0)
		if err != nil {
			t.Fatalf("SetCAS failed: %v", err)
		}
		val, _ := db.Get(ctx, "casKey")
		if val != "new" {
			t.Error("SetCAS did not update value")
		}
	})

	t.Run("SetCAS mismatch", func(t *testing.T) {
		db := newTestStore()
		defer db.Close()

		ctx := context.Background()
		_ = db.Set(ctx, "casKey2", "init", 0)
		err := db.SetCAS(ctx, "casKey2", "wrongOld", "new", 0)
		if !errors.Is(err, ErrValueMismatch) {
			t.Errorf("Expected ErrValueMismatch, got %v", err)
		}
	})
}

func TestStoreGetSet(t *testing.T) {
	t.Run("GetSet basic usage", func(t *testing.T) {
		db := newTestStore()
		defer db.Close()

		ctx := context.Background()
		oldVal, err := db.GetSet(ctx, "newKey", "newVal", 0)
		if err != nil {
			t.Fatalf("GetSet failed: %v", err)
		}
		if oldVal != nil {
			t.Errorf("Expected nil for new key, got %v", oldVal)
		}

		_, err = db.GetSet(ctx, "newKey", "updatedVal", 0)
		if err != nil {
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
		db := newTestStore()
		defer db.Close()

		val, err := db.Incr(context.Background(), "counter")
		if err != nil {
			t.Fatalf("Incr failed: %v", err)
		}
		if val != 1 {
			t.Errorf("Expected 1, got %d", val)
		}
	})

	t.Run("Incr on existing integer", func(t *testing.T) {
		db := newTestStore()
		defer db.Close()

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
		db := newTestStore()
		defer db.Close()

		ctx := context.Background()
		_ = db.Set(ctx, "str", "hello", 0)
		_, err := db.Incr(ctx, "str")
		if !errors.Is(err, ErrInvalidValueType) {
			t.Errorf("Expected ErrInvalidValueType, got %v", err)
		}
	})

	t.Run("Decr new key", func(t *testing.T) {
		db := newTestStore()
		defer db.Close()

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
		db := newTestStore()
		defer db.Close()

		ctx := context.Background()
		_ = db.LPush(ctx, "myList", 1)
		_ = db.LPush(ctx, "myList", 2)
		val, err := db.LPop(ctx, "myList")
		if err != nil {
			t.Fatalf("LPop failed: %v", err)
		}
		if val != 2 {
			t.Errorf("Expected 2, got %v", val)
		}
	})

	t.Run("RPush / RPop basic", func(t *testing.T) {
		db := newTestStore()
		defer db.Close()

		ctx := context.Background()
		_ = db.RPush(ctx, "myList", "a")
		_ = db.RPush(ctx, "myList", "b")
		val, err := db.RPop(ctx, "myList")
		if err != nil {
			t.Fatalf("RPop failed: %v", err)
		}
		if val != "b" {
			t.Errorf("Expected 'b', got %v", val)
		}
	})

	t.Run("Pop from empty list or not found key", func(t *testing.T) {
		db := newTestStore()
		defer db.Close()

		_, err := db.LPop(context.Background(), "noList")
		if !errors.Is(err, ErrKeyNotFound) {
			t.Errorf("Expected ErrKeyNotFound, got %v", err)
		}
	})
}

func TestStoreHashOperations(t *testing.T) {
	t.Run("HSet / HGet success", func(t *testing.T) {
		db := newTestStore()
		defer db.Close()

		ctx := context.Background()
		err := db.HSet(ctx, "user:1", "name", "Test", 0)
		if err != nil {
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
		db := newTestStore()
		defer db.Close()

		_, err := db.HGet(context.Background(), "nobody", "field")
		if !errors.Is(err, ErrKeyNotFound) {
			t.Errorf("Expected ErrKeyNotFound, got %v", err)
		}
	})

	t.Run("HDel field", func(t *testing.T) {
		db := newTestStore()
		defer db.Close()

		ctx := context.Background()
		_ = db.HSet(ctx, "session", "token", "abc123", 0)
		err := db.HDel(ctx, "session", "token")
		if err != nil {
			t.Fatalf("HDel failed: %v", err)
		}
		_, err = db.HGet(ctx, "session", "token")
		if !errors.Is(err, ErrKeyNotFound) {
			t.Errorf("Field should be gone, expected ErrKeyNotFound")
		}
	})

	t.Run("HGetAll", func(t *testing.T) {
		db := newTestStore()
		defer db.Close()

		ctx := context.Background()
		_ = db.HSet(ctx, "product:1", "name", "Laptop", 0)
		_ = db.HSet(ctx, "product:1", "price", 1000, 0)
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
	db := newTestStore()
	defer db.Close()

	ctx := context.Background()
	err := db.HSet(ctx, "hash", "field1", "val1", 0)
	if err != nil {
		t.Fatalf("HSet failed: %v", err)
	}
	err = db.HSet(ctx, "hash", "field2", "val2", 0)
	if err != nil {
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
	db := newTestStore()
	defer db.Close()

	ctx := context.Background()
	_ = db.Set(ctx, "check", "val", 0)
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

func TestStoreUpdateTTL(t *testing.T) {
	t.Run("UpdateTTL on existing key", func(t *testing.T) {
		db := newTestStore()
		defer db.Close()

		ctx := context.Background()
		_ = db.Set(ctx, "upKey", "val", 1)
		err := db.UpdateTTL(ctx, "upKey", 3)
		if err != nil {
			t.Fatalf("UpdateTTL failed: %v", err)
		}

		time.Sleep(1500 * time.Millisecond)
		val, err := db.Get(ctx, "upKey")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if val != "val" {
			t.Errorf("Value mismatch after UpdateTTL")
		}
	})

	t.Run("UpdateTTL on non-existent key", func(t *testing.T) {
		db := newTestStore()
		defer db.Close()

		err := db.UpdateTTL(context.Background(), "missing", 10)
		if !errors.Is(err, ErrKeyNotFound) {
			t.Errorf("Expected ErrKeyNotFound, got %v", err)
		}
	})
}

func TestStoreType(t *testing.T) {
	db := newTestStore()
	defer db.Close()

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
			t.Errorf("Type mismatch for %s: got %v, expected %v",
				tt.key, dtype, tt.expected)
		}
	}
}

func TestStoreGetWithDetails(t *testing.T) {
	db := newTestStore()
	defer db.Close()

	ctx := context.Background()

	err := db.Set(ctx, "detailed", "value", 10)
	if err != nil {
		t.Fatalf("Setup Set failed: %v", err)
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
	if !errors.Is(err, ErrKeyExpired) && !errors.Is(err, ErrKeyNotFound) {
		t.Errorf("Expected ErrKeyExpired (or ErrKeyNotFound), got %v", err)
	}
}

func TestStoreRename(t *testing.T) {
	db := newTestStore()
	defer db.Close()

	ctx := context.Background()

	err := db.Set(ctx, "oldKey", "value", 0)
	if err != nil {
		t.Fatalf("Setup Set failed: %v", err)
	}

	t.Run("Successful rename", func(t *testing.T) {
		err := db.Rename(ctx, "oldKey", "newKey")
		if err != nil {
			t.Fatalf("Rename failed: %v", err)
		}

		_, err = db.Get(ctx, "newKey")
		if err != nil {
			t.Fatalf("Get after rename failed: %v", err)
		}
	})

	t.Run("Rename to existing key", func(t *testing.T) {
		err := db.Set(ctx, "targetKey", "existing", 0)
		if err != nil {
			t.Fatalf("Setup Set failed: %v", err)
		}

		err = db.Rename(ctx, "newKey", "targetKey")
		if !errors.Is(err, ErrKeyExists) {
			t.Errorf("Expected ErrKeyExists, got %v", err)
		}
	})
}

func TestStoreFindByValue(t *testing.T) {
	db := newTestStore()
	defer db.Close()

	ctx := context.Background()
	_ = db.Set(ctx, "k1", "look", 0)
	_ = db.Set(ctx, "k2", "look", 0)
	_ = db.Set(ctx, "k3", "other", 0)

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
		db := newTestStore()
		defer db.Close()

		ctx := context.Background()
		_ = db.Set(ctx, "delKey", "someValue", 0)

		err := db.Delete(ctx, "delKey")
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		_, err = db.Get(ctx, "delKey")
		if !errors.Is(err, ErrKeyNotFound) {
			t.Errorf("Key should be deleted, expected ErrKeyNotFound")
		}
	})

	t.Run("Delete non-existent key", func(t *testing.T) {
		db := newTestStore()
		defer db.Close()

		err := db.Delete(context.Background(), "notHere")
		if !errors.Is(err, ErrKeyNotFound) {
			t.Errorf("Expected ErrKeyNotFound, got %v", err)
		}
	})
}

func TestStoreDropAll(t *testing.T) {
	t.Run("DropAll clears the DB", func(t *testing.T) {
		db := newTestStore()
		defer db.Close()

		ctx := context.Background()
		for i := 0; i < 10; i++ {
			db.Set(ctx, fmt.Sprintf("key%d", i), i, 0)
		}
		err := db.DropAll(ctx)
		if err != nil {
			t.Fatalf("DropAll failed: %v", err)
		}

		time.Sleep(50 * time.Millisecond)

		for i := 0; i < 10; i++ {
			_, err := db.Get(ctx, fmt.Sprintf("key%d", i))
			if !errors.Is(err, ErrKeyNotFound) {
				t.Errorf("Expected ErrKeyNotFound after DropAll, got %v", err)
			}
		}
	})
}

func TestStoreConcurrency(t *testing.T) {
	t.Run("Concurrent writes to same key", func(t *testing.T) {
		db := newTestStore()
		defer db.Close()

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
					if err != nil {
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

func TestStorePubSub(t *testing.T) {
	t.Run("Subscribe and Publish", func(t *testing.T) {
		db := newTestStore()
		defer db.Close()

		ch := db.pubsub.Subscribe("channel")
		defer db.pubsub.Unsubscribe("channel", ch)

		db.pubsub.Publish("channel", "hello")

		select {
		case msg := <-ch:
			if msg != "hello" {
				t.Errorf("Expected 'hello', got %s", msg)
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("Did not receive message in time")
		}
	})
}

func TestPubSubManagement(t *testing.T) {
	db := newTestStore()
	defer db.Close()

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
	db := newTestStore()
	defer db.Close()

	ctx := context.Background()

	err := db.LPush(ctx, "listKey", "item")
	if err != nil {
		t.Fatalf("Setup LPush failed: %v", err)
	}

	err = db.HSet(ctx, "listKey", "field", "val", 0)
	if !errors.Is(err, ErrInvalidType) {
		t.Errorf("Expected ErrInvalidType, got %v", err)
	}

	err = db.Set(ctx, "strKey", "val", 0)
	if err != nil {
		t.Fatalf("Setup Set failed: %v", err)
	}

	err = db.LPush(ctx, "strKey", "item")
	if !errors.Is(err, ErrInvalidType) {
		t.Errorf("Expected ErrInvalidType, got %v", err)
	}
}
