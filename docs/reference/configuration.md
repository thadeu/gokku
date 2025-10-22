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
| `build_type` | string | ❌ No | `docker` | Default build type: `docker` only |
| `lang` | string | ❌ No | `go` | Default language: `go`, `python`, `nodejs`, etc |

**Example:**
```yaml
defaults:
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
    build:
      path: ./cmd/api
```

### apps[].build

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `type` | string | ❌ No | From `defaults.build_type` | Build type: `docker` only |
| `path` | string | ✅ Yes | - | Path to app code (relative to project root) |
| `binary_name` | string | ❌ No | Same as `app.name` | Output binary name (Go only) |
| `work_dir` | string | ❌ No | `.` | Working directory for build |
| `go_version` | string | ❌ No | `1.25` | Go version (Go only) |
| `goos` | string | ❌ No | `linux` | Target OS (Go only) |
| `goarch` | string | ❌ No | `amd64` | Target architecture (Go only) |
| `cgo_enabled` | int | ❌ No | `0` | Enable CGO: `0` or `1` (Go only) |
| `dockerfile` | string | ❌ No | - | Custom Dockerfile path (Docker only) |
| `entrypoint` | string | ❌ No | Language-specific | Entrypoint file (non-Go) |
| `image` | string | ❌ No | Auto-detected | Docker base image or pre-built registry image |

### Image Configuration

The `build.image` field supports two deployment modes:

**Base Image (Local Build):**
```yaml
build:
  image: "python:3.11-slim"  # Base image for local build
  path: ./app
```

**Pre-built Registry Image (Ultra-fast Deployment):**
```yaml
build:
  image: "ghcr.io/meu-org/api:latest"  # Pre-built image from registry
```

When using a registry image (ghcr.io, ECR, docker.io, etc.), Gokku will:
1. Pull the pre-built image from the registry
2. Tag it for the application  
3. Deploy directly (no build step required)

This enables ultra-fast deployments and integrates perfectly with CI/CD pipelines.

### Automatic Version Detection

When `build.image` is not specified, Gokku automatically detects the version from project files:

**Ruby:**
- `.ruby-version` file (e.g., `3.2.0`)
- `Gemfile` (e.g., `ruby '3.1.0'`)
- Fallback: `ruby:latest`

**Go:**
- `go.mod` file (e.g., `go 1.21`)
- Fallback: `golang:latest-alpine`

**Node.js:**
- `.nvmrc` file (e.g., `18.17.0`)
- `package.json` engines field (e.g., `"node": ">=18.0.0"`)
- Fallback: `node:latest`

**Python:**
- Always uses `python:latest` as fallback


**Entrypoint Defaults:**
- Python: `main.py`
- Node.js: `index.js`
- Ruby: `app.rb`

**Example (Go + Docker):**
```yaml
build:
  path: ./cmd/api
  binary_name: api
  go_version: "1.25"
  cgo_enabled: 0
```

**Example (Python + Docker):**
```yaml
build:
  path: ./services/ml
  entrypoint: server.py
  image: python:3.11-slim
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
| `keep_releases` | int | ❌ No | `5` | Number of releases to keep |
| `keep_images` | int | ❌ No | `5` | Number of Docker images to keep |
| `restart_policy` | string | ❌ No | `always` | Container restart policy² |
| `restart_delay` | int | ❌ No | `5` | Delay between restarts (seconds) |
| `post_deploy` | array | ❌ No | `[]` | Commands to run after successful deployment |

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
  post_deploy:
    - npm run db:migrate"
    - npm run cache:warm"
```

### docker

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `registry` | array | ❌ No | `[]` | List of custom Docker registries |

**Example:**
```yaml
docker:
  registry:
    - "self-ghrc.io"  # Your own GitHub Container Registry
    - "registry.company.com"  # Company private registry
    - "harbor.example.com"  # Harbor registry
```

### User Configuration

User configuration is **automatically detected** from your git remote URL.

**Example:**
```bash
# Git remote format: user@host:path
git remote add production ubuntu@server:api
```

**No configuration needed** - Gokku automatically uses the user from your git remote.

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
  lang: go

# Applications
apps:
  # Go API with Docker
  - name: api
    build:
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
      path: ./services/ml
      entrypoint: server.py
      image: python:3.11-slim
    
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
```

## Validation

Gokku validates your configuration:

### Required Fields

- ❌ Missing `project.name`: `Error: project.name is required`
- ❌ Missing `apps[].name`: `Error: app name is required`
- ❌ Missing `apps[].build.path`: `Error: app 'api' missing build.path`

### Invalid Values

- ❌ Missing required `build.path`
- ❌ Invalid `restart_policy`: Must be `always`, `on-failure`, or `no`
- ❌ Duplicate app names: Each app must have unique name

## Environment Variables

Set via `gokku config`:

```bash
# Set variable (remote)
gokku config set KEY=value --app api --env production -a api-production

# Set variable (local, on server)
gokku config set KEY=value --app api --env production

# List variables (remote)
gokku config list --app api --env production -a api-production

# List variables (local, on server)
gokku config list --app api --env production

# Delete variable (remote)
gokku config unset KEY --app api --env production -a api-production
```

**Don't put secrets in `gokku.yml`!** Use `gokku config` instead.

## Next Steps

- [Examples](/examples/) - Real-world configurations
- [CLI Reference](/reference/cli) - Command-line tools
- [Troubleshooting](/reference/troubleshooting) - Common issues

