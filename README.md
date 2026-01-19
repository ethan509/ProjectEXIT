# ProjectEXIT - Go Server Template

This is a server template project generated for ProjectEXIT.

> **📖 새로 시작하시나요?** [SETUP.md](./SETUP.md)에서 개발 환경 설정 가이드를 확인하세요.
> **🗄️ 데이터베이스 연결?** [DB_INTEGRATION.md](./docs/DB_INTEGRATION.md)에서 PostgreSQL, MySQL, MongoDB 통합 가이드를 확인하세요.

## Features

- Graceful shutdown (max 1 minute)
- chi router
- Structured logging with per-level files and basic rotation (daily + ~1GB split)
- Request-scoped transaction ID (X-Request-Id)
- Concurrency limiting middleware
- Per-request timeout middleware
- Worker pool with channel-based communication
- Simple scheduler with daily/monthly/yearly example jobs
- JSON config with hot reload (for selected fields)
- Health (`/healthz`), readiness (`/readyz`) and metrics (`/metrics`) endpoints
- Example handlers and tests
- Docker support with multi-stage build

## Quick Start

### Prerequisites

- Go 1.24 or later
- Docker and Docker Compose (for containerized deployment)

### Run Locally

```bash
# Build and run
make build
./server

# Or run directly
make run
```

### Run with Docker

```bash
# Build and start all services (app + PostgreSQL)
make docker-up

# View logs
make docker-logs

# Stop services
make docker-down
```

### Development Mode

```bash
# Start with hot reload
make dev-up
```

## Docker

All Docker-related files are located in the `docker/` directory:
- `docker/Dockerfile` - Multi-stage build configuration
- `docker/docker-compose.yml` - Production environment
- `docker/docker-compose.dev.yml` - Development environment
- `docker/.dockerignore` - Build optimization
- `docker/DOCKER.md` - Detailed Docker guide

### Services

- **app**: ProjectEXIT server (port 8080)
- **db**: PostgreSQL 15 (port 5432)

### Available Make Commands

```bash
make help              # Show all available commands
make build             # Build Go binary
make run               # Run locally
make test              # Run tests
make docker-build      # Build Docker image
make docker-up         # Start containers
make docker-down       # Stop containers
make docker-logs       # View app logs
make docker-ps         # Show container status
make docker-restart    # Restart app container
make docker-clean      # Remove all containers, images, volumes
make dev-up            # Start development environment
make dev-down          # Stop development environment
```

## API Endpoints

- `GET /healthz` - Health check (always 200 if alive)
- `GET /readyz` - Readiness check (503 if not ready)
- `GET /metrics` - Prometheus metrics
- `GET /api/v1/ping` - Simple ping endpoint
- `POST /api/v1/echo` - Echo request body with worker processing

## Configuration

Edit `config/config.json` to customize:
- Server timeouts
- Concurrency limits
- Worker pool sizes
- Log level
- Scheduler timezone

Hot reload fields (reloaded every 10 minutes):
- `readTimeoutSec`, `writeTimeoutSec`, `idleTimeoutSec`
- `requestTimeoutSec`, `maxRequestBodyBytes`
- `logging.level`
