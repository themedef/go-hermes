## In-Memory Data Store with Transactions and Pub/Sub

A high-performance thread-safe in-memory data store with advanced features for Go applications.

## Features

- 🧩 **ACID Transactions** with rollback support
- 📡 **Publish-Subscribe** messaging pattern
- ⏲ **Automatic Expiration** (TTL) for keys
- ⚡ **Atomic Operations** (CAS, INCR/DECR, LPUSH/RPUSH)
- 🔍 **Type-Safe Operations** for lists and counters
- 📊 **Built-in Logging** with configurable output
- 🚦 **Graceful Shutdown** support

## Installation

```bash
go get github.com/themedef/go-hermes
```
## Basic Usage

```go
import (
    "context"
    "time"
    "github.com/themedef/go-hermes"
)

func main() {
    // Initialize store
    cfg := hermes.Config{
        EnableLogging:   true,
        LogFile:         "data.log",
    }
    db := hermes.NewStore(cfg)

    // Basic operations
    ctx := context.Background()
    
    // Set key with TTL
    err := db.Set(ctx, "session:123", "user_data", 3600)
    
    // Get value
    val, exists, err := db.Get(ctx, "session:123")
    
    // Delete key
    deleted, err := db.Delete(ctx, "session:123")
}
```

## Transaction Management

```go
// Start transaction
tx := db.Transaction()
tx.Begin()

// Add operations
tx.Set(ctx, "account:1", 1000, 0)
tx.Incr(ctx, "account:1")
tx.Decr(ctx, "account:1")

// Commit transaction
if err := tx.Commit(); err != nil {
    // Handle error and rollback
    tx.Rollback()
}
```

## Pub/Sub System

```go
// Subscribe to channel
messages := db.Subscribe("updates")

// Publish message
db.Publish("updates", "System upgrade scheduled")

// Receive messages
go func() {
    for msg := range messages {
        fmt.Println("Received update:", msg)
    }
}()

// Unsubscribe when done
db.Unsubscribe("updates", messages)
```

## Advanced Operations

### Atomic Counters
```go
// Increment counter
newVal, success, err := db.Incr(ctx, "page_views")

// Decrement counter
newVal, success, err := db.Decr(ctx, "stock_count")
```

### List Operations
```go
// Push to list
db.LPush(ctx, "tasks", "urgent_task")
db.RPush(ctx, "tasks", "normal_task")

// Pop from list
task, success, err := db.LPop(ctx, "tasks")
task, success, err := db.RPop(ctx, "tasks")
```

### TTL Management
```go
// Update expiration
err := db.UpdateTTL(ctx, "temp_data", 300) // 5 minutes

// Check remaining TTL
entry, exists, _ := db.Get(ctx, "temp_data")
if !entry.Expiration.IsZero() {
    ttl := time.Until(entry.Expiration)
}
```

## Performance Characteristics

| Operation           | Time Complexity | Lock Type       | Notes                          |
|---------------------|-----------------|-----------------|--------------------------------|
| Get                 | O(1)            | RLock           | Read-only access               |
| Set/Delete          | O(1)            | Lock            | Full mutex lock                |
| List operations     | O(1)            | Lock            | Head/tail ops for lists        |
| Pub/Sub             | O(n)            | RLock           | n = number of subscribers      |
| TTL Updates         | O(1)            | Lock            | Time complexity for map access |

## Performance Notes

- **Thread Safety**: All operations protected by RWMutex
- **Memory Management**: Automatic expired key cleanup
- **Batched Operations**: Efficient flush implementation
- **Non-Blocking**: Pub/Sub uses buffered channels

## License

MIT License. See [LICENSE](LICENSE) for full text.

---

**Important Notes:**
- Always use context for timeout/cancellation control
- Close Pub/Sub channels when no longer needed
- Transactions must be explicitly committed
- Default channel buffer size: 100 messages
```