package hermes

import (
	"context"
	"errors"
	"fmt"
	"github.com/themedef/go-hermes/internal/contracts"
	"sync"
)

type Transaction struct {
	mu       sync.Mutex
	db       contracts.StoreHandler
	commands []func() error
	rollback []func()
	active   bool
}

func NewTransaction(db contracts.StoreHandler) *Transaction {
	tx := &Transaction{db: db}
	_ = tx.begin()
	return tx
}

func (t *Transaction) begin() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.active {
		t.db.Logger().Warn("Transaction already active")
		return nil
	}
	t.active = true
	t.commands = nil
	t.rollback = nil
	t.db.Logger().Info("Transaction started")
	return nil
}

func (t *Transaction) Commit() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.active {
		return ErrTransactionNotActive
	}
	defer t.clear()

	for _, cmd := range t.commands {
		if err := cmd(); err != nil {
			t.rollbackCommands()
			return fmt.Errorf("%w: %v", ErrTransactionFailed, err)
		}
	}
	t.db.Logger().Info("Transaction committed")
	return nil
}

func (t *Transaction) Rollback() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.active {
		return nil
	}
	t.rollbackCommands()
	t.clear()
	t.db.Logger().Info("Transaction rolled back")
	return nil
}

func (t *Transaction) rollbackCommands() {
	for i := len(t.rollback) - 1; i >= 0; i-- {
		t.rollback[i]()
	}
}

func (t *Transaction) clear() {
	t.commands = nil
	t.rollback = nil
	t.active = false
}

func (t *Transaction) Set(ctx context.Context, key string, value interface{}, ttl int) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		return ErrTransactionNotActive
	}

	oldVal, err := t.db.Get(ctx, key)
	existed := true
	if err != nil {
		if errors.Is(err, ErrKeyNotFound) || errors.Is(err, ErrKeyExpired) {
			existed = false
		} else {
			return err
		}
	}

	t.commands = append(t.commands, func() error {
		return t.db.Set(ctx, key, value, ttl)
	})
	t.rollback = append(t.rollback, func() {
		if existed {
			_ = t.db.Set(context.Background(), key, oldVal, 0)
		} else {
			_ = t.db.Delete(context.Background(), key)
		}
	})

	return nil
}

func (t *Transaction) SetNX(ctx context.Context, key string, value interface{}, ttl int) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		return ErrTransactionNotActive
	}

	oldVal, err := t.db.Get(ctx, key)
	existed := err == nil

	t.commands = append(t.commands, func() error {
		_, setErr := t.db.SetNX(ctx, key, value, ttl)
		if setErr != nil {
			return setErr
		}
		return nil
	})
	t.rollback = append(t.rollback, func() {
		if existed {
			_ = t.db.Set(context.Background(), key, oldVal, 0)
		} else {
			_ = t.db.Delete(context.Background(), key)
		}
	})
	return nil
}

func (t *Transaction) SetXX(ctx context.Context, key string, value interface{}, ttl int) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		return ErrTransactionNotActive
	}

	oldVal, err := t.db.Get(ctx, key)
	existed := true
	if err != nil {
		if errors.Is(err, ErrKeyNotFound) || errors.Is(err, ErrKeyExpired) {
			existed = false
		} else {
			return err
		}
	}

	t.commands = append(t.commands, func() error {
		ok, xxErr := t.db.SetXX(ctx, key, value, ttl)
		if xxErr != nil {
			return xxErr
		}
		if !ok {
			return ErrKeyNotFound
		}
		return nil
	})
	t.rollback = append(t.rollback, func() {
		if existed {
			_ = t.db.Set(context.Background(), key, oldVal, 0)
		} else {
			_ = t.db.Delete(context.Background(), key)
		}
	})
	return nil
}

func (t *Transaction) Get(ctx context.Context, key string) (interface{}, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		return nil, ErrTransactionNotActive
	}
	return t.db.Get(ctx, key)
}

