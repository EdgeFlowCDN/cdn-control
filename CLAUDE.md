# cdn-control

EdgeFlow CDN control plane — configuration management REST API + gRPC service.

## Tech Stack

Go 1.23, Gin, PostgreSQL (pgx), JWT, gRPC, bcrypt

## Project Structure

```
cmd/          Entry point (main.go) with graceful shutdown
config/       YAML config loading
db/           PostgreSQL connection and migrations (embedded SQL)
model/        Data models and request/response types
handler/      Gin HTTP handlers
  router.go       Route setup with JWT middleware
  auth.go         Login + admin user init
  domain.go       Domain CRUD
  origin.go       Origin CRUD (sub-resource of domain)
  cache_rule.go   Cache rule CRUD
  purge.go        Purge/prefetch task management
  cert.go         Certificate upload/list/delete
  node.go         Node list/detail/status update
middleware/   JWT auth and password hashing
grpc/         gRPC server for edge node communication
  proto.go        Manual gRPC type definitions and registration
  service.go      Config push, purge dispatch, heartbeat collection
configs/      YAML config files
```

## Database Tables

domains, origins, cache_rules, certificates, nodes, purge_tasks, users

## API Authentication

JWT Bearer token. Default admin: `admin` / `admin123`.
All `/api/v1/*` routes require auth except `/api/v1/auth/login`.

## Running Tests

```bash
# Unit tests (no DB)
go test ./middleware/... -v

# Integration tests (needs PostgreSQL)
TEST_DATABASE_URL="postgres://..." go test ./handler/... -v
```
