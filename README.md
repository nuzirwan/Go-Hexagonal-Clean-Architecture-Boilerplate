# Go Hexagonal Clean Architecture Boilerplate

[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A lightweight Go microservice starter that's easy to learn, simple to extend, and stays out of your way. Hexagonal + clean architecture without the ceremony.

**Zero framework.** Stdlib `net/http` + clear structure. Read one file, understand the whole flow. Add a usecase in 3 files, observability included automatically.

---

## Why This Exists

Most Go boilerplates give you a flat `handler → service → repository` with logging scattered everywhere. This one gives you:

- **Zero-code tracing** — build with `otelc`, get distributed traces without writing a single span
- **Architecture enforced by linter** — not just convention, CI catches violations
- **No ORM, no framework** — stdlib only, nothing to learn except Go itself
- **Multi-entry-point binary** — HTTP, gRPC, worker, CLI — same binary, different commands
- **Resilience out of the box** — circuit breaker, retry, singleflight, pre-wired
- **~3us observability overhead** — measured, not guessed

You write business logic. Everything else is handled by the structure.

---

## How It's Organized

The project separates **what your app does** (domain) from **how it talks to the world** (adapters).

### Domain — your business logic

```
domain/
  port/       What the domain needs (interfaces)
  entity/     Your data structures + rules
  usecase/    Orchestration: "when X happens, do Y then Z"
```

No imports from outside. No HTTP, no SQL, no Redis. Pure logic. Testable with plain mocks.

### Adapters — how the world connects

Adapters are grouped by **direction** and then by **technology**:

```
adapter/
  in/                  Things that CALL your domain (entry points)
    http/              REST API (net/http stdlib)
    grpc/              gRPC (google.golang.org/grpc)
    worker/            Background job handler (hibiken/asynq)
  out/                 Things your domain CALLS (dependencies)
    postgres/          Database (database/sql + pgx)
    redis/             Cache (go-redis/v9 + circuit breaker)
    asynq/             Task publisher (hibiken/asynq)
    observability/     Tracing + metrics (OTel decorator)
```

Each folder = one technology. Need MongoDB? Add `adapter/out/mongo/`. Need Kafka? Add `adapter/in/kafka/` (consumer) and `adapter/out/kafka/` (producer). The structure scales without reorganizing existing code.

**Inbound adapters** (`in/`) are entry points — each one is a different way to trigger the same usecase. Add a new protocol? Add another adapter. The domain doesn't change.

**Outbound adapters** (`out/`) are dependencies — swap Postgres for MySQL, or Redis for Memcached, by implementing the same port interface. Again, domain untouched.

### Multi-entry-point

One binary, many ways in:

```bash
myservice serve http       # REST API on :8080
myservice serve grpc       # gRPC on :8081
myservice worker           # background jobs
myservice migrate up       # DB migrations
```

All share the same domain logic. Add `myservice serve graphql` or `myservice cron` tomorrow — just wire a new inbound adapter to the existing usecases.

---

## 30-Second Start

```bash
make init                      # install tools (otelc, air, linters)
docker compose -f deployments/docker-compose.yml up -d
cp .env.example .env
make migrate-up
make dev                       # hot reload on :8080
```

```bash
curl -X POST localhost:8080/products \
  -H "Content-Type: application/json" \
  -d '{"name":"Widget","price":99.99,"currency":"USD","stock":10}'
```

Traces appear in Jaeger at `http://localhost:16686`. No code changes.

---

## Performance

Framework overhead only (no DB, no network):

| Endpoint | Without Observability | With Decorators (no-op OTel) | Overhead |
|----------|----------------------|------------------------------|----------|
| `DELETE /products/{id}` | 0.44 us / 4 allocs | 3.1 us / 19 allocs | +2.7 us |
| `GET /products/{id}` | 3.9 us / 14 allocs | 7.2 us / 29 allocs | +3.3 us |
| `POST /products` | 15.6 us / 32 allocs | 19 us / 48 allocs | +3.4 us |
| `GET /products` (10 items) | 21 us / 50 allocs | 26 us / 66 allocs | +5 us |

Decorator overhead: ~3us per usecase call (span creation + metric recording). In production with real DB calls (2-10ms), this is <0.1% of total request time.

---

## Architecture

```
cmd/                         Wire everything here (composition root)
internal/
  domain/                    Pure business logic (ZERO external imports)
    port/                    Interfaces: what the domain needs and provides
    entity/                  Structs with behavior and validation
    usecase/                 Orchestration (calls ports, never infrastructure)
  adapter/
    in/                      Things that call your domain
      http/                  REST API (net/http stdlib)
      grpc/                  gRPC server
      worker/                Background task handlers (asynq)
    out/                     Things your domain calls
      postgres/              Repository, transactions
      redis/                 Cache with circuit breaker
      asynq/                 Task enqueuer
      observability/         Usecase tracing + metrics (decorator pattern)
  config/                    Env loading
pkg/                         Domain-free utilities
  httputil/                  JSON envelope (sonic)
  resilience/                Circuit breaker, retry, singleflight
  health/                    Liveness + readiness probes
  shutdown/                  Graceful shutdown
  dbtx/                      Transaction context helper
```

### The Rules (linter-enforced)

- `domain/` imports nothing. No `database/sql`. No `net/http`. No `pkg/`. Nothing.
- `adapter/in/` never imports `adapter/out/` (and vice versa)
- Only `cmd/` knows all concrete types

Break a rule → lint fails → CI blocks merge.

---

## Observability (Zero-Code)

You don't instrument your code. The build tool does it for you.

### How

```makefile
# This is the only change:
build:
    otelc go build -o bin/app ./main.go
```

[`otelc`](https://opentelemetry.io/blog/2026/go-compile-time-instrumentation-v1/) (CNCF, v1 July 2026) injects OpenTelemetry instrumentation at compile time. Your source code stays clean.

### What You Get For Free

| Traced automatically | How |
|---------------------|-----|
| Every HTTP request | `net/http` instrumented at build time |
| Every SQL query | `database/sql` instrumented at build time |
| Every Redis command | `go-redis/v9` instrumented at build time |
| Every gRPC call | `google.golang.org/grpc` instrumented at build time |
| Log → Trace correlation | `log/slog` gets `trace_id` injected automatically |
| Go runtime metrics | Goroutines, heap, GC — collected by default |

### What You Add (~10 lines per usecase)

Business-level spans: "which usecase ran, how long, did it fail?"

```go
// internal/adapter/out/observability/usecase_decorators.go
func (d *createProductObs) Execute(ctx context.Context, cmd port.CreateProductCommand) (port.ProductResult, error) {
    ctx, end := StartOp(ctx, "usecase.CreateProduct", A("name", cmd.Name), A("price", cmd.Price))
    result, err := d.inner.Execute(ctx, cmd)
    end(err, A("id", result.ID))
    return result, err
}
```

Wire it in `cmd/`:
```go
createProduct = obsadapter.WrapCreateProduct(createProduct)
```

### The Result

```
[HTTP: POST /products]                 ← auto (otelc)
  └── [usecase.CreateProduct]          ← your decorator (10 lines)
        ├── [db.query INSERT INTO...]  ← auto (otelc)
        └── [redis SET product:...]    ← auto (otelc)
```

### Toggle

| Want | Do |
|------|------|
| Disable all telemetry | `OTEL_SDK_DISABLED=true` (runtime, no rebuild) |
| Build without instrumentation | `go build` instead of `otelc go build` |
| Change sampling | `OTEL_TRACES_SAMPLER_ARG=0.1` (10% in prod) |
| Switch backend | Change collector config (Jaeger, Tempo, Datadog, etc.) |

### Limitations

- `hibiken/asynq` not auto-instrumented (fire-and-forget, <1ms — acceptable)
- Adapter span names follow OTel semantic conventions (`db.query`, not `postgres.FindByID`)
- Requires OTel Collector in deployment (standard infra for production observability)
- otelc adds ~2-3s to build (dev uses `make build-plain` or Air for hot-reload)
- SQL text may appear in spans — sanitize at collector if PII is a concern

---

## Resilience

All resilience lives in adapters. Domain never knows about retries or circuit breakers.

| Pattern | Where | What it does |
|---------|-------|--------------|
| Circuit breaker | Redis cache | Opens on repeated failures, returns cache-miss instead of failing the request |
| Singleflight | Postgres FindByID | Deduplicates concurrent requests for the same ID |
| Retry | Postgres (transient errors) | Retries connection resets with exponential backoff |
| Timeout | Every adapter call | Context timeout prevents hanging requests |

---

## API

```bash
POST   /products                    # Create product
GET    /products/{id}               # Get product (cache → DB fallback)
GET    /products?limit=10&cursor=x  # List (cursor pagination)
PUT    /products/{id}               # Update product
DELETE /products/{id}               # Delete product
POST   /products/{id}/discount      # Apply discount (max 50%)

GET    /healthz                     # Liveness
GET    /readyz                      # Readiness (DB + Redis)
```

Response envelope:
```json
{"status":"success","data":{...},"meta":{"request_id":"abc","timestamp":1723067890000}}
```

---

## Commands

```bash
myservice serve http       # HTTP on :8080
myservice serve grpc       # gRPC on :8081
myservice worker           # asynq background processor
myservice migrate up       # Run migrations
myservice migrate down     # Rollback
```

---

## Make Targets

```bash
make dev                   # Hot reload (Air, plain build)
make build                 # Build with otelc instrumentation
make build-plain           # Build without instrumentation (fast)
make run                   # Build + run HTTP
make test-unit             # Domain tests
make test-integration      # Adapter tests (Docker required)
make test-e2e              # End-to-end
make test-all              # All + coverage
make lint                  # golangci-lint
make fmt                   # gofumpt + goimports
make mocks                 # mockgen
make docker-up             # Start Postgres + Redis + Collector + Jaeger
make docker-down           # Stop everything
```

---

## Customize for Your Team

**Simple CRUD** — strip what you don't need:
- Remove `adapter/in/grpc/`, `adapter/in/worker/`
- Remove `adapter/out/asynq/`, `domain/event/`

**Complex service** — add as needed:
- `adapter/in/grpc/` for gRPC
- `adapter/in/worker/` for background jobs
- `adapter/out/redis/` for caching
- More usecases → more decorators (10 lines each)

---

## Tech Stack

| Purpose | Choice | Why |
|---------|--------|-----|
| HTTP | `net/http` stdlib | Zero overhead, Go 1.22 routing |
| JSON | `bytedance/sonic` | Fastest in Go (JIT) |
| DB | `database/sql` + `pgx/v5` | Fast wire protocol, swappable |
| Cache | `redis/go-redis/v9` | De facto standard |
| Async | `hibiken/asynq` | Simple Redis-backed queues |
| gRPC | `google.golang.org/grpc` | Standard |
| Observability | `otelc` + OTel SDK | Zero-code, compile-time |
| Logging | `log/slog` stdlib | Structured, trace-correlated via otelc |
| Resilience | `sony/gobreaker` + `x/sync` | Circuit breaker + singleflight |
| Validation | `go-playground/validator` | Tag-based, on port DTOs |
| IDs | `oklog/ulid` | Sortable, no DB seq |
| CLI | `cobra` | Multi-command binary |
| Testing | `testify` + `testcontainers-go` | Assertions + real Docker deps |

---

## Requirements

- Go 1.26+
- Docker
- Make
- `otelc` (installed via `make init`)

---

## License

MIT