func (t *Transaction) SetCAS(ctx context.Context, key string, oldValue, newValue interface{}, ttl int) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		return ErrTransactionNotActive
	}

	dbOldVal, err := t.db.Get(ctx, key)
	existed := true
	if err != nil {
		if errors.Is(err, ErrKeyNotFound) || errors.Is(err, ErrKeyExpired) {
			existed = false
		} else {
			return err
		}
	}

	t.commands = append(t.commands, func() error {
		return t.db.SetCAS(ctx, key, oldValue, newValue, ttl)
	})
	t.rollback = append(t.rollback, func() {
		if existed {
			_ = t.db.Set(context.Background(), key, dbOldVal, 0)
		} else {
			_ = t.db.Delete(context.Background(), key)
		}
	})

	return nil
}

func (t *Transaction) GetSet(ctx context.Context, key string, newValue interface{}, ttl int) (interface{}, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		return nil, ErrTransactionNotActive
	}

	oldVal, err := t.db.Get(ctx, key)
	existed := true
	if err != nil {
		if errors.Is(err, ErrKeyNotFound) || errors.Is(err, ErrKeyExpired) {
			existed = false
		} else {
			return nil, err
		}
	}

	t.commands = append(t.commands, func() error {
		_, err := t.db.GetSet(ctx, key, newValue, ttl)
		return err
	})
	t.rollback = append(t.rollback, func() {
		if existed {
			_ = t.db.Set(context.Background(), key, oldVal, 0)
		} else {
			_ = t.db.Delete(context.Background(), key)
		}
	})

	return oldVal, nil
}

func (t *Transaction) Incr(ctx context.Context, key string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		return ErrTransactionNotActive
	}

	oldVal, err := t.db.Get(ctx, key)
	existed := true
	if err != nil {
		if errors.Is(err, ErrKeyNotFound) || errors.Is(err, ErrKeyExpired) {
			existed = false
		} else {
			return err
		}
	}

	t.commands = append(t.commands, func() error {
		_, incrErr := t.db.Incr(ctx, key)
		return incrErr
	})
	t.rollback = append(t.rollback, func() {
		if existed {
			_ = t.db.Set(context.Background(), key, oldVal, 0)
		} else {
			_ = t.db.Delete(context.Background(), key)
		}
	})

	return nil
}

func (t *Transaction) Decr(ctx context.Context, key string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		return ErrTransactionNotActive
	}

	oldVal, err := t.db.Get(ctx, key)
	existed := true
	if err != nil {
		if errors.Is(err, ErrKeyNotFound) || errors.Is(err, ErrKeyExpired) {
			existed = false
		} else {
			return err
		}
	}

	t.commands = append(t.commands, func() error {
		_, decrErr := t.db.Decr(ctx, key)
		return decrErr
	})
	t.rollback = append(t.rollback, func() {
		if existed {
			_ = t.db.Set(context.Background(), key, oldVal, 0)
		} else {
			_ = t.db.Delete(context.Background(), key)
		}
	})

	return nil
}

func (t *Transaction) LPush(ctx context.Context, key string, values ...interface{}) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		return ErrTransactionNotActive
	}

	oldVal, err := t.db.Get(ctx, key)
	existed := true
	if err != nil {
		if errors.Is(err, ErrKeyNotFound) || errors.Is(err, ErrKeyExpired) {
			existed = false
		} else {
			return err
		}
	}

	t.commands = append(t.commands, func() error {
		return t.db.LPush(ctx, key, values...)
	})

	t.rollback = append(t.rollback, func() {
		if existed {
			if oldList, ok := oldVal.([]interface{}); ok {
				_ = t.db.LPush(context.Background(), key, oldList...)
			} else {
				_ = t.db.Set(context.Background(), key, oldVal, 0)
			}
		} else {
			_ = t.db.Delete(context.Background(), key)
		}
	})

	return nil
}

func (t *Transaction) RPush(ctx context.Context, key string, values ...interface{}) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		return ErrTransactionNotActive
	}

	oldVal, err := t.db.Get(ctx, key)
	existed := true
	if err != nil {
		if errors.Is(err, ErrKeyNotFound) || errors.Is(err, ErrKeyExpired) {
			existed = false
		} else {
			return err
		}
	}

	t.commands = append(t.commands, func() error {
		return t.db.RPush(ctx, key, values...)
	})

	t.rollback = append(t.rollback, func() {
		if existed {
			if oldList, ok := oldVal.([]interface{}); ok {
				_ = t.db.RPush(context.Background(), key, oldList...)
			} else {
				_ = t.db.Set(context.Background(), key, oldVal, 0)
			}
		} else {
			_ = t.db.Delete(context.Background(), key)
		}
	})

	return nil
}

