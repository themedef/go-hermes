package transaction

import (
	"context"
	"fmt"
	"github.com/themedef/go-hermes/internal/contracts"
	"sync"
)

type Transaction struct {
	mu       sync.Mutex
	db       contracts.Store
	commands []func() error
	rollback []func()
	active   bool
}

func NewTransaction(db contracts.Store) *Transaction {
	return &Transaction{
		db: db,
	}
}

func (t *Transaction) Begin() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.active {
		t.db.Logger().Warn("Transaction already active")
		return
	}

	t.active = true
	t.commands = nil
	t.rollback = nil
	t.db.Logger().Info("Transaction started")
}

func (t *Transaction) Set(ctx context.Context, key string, value interface{}, ttl int) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		t.db.Logger().Warn("Cannot set, no active transaction")
		return fmt.Errorf("transaction is not active")
	}

	oldValue, exists, _ := t.db.Get(ctx, key)

	t.commands = append(t.commands, func() error {
		return t.db.Set(ctx, key, value, ttl)
	})

	t.rollback = append(t.rollback, func() {
		if exists {
			err := t.db.Set(ctx, key, oldValue, 0)
			if err != nil {
				return
			}
		} else {
			_, err := t.db.Delete(ctx, key)
			if err != nil {
				return
			}
		}
	})
	return nil
}

func (t *Transaction) SetNX(ctx context.Context, key string, value interface{}, ttl int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		t.db.Logger().Warn("Cannot setNX, no active transaction")
		return
	}

	t.commands = append(t.commands, func() error {
		_, err := t.db.SetNX(ctx, key, value, ttl)
		return err
	})
}

func (t *Transaction) SetXX(ctx context.Context, key string, value interface{}, ttl int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		t.db.Logger().Warn("Cannot setXX, no active transaction")
		return
	}

	t.commands = append(t.commands, func() error {
		_, err := t.db.SetXX(ctx, key, value, ttl)
		return err
	})
}

func (t *Transaction) SetCAS(ctx context.Context, key string, oldValue, newValue interface{}, ttl int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		t.db.Logger().Warn("Cannot setCAS, no active transaction")
		return
	}

	t.commands = append(t.commands, func() error {
		_, err := t.db.SetCAS(ctx, key, oldValue, newValue, ttl)
		return err
	})
}

func (t *Transaction) Delete(ctx context.Context, key string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		t.db.Logger().Warn("Cannot delete, no active transaction")
		return
	}

	oldValue, exists, _ := t.db.Get(ctx, key)

	t.commands = append(t.commands, func() error {
		_, err := t.db.Delete(ctx, key)
		return err
	})

	t.rollback = append(t.rollback, func() {
		if exists {
			t.db.Set(ctx, key, oldValue, 0)
		}
	})
}

func (t *Transaction) Incr(ctx context.Context, key string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		t.db.Logger().Warn("Cannot incr, no active transaction")
		return
	}

	t.commands = append(t.commands, func() error {
		_, _, err := t.db.Incr(ctx, key)
		return err
	})
}

func (t *Transaction) Decr(ctx context.Context, key string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		t.db.Logger().Warn("Cannot decr, no active transaction")
		return
	}

	t.commands = append(t.commands, func() error {
		_, _, err := t.db.Decr(ctx, key)
		return err
	})
}

func (t *Transaction) LPush(ctx context.Context, key string, value interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		t.db.Logger().Warn("Cannot lpush, no active transaction")
		return
	}

	t.commands = append(t.commands, func() error {
		return t.db.LPush(ctx, key, value)
	})
}

func (t *Transaction) RPush(ctx context.Context, key string, value interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		t.db.Logger().Warn("Cannot rpush, no active transaction")
		return
	}

	t.commands = append(t.commands, func() error {
		return t.db.RPush(ctx, key, value)
	})
}

func (t *Transaction) LPop(ctx context.Context, key string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		t.db.Logger().Warn("Cannot lpop, no active transaction")
		return
	}

	t.commands = append(t.commands, func() error {
		_, _, err := t.db.LPop(ctx, key)
		return err
	})
}

func (t *Transaction) RPop(ctx context.Context, key string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		t.db.Logger().Warn("Cannot rpop, no active transaction")
		return
	}

	t.commands = append(t.commands, func() error {
		_, _, err := t.db.RPop(ctx, key)
		return err
	})
}

func (t *Transaction) UpdateTTL(ctx context.Context, key string, ttl int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		t.db.Logger().Warn("Cannot update TTL, no active transaction")
		return
	}

	t.commands = append(t.commands, func() error {
		return t.db.UpdateTTL(ctx, key, ttl)
	})
}

func (t *Transaction) Commit() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		return fmt.Errorf("no active transaction")
	}

	for _, cmd := range t.commands {
		if err := cmd(); err != nil {
			t.Rollback()
			return fmt.Errorf("transaction failed: %w", err)
		}
	}

	t.db.Logger().Info("Transaction committed")
	t.Clear()
	return nil
}

func (t *Transaction) Rollback() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		return
	}

	for _, rollbackCmd := range t.rollback {
		rollbackCmd()
	}

	t.db.Logger().Info("Transaction rolled back")
	t.Clear()
}

func (t *Transaction) Clear() {
	t.commands = nil
	t.rollback = nil
	t.active = false
	t.db.Logger().Info("Transaction cleared")
}
