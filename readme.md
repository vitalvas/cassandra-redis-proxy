# Cassandra Redis Proxy

This is a simple proxy that allows you to use Cassandra as a cache for Redis.

## Cassandra

Requires a cassandra version 3.0 or higher. Also tested with YugabyteDB with the cassandra compatibility layer.

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
