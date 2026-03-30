# Installation

## Installation

Clone the repository and build the binary:

```bash
git clone https://github.com/IdanKoblik/packster
cd packster
make build
```

The binary is written to `bin/packster`.

## Configuration

Packster is configured with a YAML file. Point to it with the `CONFIG_PATH` environment variable:

```bash
CONFIG_PATH=/etc/packster/config.yml ./bin/packster
```

**Example config:**

```yaml
file_upload_limit: 100  # MB, maximum size of an uploaded artifact

mysql:
  dsn: "root:root@tcp(localhost:3306)/packster?parseTime=true"

redis:
  addr: localhost:6379
  password: ""
  db: 0

metrics:
  addr: 0.0.0.0:9091  # Prometheus scrape endpoint (optional, defaults to 0.0.0.0:9091)
```

| Field | Description |
|---|---|
| `file_upload_limit` | Max upload size in MB |
| `mysql.dsn` | MySQL Data Source Name (e.g. `user:pass@tcp(host:3306)/dbname?parseTime=true`) |
| `redis.addr` | Redis address (`host:port`) |
| `redis.password` | Redis password (optional) |
| `redis.db` | Redis database index (optional) |
| `metrics.addr` | Address for the Prometheus `/metrics` endpoint (optional, defaults to `0.0.0.0:9091`) |

## Running

Start the server:

```bash
CONFIG_PATH=./config.yml ./bin/packster
```

To also enable the web UI:

```bash
CONFIG_PATH=./config.yml ./bin/packster --ui
```

By default the server listens on `0.0.0.0:8080`. Override with `SERVER_ADDR`:

```bash
SERVER_ADDR=0.0.0.0:9090 CONFIG_PATH=./config.yml ./bin/packster
```

**Starting dependencies with Docker Compose:**

A `docker-compose.yml` is included to spin up MySQL, Redis, Prometheus, and Grafana locally:

```bash
docker compose up -d
```

This starts:

| Service | Port | Notes |
|---|---|---|
| MySQL | 3306 | Primary data store |
| Redis | 6379 | Token cache |
| Prometheus | 9090 | Scrapes `host.docker.internal:9091` every 15 s |
| Grafana | 3000 | Pre-provisioned dashboard at `http://localhost:3000` (admin / admin) |

Prometheus scrapes the `/metrics` endpoint exposed by the running packster binary on port `9091`. Make sure packster is running before expecting data in Grafana.

## Flags

| Flag | Description |
|---|---|
| `--init-admin-token` | Generates an initial admin token on first run. No-op if an admin token already exists. |
| `--ui` | Enables the web UI served at `/ui`. Requires an admin token to log in. |

## Authentication

Every request must include an `X-Api-Token` header with a valid token.

There are two token types:

- **Admin** — can register and manage tokens, create/delete products, and perform all operations.
- **Non-admin** — access is controlled per-product through token permissions (`upload`, `download`, `delete`, `maintainer`).

### Creating the first admin token

On first run, pass `--init-admin-token` to generate an initial admin token:

```bash
CONFIG_PATH=./config.yml ./bin/packster --init-admin-token
```

The token is printed to the log output. Remove the flag after the first use — it is a no-op if an admin token already exists.

Use this token in the `X-Api-Token` header for all subsequent admin operations, such as registering additional tokens via `PUT /api/register`.

## Product Groups

Products support an optional `group_name` field, allowing the same product name to exist in multiple groups (e.g. `staging`, `production`, `test`):

```bash
# Create products in different groups
curl -X PUT http://localhost:8080/api/product/create \
  -H "X-Api-Token: <token>" \
  -H "Content-Type: application/json" \
  -d '{"name": "my-app", "group_name": "staging"}'

curl -X PUT http://localhost:8080/api/product/create \
  -H "X-Api-Token: <token>" \
  -H "Content-Type: application/json" \
  -d '{"name": "my-app", "group_name": "production"}'

# Fetch with group
curl http://localhost:8080/api/product/fetch/my-app?group=staging \
  -H "X-Api-Token: <token>"
```

Endpoints that accept a group:

| Endpoint | How to pass group |
|---|---|
| `PUT /api/product/create` | `group_name` in JSON body |
| `GET /api/product/fetch/{product}` | `?group=` query param |
| `DELETE /api/product/delete/{product}` | `?group=` query param |
| `POST /api/product/upload` | `group_name` in multipart form |
| `GET /api/product/download/{product}/{version}` | `?group=` query param |
| `DELETE /api/product/delete/{product}/{version}` | `?group=` query param |
| `POST /api/product/modify/{action}` | `group_name` in JSON body |
