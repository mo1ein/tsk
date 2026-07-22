# Task Manager

A microservices task manager API built with Go, Gin, PostgreSQL, and Redis.

## Architecture

```mermaid
flowchart TB
    Client[Client / Frontend] --> App[API Handlers]
    App --> Service[Service]
    Service --> Repo[Repository]

    Repo --> Cache{Found in Redis?}

    Cache -->|Yes| Redis[(Redis cache)]
    Cache -->|No| DB[(PostgreSQL)]
    DB --> Redis[(Redis cache)]

    App --> Metrics[Prometheus Metrics]

    subgraph "Task Manager"
        App
        Service
        Repo
    end
```

## Quick Start

### Using Docker Compose (Recommended)

```bash
# Clone the repository
git clone https://github.com/graph/tsk.git
cd tsk

# Create .env from example (if not exists)
make env

# Build and start all services
make build-up

# Check status
make status
```

The API will be available at `http://localhost:8000`.
The swagger is at the `http://0.0.0.0:8000/swagger/index.html`

### Local Development

```bash
# Prerequisites
# - Go 1.24+
# - PostgreSQL 16+
# - Redis 7+

# Start database and redis
docker compose up -d postgres redis

# Create .env and run server
make env
go run . serve
```

## API Examples

### Create a Task

```bash
curl -X POST http://localhost:8000/tasks \
  -H "Content-Type: application/json" \
  -d '{"title": "Buy groceries", "assignee": "alice"}'
```

Response:
```json
{
  "id": 1,
  "title": "Buy groceries",
  "assignee": "alice",
  "status": "pending",
  "created_at": "2025-07-20T12:00:00Z",
  "updated_at": "2025-07-20T12:00:00Z"
}
```

### List Tasks with Filtering

```bash
# List all tasks
curl http://localhost:8000/tasks

# Filter by status
curl "http://localhost:8000/tasks?status=pending"

# Filter by assignee
curl "http://localhost:8000/tasks?assignee=alice"

# Pagination
curl "http://localhost:8000/tasks?page=2&page_size=10"
```

Response:
```json
{
  "tasks": [...],
  "total": 42
}
```

### Get Task by ID

```bash
curl http://localhost:8000/tasks/1
```

### Update Task

```bash
curl -X PUT http://localhost:8000/tasks/1 \
  -H "Content-Type: application/json" \
  -d '{"status": "done"}'
```

### Delete Task

```bash
curl -X DELETE http://localhost:8000/tasks/1
```

### Health Check

```bash
curl http://localhost:8000/health
```

## Testing

```bash
# Run all tests
go test ./... -v

# Run with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run specific test
go test ./internal/repository/ -v -run TestMockTaskRepository
```

## Load Testing

```bash
# Run benchmark script
./scripts/benchmark.sh

# Or use Apache Bench directly
ab -n 1000 -c 50 http://localhost:8000/tasks
```

## Monitoring

- **Prometheus**: http://localhost:9090
- **Application Metrics**: http://localhost:8000/metrics

### Available Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `http_requests_total` | Counter | Total HTTP requests |
| `http_request_latency_seconds` | Histogram | Request latency |
| `tasks_count` | Gauge | Total tasks |
