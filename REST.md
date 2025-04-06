# REST Documentation

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
        - [Incr](#incr)
        - [Decr](#decr)
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
3. [Global Error Codes](#global-errors)
4. [Example Usage](#example-usage)

---

## 1. Initialization <a id="initialization"></a>

To initialize and run the server, call:
```go
handler := NewAPIHandler(ctx, db)
handler.RunServer("8080", "api", /* optional middlewares */)
```
- **Port**: The port on which the HTTP server listens (e.g., `8080`).
- **Prefix**: A path prefix (e.g., `api`), so that endpoints are served at `http://localhost:8080/api/...`.

*Note: Initialization does not have custom error responses beyond standard server errors (e.g., 500 Internal Server Error).*

---

## 2. Core Methods <a id="core-methods"></a>

Each endpoint below includes an **Errors** section listing the potential HTTP errors and their descriptions.

### Key-Value Operations

#### Set
**Endpoint**: `POST /set`  
**Description**: Sets a key to a given value, optionally with a TTL (in seconds).  
**Request Body**:
```json
{
  "key": "myKey",
  "value": "someValue",
  "ttl": 60
}
```
**Response** (`200 OK` on success):
```json
{
  "message": "Set OK",
  "key": "myKey",
  "ttl": 60
}
```
**Errors:**
- **400 Bad Request**: If the request body is missing required fields or is improperly formatted.
- **500 Internal Server Error**: In case of any unexpected errors.

---

#### Get
**Endpoint**: `GET /get?key=<keyName>`  
**Description**: Retrieves the value of the specified key.  
**Response** (`200 OK`):
```json
{
  "key": "myKey",
  "value": "someValue"
}
```
**Errors:**
- **404 Not Found**: If the key does not exist or is expired.
- **400 Bad Request**: If the query parameter is missing or invalid.
- **500 Internal Server Error**: For unexpected failures.

---

#### SetNX
**Endpoint**: `POST /setnx`  
**Description**: Sets a key only if it does **not** exist.  
**Request Body**:
```json
{
  "key": "nxKey",
  "value": "nxValue",
  "ttl": 30
}
```
**Response** (`200 OK`):
```json
{
  "success": true,
  "key": "nxKey",
  "ttl": 30
}
```
*Note: The response will include `"success": false` if the key already exists (logical failure, not an HTTP error).*  
**Errors:**
- **400 Bad Request**: If the request body is invalid.
- **500 Internal Server Error**: On unexpected internal errors.

---

#### SetXX
**Endpoint**: `POST /setxx`  
**Description**: Sets a key only if it **already exists**.  
**Request Body**:
```json
{
  "key": "xxKey",
  "value": "newVal",
  "ttl": 120
}
```
**Response** (`200 OK`):
```json
{
  "success": true,
  "key": "xxKey",
  "ttl": 120
}
```
*Note: The response returns `"success": false` if the key does not exist (logical failure in a 200 response).*  
**Errors:**
- **400 Bad Request**: If the request body is missing or invalid.
- **500 Internal Server Error**: For any unexpected error.

---

#### SetCAS
**Endpoint**: `POST /setcas`  
**Description**: Compare-And-Set. Updates a key’s value if its current value equals `old_value`.  
**Request Body**:
```json
{
  "key": "casKey",
  "old_value": "oldVal",
  "new_value": "newVal",
  "ttl": 60
}
```
**Response** (`200 OK`):
```json
{
  "success": true,
  "key": "casKey"
}
```
**Errors:**
- **409 Conflict**: If the current key value does not match the provided `old_value`.
- **400 Bad Request**: If the request body is invalid.
- **500 Internal Server Error**: For unexpected failures.

---

#### GetSet
**Endpoint**: `POST /getset`  
**Description**: Atomically sets a new value for a key and returns its old value.  
**Request Body**:
```json
{
  "key": "myKey",
  "new_value": "XYZ",
  "ttl": 0
}
```
**Response**:
```json
{
  "key": "myKey",
  "oldValue": "someValue",
  "newValue": "XYZ"
}
```
*Note: If the key was expired or non-existent, `oldValue` will be `null` or absent.*  
**Errors:**
- **400 Bad Request**: If the request body is invalid.
- **500 Internal Server Error**: For any unexpected errors.

---

### Atomic Counters

#### Incr
**Endpoint**: `POST /incr`  
**Description**: Increments a key’s numeric value by 1.
- If the key doesn’t exist, it is created with value `1`.
- If the key is not an integer, returns an error.

**Request Body**:
```json
{
  "key": "counterKey"
}
```
**Response**:
```json
{
  "key": "counterKey",
  "value": 2
}
```
**Errors:**
- **400 Bad Request**: If the key exists but its value is not an integer.
- **500 Internal Server Error**: For unexpected failures.

---

#### Decr
**Endpoint**: `POST /decr`  
**Description**: Decrements a key’s numeric value by 1.
- If the key doesn’t exist, it is created with value `-1`.
- If the key is not an integer, returns an error.

**Request Body**:
```json
{
  "key": "counterKey"
}
```
**Response**:
```json
{
  "key": "counterKey",
  "value": 0
}
```
**Errors:**
- **400 Bad Request**: If the key exists but its value is not an integer.
- **500 Internal Server Error**: For unexpected errors.

---

#### IncrBy
**Endpoint**: `POST /incrby`  
**Description**: Increments a key’s numeric value by a specified amount.
- If the key doesn’t exist, it is created with the increment value.
- If the key exists but is not an integer, returns an error.

**Request Body**:
```json
{
  "key": "counterKey",
  "increment": 10
}
```
**Response**:
```json
{
  "key": "counterKey",
  "value": 10
}
```
**Errors:**
- **400 Bad Request**: If the key exists but its value is not an integer.
- **500 Internal Server Error**: For unexpected failures.

---

#### DecrBy
**Endpoint**: `POST /decrby`  
**Description**: Decrements a key’s numeric value by a specified amount.
- If the key doesn’t exist, it is created with the negative of the decrement value.
- If the key exists but is not an integer, returns an error.

**Request Body**:
```json
{
  "key": "counterKey",
  "decrement": 5
}
```
**Response**:
```json
{
  "key": "counterKey",
  "value": -5
}
```
**Errors:**
- **400 Bad Request**: If the key exists but its value is not an integer.
- **500 Internal Server Error**: For unexpected failures.

---

### List Operations

#### LPush
**Endpoint**: `POST /lpush`  
**Description**: Pushes one or multiple values to the left (head) of a list.  
**Request Body**:
```json
{
  "key": "myList",
  "values": ["A", "B", 42]
}
```
**Response**:
```json
{
  "message": "LPUSH success",
  "key": "myList",
  "count": 3
}
```
**Errors:**
- **400 Bad Request**: If the request body is missing required fields or is malformed.
- **500 Internal Server Error**: For unexpected internal errors.

---

#### RPush
**Endpoint**: `POST /rpush`  
**Description**: Appends one or multiple values to the right (tail) of a list.  
**Request Body**:
```json
{
  "key": "myList",
  "values": ["X", {"obj": true}, 3.14]
}
```
**Response**:
```json
{
  "message": "RPUSH success",
  "key": "myList",
  "count": 3
}
```
**Errors:**
- **404 Not Found**: If the target list does not exist.
- **400 Bad Request**: If the request body is invalid.
- **500 Internal Server Error**: For unexpected errors.

---

#### LPop
**Endpoint**: `POST /lpop`  
**Description**: Pops (removes and returns) the first element of a list.  
**Request Body**:
```json
{
  "key": "myList"
}
```
**Response**:
```json
{
  "message": "LPOP success",
  "key": "myList",
  "value": "A"
}
```
**Errors:**
- **404 Not Found**: If the list is empty or does not exist.
- **400 Bad Request**: If the request body is missing or invalid.
- **500 Internal Server Error**: For unexpected failures.

---

#### RPop
**Endpoint**: `POST /rpop`  
**Description**: Pops (removes and returns) the last element of a list.  
**Request Body**:
```json
{
  "key": "myList"
}
```
**Response**:
```json
{
  "message": "RPOP success",
  "key": "myList",
  "value": "B"
}
```
**Errors:**
- **404 Not Found**: If the list is empty or missing.
- **400 Bad Request**: If the request body is invalid.
- **500 Internal Server Error**: For unexpected errors.

---

#### LRange
**Endpoint**: `POST /lrange`  
**Description**: Retrieves a range of elements from the list at `key`.  
**Request Body**:
```json
{
  "key": "myList",
  "start": 0,
  "end": -1
}
```
*Note: `start`/`end` are zero-based indices; negative indices read from the end (`-1` = last item).*

**Response**:
```json
{
  "key": "myList",
  "start": 0,
  "end": -1,
  "result": ["A", "B"]
}
```
**Errors:**
- **400 Bad Request**: If the indices are invalid or the request body is malformed.
- **500 Internal Server Error**: For unexpected failures.

---

#### LLen
**Endpoint**: `GET /llen?key=<listKey>`  
**Description**: Returns the length of the list.  
**Response**:
```json
{
  "key": "myList",
  "length": 2
}
```
**Errors:**
- **404 Not Found**: If the list does not exist.
- **500 Internal Server Error**: For unexpected errors.

---

### Hash Operations

#### HSet
**Endpoint**: `POST /hset`  
**Description**: Sets a `field` in a hash to a specified `value`. Optionally sets a TTL if creating a new hash.  
**Request Body**:
```json
{
  "key": "user:1",
  "field": "name",
  "value": "Test",
  "ttl": 60
}
```
**Response**:
```json
{
  "message": "HSET success",
  "key": "user:1",
  "field": "name"
}
```
**Errors:**
- **400 Bad Request**: If the request body is missing required fields or is malformed.
- **500 Internal Server Error**: For any unexpected errors.

---

#### HGet
**Endpoint**: `GET /hget?key=<hashKey>&field=<fieldName>`  
**Description**: Retrieves the value for a specific `field` in a hash.  
**Response**:
```json
{
  "key": "user:1",
  "field": "name",
  "value": "Test"
}
```
**Errors:**
- **404 Not Found**: If the key or field does not exist.
- **400 Bad Request**: If required query parameters are missing or invalid.
- **500 Internal Server Error**: For unexpected errors.

---

#### HDel
**Endpoint**: `DELETE /hdel`  
**Description**: Deletes a field from a hash. If the hash becomes empty, the entire key is removed.  
**Request Body**:
```json
{
  "key": "user:1",
  "field": "email"
}
```
**Response**:
```json
{
  "message": "HDEL success",
  "key": "user:1",
  "field": "email"
}
```
**Errors:**
- **404 Not Found**: If the key or field does not exist.
- **400 Bad Request**: If the request body is invalid.
- **500 Internal Server Error**: For unexpected errors.

---

#### HGetAll
**Endpoint**: `GET /hgetall?key=<hashKey>`  
**Description**: Retrieves all fields and values from a hash.  
**Response**:
```json
{
  "key": "user:1",
  "fields": {
    "name": "Test",
    "email": "test@example.com"
  }
}
```
**Errors:**
- **404 Not Found**: If the hash does not exist.
- **500 Internal Server Error**: For unexpected errors.

---

#### HExists
**Endpoint**: `GET /hexists?key=<hashKey>&field=<fieldName>`  
**Description**: Checks whether a specific field exists in a hash.  
**Response**:
```json
{
  "key": "user:1",
  "field": "name",
  "exists": true
}
```
**Errors:**
- **404 Not Found**: If the hash does not exist.
- **400 Bad Request**: If query parameters are missing or invalid.
- **500 Internal Server Error**: For unexpected failures.

---

#### HLen
**Endpoint**: `GET /hlen?key=<hashKey>`  
**Description**: Returns the number of fields in a hash.  
**Response**:
```json
{
  "key": "user:1",
  "length": 2
}
```
**Errors:**
- **404 Not Found**: If the hash does not exist.
- **500 Internal Server Error**: For unexpected errors.

---

### Utility Methods

#### Exists
**Endpoint**: `GET /exists?key=<keyName>`  
**Description**: Checks whether a given key exists (and is not expired).  
**Response**:
```json
{
  "key": "myKey",
  "exists": true
}
```
**Errors:**
- **500 Internal Server Error**: For unexpected failures.

---

#### Expire
**Endpoint**: `POST /expire`  
**Description**: Sets a TTL for a key.  
**Request Body**:
```json
{
  "key": "myKey",
  "ttl": 60
}
```
**Response**:
```json
{
  "key": "myKey",
  "ttl": 60,
  "success": true
}
```
**Errors:**
- **404 Not Found**: If the key does not exist.
- **400 Bad Request**: If the request body is invalid.
- **500 Internal Server Error**: For unexpected errors.

---

#### Persist
**Endpoint**: `POST /persist`  
**Description**: Removes the TTL from a key, making it persistent.  
**Request Body**:
```json
{
  "key": "myKey"
}
```
**Response**:
```json
{
  "key": "myKey",
  "success": true
}
```
**Errors:**
- **404 Not Found**: If the key does not exist or is already persistent.
- **400 Bad Request**: If the request body is invalid.
- **500 Internal Server Error**: For unexpected errors.

---

#### Type
**Endpoint**: `GET /type?key=<keyName>`  
**Description**: Returns the internal data type of the key (e.g., `String`, `List`, `Hash`, etc.).  
**Response**:
```json
{
  "key": "myKey",
  "type": 0
}
```
*Note: In the default store, `0` = String, `1` = List, `2` = Hash.*  
**Errors:**
- **404 Not Found**: If the key does not exist.
- **500 Internal Server Error**: For unexpected errors.

---

#### GetWithDetails
**Endpoint**: `GET /details?key=<keyName>`  
**Description**: Retrieves the value of a key along with its current TTL.  
**Response**:
```json
{
  "key": "myKey",
  "value": "someValue",
  "ttl": 42
}
```
*Note: If the key never expires, `ttl` will be `-1`.*  
**Errors:**
- **404 Not Found**: If the key does not exist.
- **500 Internal Server Error**: For unexpected errors.

---

#### Rename
**Endpoint**: `POST /rename`  
**Description**: Renames a key to a new name, provided that the old key exists and the new key does not.  
**Request Body**:
```json
{
  "old_key": "oldName",
  "new_key": "newName"
}
```
**Response**:
```json
{
  "message": "Rename success",
  "oldKey": "oldName",
  "newKey": "newName"
}
```
**Errors:**
- **409 Conflict**: If the `new_key` already exists.
- **404 Not Found**: If the `old_key` does not exist.
- **400 Bad Request**: If the request body is invalid.
- **500 Internal Server Error**: For unexpected errors.

---

#### FindByValue
**Endpoint**: `POST /find`  
**Description**: Returns all keys whose value matches the provided input.  
**Request Body**:
```json
{
  "value": "someValue"
}
```
**Response**:
```json
{
  "value": "someValue",
  "keys": ["key1", "key2"]
}
```
**Errors:**
- **404 Not Found**: If no matching keys are found.
- **400 Bad Request**: If the request body is invalid.
- **500 Internal Server Error**: For unexpected errors.

---

#### Delete
**Endpoint**: `POST /delete`  
**Description**: Deletes the specified key.  
**Request Body**:
```json
{
  "key": "myKey"
}
```
**Response**:
```json
{
  "message": "Deleted",
  "key": "myKey"
}
```
**Errors:**
- **404 Not Found**: If the key does not exist.
- **400 Bad Request**: If the request body is missing or invalid.
- **500 Internal Server Error**: For unexpected errors.

---

#### DropAll
**Endpoint**: `POST /dropall`  
**Description**: Removes **all keys** from the store. Use with caution!  
**Response**:
```json
{
  "message": "All keys dropped"
}
```
**Errors:**
- **500 Internal Server Error**: For unexpected errors during this dangerous operation.

---

## 3. Global Error Codes <a id="global-errors"></a>

In addition to the endpoint-specific errors above, the following HTTP error codes are commonly used throughout the API:

- **400 Bad Request**  
  Returned when the request is malformed, missing required fields, or contains invalid parameters.

- **404 Not Found**  
  Returned when a key, list element, or hash field does not exist or has expired.

- **409 Conflict**  
  Returned when there is a logical conflict (e.g., in `SetCAS` or `Rename` operations).

- **500 Internal Server Error**  
  Returned for any unexpected or internal errors.

Each error response may include a JSON body with an `"error"` field containing a plain text message (e.g., `{ "error": "Key not found" }`).

---

## 4. Example Usage <a id="example-usage"></a>

1. **Setting a key**:
   ```bash
   curl -X POST -H "Content-Type: application/json" \
   -d '{"key":"hello","value":"world","ttl":30}' \
   http://localhost:8080/api/set
   ```
   **Response**:
   ```json
   {
     "message": "Set OK",
     "key": "hello",
     "ttl": 30
   }
   ```

2. **Getting a key**:
   ```bash
   curl -X GET "http://localhost:8080/api/get?key=hello"
   ```
   **Possible Response**:
   ```json
   {
     "key": "hello",
     "value": "world"
   }
   ```

3. **Incrementing a counter**:
   ```bash
   curl -X POST -H "Content-Type: application/json" \
   -d '{"key":"counter"}' \
   http://localhost:8080/api/incr
   ```
   **Response**:
   ```json
   {
     "key": "counter",
     "value": 1
   }
   ```

4. **Incrementing by a specific value**:
   ```bash
   curl -X POST -H "Content-Type: application/json" \
   -d '{"key":"counter", "increment": 10}' \
   http://localhost:8080/api/incrby
   ```
   **Response**:
   ```json
   {
     "key": "counter",
     "value": 10
   }
   ```

5. **Decrementing by a specific value**:
   ```bash
   curl -X POST -H "Content-Type: application/json" \
   -d '{"key":"counter", "decrement": 5}' \
   http://localhost:8080/api/decrby
   ```
   **Response**:
   ```json
   {
     "key": "counter",
     "value": -5
   }
   ```

6. **Setting a TTL using Expire**:
   ```bash
   curl -X POST -H "Content-Type: application/json" \
   -d '{"key":"hello","ttl":60}' \
   http://localhost:8080/api/expire
   ```
   **Response**:
   ```json
   {
     "key": "hello",
     "ttl": 60,
     "success": true
   }
   ```

7. **Removing TTL using Persist**:
   ```bash
   curl -X POST -H "Content-Type: application/json" \
   -d '{"key":"hello"}' \
   http://localhost:8080/api/persist
   ```
   **Response**:
   ```json
   {
     "key": "hello",
     "success": true
   }
   ```

8. **Working with a list**:
   ```bash
   # LPush
   curl -X POST -H "Content-Type: application/json" \
   -d '{"key":"myList","values":["A"]}' \
   http://localhost:8080/api/lpush
   # LPop
   curl -X POST -H "Content-Type: application/json" \
   -d '{"key":"myList"}' \
   http://localhost:8080/api/lpop
   ```
   **LPop Response (if successful)**:
   ```json
   {
     "message": "LPOP success",
     "key": "myList",
     "value": "A"
   }
   ```

9. **Working with a hash**:
   ```bash
   # HSet
   curl -X POST -H "Content-Type: application/json" \
   -d '{"key":"user:1","field":"name","value":"Test","ttl":60}' \
   http://localhost:8080/api/hset

   # HGet
   curl -X GET "http://localhost:8080/api/hget?key=user:1&field=name"
   ```
   **HGet Response**:
   ```json
   {
     "key": "user:1",
     "field": "name",
     "value": "Test"
   }
   ```

10. **Finding keys by value**:
    ```bash
    curl -X POST -H "Content-Type: application/json" \
    -d '{"value":"Test"}' \
    http://localhost:8080/api/find
    ```
    **Possible Response**:
    ```json
    {
      "value": "Test",
      "keys": ["user:1", "someOtherKey"]
    }
    ```

11. **Deleting a key**:
    ```bash
    curl -X POST -H "Content-Type: application/json" \
    -d '{"key":"myKey"}' \
    http://localhost:8080/api/delete
    ```
    **Response**:
    ```json
    {
      "message": "Deleted",
      "key": "myKey"
    }
    ```

12. **Dropping all keys**:
    ```bash
    curl -X POST http://localhost:8080/api/dropall
    ```
    **Response**:
    ```json
    {
      "message": "All keys dropped"
    }
    ```
