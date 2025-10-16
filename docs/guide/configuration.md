# Configuration

Gokku is configured via a `gokku.yml` file in your project root. This file defines your apps, environments, and deployment settings.

## Minimal Configuration

The simplest `gokku.yml`:

```yaml
apps:
  - name: api
    build:
      path: ./cmd/api
```

That's it! Everything else has sensible defaults.

## Full Configuration

Here's a complete example with all options:

```yaml
project:
  name: my-project
  base_dir: /opt/gokku

defaults:
  build_type: systemd
  lang: go

apps:
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
      mise:
        plugins:
          - name: whispercpp
            url: https://github.com/thadeu/asdf-whispercpp.git
    
    environments:
      - name: production
        branch: main
        default_env_vars:
          PORT: 8080
          LOG_LEVEL: info
      
      - name: staging
        branch: develop
        default_env_vars:
          PORT: 8081
          LOG_LEVEL: debug
    
    deployment:
      keep_releases: 5
      restart_policy: always
      restart_delay: 5

docker:
  registry: ""
  base_images:
    go: "golang:1.25-alpine"
    python: "python:3.11-slim"
    nodejs: "node:20-alpine"
```

## Configuration Sections

### Project

Global project settings:

```yaml
project:
  name: my-project      # Project name
  base_dir: /opt/gokku  # Server installation directory
```

**Defaults:**
- `name`: (required)
- `base_dir`: `/opt/gokku`

### Defaults

Default values for all apps:

```yaml
defaults:
  build_type: systemd  # systemd or docker
  lang: go             # go, python, nodejs, etc
```

### Apps

The main configuration section. Each app can have:

#### Basic Settings

```yaml
apps:
  - name: api           # App name (required)
    lang: go            # Language (optional, default: go)
```

#### Build Configuration

```yaml
    build:
      type: systemd           # Build type: systemd or docker
      path: ./cmd/api         # Path to main file/directory
      binary_name: api        # Output binary name (Go only)
      work_dir: .             # Working directory for build
      
      # Go-specific settings
      go_version: "1.25"      # Go version
      goos: linux             # Target OS
      goarch: amd64           # Target architecture
      cgo_enabled: 0          # Enable CGO (0 or 1)
      
      # Docker-specific settings
      dockerfile: ./Dockerfile     # Custom Dockerfile path
      entrypoint: main.py          # Entrypoint file (non-Go)
      base_image: python:3.11-slim # Base image
      
      # Mise/asdf plugins (optional)
      mise:
        plugins:
          - name: whispercpp
            url: https://github.com/thadeu/asdf-whispercpp.git
```

**Build Type Defaults:**

| Setting | systemd | docker |
|---------|---------|--------|
| `type` | systemd | docker |
| `path` | (required) | (required) |
| `binary_name` | `{app_name}` | - |
| `work_dir` | `.` | `.` |
| `go_version` | `1.25` | - |
| `goos` | `linux` | - |
| `goarch` | `amd64` | - |
| `cgo_enabled` | `0` | - |
| `entrypoint` | - | `main.py` (python), `index.js` (node) |
| `base_image` | - | From `docker.base_images` |

#### Environments

Define deployment environments:

```yaml
    environments:
      - name: production          # Environment name
        branch: main              # Git branch
        default_env_vars:         # Default environment variables
          PORT: 8080
          LOG_LEVEL: info
      
      - name: staging
        branch: develop
        default_env_vars:
          PORT: 8081
          LOG_LEVEL: debug
```

**Defaults:**
- If no environments defined, defaults to `production` environment
- Default branches: `main` (production), `staging` (staging), `develop` (develop)

#### Deployment

Deployment settings:

```yaml
    deployment:
      keep_releases: 5        # Number of releases to keep
      keep_images: 5          # Number of Docker images to keep
      restart_policy: always  # systemd restart policy
      restart_delay: 5        # Delay between restarts (seconds)
```

**Defaults:**
- `keep_releases`: `5`
- `keep_images`: `5`
- `restart_policy`: `always`
- `restart_delay`: `5`

### Docker

Global Docker settings:

```yaml
docker:
  registry: ""                     # Docker registry URL (optional)
  base_images:
    go: "golang:1.25-alpine"
    python: "python:3.11-slim"
    nodejs: "node:20-alpine"
```

Used when `build.type: docker` and no `build.base_image` specified.

### User Configuration

User configuration is **automatically detected** from your git remote URL.

**Example:**
```bash
# Git remote format: user@host:path
git remote add production ubuntu@server:api
# The user 'ubuntu' is automatically extracted and used
```