func (t *Transaction) LPop(ctx context.Context, key string) (interface{}, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		return nil, ErrTransactionNotActive
	}

	oldVal, err := t.db.Get(ctx, key)
	existed := true
	if err != nil {
		if errors.Is(err, ErrKeyNotFound) || errors.Is(err, ErrKeyExpired) {
			existed = false
		} else {
			return nil, err
		}
	}

	var originalList []interface{}
	if existed {
		list, ok := oldVal.([]interface{})
		if !ok {
			return nil, ErrInvalidType
		}
		originalList = make([]interface{}, len(list))
		copy(originalList, list)
	}

	t.commands = append(t.commands, func() error {
		_, err := t.db.LPop(ctx, key)
		return err
	})

	t.rollback = append(t.rollback, func() {
		if existed {
			_ = t.db.Delete(context.Background(), key)
			if err := t.db.RPush(context.Background(), key, originalList...); err != nil {
				t.db.Logger().Error("Rollback RPush failed", "key", key, "error", err)
			}
		} else {
			_ = t.db.Delete(context.Background(), key)
		}
	})

	if existed && len(originalList) == 0 {
		return nil, ErrEmptyList
	}

	if existed {
		return originalList[0], nil
	}
	return nil, ErrKeyNotFound
}

func (t *Transaction) RPop(ctx context.Context, key string) (interface{}, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		return nil, ErrTransactionNotActive
	}

	oldVal, err := t.db.Get(ctx, key)
	existed := true
	if err != nil {
		if errors.Is(err, ErrKeyNotFound) || errors.Is(err, ErrKeyExpired) {
			existed = false
		} else {
			return nil, err
		}
	}

	var originalList []interface{}
	if existed {
		list, ok := oldVal.([]interface{})
		if !ok {
			return nil, ErrInvalidType
		}
		originalList = make([]interface{}, len(list))
		copy(originalList, list)
	}

	t.commands = append(t.commands, func() error {
		_, err := t.db.RPop(ctx, key)
		return err
	})

	t.rollback = append(t.rollback, func() {
		if existed {
			_ = t.db.Delete(context.Background(), key)
			if err := t.db.RPush(context.Background(), key, originalList...); err != nil {
				t.db.Logger().Error("Rollback RPush failed", "key", key, "error", err)
			}
		} else {
			_ = t.db.Delete(context.Background(), key)
		}
	})

	if existed && len(originalList) == 0 {
		return nil, ErrEmptyList
	}

	if existed {
		return originalList[len(originalList)-1], nil
	}
	return nil, ErrKeyNotFound
}

func (t *Transaction) LLen(ctx context.Context, key string) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		return 0, ErrTransactionNotActive
	}
	return t.db.LLen(ctx, key)
}

func (t *Transaction) LRange(ctx context.Context, key string, start, end int) ([]interface{}, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		return nil, ErrTransactionNotActive
	}
	return t.db.LRange(ctx, key, start, end)
}

func (t *Transaction) HSet(ctx context.Context, key, field string, value interface{}, ttl int) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		return ErrTransactionNotActive
	}

	oldValue, err := t.db.HGet(ctx, key, field)
	existed := true
	if err != nil {
		if errors.Is(err, ErrKeyNotFound) {
			existed = false
		} else {
			return err
		}
	}

	t.commands = append(t.commands, func() error {
		return t.db.HSet(ctx, key, field, value, ttl)
	})
	t.rollback = append(t.rollback, func() {
		if existed {
			_ = t.db.HSet(context.Background(), key, field, oldValue, 0)
		} else {
			_ = t.db.HDel(context.Background(), key, field)
		}
	})
	return nil
}

func (t *Transaction) HGet(ctx context.Context, key, field string) (interface{}, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		return nil, ErrTransactionNotActive
	}
	return t.db.HGet(ctx, key, field)
}

