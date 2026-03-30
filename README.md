# Packster

A self-hosted REST API service for managing versioned build artifacts. Store, retrieve, and control access to your build outputs with token-based authentication and per-product permissions.

## Requirements

- Go 1.25+
- MySQL 8.0+
- Redis 7+

## Installation

### From Source

```bash
git clone https://github.com/IdanKoblik/packster.git
cd packster
go mod tidy
make build
```

The binary will be at `bin/packster`.

### Docker Compose (Dependencies)

Spin up MySQL and Redis quickly:

```bash
docker compose up -d
```

## Configuration

Create a YAML config file (see `fixtures/example.yml` for reference):

```yaml
file_upload_limit: 20              # Max upload size in MB

mysql:
  dsn: "root:root@tcp(localhost:3306)/packster?parseTime=true"

redis:
  addr: "localhost:6379"
  password: ""
  db: 0

metrics:
  addr: "0.0.0.0:9091"
```

## Running

Set the `CONFIG_PATH` environment variable to point to your config file:

```bash
CONFIG_PATH=./config.yml ./bin/packster
```

To listen on a custom address (default is `0.0.0.0:8080`):

```bash
SERVER_ADDR=0.0.0.0:9090 CONFIG_PATH=./config.yml ./bin/packster
```

### Environment Variables

| Variable      | Description                                              |
|---------------|----------------------------------------------------------|
| `CONFIG_PATH` | Path to the YAML config file (required)                  |
| `SERVER_ADDR` | Server listen address (default: `0.0.0.0:8080`)         |

### First-Time Setup

On first run, generate an initial admin token:

```bash
CONFIG_PATH=./config.yml ./bin/packster --init-admin-token
```

This prints an admin token to the logs. Save it — you'll use it in the `X-Api-Token` header to perform admin operations. Remove the flag after the first run; it has no effect once an admin token exists.

## Usage

All requests require the `X-Api-Token` header.

For example:
```bash
curl -X PUT http://localhost:8080/api/product/create \
  -H "X-Api-Token: <admin-token>" \
  -H "Content-Type: application/json" \
  -d '{"name": "my-app"}'
```

### Product Groups

Products can be organized into groups, allowing the same product name to exist in multiple environments (e.g. `staging`, `production`, `test`):

```bash
curl -X PUT http://localhost:8080/api/product/create \
  -H "X-Api-Token: <admin-token>" \
  -H "Content-Type: application/json" \
  -d '{"name": "my-app", "group_name": "staging"}'
```

Fetch, delete, download, and delete-version endpoints accept an optional `?group=` query parameter. Upload and modify endpoints accept `group_name` in the request body.

## Permissions

Non-admin tokens have no product access by default. A product maintainer or admin must grant permissions explicitly. Available permissions: `upload`, `download`, `delete`, `maintainer`.

## Development

```bash
make build        # Build binary
make run          # Build and run
make test         # Run all tests
make test-unit    # Run unit tests only
make cover        # Generate coverage report
```
