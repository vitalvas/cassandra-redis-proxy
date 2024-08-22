package app

import (
	"log"
	"log/slog"
	"os"

	"github.com/gocql/gocql"
	"github.com/tidwall/redcon"
)

type App struct {
	logger *slog.Logger

	cassandraCluster *gocql.ClusterConfig
	cassandra        *gocql.Session

	cassandraTable string
	redisAddress   string
}

func New() *App {
	conf, err := getConfig()
	if err != nil {
		log.Fatal(err)
	}

	loggerOptions := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	switch conf.LogLevel {
	case "debug":
		loggerOptions.Level = slog.LevelDebug

	case "info":
		loggerOptions.Level = slog.LevelInfo

	case "warn":
		loggerOptions.Level = slog.LevelWarn

	case "error":
		loggerOptions.Level = slog.LevelError

	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, loggerOptions))

	cluster := gocql.NewCluster(conf.CassandraHosts...)
	cluster.Keyspace = conf.CassandraKeyspace
	cluster.Consistency = gocql.Quorum
	// https://github.com/gocql/gocql/issues/538
	cluster.ProtoVersion = 4

	return &App{
		logger:           logger,
		cassandraCluster: cluster,
		cassandraTable:   conf.CassandraTable,
		redisAddress:     conf.RedisAddress,
	}
}

func (app *App) ListenAndServe() error {
	session, err := app.cassandraCluster.CreateSession()
	if err != nil {
		return err
	}

	defer session.Close()

	app.cassandra = session

	app.logger.Info("listen redis", "address", app.redisAddress)
	return redcon.ListenAndServe(
		app.redisAddress,
		app.redisRequestHandler,
		app.redisAcceptHandler,
		app.redisClosedHandler,
	)
}
