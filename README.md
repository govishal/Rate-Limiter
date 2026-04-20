# Rate-Limited API Service (Go)

Assignment scope: **`POST /request`** with `{ "user_id", "payload" }`, **`GET /stats`** for per-user counters, **max 5 requests per user per minute**, **429** when exceeded, **in-memory** storage, **correct under concurrency** (`sync.Mutex`).

## Run

**Go 1.18+** required.

```bash
go run ./cmd/api
```

Copy **`.env.example`** to **`.env`** in the repo root and save it. **`config.LoadConfig()`** calls **`godotenv.Load()`** — run from the repo root so **`.env`** is found. **`.env`** is gitignored.

## API

### `POST /request`

```bash
curl -X POST http://localhost:8080/request \
  -H "Content-Type: application/json" \
  -d '{"user_id":"user-1","payload":{"message":"hello"}}'
```

**200** example:

```json
{
  "message": "request accepted",
  "user_id": "user-1",
  "window_count": 1,
  "payload": { "message": "hello" }
}
```

**429** (body shape; text reflects your `RATE_LIMIT_*` env):

```json
{
  "error": "rate limit exceeded: max 5 requests per user per 1m0s",
  "retry_after": "60s",
  "user_id": "user-1",
  "window_count": 5
}
```

### `GET /stats`

```bash
curl http://localhost:8080/stats
```

**200** example:

```json
{
  "window_seconds": 60,
  "max_requests": 5,
  "users": [
    {
      "user_id": "user-1",
      "accepted_requests": 5,
      "rejected_requests": 2,
      "last_request_at": "2026-04-20T15:30:00Z",
      "current_window_request_count": 5
    }
  ]
}
```

## Design decisions

* Implemented a **sliding window rate limiting algorithm** to ensure accurate, time-based request control and avoid burst issues at window boundaries
* Used **in-memory storage** for simplicity and fast access
* Ensured **concurrency safety** using per-user locking to handle parallel requests correctly
* Exposed clean REST APIs for request handling and statistics

## What I would improve with more time

* Introduce **Redis** for distributed rate limiting across multiple instances
* Add **persistent storage** for request statistics
* Implement **configurable rate limits per user/tier**
* Improve **observability** with structured logging, metrics, and monitoring
* Add **unit and load tests** for reliability
* Containerize and deploy using cloud infrastructure (e.g., Docker + Azure)
