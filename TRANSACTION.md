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
      - [IncrBy](#incrby)
      - [DecrBy](#decrby)
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
      - [Expire](#expire)
      - [Persist](#persist)
      - [Type](#type)
      - [GetWithDetails](#getwithdetails)
      - [Rename](#rename)
      - [FindByValue](#findbyvalue)
      - [Delete](#delete)
3. [Error Reference](#error-reference)
4. [Best Practices](#best-practices)

---

## 1. Overview <a id="overview"></a>

A **transaction** groups multiple operations (reads/writes) to execute atomically. Changes are buffered locally until the transaction is committed. If any operation fails during commit, all changes are rolled back automatically.

Key Features:
- **Optimistic Concurrency**: Changes remain invisible until commit.
- **Atomic Guarantee**: All operations succeed or none do.
- **Rollback Support**: Automatically restores previous state on failure.
- **Support for Complex Data Structures**: Transactions handle key-value, counters, lists, and hashes.

---

## 2. Transaction Lifecycle <a id="transaction-lifecycle"></a>

### Begin Transaction <a id="begin-transaction"></a>
```go
tx := db.Transaction()
```
Initializes a new transaction context. Must be invoked before any transactional operations.

---

### Commit <a id="commit"></a>
```go
err := tx.Commit()
```
Executes all queued operations atomically.
- Returns an error (e.g., `ErrTransactionFailed`) if any operation fails.
- Automatically rolls back changes on failure.

---

### Rollback <a id="rollback"></a>
```go
err := tx.Rollback()
```
Discards all pending operations.
- No-op if there is no active transaction.

---

## 3. Operations <a id="operations"></a>

### Key-Value Operations <a id="key-value-operations"></a>

#### Set <a id="set"></a>
```go
err := tx.Set(ctx, "key", value, ttl)
```
Creates or updates a key with the specified TTL.  
**Rollback**: Restores the previous value or deletes the key if it was newly created.

---

#### Get <a id="get"></a>
```go
val, err := tx.Get(ctx, "key")
```
Retrieves the current value of the key, including any changes queued in the transaction.  
**Errors**: `ErrKeyNotFound`, `ErrKeyExpired`.

---

#### SetNX <a id="setnx"></a>
```go
err := tx.SetNX(ctx, "key", value, ttl)
```
Sets the key only if it does not exist.  
**Rollback**: Deletes the key if it was created by this operation.

---

#### SetXX <a id="setxx"></a>
```go
err := tx.SetXX(ctx, "key", value, ttl)
```
Updates the key only if it already exists.  
**Rollback**: Restores the original value.

---

#### SetCAS <a id="setcas"></a>
```go
err := tx.SetCAS(ctx, "key", oldVal, newVal, ttl)
```
Performs a compare-and-swap operation. Updates the key only if its current value matches `oldVal`.  
**Errors**: `ErrValueMismatch`, `ErrKeyNotFound`.

---

#### GetSet <a id="getset"></a>
```go
oldVal, err := tx.GetSet(ctx, "key", newVal, ttl)
```
Atomically sets a new value and returns the old value.  
**Rollback**: Restores the previous value.

---

### Atomic Counters <a id="atomic-counters"></a>

#### Incr <a id="increment"></a>
```go
err := tx.Incr(ctx, "counter")
```
Increments an integer value by 1. If the key is missing, initializes it to 1.  
**Rollback**: Restores the previous value.

---

#### Decr <a id="decrement"></a>
```go
err := tx.Decr(ctx, "counter")
```
Decrements an integer value by 1. If the key is missing, initializes it to -1.  
**Rollback**: Restores the previous value.

---

#### IncrBy <a id="incrby"></a>
```go
err := tx.IncrBy(ctx, "counter", 10)
```
Increments an integer value by a specified amount.  
If the key does not exist, it is created with the given increment value.  
**Rollback**: Restores the previous value.

---

#### DecrBy <a id="decrby"></a>
```go
err := tx.DecrBy(ctx, "counter", 5)
```
Decrements an integer value by a specified amount.  
If the key does not exist, it is created with the negative of the decrement value.  
**Rollback**: Restores the previous value.

---

### List Operations <a id="list-operations"></a>

#### LPush <a id="lpush"></a>
```go
err := tx.LPush(ctx, "list", value)
```
Inserts a value at the head (left) of the list.  
**Rollback**: Restores the original list state.

---

#### RPush <a id="rpush"></a>
```go
err := tx.RPush(ctx, "list", value)
```
Appends a value at the tail (right) of the list.  
**Rollback**: Restores the original list state.

---

#### LPop <a id="lpop"></a>
```go
val, err := tx.LPop(ctx, "list")
```
Removes and returns the first element of the list.  
**Rollback**: Reinserts the popped element.

---

#### RPop <a id="rpop"></a>
```go
val, err := tx.RPop(ctx, "list")
```
Removes and returns the last element of the list.  
**Rollback**: Reinserts the popped element.

---

#### LRange <a id="lrange"></a>
```go
items, err := tx.LRange(ctx, "list", start, end)
```
Returns a slice of list elements between specified indices.  
Supports negative indices for counting from the end.

---

#### LLen <a id="llen"></a>
```go
length, err := tx.LLen(ctx, "list")
```
Returns the length of the list.

---

### Hash Operations <a id="hash-operations"></a>

#### HSet <a id="hset"></a>
```go
err := tx.HSet(ctx, "hash", "field", value, ttl)
```
Sets a field in a hash to the specified value.  
**Rollback**: Restores the previous field value or removes the field if newly added.

---

#### HGet <a id="hget"></a>
```go
val, err := tx.HGet(ctx, "hash", "field")
```
Retrieves the value of a hash field.  
**Errors**: `ErrKeyNotFound`.

---

#### HDel <a id="hdel"></a>
```go
err := tx.HDel(ctx, "hash", "field")
```
Deletes a field from a hash.  
**Rollback**: Restores the deleted field.

---

#### HGetAll <a id="hgetall"></a>
```go
fields, err := tx.HGetAll(ctx, "hash")
```
Returns all fields and values in the hash.  
**Errors**: `ErrKeyNotFound`.

---

#### HExists <a id="hexists"></a>
```go
exists, err := tx.HExists(ctx, "hash", "field")
```
Checks whether a specific field exists in the hash.

---

#### HLen <a id="hlen"></a>
```go
count, err := tx.HLen(ctx, "hash")
```
Returns the number of fields in the hash.  
**Errors**: `ErrKeyNotFound`.

---

### Utility Methods <a id="utility-methods"></a>

#### Exists <a id="exists"></a>
```go
exists, err := tx.Exists(ctx, "key")
```
Checks whether a key exists within the transaction context.

---

#### Expire <a id="expire"></a>
```go
err := tx.Expire(ctx, "key", ttl)
```
Sets a new TTL for an existing key within the transaction.  
**Rollback**: Restores the original expiration state.

---

#### Persist <a id="persist"></a>
```go
err := tx.Persist(ctx, "key")
```
Removes the TTL from a key, making it persistent within the transaction.  
**Rollback**: Restores the previous TTL.

---

#### Type <a id="type"></a>
```go
dataType, err := tx.Type(ctx, "key")
```
Returns the data type of the key (e.g., String, List, Hash).  
**Errors**: `ErrKeyNotFound`.

---

#### GetWithDetails <a id="getwithdetails"></a>
```go
value, ttl, err := tx.GetWithDetails(ctx, "key")
```
Returns the key’s value along with its remaining TTL (in seconds).  
TTL is `-1` if the key is persistent.

---

#### Rename <a id="rename"></a>
```go
err := tx.Rename(ctx, "oldKey", "newKey")
```
Atomically renames a key.  
**Rollback**: Restores the original key names.

---

#### FindByValue <a id="findbyvalue"></a>
```go
keys, err := tx.FindByValue(ctx, targetValue)
```
Returns all keys whose values match the specified target value.  
**Warning**: This operation can be expensive.

---

#### Delete <a id="delete"></a>
```go
err := tx.Delete(ctx, "key")
```
Deletes a key permanently.  
**Rollback**: Restores the key and its previous value.

---

## 4. Error Reference <a id="error-reference"></a>

| Error Code              | Description                                    |
|-------------------------|------------------------------------------------|
| `ErrTransactionNotActive` | Operation attempted without an active transaction. |
| `ErrTransactionFailed`    | Commit failed due to an internal error.      |
| `ErrKeyNotFound`           | The specified key does not exist.            |
| `ErrKeyExpired`            | The specified key has expired.               |
| `ErrValueMismatch`         | CAS operation failed due to a value mismatch.|
| `ErrInvalidType`           | Operation type mismatch with key's data type.|
| `ErrInvalidTTL`            | Provided TTL is negative.                    |

---

## 5. Best Practices <a id="best-practices"></a>

1. **Transaction Scope**
   - Begin a transaction and ensure rollback is called if commit is not reached:
     ```go
     tx := db.Transaction()
     defer tx.Rollback() // Safe to call; no effect if commit succeeds
     
     // Perform operations...
     if err := tx.Commit(); err != nil {
         // Handle commit error
     }
     ```
2. **Atomic Patterns**
   - Use CAS operations to update critical data safely:
     ```go
     err := tx.SetCAS(ctx, "inventory", currentStock, currentStock-1, 0)
     if err != nil {
         // Handle CAS failure
     }
     ```
3. **Batching Operations**
   - Group related operations in a transaction to minimize intermediate state:
     ```go
     tx.Set(ctx, "key1", val1, 0)
     tx.HSet(ctx, "hash1", "field", val2, 0)
     tx.Incr(ctx, "counter")
     ```
4. **Rollback Awareness**
   - Understand the rollback behavior for each operation; for example, updates, deletions, and counter modifications are fully reversible within a transaction.
5. **Performance Monitoring**
   - Track transaction commit and rollback rates as well as latency to fine-tune your usage.
6. **Error Handling**
   - Always inspect errors (using `errors.Is()` where applicable) to distinguish between expected and unexpected failures.
   - Example:
     ```go
     if errors.Is(err, ErrValueMismatch) {
         // Specific handling for CAS failure
     }
     ```
