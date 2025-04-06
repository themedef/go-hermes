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
        - [DropAll](#dropall)
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

| Parameter           | Type             | Default | Description                                                                                         |
|---------------------|------------------|---------|-----------------------------------------------------------------------------------------------------|
| `ShardCount`        | `int`            | `1`     | Number of shards. Higher values improve parallelism under high concurrency.                        |
| `CleanupInterval`   | `time.Duration`  | `1s`    | Interval for background cleanup of expired keys. Set to `0` to disable cleanup.                   |
| `EnableLogging`     | `bool`           | `false` | Enables logging of operations.                                                                      |
| `LogFile`           | `string`         | `""`    | File path for logs. If empty, logs are written to stdout.                                           |
| `LogBufferSize`     | `int`            | `1000`  | Size of the asynchronous log buffer.                                                              |
| `MinLevel`          | `logger.LogLevel`| `DEBUG` | Minimum log level. Levels: `DEBUG`, `INFO`, `WARN`, `ERROR`.                                         |
| `PubSubBufferSize`  | `int`            | `10000` | Buffer size for PubSub channels. If not provided or ≤ 0, defaults to 10000.                           |

---

## 2. Core Methods <a id="core-methods"></a>

### **Key-Value Operations** <a id="key-value-operations"></a>

#### **Set** <a id="set"></a>
```go
err := db.Set(context.Background(), "user", "Test", 60) // TTL = 60 seconds
```
**Description**:  
Sets a key to the given value with an optional TTL.

**Errors**:
- `ErrContextCanceled`, `ErrInvalidKey`, `ErrInvalidTTL`.

---

#### **Get** <a id="get"></a>
```go
value, err := db.Get(context.Background(), "user")
```
**Description**:  
Retrieves the value of the specified key.

**Errors**:
- `ErrContextCanceled`, `ErrKeyNotFound`, `ErrKeyExpired`.

---

#### **SetNX** <a id="setnx"></a>
```go
success, err := db.SetNX(context.Background(), "user", "Test", 60)
```
**Description**:  
Sets a key only if it does not already exist.

**Errors**:
- `ErrContextCanceled`, `ErrInvalidKey`, `ErrInvalidTTL`, `ErrKeyExists`.

---

#### **SetXX** <a id="setxx"></a>
```go
success, err := db.SetXX(context.Background(), "user", "Test", 60)
```
**Description**:  
Updates a key only if it already exists.

**Errors**:
- `ErrContextCanceled`, `ErrInvalidKey`, `ErrInvalidTTL`, `ErrKeyNotFound`.

---

#### **SetCAS** <a id="setcas"></a>
```go
err := db.SetCAS(context.Background(), "user", "old_value", "new_value", 60)
```
**Description**:  
Updates a key only if its current value matches `old_value`.

**Errors**:
- `ErrContextCanceled`, `ErrKeyNotFound`, `ErrValueMismatch`, `ErrInvalidTTL`.

---

#### **GetSet** <a id="getset"></a>
```go
oldVal, err := db.GetSet(context.Background(), "key", "new_value", 120)
```
**Description**:  
Atomically sets a new value and returns the previous value.  
If the key does not exist, sets it and returns `nil`.

**Errors**:
- `ErrContextCanceled`, `ErrInvalidTTL`.

---

### **Atomic Counters** <a id="atomic-counters"></a>

#### **Incr** <a id="increment"></a>
```go
newValue, err := db.Incr(context.Background(), "counter")
```
**Description**:  
Increments the integer value by 1. Initializes key to `1` if missing.

**Errors**:
- `ErrContextCanceled`, `ErrInvalidValueType`.

---

#### **Decr** <a id="decrement"></a>
```go
newValue, err := db.Decr(context.Background(), "counter")
```
**Description**:  
Decrements the integer value by 1. Initializes key to `-1` if missing.

**Errors**:
- `ErrContextCanceled`, `ErrInvalidValueType`.

---

#### **IncrBy** <a id="incrby"></a>
```go
newValue, err := db.IncrBy(context.Background(), "counter", 10)
```
**Description**:  
Increments the integer value by a specified amount.  
If the key does not exist, it is created with the increment value.

**Errors**:
- `ErrContextCanceled`, `ErrInvalidValueType`.

---

#### **DecrBy** <a id="decrby"></a>
```go
newValue, err := db.DecrBy(context.Background(), "counter", 5)
```
**Description**:  
Decrements the integer value by a specified amount.  
If the key does not exist, it is created with the negative of the decrement value.

**Errors**:
- `ErrContextCanceled`, `ErrInvalidValueType`.

---

### **List Operations** <a id="list-operations"></a>

#### **LPush** <a id="lpush"></a>
```go
err := db.LPush(context.Background(), "tasks", "task1", "task2")
```
**Description**:  
Inserts one or more elements at the beginning (left) of the list.  
Creates the list if it does not exist.

**Errors**:
- `ErrContextCanceled`, `ErrInvalidKey`, `ErrInvalidType`.

---

#### **RPush** <a id="rpush"></a>
```go
err := db.RPush(context.Background(), "tasks", "task3", "task4")
```
**Description**:  
Inserts one or more elements at the end (right) of the list.  
Creates the list if it does not exist.

**Errors**:
- `ErrContextCanceled`, `ErrInvalidType`.

---

#### **LPop** <a id="lpop"></a>
```go
value, err := db.LPop(context.Background(), "tasks")
```
**Description**:  
Removes and returns the first element of the list.

**Errors**:
- `ErrContextCanceled`, `ErrKeyNotFound`, `ErrInvalidType`, `ErrEmptyList`.

---

#### **RPop** <a id="rpop"></a>
```go
value, err := db.RPop(context.Background(), "tasks")
```
**Description**:  
Removes and returns the last element of the list.

**Errors**:
- `ErrContextCanceled`, `ErrKeyNotFound`, `ErrInvalidType`, `ErrEmptyList`.

---

#### **LRange** <a id="lrange"></a>
```go
values, err := db.LRange(context.Background(), "mylist", 0, -1)
```
**Description**:  
Returns a range of list elements. Supports negative indices for counting from the end.

**Errors**:
- `ErrContextCanceled`, `ErrEmptyList`, `ErrInvalidType`.

---

#### **LLen** <a id="llen"></a>
```go
length, err := db.LLen(context.Background(), "tasks")
```
**Description**:  
Returns the length of the list.

**Errors**:
- `ErrContextCanceled`, `ErrKeyNotFound`, `ErrInvalidType`.

---

### **Hash Operations** <a id="hash-operations"></a>

#### **HSet** <a id="hset"></a>
```go
err := db.HSet(context.Background(), "user", "name", "Test", 0)
```
**Description**:  
Sets a hash field. Creates the hash if needed.  
TTL remains unchanged if set to `0`.

**Errors**:
- `ErrContextCanceled`, `ErrInvalidType`, `ErrInvalidTTL`.

---

#### **HGet** <a id="hget"></a>
```go
value, err := db.HGet(context.Background(), "user", "name")
```
**Description**:  
Retrieves the value for a given hash field.

**Errors**:
- `ErrContextCanceled`, `ErrKeyNotFound`, `ErrInvalidType`.

---

#### **HDel** <a id="hdel"></a>
```go
err := db.HDel(context.Background(), "user", "name")
```
**Description**:  
Deletes a specific hash field. Deletes the key if the hash becomes empty.

**Errors**:
- `ErrContextCanceled`, `ErrKeyNotFound`, `ErrInvalidType`.

---

#### **HGetAll** <a id="hgetall"></a>
```go
fields, err := db.HGetAll(context.Background(), "user")
```
**Description**:  
Returns all fields and values in a hash.

**Errors**:
- `ErrContextCanceled`, `ErrKeyNotFound`, `ErrInvalidType`.

---

#### **HExists** <a id="hexists"></a>
```go
exists, err := db.HExists(context.Background(), "user", "email")
```
**Description**:  
Checks for the existence of a field in a hash.

**Errors**:
- `ErrContextCanceled`, `ErrKeyNotFound`, `ErrInvalidType`.

---

#### **HLen** <a id="hlen"></a>
```go
length, err := db.HLen(context.Background(), "user")
```
**Description**:  
Returns the number of fields in the hash.

**Errors**:
- `ErrContextCanceled`, `ErrKeyNotFound`, `ErrInvalidType`.

---

### **Utility Methods** <a id="utility-methods"></a>

#### **Exists** <a id="exists"></a>
```go
exists, err := db.Exists(context.Background(), "user")
```
**Description**:  
Checks whether the key exists and is not expired.

**Errors**:
- `ErrContextCanceled`.

---

#### **Expire** <a id="expire"></a>
```go
success, err := db.Expire(context.Background(), "user", 60)
```
**Description**:  
Sets a new TTL for an existing key.

**Response**:  
Returns a boolean indicating success.

**Errors**:
- `ErrContextCanceled`, `ErrInvalidTTL`, `ErrKeyNotFound`.

---

#### **Persist** <a id="persist"></a>
```go
success, err := db.Persist(context.Background(), "user")
```
**Description**:  
Removes the TTL from a key, making it persistent.

**Response**:  
Returns a boolean indicating success.

**Errors**:
- `ErrContextCanceled`, `ErrKeyNotFound`.

---

#### **Type** <a id="type"></a>
```go
dataType, err := db.Type(context.Background(), "user")
```
**Description**:  
Returns the data type of the key (e.g., `String`, `List`, or `Hash`).

**Errors**:
- `ErrContextCanceled`, `ErrKeyNotFound`.

---

#### **GetWithDetails** <a id="getwithdetails"></a>
```go
value, ttl, err := db.GetWithDetails(context.Background(), "user")
```
**Description**:  
Returns the value along with its TTL (in seconds). TTL is `-1` if the key never expires.

**Errors**:
- `ErrContextCanceled`, `ErrKeyNotFound`, `ErrKeyExpired`.

---

#### **Rename** <a id="rename"></a>
```go
err := db.Rename(context.Background(), "oldKey", "newKey")
```
**Description**:  
Renames an existing key. Fails if the old key doesn’t exist or the new key already exists.

**Errors**:
- `ErrContextCanceled`, `ErrInvalidKey`, `ErrKeyNotFound`, `ErrKeyExists`.

---

#### **FindByValue** <a id="findbyvalue"></a>
```go
keys, err := db.FindByValue(context.Background(), "Test")
```
**Description**:  
Finds all keys that have the specified value.

**Errors**:
- `ErrContextCanceled`, `ErrKeyNotFound`.

---

#### **Delete** <a id="delete"></a>
```go
err := db.Delete(context.Background(), "user")
```
**Description**:  
Deletes a key from the store.

**Errors**:
- `ErrKeyNotFound`.

---

#### **DropAll** <a id="dropall"></a>
```go
err := db.DropAll(context.Background())
```
**Description**:  
Deletes all keys from the store.

**Errors**:
- `ErrContextCanceled`.

---

## 3. Error Reference <a id="error-reference"></a>

| Error                | Description                                      | Example Scenario                                  |
|----------------------|--------------------------------------------------|---------------------------------------------------|
| `ErrKeyNotFound`     | Key does not exist or has expired                | Calling `Get` on a deleted or expired key         |
| `ErrInvalidKey`      | Key is empty or improperly formatted             | Calling `Set("", "value", 0)`                     |
| `ErrInvalidType`     | Operation on a key with an unexpected data type    | Using `LPush` on a key that is not a list          |
| `ErrEmptyList`       | List is empty during a pop operation              | Calling `LPop` on an empty list                    |
| `ErrInvalidTTL`      | TTL value is negative                             | Calling `Set("key", "value", -10)`                 |
| `ErrValueMismatch`   | CAS operation failed: current value ≠ old value   | `SetCAS` with an incorrect `old_value`             |
| `ErrKeyExists`       | Key already exists when using SETNX               | Calling `SetNX` on an existing key                 |
| `ErrContextCanceled` | Operation canceled via context                    | Request timeout or manual cancellation             |

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
   **Tip:** Use different timeouts for critical operations versus batch processing.

2. **Check Return Values and Errors Explicitly**  
   Always verify error returns and validate result values. This is especially important for operations that might fail silently or return nil values.  
   **Example:**
   ```go
   err := db.SetCAS(context.Background(), "account:balance", "100", "150", 300)
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
   **Tip:** Wrap errors or use `errors.Is()` to distinguish between error types.

3. **Monitor Key Expiration and Cleanup**  
   Use TTL values where appropriate and implement periodic checks if your application relies on long-lived keys.  
   **Example:**
   ```go
   err := db.Set(context.Background(), "session:abc", "data", 60)
   if err != nil {
       log.Printf("Error setting key: %v", err)
   }
   
   exists, err := db.Exists(context.Background(), "session:abc")
   if err != nil {
       log.Printf("Error checking key: %v", err)
   } else if !exists {
       log.Println("Session expired or does not exist; refresh or reinitialize session")
   }
   ```
   **Tip:** Consider client-side caching or periodic refresh logic if key expiration is critical.

4. **Use Shards to Improve Concurrency**  
   For high-concurrency environments, configure the store with a higher `ShardCount` to reduce lock contention.  
   **Example:**
   ```go
   db := hermes.NewStore(hermes.Config{
       ShardCount:      8,
       CleanupInterval: 5 * time.Second,
       EnableLogging:   true,
   })

   go func() {
       err := db.Set(context.Background(), "user:1", "Test", 60)
       if err != nil {
           log.Println("Error in goroutine 1:", err)
       }
   }()
   go func() {
       err := db.Set(context.Background(), "user:2", "Bob", 60)
       if err != nil {
           log.Println("Error in goroutine 2:", err)
       }
   }()
   ```
   **Tip:** Test different shard counts under load to determine the optimal configuration.

5. **Synchronous vs. Asynchronous Logging**  
   Adjust `LogBufferSize` and logging options based on your performance needs. Under high load, asynchronous logging (with a larger buffer) may drop messages if the buffer overflows.  
   **Example:**
   ```go
   db := hermes.NewStore(hermes.Config{
       EnableLogging:   true,
       LogFile:         "/var/log/hermes.log",
       LogBufferSize:   5000,
       MinLevel:        hermes.WARN,
   })
   ```
   **Tip:** Monitor log metrics; if you notice frequent drops, increase the buffer or adjust the log level.

6. **Working with Structured Data**  
   When dealing with complex data structures, consider the following approaches:
    - **Using Hashes:**  
      Use hash operations (`HSet`, `HGet`, `HGetAll`) to store structured data such as user profiles.
      ```go
      err := db.HSet(context.Background(), "user", "name", "Test", 0)
      err = db.HSet(context.Background(), "user", "email", "test@example.com", 0)
      fields, err := db.HGetAll(context.Background(), "user")
      ```
    - **Using JSON Encoding:**  
      Marshal complex structures to JSON and store them as strings using `Set` and retrieve using `Get`.
      ```go
      type User struct {
          Name  string `json:"name"`
          Email string `json:"email"`
      }
      user := User{Name: "Test", Email: "test@example.com"}
      jsonData, err := json.Marshal(user)
      err = db.Set(context.Background(), "user:101", string(jsonData), 0)
      data, err := db.Get(context.Background(), "user:101")
      var retrievedUser User
      err = json.Unmarshal([]byte(data.(string)), &retrievedUser)
      ```
   **Tip:** Choose the method that best fits your application's complexity and performance needs.

---
