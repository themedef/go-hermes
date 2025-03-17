package hermes

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func testStore() *DB {
	db := NewStore(Config{
		CleanupInterval: 50 * time.Millisecond,
		EnableLogging:   false,
	})
	return db
}

func TestBasicCRUDOperations(t *testing.T) {
	t.Run("Set and Get with TTL", func(t *testing.T) {
		db := testStore()
		defer db.ClosePubSub()

		ctx := context.Background()
		if err := db.Set(ctx, "key1", "value1", 1); err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		val, ok, err := db.Get(ctx, "key1")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if !ok || val != "value1" {
			t.Error("Failed to get freshly set value")
		}
	})

	t.Run("Get non-existent key", func(t *testing.T) {
		db := testStore()

		defer db.ClosePubSub()

		_, ok, err := db.Get(context.Background(), "missing")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if ok {
			t.Error("Should return not found for missing key")
		}
	})

	t.Run("Delete existing key", func(t *testing.T) {
		db := testStore()

		defer db.ClosePubSub()

		ctx := context.Background()
		if err := db.Set(ctx, "delKey", "value", 0); err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		deleted, err := db.Delete(ctx, "delKey")
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}
		if !deleted {
			t.Error("Delete should report true for existing key")
		}

		_, ok, err := db.Get(ctx, "delKey")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if ok {
			t.Error("Key should be deleted")
		}
	})
}

func TestConditionalOperations(t *testing.T) {
	t.Run("SetNX new key", func(t *testing.T) {
		db := testStore()

		defer db.ClosePubSub()

		ok, err := db.SetNX(context.Background(), "newKey", "val", 0)
		if err != nil {
			t.Fatalf("SetNX failed: %v", err)
		}
		if !ok {
			t.Error("SetNX should succeed for new key")
		}
	})

	t.Run("SetNX existing key", func(t *testing.T) {
		db := testStore()

		defer db.ClosePubSub()

		ctx := context.Background()
		if err := db.Set(ctx, "existKey", "val", 0); err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		ok, err := db.SetNX(ctx, "existKey", "newVal", 0)
		if err != nil {
			t.Fatalf("SetNX failed: %v", err)
		}
		if ok {
			t.Error("SetNX should fail for existing key")
		}
	})

	t.Run("CAS successful update", func(t *testing.T) {
		db := testStore()

		defer db.ClosePubSub()

		ctx := context.Background()
		if err := db.Set(ctx, "casKey", "old", 0); err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		ok, err := db.SetCAS(ctx, "casKey", "old", "new", 0)
		if err != nil {
			t.Fatalf("SetCAS failed: %v", err)
		}
		if !ok {
			t.Error("CAS should succeed with correct old value")
		}

		val, _, err := db.Get(ctx, "casKey")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if val != "new" {
			t.Error("Value not updated after CAS")
		}
	})
}

func TestListOperations(t *testing.T) {
	t.Run("LPush and LPop", func(t *testing.T) {
		db := testStore()

		defer db.ClosePubSub()

		ctx := context.Background()
		if err := db.LPush(ctx, "list", "a"); err != nil {
			t.Fatalf("LPush failed: %v", err)
		}
		if err := db.LPush(ctx, "list", "b"); err != nil {
			t.Fatalf("LPush failed: %v", err)
		}

		val, ok, err := db.LPop(ctx, "list")
		if err != nil {
			t.Fatalf("LPop failed: %v", err)
		}
		if !ok || val != "b" {
			t.Error("LPop should return last pushed item")
		}
	})

	t.Run("RPush and RPop", func(t *testing.T) {
		db := testStore()

		defer db.ClosePubSub()

		ctx := context.Background()
		if err := db.RPush(ctx, "list", "a"); err != nil {
			t.Fatalf("RPush failed: %v", err)
		}
		if err := db.RPush(ctx, "list", "b"); err != nil {
			t.Fatalf("RPush failed: %v", err)
		}

		val, ok, err := db.RPop(ctx, "list")
		if err != nil {
			t.Fatalf("RPop failed: %v", err)
		}
		if !ok || val != "b" {
			t.Error("RPop should return last pushed item")
		}
	})

	t.Run("Pop from empty list", func(t *testing.T) {
		db := testStore()

		defer db.ClosePubSub()

		ctx := context.Background()
		if err := db.LPush(ctx, "emptyList", "item"); err != nil {
			t.Fatalf("LPush failed: %v", err)
		}
		if _, _, err := db.LPop(ctx, "emptyList"); err != nil {
			t.Fatalf("LPop failed: %v", err)
		}

		_, ok, err := db.LPop(ctx, "emptyList")
		if err != nil {
			t.Fatalf("LPop failed: %v", err)
		}
		if ok {
			t.Error("Should not pop from empty list")
		}
	})
}

