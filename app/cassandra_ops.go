package app

import (
	"fmt"

	"github.com/gocql/gocql"
)

func (app *App) cassandraGet(key string) (string, int64, error) {
	var (
		value string
		ttl   int64
	)

	err := app.cassandra.Query(fmt.Sprintf("SELECT value,TTL(value) FROM %s WHERE key = ?", app.cassandraTable), key).Scan(&value, &ttl)
	if err != nil {
		// TTL returns -2 if the key does not exist
		if err == gocql.ErrNotFound {
			return "", -2, nil
		}

		return "", 0, err
	}

	// TTL returns -1 if the key does not have a TTL
	if ttl == 0 {
		return "", -1, nil
	}

	return value, ttl, nil
}

func (app *App) cassandraSet(key, value string, ttl uint64) error {
	if ttl == 0 {
		return app.cassandra.Query(fmt.Sprintf("INSERT INTO %s (key, value) VALUES (?, ?)", app.cassandraTable), key, value).Exec()
	}

	return app.cassandra.Query(fmt.Sprintf("INSERT INTO %s (key, value) VALUES (?, ?) USING TTL ?", app.cassandraTable), key, value, ttl).Exec()
}

func (app *App) cassandraDel(keys ...string) error {
	return app.cassandra.Query(fmt.Sprintf("DELETE FROM %s WHERE key IN ?", app.cassandraTable), keys).Exec()
}

func (app *App) cassandraExpire(key string, ttl int64) (int, error) {
	value, valueTTL, err := app.cassandraGet(key)
	if err != nil {
		return 0, err
	}
	// If the key does not exist, return 0
	if value == "" {
		return 0, nil
	}

	if valueTTL == ttl {
		return 1, nil
	}

	if err := app.cassandra.Query(fmt.Sprintf("UPDATE %s USING TTL ? SET value = ? WHERE key = ?", app.cassandraTable), ttl, value, key).Exec(); err != nil {
		return 0, err
	}

	return 1, nil
}

func (app *App) cassandraExists(keys ...string) (int, error) {
	var count int

	err := app.cassandra.Query(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE key IN ?", app.cassandraTable), keys).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (app *App) cassandraRename(oldKey, newKey string) error {
	oldKeyValue, oldKeyTTL, err := app.cassandraGet(oldKey)
	if err != nil {
		return err
	}
	// If the key does not exist, return an error
	if oldKeyValue == "" {
		return gocql.ErrNotFound
	}

	query := app.cassandra.Query(fmt.Sprintf("INSERT INTO %s (key, value) VALUES (?, ?)", app.cassandraTable), newKey, oldKeyValue)
	if oldKeyTTL > 0 {
		query = app.cassandra.Query(fmt.Sprintf("INSERT INTO %s (key, value) VALUES (?, ?) USING TTL ?", app.cassandraTable), newKey, oldKeyValue, oldKeyTTL)
	}

	if err := query.Exec(); err != nil {
		return err
	}

	if err := app.cassandra.Query(fmt.Sprintf("DELETE FROM %s WHERE key = ?", app.cassandraTable), oldKey).Exec(); err != nil {
		return err
	}

	return nil
}
