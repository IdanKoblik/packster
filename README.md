# Packster

A self-hosted REST API service for manging versioned build artifacts. 
Store, retrieve and delete your build outputs with git based sso.

## Requirements
* Go 1.25+
* Pgsql 17

## Installation

### From Source

``` shell
git clone https://github.com/IdanKoblik/packster.git
cd packster
go mod tidy
make build
```

The binary will be at `bin/packster`.

### Docker compose
Spin up MongoDB and Redis quickly:

``` shell
docker compose up -d
```

## Configuration
Create a YAML config file (see fixtures/example.yml for reference):

``` yaml
file_upload_limit: 20 # Max upload size in MB

sql:
  host: "localhost"
  port: 5432
  db: "postgres"
  user: "root"
  password: "root"

# (Optional)
gitlab:
  host: "https://gitlab.com"
  application_id: "YOUR_APPLICATION_ID"
  secert: "YOUR_SECRET"
```

## Running
Set the CONFIG_PATH environment variable to point to your config file:

``` shell
CONFIG_PATH=./config.yml ./bin/packster
```

## Environment Variables

| Variable      | Description                                              |
|---------------|----------------------------------------------------------|
| `CONFIG_PATH` | Path to the YAML config file (required)                  |
