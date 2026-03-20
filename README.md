# cdn-control

EdgeFlow CDN control plane — configuration management and API service.

## Features

- REST API (Gin) for domain, origin, cache rule, certificate, and node management
- PostgreSQL storage with auto-migration
- JWT authentication with bcrypt passwords
- Admin/user role-based access control
- Cache purge task management (URL, directory, full site)
- Cache prefetch API
- Certificate PEM parsing and storage
- gRPC server for edge node communication:
  - Config push (streaming updates)
  - Purge command dispatch
  - Node heartbeat collection

## Quick Start

```bash
# Prerequisites: PostgreSQL running on localhost:5432

# Build
go build -o bin/cdn-control ./cmd

# Run
./bin/cdn-control -config configs/control-config.yaml

# Docker
docker build -t cdn-control .
```

Default admin credentials: `admin` / `admin123`

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | /api/v1/auth/login | Login, get JWT token |
| CRUD | /api/v1/domains | Domain management |
| CRUD | /api/v1/domains/:id/origins | Origin management |
| CRUD | /api/v1/domains/:id/cache-rules | Cache rules |
| POST | /api/v1/purge/url | Purge by URL |
| POST | /api/v1/purge/dir | Purge by directory |
| POST | /api/v1/purge/all | Purge entire domain |
| POST | /api/v1/prefetch | Cache prefetch |
| CRUD | /api/v1/domains/:id/certs | Certificate management |
| GET | /api/v1/nodes | Node list |

## Testing

```bash
# Unit tests (no DB required)
go test ./middleware/... -v

# Integration tests (requires PostgreSQL)
TEST_DATABASE_URL="postgres://user:pass@localhost:5432/edgeflow_test?sslmode=disable" go test ./handler/... -v
```
