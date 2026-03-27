# Installation

## Installation

Clone the repository and build the binary:

```bash
git clone https://github.com/IdanKoblik/artifactor
cd artifactor
make build
```

The binary is written to `bin/artifactor`.

## Configuration

Artifactor is configured with a YAML file. Point to it with the `CONFIG_PATH` environment variable:

```bash
CONFIG_PATH=/etc/artifactor/config.yml ./bin/artifactor
```

**Example config:**

```yaml
file_upload_limit: 100  # MB, maximum size of an uploaded artifact
jwt_secret: "change-me"  # Secret key used to sign JWT tokens (required)

mongo:
  connection_string: mongodb://localhost:27017
  database: artifactor
  token_collection: tokens
  product_collection: products

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
| `jwt_secret` | Secret key used to sign and verify JWT tokens (required) |
| `mongo.connection_string` | MongoDB connection URI |
| `mongo.database` | Database name |
| `mongo.token_collection` | Collection used to store tokens |
| `mongo.product_collection` | Collection used to store products |
| `redis.addr` | Redis address (`host:port`) |
| `redis.password` | Redis password (optional) |
| `redis.db` | Redis database index (optional) |
| `metrics.addr` | Address for the Prometheus `/metrics` endpoint (optional, defaults to `0.0.0.0:9091`) |

## Running

Start the server:

```bash
CONFIG_PATH=./config.yml ./bin/artifactor
```

To also enable the web UI:

```bash
CONFIG_PATH=./config.yml ./bin/artifactor --ui
```

By default the server listens on `0.0.0.0:8080`. Override with `SERVER_ADDR`:

```bash
SERVER_ADDR=0.0.0.0:9090 CONFIG_PATH=./config.yml ./bin/artifactor
```

**Starting dependencies with Docker Compose:**

A `docker-compose.yml` is included to spin up MongoDB, Redis, Prometheus, and Grafana locally:

```bash
docker compose up -d
```

This starts:

| Service | Port | Notes |
|---|---|---|
| MongoDB | 27017 | Primary data store |
| Redis | 6379 | Token cache |
| Prometheus | 9090 | Scrapes `host.docker.internal:9091` every 15 s |
| Grafana | 3000 | Pre-provisioned dashboard at `http://localhost:3000` (admin / admin) |

Prometheus scrapes the `/metrics` endpoint exposed by the running artifactor binary on port `9091`. Make sure artifactor is running before expecting data in Grafana.

## Flags

| Flag | Description |
|---|---|
| `--init-admin-token` | Generates an initial admin token on first run. No-op if an admin token already exists. |
| `--ui` | Enables the web UI served at `/ui`. Requires an admin token to log in. |

## Authentication

Every request must include an `X-Api-Token` header with a valid JWT token. Tokens are signed with the `jwt_secret` from your config and encode the token's unique identifier.

There are two token types:

- **Admin** — can register and manage tokens, create/delete products, and perform all operations.
- **Non-admin** — access is controlled per-product through token permissions (`upload`, `download`, `delete`, `maintainer`).

### Creating the first admin token

On first run, pass `--init-admin-token` to generate an initial admin token:

```bash
CONFIG_PATH=./config.yml ./bin/artifactor --init-admin-token
```

The JWT token is printed to the log output. Remove the flag after the first use — it is a no-op if an admin token already exists.

Use this token in the `X-Api-Token` header for all subsequent admin operations, such as registering additional tokens via `PUT /api/register`.
