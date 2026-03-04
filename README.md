# artifactor 📦

Robust package version management service designed to simplify the way developers handle dependencies across projects

> ⚠ THIS PROJECT IS UNDER DEVELOPMENT, THERE ISN'T ANY OFFICIAL RELEASE FOR NOW.
>

## Installation

### Building
```bash
git clone https://github.com/IdanKoblik/artifactor.git
cd artifactor/

go mod tidy
make build
```

### Environment variables

| Variable    | Meaning                                                           |
|-------------|-------------------------------------------------------------------|
| CONFIG_PATH | Path to the config path, including file name (path/to/config.yml) |

### Config

Config example:
```yaml
file_upload_limit: 0 # MB (opt) 

sql:
  username: "username"
  password: "password"
  addr: "localhost:5173"
  database: "db"
```