func TestConcurrency(t *testing.T) {
	t.Run("Parallel increments", func(t *testing.T) {
		db := testStore()

		defer db.ClosePubSub()

		const workers = 10
		const iterations = 100
		var wg sync.WaitGroup
		wg.Add(workers)

		for i := 0; i < workers; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < iterations; j++ {
					if _, _, err := db.Incr(context.Background(), "counter"); err != nil {
						t.Errorf("Incr failed: %v", err)
					}
				}
			}()
		}

		wg.Wait()
		val, ok, err := db.Get(context.Background(), "counter")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if !ok || val.(int64) != int64(workers*iterations) {
			t.Errorf("Concurrency error: expected %d, got %v", workers*iterations, val)
		}
	})

	t.Run("Mixed operations", func(t *testing.T) {
		db := testStore()

		defer db.ClosePubSub()

		var wg sync.WaitGroup
		wg.Add(3)

		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				if err := db.Set(context.Background(), fmt.Sprintf("key%d", i), i, 0); err != nil {
					t.Errorf("Set failed: %v", err)
				}
			}
		}()

		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				if _, _, err := db.Get(context.Background(), fmt.Sprintf("key%d", i)); err != nil {
					t.Errorf("Get failed: %v", err)
				}
			}
		}()

		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				if _, err := db.Delete(context.Background(), fmt.Sprintf("key%d", i)); err != nil {
					t.Errorf("Delete failed: %v", err)
				}
			}
		}()

		wg.Wait()
	})
}

func TestTTLAndExpiration(t *testing.T) {
	t.Run("Key expiration", func(t *testing.T) {
		db := testStore()

		defer db.ClosePubSub()

		ctx := context.Background()
		if err := db.Set(ctx, "temp", "value", 1); err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		time.Sleep(1100 * time.Millisecond)
		_, ok, err := db.Get(ctx, "temp")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if ok {
			t.Error("Key should be expired")
		}
	})

	t.Run("Update TTL", func(t *testing.T) {
		db := testStore()

		defer db.ClosePubSub()

		ctx := context.Background()
		if err := db.Set(ctx, "key", "value", 1); err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		if err := db.UpdateTTL(ctx, "key", 2); err != nil {
			t.Fatalf("UpdateTTL failed: %v", err)
		}

		time.Sleep(1500 * time.Millisecond)
		_, ok, err := db.Get(ctx, "key")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if !ok {
			t.Error("Key should still be alive")
		}
	})
}

func TestHashOperations(t *testing.T) {
	t.Run("Basic HSET/HGET", func(t *testing.T) {
		db := testStore()
		defer db.ClosePubSub()

		ctx := context.Background()
		err := db.HSet(ctx, "user:1", "name", "Alice", 3600)
		if err != nil {
			t.Fatalf("HSET failed: %v", err)
		}

		val, ok, err := db.HGet(ctx, "user:1", "name")
		if err != nil {
			t.Fatalf("HGET failed: %v", err)
		}
		if !ok || val != "Alice" {
			t.Error("HGET returned invalid value")
		}
	})

	t.Run("HDEL field", func(t *testing.T) {
		db := testStore()
		defer db.ClosePubSub()

		ctx := context.Background()
		db.HSet(ctx, "user:1", "email", "alice@example.com", 0)

		err := db.HDel(ctx, "user:1", "email")
		if err != nil {
			t.Fatalf("HDEL failed: %v", err)
		}

		_, ok, _ := db.HGet(ctx, "user:1", "email")
		if ok {
			t.Error("Field not deleted")
		}
	})

	t.Run("HGETALL returns all fields", func(t *testing.T) {
		db := testStore()
		defer db.ClosePubSub()

		ctx := context.Background()
		db.HSet(ctx, "product:1", "title", "Laptop", 0)
		db.HSet(ctx, "product:1", "price", 999, 0)

		result, err := db.HGetAll(ctx, "product:1")
		if err != nil {
			t.Fatalf("HGETALL failed: %v", err)
		}

		if len(result) != 2 || result["title"] != "Laptop" || result["price"] != 999 {
			t.Error("HGETALL returned invalid data")
		}
	})

	t.Run("Hash expiration", func(t *testing.T) {
		db := NewStore(Config{
			CleanupInterval: 50 * time.Millisecond,
			EnableLogging:   false,
		})
		defer db.ClosePubSub()

		ctx := context.Background()
		db.HSet(ctx, "temp:hash", "data", "value", 1)

		time.Sleep(2 * time.Second)

		_, ok, _ := db.HGet(ctx, "temp:hash", "data")
		if ok {
			t.Error("Hash not expired")
		}
	})

	t.Run("Concurrent hash access", func(t *testing.T) {
		db := testStore()
		defer db.ClosePubSub()

		var wg sync.WaitGroup
		ctx := context.Background()

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				key := fmt.Sprintf("concurrent:%d", idx)
				db.HSet(ctx, key, "field", idx, 0)
				val, _, _ := db.HGet(ctx, key, "field")
				if val != idx {
					t.Errorf("Value mismatch for key %s", key)
				}
			}(i)
		}
		wg.Wait()
	})
}

