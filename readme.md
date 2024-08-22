# Cassandra Redis Proxy

This is a simple proxy that allows you to use Cassandra as a cache backend for Redis.

## Configuration

The configuration is done through environment variables:

* `PROXY_LOG_LEVEL` (default: `info`): Log level. Possible values are `debug`, `info`, `warn`, `error`.
* `PROXY_CASSANDRA_HOSTS` (default: `localhost`): Comma separated list of Cassandra hosts.
* `PROXY_CASSANDRA_KEYSPACE` (default: `key_value_store`): Cassandra keyspace.
* `PROXY_CASSANDRA_TABLE` (default: `key_value): Cassandra table.
* `PROXY_REDIS_ADDRESS` (default: `:6380`): Redis listen address.

## Cassandra

Requires a cassandra version 3.0 or higher.

### Schema

```sql
CREATE KEYSPACE key_value_store WITH replication = {'class': 'NetworkTopologyStrategy'};

CREATE TABLE key_value_store.key_value (
    key text PRIMARY KEY,
    value text
) WITH default_time_to_live = 7884000;
```

* `default_time_to_live` is set to 7884000 seconds (3 months) to avoid the cache growing indefinitely. Zero means no expiration.

## Redis

Supported commands:

* `DEL`
* `EXISTS`
* `EXPIRE`
* `GET`
* `PING`
* `PTTL`
* `QUIT`
* `RENAME`
* `SET`
* `TTL`
* `UNLINK`
