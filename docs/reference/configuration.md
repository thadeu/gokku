# Configuration Reference

Complete reference for `gokku.yml` configuration file.

## File Location

`gokku.yml` should be in your project root:

```
my-project/
├── gokku.yml       ← Here
├── cmd/
├── go.mod
└── ...
```

## Schema Overview

```yaml
project:          # Global project settings
defaults:         # Default values for all apps
apps:             # Application definitions
  - name:         # App name
    lang:         # Programming language
    build:        # Build configuration
    environments: # Deployment environments
    deployment:   # Deployment settings
docker:           # Global Docker settings
user:             # Server user settings
```

## Full Reference

### project

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `name` | string | ✅ Yes | - | Project name |
| `base_dir` | string | ❌ No | `/opt/gokku` | Installation directory on server |

**Example:**
```yaml
project:
  name: my-awesome-project
  base_dir: /opt/gokku
```

### defaults

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `build_type` | string | ❌ No | `systemd` | Default build type: `systemd` or `docker` |
| `lang` | string | ❌ No | `go` | Default language: `go`, `python`, `nodejs`, etc |

**Example:**
```yaml
defaults:
  build_type: systemd
  lang: go
```

### apps[]

Array of application definitions.

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `name` | string | ✅ Yes | - | Application name (must be unique) |
| `lang` | string | ❌ No | From `defaults.lang` | Programming language |
| `build` | object | ✅ Yes | - | Build configuration (see below) |
| `environments` | array | ❌ No | `[{name: "production", branch: "main"}]` | Deployment environments |
| `deployment` | object | ❌ No | See defaults | Deployment settings |

**Example:**
```yaml
apps:
  - name: api
    lang: go
    build:
      path: ./cmd/api
```

### apps[].build

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `type` | string | ❌ No | From `defaults.build_type` | Build type: `systemd` or `docker` |
| `path` | string | ✅ Yes | - | Path to app code (relative to project root) |
| `binary_name` | string | ❌ No | Same as `app.name` | Output binary name (Go only) |
| `work_dir` | string | ❌ No | `.` | Working directory for build |
| `go_version` | string | ❌ No | `1.25` | Go version (Go only) |
| `goos` | string | ❌ No | `linux` | Target OS (Go only) |
| `goarch` | string | ❌ No | `amd64` | Target architecture (Go only) |
| `cgo_enabled` | int | ❌ No | `0` | Enable CGO: `0` or `1` (Go only) |
| `dockerfile` | string | ❌ No | - | Custom Dockerfile path (Docker only) |
| `entrypoint` | string | ❌ No | Language-specific | Entrypoint file (non-Go) |
| `base_image` | string | ❌ No | From `docker.base_images` | Base Docker image |
| `mise` | object | ❌ No | - | Mise/asdf configuration (see below) |

**Entrypoint Defaults:**
- Python: `main.py`
- Node.js: `index.js`
- Ruby: `app.rb`

**Example (Go + systemd):**
```yaml
build:
  type: systemd
  path: ./cmd/api
  binary_name: api
  go_version: "1.25"
  cgo_enabled: 0
```

**Example (Python + Docker):**
```yaml
build:
  type: docker
  path: ./services/ml
  entrypoint: server.py
  base_image: python:3.11-slim
```

### apps[].build.mise

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `plugins` | array | ❌ No | `[]` | Mise/asdf plugins to install |

**plugins[] fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | ✅ Yes | Plugin name |
| `url` | string | ✅ Yes | Plugin repository URL |

**Example:**
```yaml
build:
  mise:
    plugins:
      - name: whispercpp
        url: https://github.com/thadeu/asdf-whispercpp.git
      - name: ffmpeg
        url: https://github.com/acj/asdf-ffmpeg.git
```

### apps[].environments[]

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `name` | string | ✅ Yes | - | Environment name (e.g., `production`, `staging`) |
| `branch` | string | ❌ No | Smart default¹ | Git branch for this environment |
| `default_env_vars` | object | ❌ No | `{}` | Default environment variables |

¹ **Branch Defaults:**
- `production` → `main`
- `staging` → `staging`
- `develop` → `develop`
- Other → Same as environment name

**Example:**
```yaml
environments:
  - name: production
    branch: main
    default_env_vars:
      LOG_LEVEL: info
      WORKERS: 4
  
  - name: staging
    branch: staging
    default_env_vars:
      LOG_LEVEL: debug
      WORKERS: 2
```