func (t *Transaction) HDel(ctx context.Context, key, field string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		return ErrTransactionNotActive
	}

	oldVal, err := t.db.HGet(ctx, key, field)
	existed := true
	if err != nil {
		if errors.Is(err, ErrKeyNotFound) {
			existed = false
		} else {
			return err
		}
	}

	t.commands = append(t.commands, func() error {
		return t.db.HDel(ctx, key, field)
	})
	t.rollback = append(t.rollback, func() {
		if existed {
			_ = t.db.HSet(context.Background(), key, field, oldVal, 0)
		}
	})
	return nil
}
func (t *Transaction) HGetAll(ctx context.Context, key string) (map[string]interface{}, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		return nil, ErrTransactionNotActive
	}
	return t.db.HGetAll(ctx, key)
}

func (t *Transaction) HExists(ctx context.Context, key, field string) (bool, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		return false, ErrTransactionNotActive
	}
	return t.db.HExists(ctx, key, field)
}

func (t *Transaction) HLen(ctx context.Context, key string) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		return 0, ErrTransactionNotActive
	}
	return t.db.HLen(ctx, key)
}

func (t *Transaction) Exists(ctx context.Context, key string) (bool, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		return false, ErrTransactionNotActive
	}
	return t.db.Exists(ctx, key)
}

func (t *Transaction) UpdateTTL(ctx context.Context, key string, ttl int) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		return ErrTransactionNotActive
	}

	oldVal, oldTtl, err := t.db.GetWithDetails(ctx, key)
	existed := true
	if err != nil {
		if errors.Is(err, ErrKeyNotFound) || errors.Is(err, ErrKeyExpired) {
			existed = false
		} else {
			return err
		}
	}

	t.commands = append(t.commands, func() error {
		return t.db.UpdateTTL(ctx, key, ttl)
	})
	t.rollback = append(t.rollback, func() {
		if existed {
			_ = t.db.Set(context.Background(), key, oldVal, oldTtl)
		} else {
			_ = t.db.Delete(context.Background(), key)
		}
	})
	return nil
}

func (t *Transaction) Type(ctx context.Context, key string) (interface{}, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		return -1, ErrTransactionNotActive
	}
	return t.db.Type(ctx, key)
}

func (t *Transaction) GetWithDetails(ctx context.Context, key string) (interface{}, int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		return nil, 0, ErrTransactionNotActive
	}
	return t.db.GetWithDetails(ctx, key)
}

func (t *Transaction) Rename(ctx context.Context, oldKey, newKey string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		return ErrTransactionNotActive
	}

	oldVal, oldTtl, err := t.db.GetWithDetails(ctx, oldKey)
	if err != nil {
		return err
	}

	newExists, err := t.db.Exists(ctx, newKey)
	if err != nil {
		return err
	}

	var newVal interface{}
	var newTtl int
	if newExists {
		newVal, newTtl, err = t.db.GetWithDetails(ctx, newKey)
		if err != nil {
			return err
		}
	}

	t.commands = append(t.commands, func() error {
		return t.db.Rename(ctx, oldKey, newKey)
	})
	t.rollback = append(t.rollback, func() {
		_ = t.db.Set(context.Background(), oldKey, oldVal, oldTtl)
		if newExists {
			_ = t.db.Set(context.Background(), newKey, newVal, newTtl)
		} else {
			_ = t.db.Delete(context.Background(), newKey)
		}
	})

	return nil
}

func (t *Transaction) FindByValue(ctx context.Context, value interface{}) ([]string, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		return nil, ErrTransactionNotActive
	}
	return t.db.FindByValue(ctx, value)
}

func (t *Transaction) Delete(ctx context.Context, key string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		return ErrTransactionNotActive
	}

	oldVal, err := t.db.Get(ctx, key)
	existed := true
	if err != nil {
		if errors.Is(err, ErrKeyNotFound) || errors.Is(err, ErrKeyExpired) {
			existed = false
		} else {
			return err
		}
	}

	t.commands = append(t.commands, func() error {
		return t.db.Delete(ctx, key)
	})
	if existed {
		t.rollback = append(t.rollback, func() {
			_ = t.db.Set(context.Background(), key, oldVal, 0)
		})
	} else {
		t.rollback = append(t.rollback, func() {})
	}
	return nil
}
