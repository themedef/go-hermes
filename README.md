## In-Memory Data Store with Transactions and Pub/Sub

A high-performance thread-safe in-memory data store with advanced features for Go applications.

## Features

- 🧩 **ACID Transactions** with rollback support
- 📡 **Publish-Subscribe** messaging pattern
- ⏲ **Automatic Expiration** (TTL) for keys
- ⚡ **Atomic Operations** (CAS, INCR/DECR, LPUSH/RPUSH)
- 🔍 **Type-Safe Operations** for lists and counters
- 📊 **Built-in Logging** with configurable output

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
    val, err := db.Get(ctx, "session:123")
    
    // Delete key
    deleted, err := db.Delete(ctx, "session:123")

    // Ensure the store is closed properly on application exit
    defer func() {
        if err := db.Close(); err != nil {
            log.Printf("Error while closing Hermes store: %v\n", err)
        }
    }()
}
```
[STORE Documentation](STORE.md)

## Transaction Management

```go
// Start transaction
tx := db.Transaction()

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
[TRANSACTION Documentation](TRANSACTION.md)

## Pub/Sub System

```go
// Subscribe to channel
messages := db.Subscribe("updates")

// Receive messages
go func() {
    for msg := range messages {
        fmt.Println("Received update:", msg)
    }
}()

// Unsubscribe when done
db.Unsubscribe("updates", messages)
```
[PUBSUB Documentation](PUBSUB.md)


##  Benchmarks

Performance tested on a single-core execution using Go's `testing` and `runtime` packages.

| Operation | Ops Count | Time (sec) | RPS (req/sec) |
|-----------|-----------|------------|---------------|
| `Get`     | 100,000   | 0.15       | **3,991,454** |
| `Set`     | 100,000   | 0.24       | **1,241,792** |

> ️ Test environment: **1 CPU core**, TTL disabled, logging disabled.
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


**Important Notes:**
- Always use context for timeout/cancellation control
- Close Pub/Sub channels when no longer needed
- Transactions must be explicitly committed
- Default channel buffer size: 10000 messages

## License

MIT License. See [LICENSE](LICENSE) for full text.

---