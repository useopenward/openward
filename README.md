# OpenWard

A self-hosted reverse proxy that enforces rate limiting on your upstream APIs. Single Go binary, SQLite-backed, zero external dependencies at runtime.

> **Status:** Early development. Core proxy and rate limiting engine are functional. REST API and dashboard are not yet built.

---

## How it works

OpenWard sits in front of your upstream API. Clients send requests to OpenWard with an API key — OpenWard authenticates the key, checks the rate limit, and forwards the request if allowed.

```
client → OpenWard (auth + rate limit) → upstream API
```

Each **project** maps one OpenWard API key to one upstream target with its own rate limiting configuration. Upstream API keys (e.g. your OpenAI key) pass through untouched in the proxied request — OpenWard never sees or stores them.

---

## Rate limiting algorithms

**Fixed window** — allows N requests per fixed time window. Simple and cheap. Susceptible to burst at window boundaries.

**Sliding window** — approximates a rolling window using a weighted count of the previous and current window. Smoother than fixed, same storage cost.

**Token bucket** — allows bursts up to a capacity, then refills at a steady rate. Best for APIs that tolerate short bursts but need a sustained rate enforced.

---

## Architecture

- **Language:** Go
- **Storage:** SQLite (WAL mode) — no Redis, no Postgres, no external runtime dependencies
- **Proxy:** `net/http/httputil.ReverseProxy`
- **Project cache:** in-memory with 30s TTL, avoids hitting SQLite on every request
- **Log writes:** async, separate writer connection from reader pool

Inspired by [PocketBase](https://pocketbase.io) — the goal is a single binary you can drop on a server and run.

---
## Benchmarks

Benchmarks were performed against a local `httpbin` instance using **100 concurrent virtual users** over **30 seconds**.

| Metric          |       Direct | Through OpenWard |
| --------------- | -----------: | ---------------: |
| Average Latency | **62.81 ms** |     **63.98 ms** |
| Median Latency  |     62.03 ms |         63.04 ms |
| P90             |     69.42 ms |         74.72 ms |
| P95             |     72.75 ms |         85.10 ms |
| Maximum         |    146.92 ms |        135.09 ms |

**Throughput**

```text
Requests/sec : ~1,574
Total Requests: 47,307
Failed Requests: 0
```

The additional latency introduced by OpenWard is approximately **1.2 ms on average**, while sustaining over **1,500 requests per second** with **zero failed requests**.

These benchmarks represent the current state of OpenWard during active development. Several performance optimizations are still planned and have not yet been implemented, including improvements to the proxy pipeline, request processing, and other core components. As development continues, these results are expected to improve. You can follow the project's roadmap and GitHub timeline to track upcoming performance-related work.



---

## Project structure

```
openward/
├── cmd/
│   ├── main.go           # entry point
│   └── seed/             # temporary seed script for dev
│       └── main.go
├── internal/
│   ├── core/
│   │   └── project.go    # domain types — Project, rate limit configs
│   ├── db/
│   │   ├── db.go         # SQLite init, WAL setup, migrations
│   │   └── projects.go   # CRUD for projects
│   ├── limiter/
│   │   ├── limiter.go    # Limiter interface + factory
│   │   ├── fixed.go      # fixed window
│   │   ├── sliding.go    # sliding window
│   │   └── token.go      # token bucket
│   └── proxy/
│       └── proxy.go      # reverse proxy handler
├── benchmarks/     
│   └── overhead.js       # k6 proxy overhead benchmark
├── go.mod
└── go.sum
```

---

## Running locally

**Prerequisites:** Go 1.21+, [k6](https://k6.io) (optional, for benchmarks), Docker (optional, for local httpbin)

```bash
git clone https://github.com/useopenward/openward
cd openward
go mod download

# seed a test project
go run ./cmd/seed/main.go
# outputs: created project, api key: ow_...

# start the proxy
go run ./cmd
# listening on :8080

# test it
curl -H "X-API-Key: ow_yourkey" http://localhost:8080/get
```

**Optional: run a local httpbin for testing**
```bash
docker run -p 9999:80 kennethreitz/httpbin
```
Update the upstream in `cmd/seed/main.go` to `http://localhost:9999`.

**Run benchmarks**
```bash
k6 run benchmarks/overhead.js
```

---

## Roadmap

- [ ] REST API for project management
- [ ] Embedded SvelteKit dashboard
- [ ] `X-RateLimit-Remaining` and `X-RateLimit-Reset` response headers
- [ ] Buffered log mode (in-memory flush) vs sync mode, configurable per project
- [ ] Single binary with embedded frontend (à la PocketBase)

---

## License

MIT