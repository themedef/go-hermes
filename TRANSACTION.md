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
        - [LTrim](#ltrim)
    - [Hash Operations](#hash-operations)
        - [HSet](#hset)
        - [HGet](#hget)
        - [HDel](#hdel)
        - [HGetAll](#hgetall)
        - [HExists](#hexists)
        - [HLen](#hlen)
    - [Set Operations](#set-operations)
        - [SAdd](#sadd)
        - [SRem](#srem)
        - [SMembers](#smembers)
        - [SIsMember](#sismember)
        - [SCard](#scard)
    - [Utility Methods](#utility-methods)
        - [Exists](#exists)
        - [Expire](#expire)
        - [Persist](#persist)
        - [Type](#type)
        - [GetWithDetails](#getwithdetails)
        - [Rename](#rename)
        - [FindByValue](#findbyvalue)
        - [Delete](#delete)
4. [Error Reference](#error-reference)
5. [Best Practices](#best-practices)

---

## 1. Overview <a id="overview"></a>

A **transaction** in the Hermes store aggregates multiple read/write operations so they execute atomically. Within a transaction, changes are buffered locally and remain invisible to the outside world until the transaction is committed. If any operation fails during commit, a rollback mechanism automatically restores the previous state.

**Key Features:**
- **Optimistic Concurrency:** Operations are queued and only become visible upon commit.
- **Atomic Guarantee:** Either all operations succeed or none of them take effect.
- **Rollback Support:** Each operation registers a rollback function to undo the change if needed.
- **Support for Complex Data Structures:** Transactions cover key-value pairs, counters, lists, hashes, and sets.

---

## 2. Transaction Lifecycle <a id="transaction-lifecycle"></a>

### Begin Transaction <a id="begin-transaction"></a>
```go
tx := db.Transaction()
```
**Description:**  
Initializes a new transaction context. This must be invoked before executing any transactional operations.

---

### Commit <a id="commit"></a>
```go
err := tx.Commit()
```
**Description:**  
Executes all queued transaction operations atomically. If any command fails during commit, the transaction automatically calls all rollback functions, reverting all changes.

**On Success:**
- All operations are applied.
- The transaction buffers are cleared.

**On Failure:**
- Returns an error (for example, `ErrTransactionFailed`) and reverts all changes.

---

### Rollback <a id="rollback"></a>
```go
err := tx.Rollback()
```
**Description:**  
Discards all queued operations and invokes rollback callbacks to restore the previous state. If no transaction is active, this is a no-op.

---

## 3. Operations <a id="operations"></a>

### Key-Value Operations <a id="key-value-operations"></a>

#### Set <a id="set"></a>
```go
err := tx.Set(ctx, "key", value, ttl)
```
**Description:**  
Creates or updates a key with the specified value and TTL.  
**Rollback:** If the key previously existed, its former value is restored; otherwise, the key is deleted.

---

#### Get <a id="get"></a>
```go
val, err := tx.Get(ctx, "key")
```
**Description:**  
Retrieves the current value of a key, including any buffered changes in the transaction.  
**Errors:**
- `ErrKeyNotFound` if the key does not exist or has expired.

---

#### SetNX <a id="setnx"></a>
```go
err := tx.SetNX(ctx, "key", value, ttl)
```
**Description:**  
Sets the key only if it does not already exist.  
**Rollback:** Deletes the key if it was newly created by this operation.

---

#### SetXX <a id="setxx"></a>
```go
err := tx.SetXX(ctx, "key", value, ttl)
```
**Description:**  
Updates the key only if it already exists.  
**Rollback:** Restores the previous value if the operation is later rolled back.

---

#### SetCAS <a id="setcas"></a>
```go
err := tx.SetCAS(ctx, "key", oldVal, newVal, ttl)
```
**Description:**  
Performs a compare-and-swap update – sets the key to a new value only if the current value equals `oldVal`.  
**Errors:**
- `ErrValueMismatch` if the current value does not equal `oldVal`.
- `ErrKeyNotFound` if the key does not exist.  
  **Rollback:** Restores the previous state on failure.

---

#### GetSet <a id="getset"></a>
```go
oldVal, err := tx.GetSet(ctx, "key", newVal, ttl)
```
**Description:**  
Atomically sets a new value for a key and returns the old value.  
**Rollback:** Restores the previous value if necessary.

---

### Atomic Counters <a id="atomic-counters"></a>

#### Incr <a id="increment"></a>
```go
err := tx.Incr(ctx, "counter")
```
**Description:**  
Increments an integer value by 1. If the key does not exist, it is initialized to `1`.  
**Rollback:** Restores the original counter value.

---

#### Decr <a id="decrement"></a>
```go
err := tx.Decr(ctx, "counter")
```
**Description:**  
Decrements an integer value by 1. If the key does not exist, it is initialized to `-1`.  
**Rollback:** Restores the previous value.

---

#### IncrBy <a id="incrby"></a>
```go
err := tx.IncrBy(ctx, "counter", 10)
```
**Description:**  
Increments an integer counter by a specified amount. If the key is missing, it is created with the provided increment value.  
**Rollback:** Restores the previous counter value.

---

#### DecrBy <a id="decrby"></a>
```go
err := tx.DecrBy(ctx, "counter", 5)
```
**Description:**  
Decrements an integer counter by a specified amount. If the key is missing, it is created with the negative of the specified value.  
**Rollback:** Restores the previous counter value.

---

### List Operations <a id="list-operations"></a>

#### LPush <a id="lpush"></a>
```go
err := tx.LPush(ctx, "list", value1, value2)
```
**Description:**  
Inserts one or more elements at the beginning (left side) of the list.  
**Rollback:** Restores the original list state.

---

#### RPush <a id="rpush"></a>
```go
err := tx.RPush(ctx, "list", value1, value2)
```
**Description:**  
Appends one or more elements at the end (right side) of the list.  
**Rollback:** Restores the original list state.

---

#### LPop <a id="lpop"></a>
```go
val, err := tx.LPop(ctx, "list")
```
**Description:**  
Removes and returns the first element of the list.  
**Rollback:** Reinserts the popped element into the list.

---

#### RPop <a id="rpop"></a>
```go
val, err := tx.RPop(ctx, "list")
```
**Description:**  
Removes and returns the last element of the list.  
**Rollback:** Reinserts the popped element into the list.

---

#### LRange <a id="lrange"></a>
```go
items, err := tx.LRange(ctx, "list", start, end)
```
**Description:**  
Retrieves a slice of elements from the list between the given start and end indices. Supports negative indices.

---

#### LLen <a id="llen"></a>
```go
length, err := tx.LLen(ctx, "list")
```
**Description:**  
Returns the number of elements in the list.

---

#### LTrim <a id="ltrim"></a>
```go
err := tx.LTrim(ctx, "list", start, stop)
```
**Description:**  
Trims the list to contain only the elements within the specified range. If the resulting list is empty, the key is removed.  
**Rollback:** Restores the original list.

---

### Hash Operations <a id="hash-operations"></a>

#### HSet <a id="hset"></a>
```go
err := tx.HSet(ctx, "hash", "field", value, ttl)
```
**Description:**  
Sets a field in the hash to the specified value.  
**Rollback:** If the field existed before, its value is restored; otherwise, it is removed.

---

#### HGet <a id="hget"></a>
```go
val, err := tx.HGet(ctx, "hash", "field")
```
**Description:**  
Retrieves the value associated with a specific field in the hash.  
**Errors:**
- `ErrKeyNotFound` if the hash or field is absent.

---

#### HDel <a id="hdel"></a>
```go
err := tx.HDel(ctx, "hash", "field")
```
**Description:**  
Deletes a field from the hash.  
**Rollback:** Restores the deleted field with its previous value if necessary.

---

#### HGetAll <a id="hgetall"></a>
```go
fields, err := tx.HGetAll(ctx, "hash")
```
**Description:**  
Returns all fields and their corresponding values from the hash.  
**Errors:**
- `ErrKeyNotFound` if the hash does not exist.

---

#### HExists <a id="hexists"></a>
```go
exists, err := tx.HExists(ctx, "hash", "field")
```
**Description:**  
Checks for the existence of the specified field in the hash.

---

#### HLen <a id="hlen"></a>
```go
count, err := tx.HLen(ctx, "hash")
```
**Description:**  
Returns the number of fields in the hash.  
**Errors:**
- `ErrKeyNotFound` if the hash is missing.

---

### Set Operations <a id="set-operations"></a>

#### SAdd <a id="sadd"></a>
```go
err := tx.SAdd(ctx, "setKey", member1, member2)
```
**Description:**  
Adds one or more members to a set. If the set does not exist, it is created.  
**Rollback:** Restores the previous set state (or deletes the key if it was newly created).

---

#### SRem <a id="srem"></a>
```go
err := tx.SRem(ctx, "setKey", member1)
```
**Description:**  
Removes the specified members from the set. If the set becomes empty, the key is deleted.  
**Rollback:** Restores the original set contents.

---

#### SMembers <a id="smembers"></a>
```go
members, err := tx.SMembers(ctx, "setKey")
```
**Description:**  
Retrieves all members of the set.  
**Errors:**
- `ErrKeyNotFound` if the set does not exist.

---

#### SIsMember <a id="sismember"></a>
```go
exists, err := tx.SIsMember(ctx, "setKey", member)
```
**Description:**  
Checks whether the provided member exists within the set.

---

#### SCard <a id="scard"></a>
```go
cardinality, err := tx.SCard(ctx, "setKey")
```
**Description:**  
Returns the number of members in the set.  
**Errors:**
- `ErrKeyNotFound` if the set is missing.

---

### Utility Methods <a id="utility-methods"></a>

#### Exists <a id="exists"></a>
```go
exists, err := tx.Exists(ctx, "key")
```
**Description:**  
Checks whether the specified key exists within the current transaction.

---

#### Expire <a id="expire"></a>
```go
err := tx.Expire(ctx, "key", ttl)
```
**Description:**  
Sets a new time-to-live (TTL) for an existing key.  
**Rollback:** Restores the original TTL setting.

---

#### Persist <a id="persist"></a>
```go
err := tx.Persist(ctx, "key")
```
**Description:**  
Removes the TTL from a key, making it persistent.  
**Rollback:** Restores the previous TTL if needed.

---

#### Type <a id="type"></a>
```go
dataType, err := tx.Type(ctx, "key")
```
**Description:**  
Returns the data type of the specified key (such as String, List, Hash, or Set).  
**Errors:**
- `ErrKeyNotFound` if the key does not exist.

---

#### GetWithDetails <a id="getwithdetails"></a>
```go
value, ttl, err := tx.GetWithDetails(ctx, "key")
```
**Description:**  
Retrieves the key’s value along with its remaining TTL in seconds. If the key is persistent, TTL is `-1`.

---

#### Rename <a id="rename"></a>
```go
err := tx.Rename(ctx, "oldKey", "newKey")
```
**Description:**  
Atomically renames a key.  
**Rollback:** Restores the original key(s) if the rename operation is undone.

---

#### FindByValue <a id="findbyvalue"></a>
```go
keys, err := tx.FindByValue(ctx, targetValue)
```
**Description:**  
Finds and returns all keys whose values match the specified target value.  
**Note:** This operation can be expensive on large datasets.

---

#### Delete <a id="delete"></a>
```go
err := tx.Delete(ctx, "key")
```
**Description:**  
Deletes the specified key permanently.  
**Rollback:** Restores the key along with its previous value if required.

---

## 4. Error Reference <a id="error-reference"></a>

| Error Code               | Description                                                              |
|--------------------------|--------------------------------------------------------------------------|
| **ErrTransactionNotActive** | Operation attempted without an active transaction.                     |
| **ErrTransactionFailed**    | Commit failed due to an error in one or more operations.                 |
| **ErrKeyNotFound**          | The specified key does not exist or has expired.                         |
| **ErrKeyExpired**           | The specified key exists but its TTL has expired.                        |
| **ErrValueMismatch**        | The compare-and-swap (CAS) operation failed because the current value did not match the expected value. |
| **ErrInvalidType**          | The operation was executed on a key whose data type does not match the expected type.  |
| **ErrInvalidTTL**           | The provided TTL is negative or otherwise invalid.                     |

---

## 5. Best Practices <a id="best-practices"></a>

1. **Transaction Scope and Cleanup**
    - Begin a transaction and ensure you call rollback if commit is not reached:
      ```go
      tx := db.Transaction()
      defer tx.Rollback() // Safe to call; has no effect if commit succeeds
      
      // Execute multiple operations within the transaction...
      if err := tx.Commit(); err != nil {
          // Handle commit error accordingly
      }
      ```

2. **Atomic Patterns for Critical Data**
    - Utilize CAS operations to safely update critical values:
      ```go
      err := tx.SetCAS(ctx, "inventory", currentStock, currentStock-1, 0)
      if err != nil {
          // Handle CAS failure (e.g., inventory conflict)
      }
      ```

3. **Batch Related Operations**
    - Group multiple operations within a single transaction to maintain consistency:
      ```go
      tx.Set(ctx, "user:balance", 150, 0)
      tx.HSet(ctx, "user:profile", "email", "user@example.com", 0)
      tx.Incr(ctx, "user:loginCount")
      ```

4. **Rollback Awareness**
    - Understand each operation's rollback behavior. Updates, deletions, and counter modifications are fully reversible within a transaction.

5. **Error Handling**
    - Always check and distinguish errors using patterns like `errors.Is()`:
      ```go
      if err := tx.Commit(); err != nil {
          if errors.Is(err, ErrValueMismatch) {
              // Handle specific CAS failure
          } else {
              // General error handling
          }
      }
      ```

6. **Monitor Transaction Performance**
    - Track commit/rollback latency and failure rates to fine-tune TTL settings, shard counts, and overall system performance.

---

This documentation provides a comprehensive guide to using transactions in the Hermes store, including all methods (key-value, atomic counters, lists, hashes, sets, and utility functions) with full descriptions, rollback behavior, error details, and recommended practices. If you have any questions or need further clarification, please feel free to ask!