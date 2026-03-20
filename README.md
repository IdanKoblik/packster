# Artifactor

A self-hosted REST API service for managing versioned build artifacts. Store, retrieve, and control access to your build outputs with token-based authentication and per-product permissions.

## Requirements

- Go 1.25+
- MongoDB 7.0+
- Redis 7+

## Installation

### From Source

```bash
git clone https://github.com/IdanKoblik/artifactor.git
cd artifactor
go mod tidy
make build
```

The binary will be at `bin/artifactor`.

### Docker Compose (Dependencies)

Spin up MongoDB and Redis quickly:

```bash
docker compose up -d
```

## Configuration

Create a YAML config file (see `fixtures/example.yml` for reference):

```yaml
file_upload_limit: 20              # Max upload size in MB

mongo:
  connection_string: "mongodb://localhost:27017/"
  database: "artifactor"
  token_collection: "tokens"
  product_collection: "products"

redis:
  addr: "localhost:6379"
  password: ""
  db: 0
```

## Running

Set the `CONFIG_PATH` environment variable to point to your config file:

```bash
CONFIG_PATH=./config.yml ./bin/artifactor
```

To listen on a custom address (default is `0.0.0.0:8080`):

```bash
SERVER_ADDR=0.0.0.0:9090 CONFIG_PATH=./config.yml ./bin/artifactor
```

### Environment Variables

| Variable      | Description                                              |
|---------------|----------------------------------------------------------|
| `CONFIG_PATH` | Path to the YAML config file (required)                  |
| `SERVER_ADDR` | Server listen address (default: `0.0.0.0:8080`)         |

### First-Time Setup

On first run, generate an initial admin token:

```bash
CONFIG_PATH=./config.yml ./bin/artifactor --init-admin-token
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
