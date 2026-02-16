# RediSolar Go

A Go port of the sample application codebase for RU102PY, [Redis for Python Developers](https://github.com/redislabs-training/ru-dev-path-py.git) on [Redis University](https://university.redis.com).

This project implements the same RediSolar solar energy monitoring application using Go instead of Python/Flask. It is wire-compatible with the original Python version -- same Redis key prefix (`ru102py-app`), same REST API contract, and the same Vue.js frontend.

## Branches

| Branch     | Description                                                                 |
| ---------- | --------------------------------------------------------------------------- |
| `main`     | Complete, working application with all challenges solved                    |
| `learning` | Challenge stubs for hands-on learning (see [Learning Mode](#learning-mode)) |

If you want to learn Redis by solving challenges (recommended), switch to the `learning` branch:

```
$ git checkout learning
```

Then follow the instructions in [challenge.md](challenge.md) to work through 7 progressive challenges.

If you just want to run the finished application, stay on `main`.

## Setup

![Preview of running application - Solar Site Map with markers](preview.png)

### Prerequisites

To start and run this application, you will need:

- [Go 1.21+](https://go.dev/dl/)
- [Node.js and npm](https://nodejs.org/) (for building the frontend)
- Access to a local or remote installation of [Redis](https://redis.io/download) version 5 or newer
- Your Redis installation should have the RedisTimeSeries module installed. You can find the installation instructions at: https://oss.redis.com/redistimeseries/

**Note**: If you don't have Redis installed but do have Docker, you can start a Redis container with RedisTimeSeries:

```
$ docker run -p 6379:6379 redislabs/redistimeseries
```

### Install dependencies

```
$ make deps
```

This runs `go mod tidy` and `npm install` for the frontend.

### Redis

This project requires a connection to Redis. The default settings are configured via environment variables. If not set, the defaults are:

| Variable           | Default                            |
| ------------------ | ---------------------------------- |
| `REDIS_HOST`       | `redis-19256.redis.alldataint.com` |
| `REDIS_PORT`       | `19256`                            |
| `REDIS_KEY_PREFIX` | `ru102py-app`                      |
| `USE_GEO_SITE_API` | `true`                             |
| `SERVER_PORT`      | `8081`                             |

You can change these defaults in `internal/config/config.go`, or override them with environment variables:

```
$ REDIS_HOST=localhost REDIS_PORT=6379 make dev
```

#### Username and password protection

If you use Redis with a username (via the ACL system in Redis 6+) and/or a password, set the following environment variables:

```
$ export REDISOLAR_REDIS_USERNAME=your-username
$ export REDISOLAR_REDIS_PASSWORD=your-password
```

The project is configured to read environment variables from a `.env` file. If you need credentials, we recommend adding them there.

**Note**: The `.env` file should not be committed to git to avoid leaking credentials.

#### Key prefixes

This project prefixes all keys with a string. By default, the dev server and sample data loader use the prefix `ru102py-app:`, while the test suite uses `ru102py-test:`.

When you run the tests, they add keys to the same Redis database that the running application uses, but with the prefix `ru102py-test:`. The test runner deletes all keys with that prefix after each test.

## Loading sample data

Before the example app will do anything interesting, it needs data.

```
$ make load
```

This loads solar sites from `fixtures/sites.json` and generates example meter readings. It uses the Redis connection configured via environment variables.

## Running the dev server

Build the frontend and start the server:

```
$ make dev
```

After running `make dev`, access http://localhost:8081 to see the app.

**Don't see any data?** The first time you run `make dev`, you may see a map with nothing on it. You'll need to:

1. Load data with `make load`
2. If on the `learning` branch, complete Challenge #1

## Running tests

Run all tests:

```
$ make test
```

Run tests for a specific package:

```
$ go test ./internal/dao/redis/ -v
```

Run a specific test:

```
$ go test ./internal/dao/redis/ -run TestSite_FindAll -v
```

**Note**: The DAO integration tests require a running Redis instance. Unit tests (like keyschema tests) run without Redis.

## Project structure

```
.
├── cmd/
│   ├── server/         # HTTP server entry point
│   └── loader/         # Data loader entry point
├── internal/
│   ├── api/            # HTTP handlers, router, middleware, DTOs
│   ├── config/         # Environment-based configuration
│   ├── dao/            # DAO interfaces and errors
│   │   └── redis/      # Redis DAO implementations (challenges live here)
│   ├── datagen/        # Sample data generator
│   ├── keyschema/      # Redis key naming patterns
│   ├── models/         # Domain models and conversion functions
│   └── scripts/        # Embedded Lua scripts for atomic operations
├── fixtures/           # Sample site data (sites.json)
├── frontend/           # Vue.js frontend source
├── static/             # Built frontend assets (generated by make frontend)
├── challenge.md        # Challenge instructions for learning branch
├── Makefile
└── go.mod
```

## Learning Mode

To learn Redis hands-on by working through progressive challenges, switch to the `learning` branch:

```
$ git checkout learning
```

The `learning` branch has 7 challenges where key Redis operations are stubbed out with TODO comments. Your job is to implement them using the go-redis client library.

### How it works

1. Open [challenge.md](challenge.md) and read the instructions for the current challenge
2. Find the `// START Challenge #N` / `// END Challenge #N` markers in the source code
3. Implement the Redis operations described in the TODO comments
4. Remove the `t.Skip(...)` line from the corresponding test
5. Run the test to verify your solution:
   ```
   $ go test ./internal/dao/redis/ -run TestName -v
   ```
6. Move on to the next challenge

### Challenge overview

| #   | File                                 | Method                  | Redis Commands                | What you'll learn               |
| --- | ------------------------------------ | ----------------------- | ----------------------------- | ------------------------------- |
| 1   | `site.go`                            | `FindAll`               | SMEMBERS, HGETALL             | Sets and Hashes                 |
| 2   | `metric.go`                          | `insertMetric`          | ZADD, EXPIRE                  | Sorted Sets with expiry         |
| 3   | `site_stats.go` + `meter_reading.go` | `UpdateWithPipeline`    | Pipeline, Lua scripts         | Pipelines and atomic operations |
| 4   | `capacity_report.go`                 | `GetRank`               | ZREVRANK                      | Sorted Set ranking              |
| 5   | `site_geo.go`                        | `findByGeoWithCapacity` | GEORADIUS, ZSCORE             | Geo queries with pipeline       |
| 6   | `feed.go`                            | `InsertWithPipeline`    | XADD                          | Redis Streams                   |
| 7   | `sliding_window_rate_limiter.go`     | `Hit`                   | ZADD, ZREMRANGEBYSCORE, ZCARD | Sliding window rate limiting    |

All challenge files are in `internal/dao/redis/`. Tests are in the same package with `_test.go` suffix.

### Switching back to the complete solution

If you get stuck or want to see the finished code, you can always check the `main` branch:

```
$ git diff main -- internal/dao/redis/site.go    # see the solution for a specific file
$ git checkout main                                # switch to the complete solution
```

## Makefile targets

| Target          | Description                                    |
| --------------- | ---------------------------------------------- |
| `make deps`     | Install Go and frontend dependencies           |
| `make build`    | Compile server and loader binaries to `bin/`   |
| `make test`     | Run all Go tests                               |
| `make frontend` | Build the Vue.js frontend                      |
| `make load`     | Load sample data into Redis                    |
| `make dev`      | Build frontend and start the dev server        |
| `make run`      | Build everything and run the production binary |
| `make clean`    | Remove built artifacts                         |

## Optional (But recommended): RedisInsight

RedisInsight is a graphical tool for viewing data in Redis and managing Redis server instances. You don't need to install it to be successful with this course, but we recommend it as a good way of viewing data stored in Redis.

To use RedisInsight, [download it](https://redis.io/docs/ui/insight/) then point it at your Redis instance.

## FAQ

### Why do I get a connection error when I run the tests or dev server?

Redis is not running or not reachable at the configured host/port. Make sure Redis is running and the `REDIS_HOST` and `REDIS_PORT` environment variables (or defaults in `internal/config/config.go`) are correct.

### Why do I get an "Authentication required" error?

Your Redis instance requires credentials. Set `REDISOLAR_REDIS_USERNAME` and/or `REDISOLAR_REDIS_PASSWORD` environment variables. See the "Username and password protection" section above.

### Why do I get an "unknown command `TS.ADD`" when I try to run the tests?

Your Redis instance does not have the RedisTimeSeries module installed. See the Prerequisites section.

### Why do tests show "SKIP" for some tests?

On the `learning` branch, challenge tests are skipped by default with `t.Skip("Remove for Challenge #N")`. Remove the `t.Skip` line after you've implemented the challenge. See [challenge.md](challenge.md) for details.

## License

This project is released under the MIT license.
