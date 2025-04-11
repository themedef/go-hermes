# Store Documentation

---

## Table of Contents

1. [Initialization](#initialization)
2. [Core Methods](#core-methods)
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
      - [DropAll](#dropall)
      - [GetRawEntry](#getrawentry)
      - [RestoreRawEntry](#restorerawentry)
   - [PubSub / Subscription Methods](#pubsub-methods)
      - [Subscribe](#subscribe)
      - [Unsubscribe](#unsubscribe)
      - [ListSubscriptions](#listsubscriptions)
      - [CloseAllSubscriptionsForKey](#closeallsubscriptionsforkey)
   - [Accessor Methods](#accessor-methods)
      - [Logger](#logger)
      - [Commands](#commands)
      - [Transaction](#transaction)
   - [Shutdown](#shutdown)
3. [Error Reference](#error-reference)
4. [Best Practices](#best-practices)

---

## 1. Initialization <a id="initialization"></a>

### Creating a Store Instance

To initialize the store, call the constructor `NewStore` and pass an optional configuration:

```go
import (
    "github.com/themedef/go-hermes"
    "log"
    "time"
)

func main() {
    // Minimal configuration (default values)
    db := hermes.NewStore(hermes.Config{})

    // Custom configuration with all parameters, including shard count and PubSub buffer size:
    db = hermes.NewStore(hermes.Config{
        ShardCount:       4,                // Number of shards. Increases parallelism and reduces lock contention.
        CleanupInterval:  5 * time.Second,  // Auto-delete expired keys every 5 seconds.
        EnableLogging:    true,             // Enable operation logging.
        LogFile:          "hermes.log",     // Log file path (empty = stdout).
        LogBufferSize:    2000,             // Buffer 2000 log entries.
        MinLevel:         hermes.INFO,      // Log INFO and higher levels.
        PubSubBufferSize: 5000,             // Buffer size for PubSub channels.
    })

    // Ensure the store is closed properly on application exit.
    defer func() {
        if err := db.Close(); err != nil {
            log.Printf("Error while closing Hermes store: %v\n", err)
        }
    }()
}
```

### Configuration Options

| Parameter           | Type              | Default | Description                                                                                         |
|---------------------|-------------------|---------|-----------------------------------------------------------------------------------------------------|
| `ShardCount`        | `int`             | `1`     | Number of shards. Higher values improve parallelism under high concurrency.                        |
| `CleanupInterval`   | `time.Duration`   | `1s`    | Interval for background cleanup of expired keys. Set to `0` to disable cleanup.                   |
| `EnableLogging`     | `bool`            | `false` | Enables logging of operations.                                                                      |
| `LogFile`           | `string`          | `""`    | File path for logs. If empty, logs are written to stdout.                                           |
| `LogBufferSize`     | `int`             | `1000`  | Size of the asynchronous log buffer.                                                              |
| `MinLevel`          | `logger.LogLevel` | `DEBUG` | Minimum log level. Levels: `DEBUG`, `INFO`, `WARN`, `ERROR`.                                         |
| `PubSubBufferSize`  | `int`             | `10000` | Buffer size for PubSub channels. If not provided or ≤ 0, defaults to 10000.                           |

---

## 2. Core Methods <a id="core-methods"></a>

### 2.1 Key-Value Operations <a id="key-value-operations"></a>

#### **Set** <a id="set"></a>
```go
err := db.Set(context.Background(), "user", "Test", 60) // TTL = 60 seconds
```
**Description:**  
Sets a key to the given value with an optional TTL.

**Errors:**
- `ErrContextCanceled`: operation canceled via context.
- `ErrInvalidKey`: key is empty or improperly formatted.
- `ErrInvalidTTL`: TTL value is negative or otherwise invalid.

---

#### **Get** <a id="get"></a>
```go
value, err := db.Get(context.Background(), "user")
```
**Description:**  
Retrieves the value of the specified key. If the key is missing or expired, returns `ErrKeyNotFound`.

**Errors:**
- `ErrContextCanceled`
- `ErrKeyNotFound`

---

#### **SetNX** <a id="setnx"></a>
```go
success, err := db.SetNX(context.Background(), "user", "Test", 60)
```
**Description:**  
Sets the key only if it does not already exist.

**Errors:**
- `ErrContextCanceled`
- `ErrInvalidKey`
- `ErrInvalidTTL`
- `ErrKeyExists`

---

#### **SetXX** <a id="setxx"></a>
```go
success, err := db.SetXX(context.Background(), "user", "Test", 60)
```
**Description:**  
Updates the key only if it already exists.

**Errors:**
- `ErrContextCanceled`
- `ErrInvalidKey`
- `ErrInvalidTTL`
- `ErrKeyNotFound`

---

#### **SetCAS** <a id="setcas"></a>
```go
err := db.SetCAS(context.Background(), "user", "old_value", "new_value", 60)
```
**Description:**  
Atomically updates a key only if its current value matches the given `old_value`.

**Errors:**
- `ErrContextCanceled`
- `ErrKeyNotFound`
- `ErrValueMismatch`
- `ErrInvalidTTL`

---

#### **GetSet** <a id="getset"></a>
```go
oldVal, err := db.GetSet(context.Background(), "key", "new_value", 120)
```
**Description:**  
Atomically sets a new value and returns the previous one. If the key does not exist, it is set and `nil` is returned.

**Errors:**
- `ErrContextCanceled`
- `ErrInvalidTTL`

---

### 2.2 Atomic Counters <a id="atomic-counters"></a>

#### **Incr** <a id="increment"></a>
```go
newValue, err := db.Incr(context.Background(), "counter")
```
**Description:**  
Increments the integer value by 1. If missing, the key is created with the value `1`.

**Errors:**
- `ErrContextCanceled`
- `ErrInvalidValueType` – if the current value isn’t an `int64`.

---

#### **Decr** <a id="decrement"></a>
```go
newValue, err := db.Decr(context.Background(), "counter")
```
**Description:**  
Decrements the integer value by 1. If missing, the key is created with the value `-1`.

**Errors:**
- `ErrContextCanceled`
- `ErrInvalidValueType`

---

#### **IncrBy** <a id="incrby"></a>
```go
newValue, err := db.IncrBy(context.Background(), "counter", 10)
```
**Description:**  
Increments the integer value by the specified amount. If the key is missing, it is created with the specified increment value.

**Errors:**
- `ErrContextCanceled`
- `ErrInvalidValueType`

---

#### **DecrBy** <a id="decrby"></a>
```go
newValue, err := db.DecrBy(context.Background(), "counter", 5)
```
**Description:**  
Decrements the integer value by the specified amount. If the key is missing, it is created with the negative value of the decrement.

**Errors:**
- `ErrContextCanceled`
- `ErrInvalidValueType`

---

### 2.3 List Operations <a id="list-operations"></a>

#### **LPush** <a id="lpush"></a>
```go
err := db.LPush(context.Background(), "tasks", "task1", "task2")
```
**Description:**  
Inserts one or more elements at the beginning (left) of the list. Creates the list if it doesn’t exist.

**Errors:**
- `ErrContextCanceled`
- `ErrInvalidKey`
- `ErrInvalidType` – if the key exists but isn’t a list.
- `ErrEmptyValues` – if no values are provided.

---

#### **RPush** <a id="rpush"></a>
```go
err := db.RPush(context.Background(), "tasks", "task3", "task4")
```
**Description:**  
Inserts one or more elements at the end (right) of the list. Creates the list if it doesn’t exist.

**Errors:**
- `ErrContextCanceled`
- `ErrInvalidType`
- `ErrEmptyValues`

---

#### **LPop** <a id="lpop"></a>
```go
value, err := db.LPop(context.Background(), "tasks")
```
**Description:**  
Removes and returns the first element of the list.

**Errors:**
- `ErrContextCanceled`
- `ErrKeyNotFound`
- `ErrInvalidType`
- `ErrEmptyList`

---

#### **RPop** <a id="rpop"></a>
```go
value, err := db.RPop(context.Background(), "tasks")
```
**Description:**  
Removes and returns the last element of the list.

**Errors:**
- `ErrContextCanceled`
- `ErrKeyNotFound`
- `ErrInvalidType`
- `ErrEmptyList`

---

#### **LRange** <a id="lrange"></a>
```go
values, err := db.LRange(context.Background(), "mylist", 0, -1)
```
**Description:**  
Returns a slice of list elements from index `start` to `end`. Negative indices are supported for counting from the end.

**Errors:**
- `ErrContextCanceled`
- `ErrKeyNotFound`
- `ErrInvalidType`

---

#### **LLen** <a id="llen"></a>
```go
length, err := db.LLen(context.Background(), "tasks")
```
**Description:**  
Returns the number of elements in the list.

**Errors:**
- `ErrContextCanceled`
- `ErrKeyNotFound`
- `ErrInvalidType`

---

#### **LTrim** <a id="ltrim"></a>
```go
err := db.LTrim(context.Background(), "tasks", 1, 3)
```
**Description:**  
Trims the list, keeping only the elements between the specified start and stop indices.  
If the resulting list is empty, the key is removed.

**Errors:**
- `ErrContextCanceled`
- `ErrKeyNotFound`
- `ErrInvalidType`

---

### 2.4 Hash Operations <a id="hash-operations"></a>

#### **HSet** <a id="hset"></a>
```go
err := db.HSet(context.Background(), "user", "name", "Test", 0)
```
**Description:**  
Sets a field in the hash. Creates the hash if needed. If TTL is set to 0, the expiration remains unchanged.

**Errors:**
- `ErrContextCanceled`
- `ErrInvalidType`
- `ErrInvalidTTL`

---

#### **HGet** <a id="hget"></a>
```go
value, err := db.HGet(context.Background(), "user", "name")
```
**Description:**  
Retrieves the value for a given hash field.

**Errors:**
- `ErrContextCanceled`
- `ErrKeyNotFound`
- `ErrInvalidType`

---

#### **HDel** <a id="hdel"></a>
```go
err := db.HDel(context.Background(), "user", "name")
```
**Description:**  
Deletes a specific hash field. If the hash becomes empty after deletion, the key is removed.

**Errors:**
- `ErrContextCanceled`
- `ErrKeyNotFound`
- `ErrInvalidType`

---

#### **HGetAll** <a id="hgetall"></a>
```go
fields, err := db.HGetAll(context.Background(), "user")
```
**Description:**  
Returns all fields and values in a hash.

**Errors:**
- `ErrContextCanceled`
- `ErrKeyNotFound`
- `ErrInvalidType`

---

#### **HExists** <a id="hexists"></a>
```go
exists, err := db.HExists(context.Background(), "user", "email")
```
**Description:**  
Checks for the existence of a field in a hash.

**Errors:**
- `ErrContextCanceled`
- `ErrKeyNotFound`
- `ErrInvalidType`

---

#### **HLen** <a id="hlen"></a>
```go
length, err := db.HLen(context.Background(), "user")
```
**Description:**  
Returns the number of fields in the hash.

**Errors:**
- `ErrContextCanceled`
- `ErrKeyNotFound`
- `ErrInvalidType`

---

### 2.5 Set Operations <a id="set-operations"></a>

#### **SAdd** <a id="sadd"></a>
```go
err := db.SAdd(context.Background(), "tags", "go", "redis", "hermes")
```
**Description:**  
Adds one or more members to a set. If the set does not exist, it is created.

**Errors:**
- `ErrContextCanceled`
- `ErrInvalidKey`
- `ErrEmptyValues`
- `ErrInvalidType`

---

#### **SRem** <a id="srem"></a>
```go
err := db.SRem(context.Background(), "tags", "redis")
```
**Description:**  
Removes the specified members from the set. If the set becomes empty, the key is deleted.

**Errors:**
- `ErrContextCanceled`
- `ErrKeyNotFound`
- `ErrInvalidType`

---

#### **SMembers** <a id="smembers"></a>
```go
members, err := db.SMembers(context.Background(), "tags")
```
**Description:**  
Returns all members of the set.

**Errors:**
- `ErrContextCanceled`
- `ErrKeyNotFound`
- `ErrInvalidType`

---

#### **SIsMember** <a id="sismember"></a>
```go
exists, err := db.SIsMember(context.Background(), "tags", "go")
```
**Description:**  
Checks whether the given member is a member of the set.

**Errors:**
- `ErrContextCanceled`
- `ErrKeyNotFound`
- `ErrInvalidType`

---

#### **SCard** <a id="scard"></a>
```go
cardinality, err := db.SCard(context.Background(), "tags")
```
**Description:**  
Returns the number of members in the set.

**Errors:**
- `ErrContextCanceled`
- `ErrKeyNotFound`
- `ErrInvalidType`

---

### 2.6 Utility Methods <a id="utility-methods"></a>

#### **Exists** <a id="exists"></a>
```go
exists, err := db.Exists(context.Background(), "user")
```
**Description:**  
Checks if a key exists and its TTL has not expired.

**Errors:**
- `ErrContextCanceled`

---

#### **Expire** <a id="expire"></a>
```go
success, err := db.Expire(context.Background(), "user", 60)
```
**Description:**  
Sets a new TTL for an existing key.

**Response:**  
Returns `true` if successful.

**Errors:**
- `ErrContextCanceled`
- `ErrInvalidTTL`
- `ErrKeyNotFound`

---

#### **Persist** <a id="persist"></a>
```go
success, err := db.Persist(context.Background(), "user")
```
**Description:**  
Removes the TTL from a key, making it persistent.

**Response:**  
Returns `true` if successful.

**Errors:**
- `ErrContextCanceled`
- `ErrKeyNotFound`

---

#### **Type** <a id="type"></a>
```go
dataType, err := db.Type(context.Background(), "user")
```
**Description:**  
Returns the data type of the specified key (e.g., `String`, `List`, `Hash`, `Set`).

**Errors:**
- `ErrContextCanceled`
- `ErrKeyNotFound`

---

#### **GetWithDetails** <a id="getwithdetails"></a>
```go
value, ttl, err := db.GetWithDetails(context.Background(), "user")
```
**Description:**  
Returns the value along with its remaining TTL (in seconds). If the key does not expire, TTL is `-1`.

**Errors:**
- `ErrContextCanceled`
- `ErrKeyNotFound`

---

#### **Rename** <a id="rename"></a>
```go
err := db.Rename(context.Background(), "oldKey", "newKey")
```
**Description:**  
Renames an existing key. Fails if the old key doesn’t exist or the new key already exists.

**Errors:**
- `ErrContextCanceled`
- `ErrInvalidKey`
- `ErrKeyNotFound`
- `ErrKeyExists`

---

#### **FindByValue** <a id="findbyvalue"></a>
```go
keys, err := db.FindByValue(context.Background(), "Test")
```
**Description:**  
Finds all keys that have the specified value.

**Errors:**
- `ErrContextCanceled`
- `ErrKeyNotFound`

---

#### **Delete** <a id="delete"></a>
```go
err := db.Delete(context.Background(), "user")
```
**Description:**  
Deletes the specified key from the store.

**Errors:**
- `ErrKeyNotFound`

---

#### **DropAll** <a id="dropall"></a>
```go
err := db.DropAll(context.Background())
```
**Description:**  
Deletes all keys from the store.

**Errors:**
- `ErrContextCanceled`

---

#### **GetRawEntry** <a id="getrawentry"></a>
```go
entry, err := db.GetRawEntry(context.Background(), "user")
```
**Description:**  
Returns the raw internal entry (of type `types.Entry`) for the given key. Useful for debugging or advanced operations.

**Errors:**
- `ErrContextCanceled`
- `ErrKeyNotFound`

---

#### **RestoreRawEntry** <a id="restorerawentry"></a>
```go
err := db.RestoreRawEntry(context.Background(), "user", entry)
```
**Description:**  
Restores a raw entry for a key into the store. This can be used for state migration or recovery.

**Errors:**
- `ErrContextCanceled`

---

### 2.7 PubSub / Subscription Methods <a id="pubsub-methods"></a>

#### **Subscribe** <a id="subscribe"></a>
```go
ch := db.Subscribe("user")
```
**Description:**  
Creates a new subscription for events on the specified key. Returns a channel (`chan string`) to receive notifications (e.g., for operations like `SET`, `DELETE`, etc.).

---

#### **Unsubscribe** <a id="unsubscribe"></a>
```go
db.Unsubscribe("user", ch)
```
**Description:**  
Removes the specified subscription channel for the given key.

---

#### **ListSubscriptions** <a id="listsubscriptions"></a>
```go
keys := db.ListSubscriptions()
```
**Description:**  
Returns a list of keys (or topics) for which subscriptions currently exist.

---

#### **CloseAllSubscriptionsForKey** <a id="closeallsubscriptionsforkey"></a>
```go
db.CloseAllSubscriptionsForKey("user")
```
**Description:**  
Closes all subscriptions associated with the given key. Useful for resource cleanup when a key is deleted.

---

### 2.8 Accessor Methods <a id="accessor-methods"></a>

#### **Logger** <a id="logger"></a>
```go
loggerHandler := db.Logger()
```
**Description:**  
Returns the store's internal logger (an instance of `contracts.LoggerHandler`), allowing integration with external logging systems.

---

#### **Commands** <a id="commands"></a>
```go
commandsHandler := db.Commands()
```
**Description:**  
Returns the commands handler for executing advanced or composite commands. (Refer to the commands documentation for more details on the supported API.)

---

#### **Transaction** <a id="transaction"></a>
```go
tx := db.Transaction()
```
**Description:**  
Returns a transaction handler to group multiple operations atomically.

---

### 2.9 Shutdown <a id="shutdown"></a>

#### **Close**
```go
err := db.Close()
```
**Description:**  
Performs a graceful shutdown of the store:
- Stops the background expiration cleanup.
- Closes the PubSub system.
- Closes the logger.

**Errors:**
- May return an error if the logger fails to close properly.

---

## 3. Error Reference <a id="error-reference"></a>

The current implementation uses the following error types:

| Error                 | Description                                                                                              | Example Scenario                                    |
|-----------------------|----------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| **ErrContextCanceled**    | The operation was canceled via context (e.g., due to a timeout).                                      | Request timed out or was manually canceled.         |
| **ErrInvalidKey**         | The key is empty or improperly formatted.                                                           | Calling `Set("", value, ttl)` results in an error.    |
| **ErrInvalidTTL**         | The TTL (time-to-live) value is negative or otherwise invalid.                                      | Calling `Set("key", value, -10)`                     |
| **ErrKeyNotFound**        | The key does not exist or has expired.                                                              | Calling `Get("nonexistent_key")`                     |
| **ErrValueMismatch**      | CAS operation failed because the current value does not match the expected one.                       | Calling `SetCAS` with an incorrect expected value.   |
| **ErrKeyExists**          | A key already exists when using conditional operations (e.g., SETNX).                                 | Calling `SetNX` on an existing key.                  |
| **ErrInvalidType**        | The operation was performed on a key with a different data type (for example, trying LPush on a non-list).| Calling `LPush("user", ...)` when `user` is not a list.|
| **ErrEmptyList**          | An attempt was made to pop an element from an empty list.                                            | Calling `LPop` on an empty list.                     |
| **ErrInvalidValueType**   | The value type is not as expected (e.g., a counter operation was applied to a non-`int64` value).        | Calling `Incr` on a key containing a string.         |
| **ErrEmptyValues**        | No values or members were provided for an operation that requires them.                              | Calling `LPush("tasks")` without any arguments.      |

*Note:* Some errors have been consolidated. For example, a separate error for an expired key is now merged with `ErrKeyNotFound` for simplicity.

---

## 4. Best Practices <a id="best-practices"></a>

1. **Use Contexts for Timeouts and Cancellations**  
   Always pass a context with a proper timeout to avoid operations hanging indefinitely.  
   **Example:**
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
   defer cancel()
   
   value, err := db.Get(ctx, "user:123")
   if err != nil {
       log.Printf("Error fetching key: %v", err)
   } else {
       fmt.Println("Value:", value)
   }
   ```

2. **Check Return Values and Errors Explicitly**  
   Always verify error returns and validate result values. This is especially important for operations that might fail silently or return nil values.  
   **Example:**
   ```go
   err := db.SetCAS(context.Background(), "balance", "100", "150", 300)
   if err != nil {
       if errors.Is(err, hermes.ErrValueMismatch) {
           log.Println("CAS failed: value mismatch")
       } else {
           log.Printf("CAS failed: %v", err)
       }
   } else {
       fmt.Println("CAS successful")
   }
   ```

3. **Monitor Key Expiration and Cleanup**  
   Use TTL values where appropriate and implement periodic checks if your application relies on long-lived keys. Use Expire and Persist to manage key lifetimes.

4. **Use Sharding to Improve Concurrency**  
   In high-concurrency environments, increase the `ShardCount` to reduce lock contention.  
   **Example:**
   ```go
   db := hermes.NewStore(hermes.Config{
       ShardCount:      8,
       CleanupInterval: 5 * time.Second,
       EnableLogging:   true,
   })
   ```

5. **Leverage PubSub for Real-Time Notifications**  
   Use the subscription methods to listen for events on keys (such as updates, deletions, or renames) to drive reactive application logic.  
   **Example:**
   ```go
   ch := db.Subscribe("user")
   go func() {
       for msg := range ch {
           fmt.Println("Received update:", msg)
       }
   }()
   // To unsubscribe:
   db.Unsubscribe("user", ch)
   ```

6. **Group Related Operations in Transactions**  
   For complex operations that need to be executed atomically, use the Transaction API. This ensures all operations in a transaction are committed together.

7. **Configure Logging Appropriately**  
   Adjust `LogBufferSize` and log levels according to your application's performance and visibility requirements. Monitor log throughput and adjust if necessary.

---

This updated English documentation now includes all the new methods and provides an updated error reference, while preserving the original content. If you have further questions or need clarifications on any of the sections, please let me know!