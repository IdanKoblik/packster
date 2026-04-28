# Packster

A self-hosted REST API and web UI for managing versioned build artifacts.
Import projects from GitLab, organize them into products, upload versions, and
share access with other users via fine-grained permissions.

## Requirements
* Go 1.26+
* PostgreSQL 17
* Node.js (for building the web UI)

## Installation

### From Source

``` shell
git clone https://github.com/IdanKoblik/packster.git
cd packster
go mod tidy
make build
```

`make build` also builds the React web UI and embeds it into the binary. The
output binary is at `bin/packster`.

### Docker Compose
Spin up a local PostgreSQL instance (with the schema pre-loaded from
`scheme.sql`):

``` shell
docker compose up -d
```

## Configuration
Create a YAML config file (see `fixtures/example.yml` for a minimal example):

``` yaml
file_upload_limit: 20    # Max upload size in MB
secret: "change-me"      # Secret used to sign session tokens

sql:
  host: "localhost"
  port: 5432
  db: "packster"
  user: "root"
  password: "root"
  ssl: false

storage:
  path: "./data"         # Directory where uploaded artifacts are stored

# (Optional) GitLab SSO. One or more hosts may be configured.
gitlab:
  hosts:
    "https://gitlab.com":
      id: "YOUR_APPLICATION_ID"
      secret: "YOUR_SECRET"
```

Each GitLab host listed under `gitlab.hosts` must already exist in the `host`
table in the database for it to be loaded at startup.

## Running
Set the `CONFIG_PATH` environment variable to point at your config file:

``` shell
CONFIG_PATH=./config.yml ./bin/packster
```

The server listens on `0.0.0.0:8080` by default. Override the address with the
`ADDR` environment variable.

## Environment Variables

| Variable      | Description                                              |
|---------------|----------------------------------------------------------|
| `CONFIG_PATH` | Path to the YAML config file (required)                  |
| `ADDR`        | Listen address (default `0.0.0.0:8080`)                  |

## API
All endpoints are mounted under `/api`.

| Method | Path                                                          | Description                    |
|--------|---------------------------------------------------------------|--------------------------------------|
| GET    | `/health`                                                     | Health check                    |
| GET    | `/hosts`                                                      | List configured SSO hosts            |
| GET    | `/auth/gitlab/redirect`                                       | Start GitLab OAuth flow              |
| GET    | `/auth/gitlab/callback`                                       | GitLab OAuth callback                |
| GET    | `/auth/session`                                               | Current session info                 |
| GET    | `/user/candidates`                                            | List candidate users                 |
| GET    | `/user/projects`                                              | List the user's imported projects    |
| POST   | `/user/projects`                                              | Import a project                    |
| DELETE | `/projects/:id`                                               | Delete a project                    |
| GET    | `/projects/:id/permissions`                                   | List project permissions             |
| PUT    | `/projects/:id/permissions`                                   | Set a permission                    |
| DELETE | `/projects/:id/permissions/:user_id`                          | Revoke a permission                  |
| GET    | `/projects/:id/permissions/candidates`                        | Search users forpermission grants   |
| GET    | `/projects/:id/products`                                      | List products ina project           |
| POST   | `/projects/:id/products`                                      | Create a product                    |
| DELETE | `/projects/:id/products/:product_id`                          | Delete a product                    |
| GET    | `/products/:product_id/versions`                              | List versions ofa product           |
| POST   | `/products/:product_id/versions`                              | Upload a new version                 |
| GET    | `/versions/:version_id`                                       | Download a version by id             |
| DELETE | `/versions/:version_id`                                       | Delete a version                    |
| GET    | `/projects/:id/products/:product_name/versions/:version_name` | Download a version by name           |

## Development

* `make ui` — install web deps and build the UI into `internal/ui/static`
* `make build` — build UI then compile the Go binary into `bin/packster`
* `make run` — build and run
* `make test` — run all Go tests
* `make cover-integration` — run integration tests with coverage report
* `make clean` — remove build artifacts
