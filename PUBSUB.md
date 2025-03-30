# Pub/Sub Documentation

---

## Table of Contents

1. [Accessing PubSub](#accessing-pubsub)
2. [Core Methods](#core-methods)
    - [Subscribe](#subscribe)
    - [Unsubscribe](#unsubscribe)
    - [ListSubscribers](#listsubscribers)
    - [UnsubscribeAllForKey](#unsubscribeallforkey)
3. [Example Usage](#example-usage)
4. [Best Practices](#best-practices)

---
## 1. Accessing PubSub <a id="accessing-pubsub"></a>

The Pub/Sub package provides an in-memory publish-subscribe system that facilitates decoupled communication between different parts of your application. Key features include:

- **In-memory Operation:** Messages are stored only in memory and are not persisted, ensuring high performance.
- **Buffered Channels:** Subscribers receive messages via buffered channels (default buffer size is 10,000) to avoid blocking during heavy loads.
- **Thread-Safe:** Internal synchronization is handled using read/write mutexes to safely manage concurrent access.
- **Flexible Subscriptions:** Easily subscribe, unsubscribe, or list active subscriptions by topic.

---

## 2. Core Methods

### **Subscribe** <a id="subscribe"></a>

```go
ch := ps.Subscribe("key")
```

- Subscribes to a key/topic.
- Returns a buffered channel (`chan string`) that receives published messages.
- Buffer size is 10,000 by default.

---

### **Unsubscribe** <a id="unsubscribe"></a>

```go
ps.Unsubscribe("key", ch)
```

- Removes a specific subscription.
- Automatically closes the channel.

---


### **ListSubscribers** <a id="listsubscribers"></a>

```go
keys := ps.ListSubscribers()
```

- Returns a slice of all keys with at least one active subscriber.

---

### **UnsubscribeAllForKey** <a id="unsubscribeallforkey"></a>

```go
ps.UnsubscribeAllForKey("key")
```

- Unsubscribes and closes **all** channels for the given key.

---


## 3. Example Usage <a id="example-usage"></a>

```go
ch := db.Subscribe("notifications:user:123")

go func() {
    for msg := range ch {
        fmt.Println("Received:", msg)
    }
}()

ps.Publish("notifications:user:123", "You have a new message!")

// Clean up when done
defer ps.Unsubscribe("notifications:user:123", ch)
```

---

## 4. Best Practices <a id="best-practices"></a>

- **Always unsubscribe channels when done**:

  ```go
  defer db.Unsubscribe("key", ch)
  ```

- **Guard against dropped messages due to buffer overflow**:

  ```go
  select {
  case ch <- msg:
  default:
      log.Println("Dropped message")
  }
  ```

- **Close all subscriptions for a key if it's no longer relevant**:

  ```go
  db.UnsubscribeAllForKey("deprecated:key")
  ```

- **Use context cancellation for dynamic lifetime control**:

  ```go
  ctx, cancel := context.WithCancel(context.Background())
  go func() {
      <-ctx.Done()
      db.Unsubscribe("key", ch)
  }()
  ```