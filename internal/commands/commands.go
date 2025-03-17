package commands

import (
	"context"
	"fmt"
	"github.com/themedef/go-hermes/internal/contracts"
	"strconv"
	"strings"
)

type CommandAPI struct {
	db contracts.Store
}

func NewCommandAPI(db contracts.Store) *CommandAPI {
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
			ttl, _ = strconv.Atoi(parts[3])
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
		val, found, err := c.db.Get(ctx, key)
		if err != nil {
			return "", err
		}
		if !found {
			return "(nil)", nil
		}
		return fmt.Sprintf("\"%v\"", val), nil

	case "DEL":
		if len(parts) < 2 {
			return "", fmt.Errorf("Usage: DEL key")
		}
		key := parts[1]
		deleted, err := c.db.Delete(ctx, key)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%v", deleted), nil

	case "INCR":
		if len(parts) < 2 {
			return "", fmt.Errorf("Usage: INCR key")
		}
		key := parts[1]
		newVal, ok, err := c.db.Incr(ctx, key)
		if err != nil {
			return "", err
		}
		if !ok {
			return "(not a number?)", nil
		}
		return fmt.Sprintf("%d", newVal), nil

	case "DECR":
		if len(parts) < 2 {
			return "", fmt.Errorf("Usage: DECR key")
		}
		key := parts[1]
		newVal, ok, err := c.db.Decr(ctx, key)
		if err != nil {
			return "", err
		}
		if !ok {
			return "(not a number?)", nil
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
		val, found, err := c.db.LPop(ctx, key)
		if err != nil {
			return "", err
		}
		if !found {
			return "(nil)", nil
		}
		return fmt.Sprintf("%v", val), nil

	case "RPOP":
		if len(parts) < 2 {
			return "", fmt.Errorf("Usage: RPOP key")
		}
		key := parts[1]
		val, found, err := c.db.RPop(ctx, key)
		if err != nil {
			return "", err
		}
		if !found {
			return "(nil)", nil
		}
		return fmt.Sprintf("%v", val), nil

	case "EXPIRE":
		if len(parts) < 3 {
			return "", fmt.Errorf("Usage: EXPIRE key seconds")
		}
		key := parts[1]
		ttl, _ := strconv.Atoi(parts[2])
		if err := c.db.UpdateTTL(ctx, key, ttl); err != nil {
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
			return "", err
		}
		return fmt.Sprintf("Keys: %v", keys), nil

	case "TTL":
		return "Not implemented TTL command in DB. Implement if needed.", nil

	case "HELP":
		return `
Commands:
  SET key value [ttl]
  GET key
  DEL key
  INCR key
  DECR key
  LPUSH key value
  RPUSH key value
  LPOP key
  RPOP key
  EXPIRE key seconds
  FIND value
  TTL (not implemented)
  HELP
  QUIT / EXIT
`, nil

	case "QUIT", "EXIT":
		return "Bye!", nil

	default:
		return "", fmt.Errorf("unknown command: %s", cmd)
	}
}
