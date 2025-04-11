package hermes

import (
	"context"
	"fmt"
	"github.com/themedef/go-hermes/internal/types"
	"sync"

	"github.com/themedef/go-hermes/internal/contracts"
)

type Transaction struct {
	mu       sync.Mutex
	db       contracts.StoreHandler
	commands []func() error
	rollback []func()
	active   bool
}

func NewTransaction(db contracts.StoreHandler) contracts.TransactionHandler {
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

func (t *Transaction) getRawEntryOrNil(ctx context.Context, key string) (entry types.Entry, existed bool, err error) {
	entry, err = t.db.GetRawEntry(ctx, key)
	if err != nil {
		if IsKeyNotFound(err) || IsKeyExpired(err) {
			return types.Entry{}, false, nil
		}
		return types.Entry{}, false, err
	}
	return entry, true, nil
}

func (t *Transaction) Set(ctx context.Context, key string, value interface{}, ttl int) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.active {
		return ErrTransactionNotActive
	}
	oldEntry, existed, err := t.getRawEntryOrNil(ctx, key)
	if err != nil {
		return err
	}
	t.commands = append(t.commands, func() error {
		return t.db.Set(ctx, key, value, ttl)
	})
	t.rollback = append(t.rollback, func() {
		if existed {
			_ = t.db.RestoreRawEntry(context.Background(), key, oldEntry)
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
	oldEntry, existed, err := t.getRawEntryOrNil(ctx, key)
	if err != nil {
		return err
	}
	t.commands = append(t.commands, func() error {
		_, setErr := t.db.SetNX(ctx, key, value, ttl)
		return setErr
	})
	t.rollback = append(t.rollback, func() {
		if existed {
			_ = t.db.RestoreRawEntry(context.Background(), key, oldEntry)
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
	oldEntry, existed, err := t.getRawEntryOrNil(ctx, key)
	if err != nil {
		return err
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
			_ = t.db.RestoreRawEntry(context.Background(), key, oldEntry)
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
	oldEntry, existed, err := t.getRawEntryOrNil(ctx, key)
	if err != nil {
		return err
	}
	t.commands = append(t.commands, func() error {
		return t.db.SetCAS(ctx, key, oldValue, newValue, ttl)
	})
	t.rollback = append(t.rollback, func() {
		if existed {
			_ = t.db.RestoreRawEntry(context.Background(), key, oldEntry)
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
	oldEntry, existed, err := t.getRawEntryOrNil(ctx, key)
	if err != nil {
		return nil, err
	}
	t.commands = append(t.commands, func() error {
		_, err := t.db.GetSet(ctx, key, newValue, ttl)
		return err
	})
	t.rollback = append(t.rollback, func() {
		if existed {
			_ = t.db.RestoreRawEntry(context.Background(), key, oldEntry)
		} else {
			_ = t.db.Delete(context.Background(), key)
		}
	})
	if !existed {
		return nil, nil
	}
	return oldEntry.Value, nil
}

func (t *Transaction) Incr(ctx context.Context, key string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.active {
		return ErrTransactionNotActive
	}
	oldEntry, existed, err := t.getRawEntryOrNil(ctx, key)
	if err != nil {
		return err
	}
	t.commands = append(t.commands, func() error {
		_, incrErr := t.db.Incr(ctx, key)
		return incrErr
	})
	t.rollback = append(t.rollback, func() {
		if existed {
			_ = t.db.RestoreRawEntry(context.Background(), key, oldEntry)
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
	oldEntry, existed, err := t.getRawEntryOrNil(ctx, key)
	if err != nil {
		return err
	}
	t.commands = append(t.commands, func() error {
		_, decrErr := t.db.Decr(ctx, key)
		return decrErr
	})
	t.rollback = append(t.rollback, func() {
		if existed {
			_ = t.db.RestoreRawEntry(context.Background(), key, oldEntry)
		} else {
			_ = t.db.Delete(context.Background(), key)
		}
	})
	return nil
}

func (t *Transaction) IncrBy(ctx context.Context, key string, increment int64) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.active {
		return ErrTransactionNotActive
	}
	oldEntry, existed, err := t.getRawEntryOrNil(ctx, key)
	if err != nil {
		return err
	}
	t.commands = append(t.commands, func() error {
		_, err := t.db.IncrBy(ctx, key, increment)
		return err
	})
	t.rollback = append(t.rollback, func() {
		if existed {
			_ = t.db.RestoreRawEntry(context.Background(), key, oldEntry)
		} else {
			_ = t.db.Delete(context.Background(), key)
		}
	})
	return nil
}

func (t *Transaction) DecrBy(ctx context.Context, key string, decrement int64) error {
	return t.IncrBy(ctx, key, -decrement)
}

func (t *Transaction) LPush(ctx context.Context, key string, values ...interface{}) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.active {
		return ErrTransactionNotActive
	}
	oldEntry, existed, err := t.getRawEntryOrNil(ctx, key)
	if err != nil {
		return err
	}
	t.commands = append(t.commands, func() error {
		return t.db.LPush(ctx, key, values...)
	})
	t.rollback = append(t.rollback, func() {
		if existed {
			_ = t.db.RestoreRawEntry(context.Background(), key, oldEntry)
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
	oldEntry, existed, err := t.getRawEntryOrNil(ctx, key)
	if err != nil {
		return err
	}
	t.commands = append(t.commands, func() error {
		return t.db.RPush(ctx, key, values...)
	})
	t.rollback = append(t.rollback, func() {
		if existed {
			_ = t.db.RestoreRawEntry(context.Background(), key, oldEntry)
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
	oldEntry, existed, err := t.getRawEntryOrNil(ctx, key)
	if err != nil {
		return nil, err
	}
	t.commands = append(t.commands, func() error {
		_, err := t.db.LPop(ctx, key)
		return err
	})
	t.rollback = append(t.rollback, func() {
		if existed {
			_ = t.db.RestoreRawEntry(context.Background(), key, oldEntry)
		} else {
			_ = t.db.Delete(context.Background(), key)
		}
	})
	if !existed {
		return nil, ErrKeyNotFound
	}
	list, ok := oldEntry.Value.([]interface{})
	if !ok || len(list) == 0 {
		return nil, ErrEmptyList
	}
	return list[0], nil
}

func (t *Transaction) RPop(ctx context.Context, key string) (interface{}, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.active {
		return nil, ErrTransactionNotActive
	}
	oldEntry, existed, err := t.getRawEntryOrNil(ctx, key)
	if err != nil {
		return nil, err
	}
	t.commands = append(t.commands, func() error {
		_, err := t.db.RPop(ctx, key)
		return err
	})
	t.rollback = append(t.rollback, func() {
		if existed {
			_ = t.db.RestoreRawEntry(context.Background(), key, oldEntry)
		} else {
			_ = t.db.Delete(context.Background(), key)
		}
	})
	if !existed {
		return nil, ErrKeyNotFound
	}
	list, ok := oldEntry.Value.([]interface{})
	if !ok || len(list) == 0 {
		return nil, ErrEmptyList
	}
	return list[len(list)-1], nil
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

func (t *Transaction) LTrim(ctx context.Context, key string, start, stop int) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.active {
		return ErrTransactionNotActive
	}
	oldEntry, existed, err := t.getRawEntryOrNil(ctx, key)
	if err != nil {
		return err
	}
	t.commands = append(t.commands, func() error {
		return t.db.LTrim(ctx, key, start, stop)
	})
	t.rollback = append(t.rollback, func() {
		if existed {
			_ = t.db.RestoreRawEntry(context.Background(), key, oldEntry)
		} else {
			_ = t.db.Delete(context.Background(), key)
		}
	})
	return nil
}

func (t *Transaction) HSet(ctx context.Context, key, field string, value interface{}, ttl int) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.active {
		return ErrTransactionNotActive
	}
	oldEntry, existed, err := t.getRawEntryOrNil(ctx, key)
	if err != nil {
		return err
	}
	t.commands = append(t.commands, func() error {
		return t.db.HSet(ctx, key, field, value, ttl)
	})
	t.rollback = append(t.rollback, func() {
		if existed {
			_ = t.db.RestoreRawEntry(context.Background(), key, oldEntry)
		} else {
			_ = t.db.Delete(context.Background(), key)
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
	oldEntry, existed, err := t.getRawEntryOrNil(ctx, key)
	if err != nil {
		return err
	}
	t.commands = append(t.commands, func() error {
		return t.db.HDel(ctx, key, field)
	})
	t.rollback = append(t.rollback, func() {
		if existed {
			_ = t.db.RestoreRawEntry(context.Background(), key, oldEntry)
		} else {
			_ = t.db.Delete(context.Background(), key)
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

func (t *Transaction) SAdd(ctx context.Context, key string, members ...interface{}) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.active {
		return ErrTransactionNotActive
	}
	oldEntry, existed, err := t.getRawEntryOrNil(ctx, key)
	if err != nil {
		return err
	}
	t.commands = append(t.commands, func() error {
		return t.db.SAdd(ctx, key, members...)
	})
	t.rollback = append(t.rollback, func() {
		if existed {
			_ = t.db.RestoreRawEntry(context.Background(), key, oldEntry)
		} else {
			_ = t.db.Delete(context.Background(), key)
		}
	})
	return nil
}

func (t *Transaction) SRem(ctx context.Context, key string, members ...interface{}) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.active {
		return ErrTransactionNotActive
	}
	oldEntry, existed, err := t.getRawEntryOrNil(ctx, key)
	if err != nil {
		return err
	}
	t.commands = append(t.commands, func() error {
		return t.db.SRem(ctx, key, members...)
	})
	t.rollback = append(t.rollback, func() {
		if existed {
			_ = t.db.RestoreRawEntry(context.Background(), key, oldEntry)
		} else {
			_ = t.db.Delete(context.Background(), key)
		}
	})
	return nil
}

func (t *Transaction) SMembers(ctx context.Context, key string) ([]interface{}, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.active {
		return nil, ErrTransactionNotActive
	}
	return t.db.SMembers(ctx, key)
}

func (t *Transaction) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.active {
		return false, ErrTransactionNotActive
	}
	return t.db.SIsMember(ctx, key, member)
}

func (t *Transaction) SCard(ctx context.Context, key string) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.active {
		return 0, ErrTransactionNotActive
	}
	return t.db.SCard(ctx, key)
}

func (t *Transaction) Exists(ctx context.Context, key string) (bool, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.active {
		return false, ErrTransactionNotActive
	}
	return t.db.Exists(ctx, key)
}

func (t *Transaction) Expire(ctx context.Context, key string, ttl int) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.active {
		return ErrTransactionNotActive
	}
	oldEntry, existed, err := t.getRawEntryOrNil(ctx, key)
	if err != nil {
		return err
	}
	t.commands = append(t.commands, func() error {
		ok, err := t.db.Expire(ctx, key, ttl)
		if err != nil {
			return err
		}
		if !ok {
			return ErrKeyNotFound
		}
		return nil
	})
	t.rollback = append(t.rollback, func() {
		if existed {
			_ = t.db.RestoreRawEntry(context.Background(), key, oldEntry)
		} else {
			_ = t.db.Delete(context.Background(), key)
		}
	})
	return nil
}

func (t *Transaction) Persist(ctx context.Context, key string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.active {
		return ErrTransactionNotActive
	}
	oldEntry, existed, err := t.getRawEntryOrNil(ctx, key)
	if err != nil {
		return err
	}
	t.commands = append(t.commands, func() error {
		ok, err := t.db.Persist(ctx, key)
		if err != nil {
			return err
		}
		if !ok {
			return ErrKeyNotFound
		}
		return nil
	})
	t.rollback = append(t.rollback, func() {
		if existed {
			_ = t.db.RestoreRawEntry(context.Background(), key, oldEntry)
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
	oldKeyEntry, oldExisted, err := t.getRawEntryOrNil(ctx, oldKey)
	if err != nil {
		return err
	}
	newKeyEntry, newExisted, err := t.getRawEntryOrNil(ctx, newKey)
	if err != nil {
		return err
	}
	t.commands = append(t.commands, func() error {
		return t.db.Rename(ctx, oldKey, newKey)
	})
	t.rollback = append(t.rollback, func() {
		if oldExisted {
			_ = t.db.RestoreRawEntry(context.Background(), oldKey, oldKeyEntry)
		} else {
			_ = t.db.Delete(context.Background(), oldKey)
		}
		if newExisted {
			_ = t.db.RestoreRawEntry(context.Background(), newKey, newKeyEntry)
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
	oldEntry, existed, err := t.getRawEntryOrNil(ctx, key)
	if err != nil {
		return err
	}
	t.commands = append(t.commands, func() error {
		return t.db.Delete(ctx, key)
	})
	t.rollback = append(t.rollback, func() {
		if existed {
			_ = t.db.RestoreRawEntry(context.Background(), key, oldEntry)
		} else {
			_ = t.db.Delete(context.Background(), key)
		}
	})
	return nil
}