func TestEdgeCasesForHashes(t *testing.T) {
	t.Run("Empty field name", func(t *testing.T) {
		db := testStore()
		defer db.ClosePubSub()

		ctx := context.Background()
		err := db.HSet(ctx, "empty", "", "value", 0)
		if err != nil {
			t.Fatalf("HSET failed: %v", err)
		}

		val, ok, _ := db.HGet(ctx, "empty", "")
		if !ok || val != "value" {
			t.Error("Empty field not handled")
		}
	})

	t.Run("Nonexistent hash key", func(t *testing.T) {
		db := testStore()
		defer db.ClosePubSub()

		ctx := context.Background()
		_, ok, err := db.HGet(ctx, "nonexistent", "field")
		if err != nil || ok {
			t.Error("Should handle nonexistent keys")
		}
	})

	t.Run("Update TTL for hash", func(t *testing.T) {
		db := testStore()
		defer db.ClosePubSub()

		ctx := context.Background()
		db.HSet(ctx, "update:ttl", "field", "value", 1)
		db.UpdateTTL(ctx, "update:ttl", 3600)

		time.Sleep(2 * time.Second)
		_, ok, _ := db.HGet(ctx, "update:ttl", "field")
		if !ok {
			t.Error("TTL update failed")
		}
	})
}

func TestPubSubSystem(t *testing.T) {
	t.Run("Basic publish-subscribe", func(t *testing.T) {
		db := testStore()

		defer db.ClosePubSub()

		ch := db.Subscribe("test")
		defer db.Unsubscribe("test", ch)

		db.Publish("test", "message")

		select {
		case msg := <-ch:
			if msg != "message" {
				t.Errorf("Received wrong message: %s", msg)
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("Message not received")
		}
	})

	t.Run("Multiple channels", func(t *testing.T) {
		db := testStore()

		defer db.ClosePubSub()

		ch1 := db.Subscribe("chan1")
		ch2 := db.Subscribe("chan2")
		defer db.Unsubscribe("chan1", ch1)
		defer db.Unsubscribe("chan2", ch2)

		db.Publish("chan1", "msg1")
		db.Publish("chan2", "msg2")

		received := 0
		for i := 0; i < 2; i++ {
			select {
			case <-ch1:
				received++
			case <-ch2:
				received++
			case <-time.After(100 * time.Millisecond):
			}
		}

		if received != 2 {
			t.Errorf("Expected 2 messages, got %d", received)
		}
	})
}

func TestTransactions(t *testing.T) {
	t.Run("Commit transaction", func(t *testing.T) {
		db := testStore()

		defer db.ClosePubSub()

		ctx := context.Background()
		tx := db.Transaction()
		tx.Begin()
		if err := tx.Set(ctx, "txKey", "value", 0); err != nil {
			t.Fatalf("Set in transaction failed: %v", err)
		}
		if err := tx.Commit(); err != nil {
			t.Fatalf("Transaction commit failed: %v", err)
		}

		val, ok, err := db.Get(ctx, "txKey")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if !ok || val != "value" {
			t.Error("Transaction commit failed")
		}
	})

	t.Run("Rollback transaction", func(t *testing.T) {
		db := NewStore(Config{
			CleanupInterval: 50 * time.Millisecond,
			EnableLogging:   false,
		})
		defer db.ClosePubSub()

		ctx := context.Background()
		tx := db.Transaction()
		tx.Begin()
		if err := tx.Set(ctx, "txKey", "value", 0); err != nil {
			t.Fatalf("Set in transaction failed: %v", err)
		}
		tx.Rollback()

		_, ok, err := db.Get(ctx, "txKey")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if ok {
			t.Error("Transaction rollback failed")
		}
	})
}

func TestAdvancedOperations(t *testing.T) {
	t.Run("Find by value", func(t *testing.T) {
		db := testStore()
		defer db.ClosePubSub()

		ctx := context.Background()
		if err := db.Set(ctx, "k1", "findMe", 0); err != nil {
			t.Fatalf("Set failed: %v", err)
		}
		if err := db.Set(ctx, "k2", "findMe", 0); err != nil {
			t.Fatalf("Set failed: %v", err)
		}
		if err := db.Set(ctx, "k3", "other", 0); err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		keys, err := db.FindByValue(ctx, "findMe")
		if err != nil {
			t.Fatalf("FindByValue failed: %v", err)
		}
		if len(keys) != 2 {
			t.Errorf("Expected 2 keys, got %d", len(keys))
		}
	})

}

func TestEdgeCases(t *testing.T) {
	t.Run("Empty key operations", func(t *testing.T) {
		db := testStore()

		defer db.ClosePubSub()

		ctx := context.Background()
		if err := db.Set(ctx, "", "value", 0); err != nil {
			t.Fatalf("Set failed: %v", err)
		}
		val, ok, err := db.Get(ctx, "")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if !ok || val != "value" {
			t.Error("Should handle empty key")
		}
	})

	t.Run("Very long TTL", func(t *testing.T) {
		db := testStore()

		defer db.ClosePubSub()

		ctx := context.Background()
		if err := db.Set(ctx, "long", "value", 365*24*3600); err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		_, ok, err := db.Get(ctx, "long")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if !ok {
			t.Error("Key with long TTL should exist")
		}
	})
}
