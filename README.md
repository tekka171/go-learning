# go-learning: a Redis-compatible key-value server

A small in-memory key-value store written from scratch in Go, served over TCP
using the Redis protocol (RESP), so the standard `redis-cli` can talk to it.

Built as a learning project, milestone by milestone, using only the Go
standard library.

## Features

- Thread-safe in-memory store with per-key TTL
- Lazy expiration on read, plus a background janitor that sweeps expired keys
- Optional capacity limit with O(1) least-recently-used (LRU) eviction
- Binary-safe RESP parser with limits against oversized input
- One goroutine per client connection
- Graceful shutdown: stops accepting, waits for clients, then force-closes
  any that remain when the deadline passes
- Ctrl+C / SIGTERM handling in the server binary
- Sharded locking: Keys are spread across independent shards, each with its own lock and LRU list, so concurrent clients rarely wait on each other

## Requirements

Go 1.25 or later (the server uses `sync.WaitGroup.Go`).

## Project layout

```
go-learning/
├── kvstore/        the store: TTL, janitor, LRU eviction
├── periodic/       reusable "run a function every interval" worker
├── resp/           RESP protocol: command parser and reply writers
├── server/         command handling and the TCP server
└── cmd/kvserver/   the runnable server binary
```

| Package | Responsibility |
|---|---|
| `kvstore` | Stores keys and values. Knows nothing about the network. |
| `periodic` | Runs a function on a ticker until stopped. Used for the janitor. |
| `resp` | Converts between bytes and commands/replies. Knows nothing about the store. |
| `server` | Turns parsed commands into store calls, and manages connections. |
| `cmd/kvserver` | Wires everything together, reads flags, handles signals. |

## Running

```bash
go run ./cmd/kvserver
```

### Flags

| Flag | Default | Description |
|---|---|---|
| `-addr` | `:6379` | Address to listen on. `:6379` means all interfaces; use `127.0.0.1:6379` for local only. |
| `-cleanup` | `1s` | Interval between expired-key sweeps. `0` disables the janitor (lazy expiry only). |
| `-max-entries` | `0` | Maximum number of keys. `0` means unlimited. |
| `-shards` | `16` | Number of independent partitions. More shards reduce lock contention; LRU order is per shard. |

Example:

```bash
go run ./cmd/kvserver -addr 127.0.0.1:7000 -cleanup 500ms -max-entries 1000
```

Run with `-h` to print all flags.

### Trying it with redis-cli

Start the server, then in another terminal:

```bash
redis-cli -p 6379
```

On Windows, a community build of `redis-cli.exe` (for example from
`tporadowski/redis` releases) works; run it with `.\redis-cli.exe -p 6379`.
Don't start the bundled `redis-server.exe`, which would compete for the port.

```
127.0.0.1:6379> SET name john
OK
127.0.0.1:6379> GET name
"john"
127.0.0.1:6379> SET temp hello EX 10
OK
127.0.0.1:6379> TTL temp
(integer) 9
```

Plain `telnet` or `nc` sessions won't work: the server only accepts commands
encoded as RESP arrays, which is what `redis-cli` sends.

## Supported commands

Command names and the `EX` option are case-insensitive.

| Command | Reply |
|---|---|
| `PING` | `PONG` |
| `SET key value [EX seconds]` | `OK` |
| `GET key` | the value, or nil if missing or expired |
| `DEL key [key ...]` | number of keys removed |
| `EXPIRE key seconds` | `1` if the key existed, `0` otherwise |
| `TTL key` | seconds remaining, `-1` if no expiry, `-2` if missing |

Errors follow Redis's wording, for example
`ERR wrong number of arguments for 'get' command` and
`ERR value is not an integer or out of range`.

## Behaviour notes

- **`SET ... EX`** requires a positive number of seconds; `0`, negative, or
  non-numeric values are rejected.
- **`EXPIRE` with zero or a negative value deletes the key**, as in Redis.
- **TTL values are bounded**: seconds that would overflow a `time.Duration`
  (about 292 years) are rejected rather than wrapping to a negative duration.
- **`TTL` rounds down** to whole seconds.
- **LRU eviction**: `GET` (on a hit) and `SET` count as a use; `DEL`, `EXPIRE`,
  and `TTL` do not. Overwriting an existing key never evicts.
- `MaxEntries` is divided evenly across shards, rounded up, so the store may
  hold slightly more than `MaxEntries` in total.
- **Eviction is per shard**: when a key's shard is full, that shard's least
  recently used entry is evicted, even if the store as a whole is below the
  limit. Use `-shards 1` for one exact, global LRU order.
- Expired keys count toward the limit until they are removed.
- **Protocol errors** (input that isn't valid RESP) get an
  `ERR protocol error` reply, and the connection is then closed.
- **Input limits**: at most 1024 elements per command, and 512 MiB per
  bulk string.

## Shutdown

On Ctrl+C or SIGTERM the server:

1. stops accepting new connections,
2. waits up to 5 seconds for connected clients to disconnect,
3. closes any connections still open, and waits for their goroutines to exit.

The process exits with status `1` if the server stopped because of an error,
and `0` after a normal shutdown.

## Using the packages

```go
store := kvstore.NewStore(kvstore.Config{
	CleanupInterval: time.Second,
	MaxEntries:      10_000,
  Shards:          16,
})
defer store.Close()

ln, err := net.Listen("tcp", "127.0.0.1:6379")
if err != nil {
	log.Fatal(err)
}

srv := server.New(store)
go srv.Serve(ln) // returns server.ErrServerClosed after Shutdown

// later
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
srv.Shutdown(ctx)
```

## Testing

```bash
go test -race ./...
```

Add `-count=1` to bypass the test cache and `-v` for per-test output.

The tests use a few techniques worth knowing:

- **Injected clock** in `kvstore`, so expiry is tested without sleeping.
- **`net.Pipe`** to test a single connection in memory, without ports.
- **Port `0`** (`127.0.0.1:0`) so the OS picks a free port for server tests.
- **Read deadlines** on every test connection, so a bug fails the test instead
  of hanging it.

## Performance

Measured on an Intel i7-8750H (6 cores, 12 threads), 100,000 keys, random reads:

| Benchmark | 1 shard | 16 shards |
|---|---|---|
| `Get`, 1 goroutine | 486 ns/op | 543 ns/op |
| `Get`, 12 goroutines | 607 ns/op | 170 ns/op |

With a single lock, adding goroutines made the store *slower*: profiling showed
most CPU time spent spinning and parking on the mutex. Sharding cut concurrent
latency by 72%, at the cost of ~10% on a single goroutine (an extra hash to
pick the shard). Switching the store from `sync.RWMutex` to `sync.Mutex` saved
a further ~20 ns per operation, since every hot path needs exclusive access.

Run the benchmarks with:

    go test ./kvstore/ -run='^$' -bench='.' -benchmem

## Possible next steps

- An expiry heap so the janitor only touches keys that have expired
- Persistence with an append-only log replayed at startup
- More commands (`EXISTS`, `INCR`, `MGET`, ...)
