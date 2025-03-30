# Transaction Documentation

---

## Table of Contents

1. [Overview](#overview)
2. [Transaction Lifecycle](#transaction-lifecycle)
   - [Begin Transaction](#begin-transaction)
   - [Commit](#commit)
   - [Rollback](#rollback)
3. [Operations](#operations)
    - [Key-Value Operations](#key-value-operations)
        - [Set](#set)
        - [Get](#get)
        - [SetNX](#setnx)
        - [SetXX](#setxx)
        - [SetCAS](#setcas)
        - [GetSet](#getset)
    - [Atomic Counters](#atomic-counters)
        - [Incr](#increment)
        - [Decr](#decrement)
    - [List Operations](#list-operations)
        - [LPush](#lpush)
        - [RPush](#rpush)
        - [LPop](#lpop)
        - [RPop](#rpop)
        - [LRange](#lrange)
        - [LLen](#llen)
    - [Hash Operations](#hash-operations)
        - [HSet](#hset)
        - [HGet](#hget)
        - [HDel](#hdel)
        - [HGetAll](#hgetall)
        - [HExists](#hexists)
        - [HLen](#hlen)
    - [Utility Methods](#utility-methods)
        - [Exists](#exists)
        - [UpdateTTL](#updatettl)
        - [Type](#type)
        - [GetWithDetails](#getwithdetails)
        - [Rename](#rename)
        - [FindByValue](#findbyvalue)
        - [Delete](#delete)
4. [Error Reference](#error-reference)
5. [Best Practices](#best-practices)

---

## 1. Overview <a id="overview"></a>

A **transaction** groups multiple operations (reads/writes) to execute atomically. Changes are buffered locally until committed. If any operation fails during commit, all changes are rolled back automatically.

Key Features:
- **Optimistic Concurrency**: Changes are visible only after commit
- **Atomic Guarantee**: All operations succeed or fail together
- **Rollback Support**: Full state restoration on failure
- **Nested Operations**: Supports complex data structure operations

---

## 2. Transaction Lifecycle <a id="transaction-lifecycle"></a>

### Begin Transaction <a id="begin-transaction"></a>
```go
tx := db.Transaction()
```
Initializes a new transaction context. Must be called before any operations.

---

### Commit <a id="commit"></a>
```go
err := tx.Commit()
```
- Executes all queued operations atomically
- Returns `ErrTransactionFailed` if any operation fails
- Automatically rolls back on failure

---

### Rollback <a id="rollback"></a>
```go
err := tx.Rollback()
```
- Discards all pending changes
- No-op if no active transaction

---

## 3. Operations <a id="operations"></a>

### Key-Value Operations <a id="key-value-operations"></a>

#### Set <a id="set"></a>
```go
err := tx.Set(ctx, "key", value, ttl)
```
- Creates/updates key with TTL
- **Rollback**: Restores previous value or deletes new key

#### Get <a id="get"></a>
```go
val, err := tx.Get(ctx, "key")
```
- Returns current value (including transaction changes)
- **Errors**: `ErrKeyNotFound`, `ErrKeyExpired`

#### SetNX <a id="setnx"></a>
```go
err := tx.SetNX(ctx, "key", value, ttl)
```
- Sets value only if key doesn't exist
- **Rollback**: Deletes key if created

#### SetXX <a id="setxx"></a>
```go
err := tx.SetXX(ctx, "key", value, ttl)
```
- Updates value only if key exists
- **Rollback**: Restores previous value

#### SetCAS <a id="setcas"></a>
```go
err := tx.SetCAS(ctx, "key", oldVal, newVal, ttl)
```
- Atomic compare-and-swap
- **Errors**: `ErrValueMismatch`, `ErrKeyNotFound`

#### GetSet <a id="getset"></a>
```go
oldVal, err := tx.GetSet(ctx, "key", newVal, ttl)
```
- Returns previous value while setting new
- **Rollback**: Restores original value

---

### Atomic Counters <a id="atomic-counters"></a>

#### Incr <a id="increment"></a>
```go
err := tx.Incr(ctx, "counter")
```
- Increments integer value (initializes to 1 if missing)
- **Rollback**: Restores previous value

#### Decr <a id="decrement"></a>
```go
err := tx.Decr(ctx, "counter")
```
- Decrements integer value (initializes to -1 if missing)
- **Rollback**: Restores previous value

---

### List Operations <a id="list-operations"></a>

#### LPush <a id="lpush"></a>
```go
err := tx.LPush(ctx, "list", value)
```
- Inserts value at list head
- **Rollback**: Restores original list state

#### RPush <a id="rpush"></a>
```go
err := tx.RPush(ctx, "list", value)
```
- Appends value at list tail
- **Rollback**: Restores original list state

#### LPop <a id="lpop"></a>
```go
val, err := tx.LPop(ctx, "list")
```
- Removes/returns head element
- **Rollback**: Reinserts popped element

#### RPop <a id="rpop"></a>
```go
val, err := tx.RPop(ctx, "list")
```
- Removes/returns tail element
- **Rollback**: Reinserts popped element

#### LRange <a id="lrange"></a>
```go
items, err := tx.LRange(ctx, "list", start, end)
```
- Returns slice of elements
- Supports negative indices

#### LLen <a id="llen"></a>
```go
length, err := tx.LLen(ctx, "list")
```
- Returns list length
- **Errors**: `ErrInvalidType`

---

### Hash Operations <a id="hash-operations"></a>

#### HSet <a id="hset"></a>
```go
err := tx.HSet(ctx, "hash", "field", value, ttl)
```
- Sets hash field value
- **Rollback**: Restores previous field state

#### HGet <a id="hget"></a>
```go
val, err := tx.HGet(ctx, "hash", "field")
```
- Returns field value
- **Errors**: `ErrKeyNotFound`

#### HDel <a id="hdel"></a>
```go
err := tx.HDel(ctx, "hash", "field")
```
- Deletes hash field
- **Rollback**: Restores deleted field

#### HGetAll <a id="hgetall"></a>
```go
fields, err := tx.HGetAll(ctx, "hash")
```
- Returns all fields/values
- **Errors**: `ErrKeyNotFound`

#### HExists <a id="hexists"></a>
```go
exists, err := tx.HExists(ctx, "hash", "field")
```
- Checks field existence
- **Errors**: `ErrInvalidType`

#### HLen <a id="hlen"></a>
```go
count, err := tx.HLen(ctx, "hash")
```
- Returns field count
- **Errors**: `ErrKeyNotFound`

---

### Utility Methods <a id="utility-methods"></a>

#### Exists <a id="exists"></a>
```go
exists, err := tx.Exists(ctx, "key")
```
- Checks key existence
- Returns boolean

#### UpdateTTL <a id="updatettl"></a>
```go
err := tx.UpdateTTL(ctx, "key", newTTL)
```
- Updates key expiration
- **Rollback**: Restores original TTL

#### Type <a id="type"></a>
```go
dataType, err := tx.Type(ctx, "key")
```
- Returns key's data type (String/List/Hash)
- **Errors**: `ErrKeyNotFound`

#### GetWithDetails <a id="getwithdetails"></a>
```go
value, ttl, err := tx.GetWithDetails(ctx, "key")
```
- Returns value + remaining TTL
- TTL = -1 for persistent keys

#### Rename <a id="rename"></a>
```go
err := tx.Rename(ctx, "oldKey", "newKey")
```
- Atomically renames key
- **Rollback**: Restores original names

#### FindByValue <a id="findbyvalue"></a>
```go
keys, err := tx.FindByValue(ctx, targetValue)
```
- Returns all keys with matching value
- **WARNING**: Expensive operation

#### Delete <a id="delete"></a>
```go
err := tx.Delete(ctx, "key")
```
- Permanent key deletion
- **Rollback**: Restores deleted key

---

## 4. Error Reference <a id="error-reference"></a>

| Error Code | Description |
|------------|-------------|
| `ErrTransactionNotActive` | Operation attempted without active transaction |
| `ErrTransactionFailed` | Commit failed due to internal error |
| `ErrKeyNotFound` | Specified key doesn't exist |
| `ErrKeyExpired` | Key has expired |
| `ErrValueMismatch` | CAS operation failed |
| `ErrInvalidType` | Operation mismatch with data type |
| `ErrInvalidTTL` | Negative TTL provided |

---

## 5. Best Practices <a id="best-practices"></a>

1. **Transaction Scope**
   ```go
   tx := db.Transaction()
   defer tx.Rollback()
   
   // Add operations
   if err := tx.Commit(); err != nil {
       // Handle error
   }
   ```

2. **Atomic Patterns**
    - Use `SetCAS` for inventory management:
   ```go
   tx.SetCAS(ctx, "item:123", currentStock, currentStock-1, 0)
   ```

3. **List Management**
    - Implement queues with `LPush`/`RPop`:
   ```go
   tx.LPush(ctx, "queue", task)
   tx.RPop(ctx, "queue")
   ```

4. **Hash Operations**
    - Store objects as hashes:
   ```go
   tx.HSet(ctx, "user:42", "name", "Anton", 3600)
   tx.HSet(ctx, "user:42", "email", "Anton@example.com", 3600)
   ```

5. **Monitoring**
    - Track transaction metrics:
        - Commit success rate
        - Rollback frequency
        - Average transaction duration

6. **Error Handling**
   ```go
   if errors.Is(err, ErrValueMismatch) {
       // Handle CAS failure
   }
   ```

7. **Performance**
    - Batch related operations:
   ```go
   tx.Set(ctx, "key1", val1, 0)
   tx.HSet(ctx, "hash1", "field", val2, 0)
   tx.Incr(ctx, "counter")
   ```