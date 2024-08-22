package app

import (
	"errors"
	"slices"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	LogLevel string `split_words:"true" default:"info"`

	CassandraHosts    []string `split_words:"true" default:"127.0.0.1"`
	CassandraKeyspace string   `split_words:"true" default:"key_value_store"`
	CassandraTable    string   `split_words:"true" default:"key_value"`

	RedisAddress string `split_words:"true" default:":6380"`
}

func getConfig() (*Config, error) {
	var cfg Config

	err := envconfig.Process("proxy", &cfg)
	if err != nil {
		return nil, err
	}

	if !slices.Contains([]string{"debug", "info", "warn", "error"}, cfg.LogLevel) {
		return nil, errors.New("invalid log level")

	}

	return &cfg, nil
}
