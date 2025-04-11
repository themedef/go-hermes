package hermes

import (
	"context"
	"fmt"
	"github.com/themedef/go-hermes/internal/contracts"
	"github.com/themedef/go-hermes/internal/types"
	"strconv"
	"strings"
)

type CommandAPI struct {
	db contracts.StoreHandler
}

func NewCommandAPI(db contracts.StoreHandler) contracts.CommandsHandler {
	return &CommandAPI{db: db}
}

func (c *CommandAPI) Execute(ctx context.Context, parts []string) (string, error) {
	if len(parts) == 0 {
		return "", nil
	}
	cmd := strings.ToUpper(parts[0])

	switch cmd {
	case "SET":
		if len(parts) < 3 {
			return "", fmt.Errorf("Usage: SET key value [ttlSeconds]")
		}
		key := parts[1]
		value := parts[2]
		ttl := 0
		if len(parts) >= 4 {
			tmp, err := strconv.Atoi(parts[3])
			if err != nil {
				return "", fmt.Errorf("invalid TTL: %v", parts[3])
			}
			ttl = tmp
		}
		if err := c.db.Set(ctx, key, value, ttl); err != nil {
			return "", err
		}
		return "OK", nil

	case "GET":
		if len(parts) < 2 {
			return "", fmt.Errorf("Usage: GET key")
		}
		key := parts[1]
		val, err := c.db.Get(ctx, key)
		if err != nil {
			if IsKeyNotFound(err) || IsKeyExpired(err) {
				return "(nil)", nil
			}
			return "", err
		}
		return fmt.Sprintf("\"%v\"", val), nil

	case "SETNX":
		if len(parts) < 3 {
			return "", fmt.Errorf("Usage: SETNX key value [ttlSeconds]")
		}
		key := parts[1]
		value := parts[2]
		ttl := 0
		if len(parts) >= 4 {
			tmp, err := strconv.Atoi(parts[3])
			if err != nil {
				return "", fmt.Errorf("invalid TTL: %v", parts[3])
			}
			ttl = tmp
		}
		ok, err := c.db.SetNX(ctx, key, value, ttl)
		if err != nil {
			return "", err
		}
		if ok {
			return "1", nil
		}
		return "0", nil

	case "SETXX":
		if len(parts) < 3 {
			return "", fmt.Errorf("Usage: SETXX key value [ttlSeconds]")
		}
		key := parts[1]
		value := parts[2]
		ttl := 0
		if len(parts) >= 4 {
			tmp, err := strconv.Atoi(parts[3])
			if err != nil {
				return "", fmt.Errorf("invalid TTL: %v", parts[3])
			}
			ttl = tmp
		}
		ok, err := c.db.SetXX(ctx, key, value, ttl)
		if err != nil {
			return "", err
		}
		if ok {
			return "1", nil
		}
		return "0", nil

	case "SETCAS":
		if len(parts) < 4 {
			return "", fmt.Errorf("Usage: SETCAS key old_value new_value [ttlSeconds]")
		}
		key := parts[1]
		oldValue := parts[2]
		newValue := parts[3]
		ttl := 0
		if len(parts) >= 5 {
			tmp, err := strconv.Atoi(parts[4])
			if err != nil {
				return "", fmt.Errorf("invalid TTL: %v", parts[4])
			}
			ttl = tmp
		}
		err := c.db.SetCAS(ctx, key, oldValue, newValue, ttl)
		if err != nil {
			return "", err
		}
		return "OK", nil

	case "GETSET":
		if len(parts) < 3 {
			return "", fmt.Errorf("Usage: GETSET key new_value [ttlSeconds]")
		}
		key := parts[1]
		newValue := parts[2]
		ttl := 0
		if len(parts) >= 4 {
			tmp, err := strconv.Atoi(parts[3])
			if err != nil {
				return "", fmt.Errorf("invalid TTL: %v", parts[3])
			}
			ttl = tmp
		}
		oldVal, err := c.db.GetSet(ctx, key, newValue, ttl)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%v", oldVal), nil

	case "INCR":
		if len(parts) < 2 {
			return "", fmt.Errorf("Usage: INCR key")
		}
		key := parts[1]
		newVal, err := c.db.Incr(ctx, key)
		if err != nil {
			if IsInvalidValueType(err) {
				return "(error) value is not an integer", nil
			}
			return "", err
		}
		return fmt.Sprintf("%d", newVal), nil

	case "DECR":
		if len(parts) < 2 {
			return "", fmt.Errorf("Usage: DECR key")
		}
		key := parts[1]
		newVal, err := c.db.Decr(ctx, key)
		if err != nil {
			if IsInvalidValueType(err) {
				return "(error) value is not an integer", nil
			}
			return "", err
		}
		return fmt.Sprintf("%d", newVal), nil

	case "INCRBY":
		if len(parts) < 3 {
			return "", fmt.Errorf("Usage: INCRBY key increment")
		}
		key := parts[1]
		inc, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			return "", fmt.Errorf("invalid increment: %v", parts[2])
		}
		newVal, err := c.db.IncrBy(ctx, key, inc)
		if err != nil {
			if IsInvalidValueType(err) {
				return "(error) value is not an integer", nil
			}
			return "", err
		}
		return fmt.Sprintf("%d", newVal), nil

	case "DECRBY":
		if len(parts) < 3 {
			return "", fmt.Errorf("Usage: DECRBY key decrement")
		}
		key := parts[1]
		dec, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			return "", fmt.Errorf("invalid decrement: %v", parts[2])
		}
		newVal, err := c.db.DecrBy(ctx, key, dec)
		if err != nil {
			if IsInvalidValueType(err) {
				return "(error) value is not an integer", nil
			}
			return "", err
		}
		return fmt.Sprintf("%d", newVal), nil

	case "LPUSH":
		if len(parts) < 3 {
			return "", fmt.Errorf("Usage: LPUSH key value")
		}
		key := parts[1]
		value := parts[2]
		if err := c.db.LPush(ctx, key, value); err != nil {
			return "", err
		}
		return "OK", nil

	case "RPUSH":
		if len(parts) < 3 {
			return "", fmt.Errorf("Usage: RPUSH key value")
		}
		key := parts[1]
		value := parts[2]
		if err := c.db.RPush(ctx, key, value); err != nil {
			return "", err
		}
		return "OK", nil

	case "LPOP":
		if len(parts) < 2 {
			return "", fmt.Errorf("Usage: LPOP key")
		}
		key := parts[1]
		val, err := c.db.LPop(ctx, key)
		if err != nil {
			if IsKeyNotFound(err) || IsEmptyList(err) {
				return "(nil)", nil
			}
			return "", err
		}
		return fmt.Sprintf("%v", val), nil

	case "RPOP":
		if len(parts) < 2 {
			return "", fmt.Errorf("Usage: RPOP key")
		}
		key := parts[1]
		val, err := c.db.RPop(ctx, key)
		if err != nil {
			if IsKeyNotFound(err) || IsEmptyList(err) {
				return "(nil)", nil
			}
			return "", err
		}
		return fmt.Sprintf("%v", val), nil

	case "LLEN":
		if len(parts) < 2 {
			return "", fmt.Errorf("Usage: LLEN key")
		}
		key := parts[1]
		length, err := c.db.LLen(ctx, key)
		if err != nil {
			return "", err
		}
		return strconv.Itoa(length), nil

	case "LRANGE":
		if len(parts) < 4 {
			return "", fmt.Errorf("Usage: LRANGE key start end")
		}
		key := parts[1]
		start, err := strconv.Atoi(parts[2])
		if err != nil {
			return "", fmt.Errorf("invalid start: %v", parts[2])
		}
		end, err := strconv.Atoi(parts[3])
		if err != nil {
			return "", fmt.Errorf("invalid end: %v", parts[3])
		}
		result, err := c.db.LRange(ctx, key, start, end)
		if err != nil {
			return "", err
		}
		if len(result) == 0 {
			return "(empty list)", nil
		}
		var elems []string
		for _, v := range result {
			elems = append(elems, fmt.Sprintf("%v", v))
		}
		return fmt.Sprintf("[%s]", strings.Join(elems, ", ")), nil

	case "HSET":
		if len(parts) < 4 {
			return "", fmt.Errorf("Usage: HSET key field value [ttl]")
		}
		key := parts[1]
		field := parts[2]
		value := parts[3]
		ttl := 0
		if len(parts) >= 5 {
			tmp, err := strconv.Atoi(parts[4])
			if err != nil {
				return "", fmt.Errorf("invalid TTL: %v", parts[4])
			}
			ttl = tmp
		}
		if err := c.db.HSet(ctx, key, field, value, ttl); err != nil {
			return "", err
		}
		return "1", nil

	case "HGET":
		if len(parts) < 3 {
			return "", fmt.Errorf("Usage: HGET key field")
		}
		key := parts[1]
		field := parts[2]
		val, err := c.db.HGet(ctx, key, field)
		if err != nil {
			if IsKeyNotFound(err) {
				return "(nil)", nil
			}
			return "", err
		}
		return fmt.Sprintf("%v", val), nil

	case "HDEL":
		if len(parts) < 3 {
			return "", fmt.Errorf("Usage: HDEL key field")
		}
		key := parts[1]
		field := parts[2]
		if err := c.db.HDel(ctx, key, field); err != nil {
			if IsKeyNotFound(err) {
				return "0", nil
			}
			return "", err
		}
		return "1", nil

	case "HGETALL":
		if len(parts) < 2 {
			return "", fmt.Errorf("Usage: HGETALL key")
		}
		key := parts[1]
		hash, err := c.db.HGetAll(ctx, key)
		if err != nil {
			if IsKeyNotFound(err) {
				return "(empty list or set)", nil
			}
			return "", err
		}
		if len(hash) == 0 {
			return "(empty list or set)", nil
		}
		var result []string
		for k, v := range hash {
			result = append(result, fmt.Sprintf("%q: %q", k, v))
		}
		return strings.Join(result, "\n"), nil

	case "HEXISTS":
		if len(parts) < 3 {
			return "", fmt.Errorf("Usage: HEXISTS key field")
		}
		key := parts[1]
		field := parts[2]
		exists, err := c.db.HExists(ctx, key, field)
		if err != nil {
			return "", err
		}
		if exists {
			return "1", nil
		}
		return "0", nil

	case "HLEN":
		if len(parts) < 2 {
			return "", fmt.Errorf("Usage: HLEN key")
		}
		key := parts[1]
		length, err := c.db.HLen(ctx, key)
		if err != nil {
			return "", err
		}
		return strconv.Itoa(length), nil

	case "EXPIRE":
		if len(parts) < 3 {
			return "", fmt.Errorf("Usage: EXPIRE key seconds")
		}
		key := parts[1]
		ttl, err := strconv.Atoi(parts[2])
		if err != nil {
			return "", fmt.Errorf("invalid TTL: %v", parts[2])
		}
		ok, err := c.db.Expire(ctx, key, ttl)
		if err != nil {
			return "", err
		}
		if !ok {
			return "false", nil
		}
		return "OK", nil

	case "PERSIST":
		if len(parts) < 2 {
			return "", fmt.Errorf("Usage: PERSIST key")
		}
		key := parts[1]
		ok, err := c.db.Persist(ctx, key)
		if err != nil {
			return "", err
		}
		if !ok {
			return "false", nil
		}
		return "OK", nil

	case "TTL":
		if len(parts) < 2 {
			return "", fmt.Errorf("Usage: TTL key")
		}
		key := parts[1]
		_, ttl, err := c.db.GetWithDetails(ctx, key)
		if err != nil {
			if IsKeyNotFound(err) || IsKeyExpired(err) {
				return "-2", nil
			}
			return "", err
		}
		if ttl == -1 {
			return "-1", nil
		}
		return strconv.Itoa(ttl), nil

	case "TYPE":
		if len(parts) < 2 {
			return "", fmt.Errorf("Usage: TYPE key")
		}
		key := parts[1]
		dtype, err := c.db.Type(ctx, key)
		if err != nil {
			if IsKeyNotFound(err) {
				return "(nil)", nil
			}
			return "", err
		}
		dt, ok := dtype.(types.DataType)
		if !ok {
			return "", fmt.Errorf("unexpected type returned")
		}
		var typeStr string
		switch dt {
		case types.String:
			typeStr = "string"
		case types.List:
			typeStr = "list"
		case types.Hash:
			typeStr = "hash"
		default:
			typeStr = "unknown"
		}
		return typeStr, nil

	case "GETWITHDETAILS":
		if len(parts) < 2 {
			return "", fmt.Errorf("Usage: GETWITHDETAILS key")
		}
		key := parts[1]
		val, ttl, err := c.db.GetWithDetails(ctx, key)
		if err != nil {
			if IsKeyNotFound(err) || IsKeyExpired(err) {
				return "(nil)", nil
			}
			return "", err
		}
		return fmt.Sprintf("Value: %v, TTL: %d", val, ttl), nil

	case "RENAME":
		if len(parts) < 3 {
			return "", fmt.Errorf("Usage: RENAME old_key new_key")
		}
		oldKey := parts[1]
		newKey := parts[2]
		if err := c.db.Rename(ctx, oldKey, newKey); err != nil {
			if IsKeyNotFound(err) {
				return "(nil)", nil
			} else if IsKeyExists(err) {
				return "(error) new key exists", nil
			}
			return "", err
		}
		return "OK", nil

	case "FIND":
		if len(parts) < 2 {
			return "", fmt.Errorf("Usage: FIND value")
		}
		value := parts[1]
		keys, err := c.db.FindByValue(ctx, value)
		if err != nil {
			if IsKeyNotFound(err) {
				return "Keys: []", nil
			}
			return "", err
		}
		if len(keys) == 0 {
			return "Keys: []", nil
		}
		return fmt.Sprintf("Keys: [%s]", strings.Join(keys, ",")), nil

	case "EXISTS":
		if len(parts) < 2 {
			return "", fmt.Errorf("Usage: EXISTS key")
		}
		key := parts[1]
		exists, err := c.db.Exists(ctx, key)
		if err != nil {
			return "", err
		}
		if exists {
			return "1", nil
		}
		return "0", nil

	case "DEL":
		if len(parts) < 2 {
			return "", fmt.Errorf("Usage: DEL key")
		}
		key := parts[1]
		err := c.db.Delete(ctx, key)
		if err != nil {
			if IsKeyNotFound(err) {
				return "false", nil
			}
			return "", err
		}
		return "true", nil

	case "DROPALL":
		if err := c.db.DropAll(ctx); err != nil {
			return "", err
		}
		return "OK", nil

	case "EXEC":
		return "EXEC not implemented", nil

	case "DISCARD":
		return "DISCARD executed", nil

	case "HELP":
		return `
Available Commands:
  SET key value [ttl]
  GET key
  SETNX key value [ttl]
  SETXX key value [ttl]
  SETCAS key old_value new_value [ttl]
  GETSET key new_value [ttl]
  INCR key
  DECR key
  INCRBY key increment
  DECRBY key decrement
  LPUSH key value
  RPUSH key value
  LPOP key
  RPOP key
  LLEN key
  LRANGE key start end
  HSET key field value [ttl]
  HGET key field
  HDEL key field
  HGETALL key
  HEXISTS key field
  HLEN key
  EXISTS key
  EXPIRE key seconds
  PERSIST key
  TTL key
  TYPE key
  GETWITHDETAILS key
  RENAME old_key new_key
  FIND value
  DEL key
  DROPALL
  EXEC
  DISCARD
  HELP
  QUIT / EXIT
`, nil

	case "QUIT", "EXIT":
		return "Bye!", nil

	default:
		return "", fmt.Errorf("unknown command: %s", cmd)
	}
}