### apps[].deployment

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `keep_releases` | int | ❌ No | `5` | Number of releases to keep (systemd) |
| `keep_images` | int | ❌ No | `5` | Number of Docker images to keep |
| `restart_policy` | string | ❌ No | `always` | Systemd restart policy² |
| `restart_delay` | int | ❌ No | `5` | Delay between restarts (seconds) |

² **Restart Policies:**
- `always` - Always restart
- `on-failure` - Restart only on failure
- `no` - Never restart

**Example:**
```yaml
deployment:
  keep_releases: 10
  restart_policy: on-failure
  restart_delay: 10
```

### docker

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `registry` | string | ❌ No | `""` | Docker registry URL |
| `base_images` | object | ❌ No | See below | Default base images per language |

**base_images defaults:**
```yaml
base_images:
  go: "golang:1.25-alpine"
  python: "python:3.11-slim"
  nodejs: "node:20-alpine"
  ruby: "ruby:3.2-slim"
```

**Example:**
```yaml
docker:
  registry: "registry.example.com"
  base_images:
    go: "golang:1.25-alpine"
    python: "python:3.12-slim"
```

### user

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `deploy_user` | string | ❌ No | `ubuntu` | SSH user for deployment |
| `deploy_group` | string | ❌ No | `ubuntu` | User group |

**Example:**
```yaml
user:
  deploy_user: deploy
  deploy_group: deploy
```

## Minimal Examples

### Minimal Go App

```yaml
apps:
  - name: api
    build:
      path: ./cmd/api
```

### Minimal Python App

```yaml
apps:
  - name: app
    lang: python
    build:
      type: docker
      path: .
```

## Complete Example

```yaml
# Project settings
project:
  name: my-project
  base_dir: /opt/gokku

# Global defaults
defaults:
  build_type: systemd
  lang: go

# Applications
apps:
  # Go API with systemd
  - name: api
    lang: go
    build:
      type: systemd
      path: ./cmd/api
      binary_name: api
      work_dir: .
      go_version: "1.25"
      goos: linux
      goarch: amd64
      cgo_enabled: 0
    
    environments:
      - name: production
        branch: main
        default_env_vars:
          PORT: 8080
          LOG_LEVEL: info
      
      - name: staging
        branch: staging
        default_env_vars:
          PORT: 8080
          LOG_LEVEL: debug
    
    deployment:
      keep_releases: 5
      restart_policy: always
      restart_delay: 5
  
  # Python ML service with Docker
  - name: ml-service
    lang: python
    build:
      type: docker
      path: ./services/ml
      entrypoint: server.py
      base_image: python:3.11-slim
      mise:
        plugins:
          - name: whispercpp
            url: https://github.com/thadeu/asdf-whispercpp.git
    
    environments:
      - name: production
        branch: main
        default_env_vars:
          PORT: 8082
    
    deployment:
      keep_images: 5
      restart_policy: always
      restart_delay: 10

# Docker settings
docker:
  registry: ""
  base_images:
    go: "golang:1.25-alpine"
    python: "python:3.11-slim"
    nodejs: "node:20-alpine"

# Server user
user:
  deploy_user: ubuntu
  deploy_group: ubuntu
```

## Validation

Gokku validates your configuration:

### Required Fields

- ❌ Missing `project.name`: `Error: project.name is required`
- ❌ Missing `apps[].name`: `Error: app name is required`
- ❌ Missing `apps[].build.path`: `Error: app 'api' missing build.path`

### Invalid Values

- ❌ Invalid `build.type`: Must be `systemd` or `docker`
- ❌ Invalid `restart_policy`: Must be `always`, `on-failure`, or `no`
- ❌ Duplicate app names: Each app must have unique name

## Environment Variables

Set via `env-manager` on server:

```bash
# Set variable
./env-manager --app api --env production set KEY=value

# List variables
./env-manager --app api --env production list

# Delete variable
./env-manager --app api --env production del KEY
```

**Don't put secrets in `gokku.yml`!** Use `env-manager` instead.

## Next Steps

- [Examples](/examples/) - Real-world configurations
- [CLI Reference](/reference/cli) - Command-line tools
- [Troubleshooting](/reference/troubleshooting) - Common issues