**No configuration needed** - Gokku automatically uses the user from your git remote.

## Configuration by Language

### Go Application

```yaml
apps:
  - name: api
    lang: go
    build:
      type: systemd
      path: ./cmd/api
      binary_name: api
      go_version: "1.25"
```

### Python Application

```yaml
apps:
  - name: worker
    lang: python
    build:
      type: docker
      path: ./apps/worker
      entrypoint: main.py
      base_image: python:3.11-slim
```

### Node.js Application

```yaml
apps:
  - name: frontend
    lang: nodejs
    build:
      type: docker
      path: ./apps/frontend
      entrypoint: index.js
      base_image: node:20-alpine
```

## Multi-App Configuration

Deploy multiple apps from one repository:

```yaml
project:
  name: my-monorepo

apps:
  - name: api
    lang: go
    build:
      type: systemd
      path: ./cmd/api
  
  - name: worker
    lang: go
    build:
      type: systemd
      path: ./cmd/worker
  
  - name: ml-service
    lang: python
    build:
      type: docker
      path: ./services/ml
      entrypoint: server.py
```

Each app gets:
- Separate Git repository
- Independent environments
- Isolated deployments

## Environment-Specific Configuration

Different settings per environment:

```yaml
apps:
  - name: api
    environments:
      - name: production
        branch: main
        default_env_vars:
          DATABASE_URL: postgres://prod-db
          CACHE_TTL: 3600
          WORKERS: 4
      
      - name: staging
        branch: staging
        default_env_vars:
          DATABASE_URL: postgres://staging-db
          CACHE_TTL: 60
          WORKERS: 2
```

## Mise/ASDF Integration

Manage tool versions with mise:

```yaml
apps:
  - name: whisper
    lang: python
    build:
      type: docker
      path: ./apps/whisper
      mise:
        plugins:
          - name: whispercpp
            url: https://github.com/thadeu/asdf-whispercpp.git
          - name: ffmpeg
            url: https://github.com/acj/asdf-ffmpeg.git
```

Then create `.tool-versions` in `./apps/whisper/`:

```
python 3.11
whispercpp 1.5.0
ffmpeg 8.0
```

Gokku automatically:
1. Installs mise plugins
2. Runs `mise install`
3. Uses correct versions

See [Mise Integration](/guide/mise) for details.

## Validation

Gokku validates your configuration on deployment. Common errors:

### Missing Required Fields

```
Error: app 'api' missing required field: build.path
```

**Fix:** Add `build.path`:

```yaml
apps:
  - name: api
    build:
      path: ./cmd/api  # Required!
```

### Invalid Build Type

```
Error: invalid build type 'systemd-docker' for app 'api'
```

**Fix:** Use `systemd` or `docker`:

```yaml
apps:
  - name: api
    build:
      type: systemd  # or docker
```

### Duplicate App Names

```
Error: duplicate app name 'api'
```

**Fix:** Each app needs a unique name:

```yaml
apps:
  - name: api-v1
  - name: api-v2
```

## Best Practices

### 1. Use Sensible Defaults

Don't repeat defaults:

✅ **Good:**
```yaml
apps:
  - name: api
    build:
      path: ./cmd/api
```

❌ **Bad:**
```yaml
apps:
  - name: api
    lang: go
    build:
      type: systemd
      path: ./cmd/api
      binary_name: api
      go_version: "1.25"
      goos: linux
      goarch: amd64
      cgo_enabled: 0
```

### 2. Keep Environments Simple

Only define what's different:

✅ **Good:**
```yaml
environments:
  - name: production
    branch: main
  - name: staging
    branch: staging
```

❌ **Bad:**
```yaml
environments:
  - name: production
    branch: main
    default_env_vars:
      # Don't put secrets in config!
      DATABASE_URL: postgres://...
```

Use `gokku config` for secrets instead.

### 3. Version Control

Always commit `gokku.yml`:

```bash
git add gokku.yml
git commit -m "feat: add gokku configuration"
```

### 4. Document Custom Settings

Add comments for non-obvious settings:

```yaml
apps:
  - name: api
    build:
      cgo_enabled: 1  # Required for SQLite
```

## Examples

See [Examples](/examples/) for real-world configurations:

- [Go Application](/examples/go-app)
- [Python Application](/examples/python-app)
- [Docker Application](/examples/docker-app)
- [Multi-App Project](/examples/multi-app)

## Reference

Full configuration reference: [Configuration Reference](/reference/configuration)

