package app

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gocql/gocql"
	"github.com/tidwall/redcon"
)

var redisCommands = map[string]func(*App, redcon.Conn, []string) error{
	"ping":   redisCmdPing,
	"quit":   redisCmdQuit,
	"get":    redisCmdGet,
	"set":    redisCmdSet,
	"del":    redisCmdDel,
	"unlink": redisCmdDel,
	"ttl":    redisCmdTTL,
	"pttl":   redisCmdTTL,
	"expire": redisCmdExpire,
	"exists": redisCmdExists,
	"rename": redisCmdRename,
}

func redisCmdPing(_ *App, conn redcon.Conn, _ []string) error {
	conn.WriteString("PONG")
	return nil
}

func redisCmdQuit(_ *App, conn redcon.Conn, _ []string) error {
	conn.WriteString("OK")
	conn.Close()
	return nil
}

func redisCmdGet(app *App, conn redcon.Conn, args []string) error {
	if len(args) != 2 {
		return errors.New("wrong number of arguments for 'get' command")
	}

	key := args[1]

	value, _, err := app.cassandraGet(key)

	if err != nil {
		return fmt.Errorf("cassandra query: %w", err)
	}

	if value == "" {
		conn.WriteNull()
	} else {
		conn.WriteString(value)
	}

	return nil
}

func redisCmdSet(app *App, conn redcon.Conn, args []string) error {
	if len(args) != 3 && len(args) != 5 {
		return errors.New("wrong number of arguments for 'set' command")
	}

	var cacheTTL uint64

	if len(args) == 5 {
		if strings.ToLower(args[3]) != "ex" {
			return errors.New("syntax error")
		}

		if ttl, err := strconv.ParseUint(args[4], 10, 32); err != nil || ttl > 630720000 {
			return errors.New("value is not an integer or out of range")
		} else if ttl > 0 {
			cacheTTL = ttl
		}
	}

	key := args[1]
	value := args[2]

	if err := app.cassandraSet(key, value, cacheTTL); err != nil {
		return fmt.Errorf("cassandra query: %w", err)
	}

	conn.WriteString("OK")

	return nil
}

func redisCmdDel(app *App, conn redcon.Conn, args []string) error {
	if len(args) < 2 {
		return errors.New("wrong number of arguments for 'del' command")
	}

	keys := make([]string, len(args)-1)

	for i, key := range args[1:] {
		keys[i] = string(key)
	}

	if err := app.cassandraDel(keys...); err != nil {
		return fmt.Errorf("cassandra query: %w", err)
	}

	// cassandra doesn't return the number of deleted keys, so we return the number of keys we received
	conn.WriteInt(len(keys))

	return nil
}

func redisCmdTTL(app *App, conn redcon.Conn, args []string) error {
	if len(args) != 2 {
		return errors.New("wrong number of arguments for 'ttl' command")
	}

	key := args[1]

	_, ttl, err := app.cassandraGet(key)
	if err != nil {
		return fmt.Errorf("cassandra query: %w", err)
	}

	// pttl returns the TTL in milliseconds
	if args[0] == "pttl" && ttl > 0 {
		ttl *= 1000
	}

	conn.WriteInt64(ttl)

	return nil
}

func redisCmdExpire(app *App, conn redcon.Conn, args []string) error {
	if len(args) != 3 {
		return errors.New("wrong number of arguments for 'expire' command")
	}

	ttl, err := strconv.ParseInt(args[2], 10, 32)

	if err != nil || ttl < 0 || ttl > 630720000 {
		return errors.New("value is not an integer or out of range")
	}

	count, err := app.cassandraExpire(args[1], ttl)
	if err != nil {
		return fmt.Errorf("cassandra query: %w", err)
	}

	conn.WriteInt(count)

	return nil
}

func redisCmdExists(app *App, conn redcon.Conn, args []string) error {
	if len(args) < 2 {
		return errors.New("wrong number of arguments for 'exists' command")
	}

	keys := make([]string, len(args)-1)
	for i, key := range args[1:] {
		keys[i] = string(key)
	}

	count, err := app.cassandraExists(keys...)
	if err != nil {
		return fmt.Errorf("cassandra query: %w", err)
	}

	conn.WriteInt(count)

	return nil
}

func redisCmdRename(app *App, conn redcon.Conn, args []string) error {
	if len(args) != 3 {
		return errors.New("wrong number of arguments for 'rename' command")
	}

	oldKey := args[1]
	newKey := args[2]

	if count, err := app.cassandraExists(oldKey); err != nil {
		return fmt.Errorf("cassandra query: %w", err)
	} else if count == 0 {
		return errors.New("no such key")
	}

	if err := app.cassandraRename(oldKey, newKey); err != nil {
		if err == gocql.ErrNotFound {
			return errors.New("no such key")
		}

		return fmt.Errorf("cassandra query: %w", err)
	}

	conn.WriteString("OK")

	return nil
}
