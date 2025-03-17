package transaction

import (
	"context"
	"github.com/themedef/go-hermes"
	"github.com/themedef/go-hermes/internal/contracts"
	"testing"
	"time"
)

func setupTestDB() contracts.Store {
	config := hermes.Config{
		CleanupInterval: time.Second * 10,
		EnableLogging:   false,
		LogFile:         "",
	}
	return hermes.NewStore(config)
}

func TestTransactionSetCommit(t *testing.T) {
	db := setupTestDB()
	tx := NewTransaction(db)
	ctx := context.Background()

	tx.Begin()
	tx.Set(ctx, "test_key", "value", 60)
	err := tx.Commit()

	if err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	value, exists, _ := db.Get(ctx, "test_key")
	if !exists || value != "value" {
		t.Fatalf("Expected 'value', got %v", value)
	}
}

func TestTransactionSetRollback(t *testing.T) {
	db := setupTestDB()
	tx := NewTransaction(db)
	ctx := context.Background()

	tx.Begin()
	err := tx.Set(ctx, "test_key", "value", 60)
	if err != nil {
		return
	}
	tx.Rollback()

	_, exists, _ := db.Get(ctx, "test_key")
	if exists {
		t.Fatalf("Expected key to be absent after rollback")
	}
}

func TestTransactionDeleteCommit(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()
	_ = db.Set(ctx, "test_key", "value", 60)

	tx := NewTransaction(db)
	tx.Begin()
	tx.Delete(ctx, "test_key")
	err := tx.Commit()

	if err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	_, exists, _ := db.Get(ctx, "test_key")
	if exists {
		t.Fatalf("Expected key to be deleted")
	}
}

func TestTransactionDeleteRollback(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()
	_ = db.Set(ctx, "test_key", "value", 60)

	tx := NewTransaction(db)
	tx.Begin()
	tx.Delete(ctx, "test_key")
	tx.Rollback()

	value, exists, _ := db.Get(ctx, "test_key")
	if !exists || value != "value" {
		t.Fatalf("Expected key to be restored after rollback")
	}
}

func TestTransactionIncrCommit(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()
	_ = db.Set(ctx, "counter", int64(1), 60)

	tx := NewTransaction(db)
	tx.Begin()
	tx.Incr(ctx, "counter")
	err := tx.Commit()

	if err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	value, _, _ := db.Get(ctx, "counter")
	if value.(int64) != 2 {
		t.Fatalf("Expected 2, got %v", value)
	}
}

func TestTransactionDecrCommit(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()
	_ = db.Set(ctx, "counter", int64(1), 60)

	tx := NewTransaction(db)
	tx.Begin()
	tx.Decr(ctx, "counter")
	err := tx.Commit()

	if err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	value, _, _ := db.Get(ctx, "counter")
	if value.(int64) != 0 {
		t.Fatalf("Expected 0, got %v", value)
	}
}

func TestTransactionIncrRollback(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()
	_ = db.Set(ctx, "counter", int64(1), 60)

	tx := NewTransaction(db)
	tx.Begin()
	tx.Incr(ctx, "counter")
	tx.Rollback()

	value, _, _ := db.Get(ctx, "counter")
	if value.(int64) != 1 {
		t.Fatalf("Expected 1 after rollback, got %v", value)
	}
}

func TestTransactionDecrRollback(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()
	_ = db.Set(ctx, "counter", int64(1), 60)

	tx := NewTransaction(db)
	tx.Begin()
	tx.Decr(ctx, "counter")
	tx.Rollback()

	value, _, _ := db.Get(ctx, "counter")
	if value.(int64) != 1 {
		t.Fatalf("Expected 1 after rollback, got %v", value)
	}
}
