# Gokku Deployment System

A **100% generic** git-push deployment system for multi-language applications. No hard-coded app names, ports, or paths. Everything is configurable via `gokku.yml`.

**Gokku** = Go + Dokku - A lightweight alternative to Dokku, focused on Go applications with multi-language support.

## Key Features

✅ **Zero Hard-coding** - Everything configured via `gokku.yml`  
✅ **Multi-Language** - Go, Python, Node.js (extensible)  
✅ **Multi-Runtime** - Systemd or Docker deployment  
✅ **Portable** - Can be extracted to separate repository  
✅ **Config-Driven** - Apps, environments, ports all in config  
✅ **K3s-Style Installer** - One-line installation  
✅ **Auto Dockerfile** - Generates Dockerfile if not exists  

---

## Configuration File (`gokku.yml`)

All project-specific settings are in one place.

### Minimal Configuration

Most fields are optional with sensible defaults:

```yaml
project:
  name: my-project

apps:
  - name: api
    build:
      path: ./cmd/api
      binary_name: api
```

This minimal config will use defaults:
- `lang: go` (default language)
- `build.type: systemd` (default build type)
- `environments: [production]` (default environment)
- `branch: main` (default branch for production)
- `deployment.keep_releases: 5` (default)
- `deployment.restart_policy: always` (default)

### Full Configuration Example

```yaml
project:
  name: my-project
  base_dir: /opt/my-project

apps:
  - name: api-server
    lang: go
    build:
      type: systemd
      path: ./cmd/api
      binary_name: api-server
      work_dir: .
      go_version: "1.25"
      goos: linux
      goarch: amd64
      cgo_enabled: 0
    environments:
      - name: production
        branch: main
        default_env_vars:
          LOG_LEVEL: info
      - name: staging
        branch: staging
        default_env_vars:
          LOG_LEVEL: debug
    deployment:
      keep_releases: 5
      restart_policy: always
      restart_delay: 5
    
  - name: worker
    lang: go
    build:
      type: systemd
      path: ./cmd/worker
      binary_name: worker
      work_dir: .
      go_version: "1.25"
      goos: linux
      goarch: amd64
      cgo_enabled: 0
    environments:
      - name: production
        branch: main
    deployment:
      keep_releases: 3
      restart_policy: always
      restart_delay: 5
    
  - name: ml-service
    lang: python
    build:
      type: docker
      path: ./services/ml
      dockerfile: ./services/ml/Dockerfile  # optional
      entrypoint: main.py
      base_image: "python:3.11-slim"
    environments:
      - name: production
        branch: main
        default_env_vars:
          PORT: 8080
    deployment:
      keep_releases: 3
      keep_images: 5
      restart_policy: always
      restart_delay: 10

port_strategy: manual  # or 'auto' for sequential ports

docker:
  registry: ""  # empty = local, or docker.io, ghcr.io
  base_images:
    go: "golang:1.25-alpine"
    python: "python:3.11-slim"
    nodejs: "node:20-alpine"

user:
  deploy_user: ubuntu
  deploy_group: ubuntu
```

---

## Configuration Defaults

All configuration fields are optional with sensible defaults:

### Build Configuration

| Field | Default | Description |
|-------|---------|-------------|
| `lang` | `go` | Programming language |
| `build.type` | `systemd` | Build type (systemd or docker) |
| `build.work_dir` | `.` | Working directory for build |
| `build.go_version` | `1.25` | Go version (for Go apps) |
| `build.goos` | `linux` | Target OS |
| `build.goarch` | `amd64` | Target architecture |
| `build.cgo_enabled` | `0` | CGO enabled flag |
| `build.entrypoint` | `main.py` (Python)<br>`index.js` (Node.js) | Application entrypoint |
| `build.base_image` | `golang:1.25-alpine` (Go)<br>`python:3.11-slim` (Python)<br>`node:20-alpine` (Node.js) | Docker base image |

### Environment Configuration

| Field | Default | Description |
|-------|---------|-------------|
| `environments` | `[production]` | List of environments |
| `branch` | `main` (production)<br>`staging` (staging)<br>`develop` (dev) | Git branch for environment |

### Deployment Configuration

| Field | Default | Description |
|-------|---------|-------------|
| `deployment.keep_releases` | `5` | Number of releases to keep |
| `deployment.keep_images` | `5` | Number of Docker images to keep |
| `deployment.restart_policy` | `always` | Systemd restart policy |
| `deployment.restart_delay` | `5` | Restart delay in seconds |

### Examples

**Minimal Go app (all defaults):**
```yaml
apps:
  - name: api
    build:
      path: ./cmd/api
      binary_name: api
```

**Minimal Python app (Docker):**
```yaml
apps:
  - name: ml-service
    lang: python
    build:
      type: docker
      path: ./services/ml
```

**Custom everything:**
```yaml
apps:
  - name: custom-app
    lang: go
    build:
      type: systemd
      path: ./cmd/custom
      binary_name: custom
      go_version: "1.24"
      goos: linux
      goarch: arm64
    environments:
      - name: production
        branch: main
      - name: staging
        branch: develop
    deployment:
      keep_releases: 10
      restart_policy: on-failure
      restart_delay: 10
```

---

## Files Overview

### Core Files

- `gokku.yml` - **Main configuration file**
- `config-loader.sh` - Reads config and exports variables
- `deploy-server-setup.sh` - Generic setup script
- `env-manager.go` - Generic env var manager

### Installers

- `install-server.sh` - One-line server installer
- `install-client.sh` - One-line client installer
- `gokku-cli.go` - CLI source code

### Documentation

- `README.md` - This file
- `TESTING.md` - Testing guide
- `INSTALLERS.md` - Installer documentation

---

## Usage

### 1. Configure Your Project

Edit `gokku.yml` with your apps and environments:

```yaml
project:
  name: awesome-api

apps:
  - name: api
    build_path: ./cmd/api
    binary_name: api
    
  - name: worker
    build_path: ./cmd/worker
    binary_name: worker

environments:
  - name: production
    branch: main
  - name: staging
    branch: develop
```

### 2. Setup Server

```bash
# With config file
./deploy-server-setup.sh api production

# Or set PAPIS_CONFIG environment variable
export PAPIS_CONFIG=/path/to/gokku.yml
./deploy-server-setup.sh api production
```

### 3. Manage Environment Variables

```bash
# Build env-manager
cd infra
go build -o env-manager env-manager.go

# Use it
./env-manager --app api --env production set API_KEY=xxx
./env-manager --app api --env production list
```

### 4. Deploy

```bash
git push api-production main
git push worker-staging develop
```

---

## Port Management

### Manual Strategy (Recommended)

Ports are defined per app/env in `.env` files:

```bash
env-manager --app api --env production set PORT=8080
env-manager --app api --env staging set PORT=8081
env-manager --app worker --env production set PORT=9090
```

**Advantages:**
- Full control
- Clear and explicit
- Easy to document
- No conflicts

### Auto Strategy

Sequential ports assigned automatically:

```yaml
port_strategy: auto
base_port: 8000
```

Results in:
- `api-production` → 8000
- `api-staging` → 8001
- `worker-production` → 8002
- `worker-staging` → 8003

**Advantages:**
- No manual assignment
- Predictable
- Quick setup

---

## Docker Support

### Build Types

Each app has its own `build` configuration with type `systemd` (binary) or `docker` (container):

```yaml
apps:
  - name: api
    lang: go
    build:
      type: systemd  # Compiles Go binary, runs with systemd
      path: ./cmd/api
      binary_name: api
      go_version: "1.25"
      goos: linux
      goarch: amd64
      cgo_enabled: 0
    
  - name: ml-service
    lang: python
    build:
      type: docker   # Builds Docker image, runs in container
      path: ./services/ml
      entrypoint: main.py
      base_image: "python:3.11-slim"
```

### Automatic Dockerfile Generation

If `build.type: docker` and no Dockerfile exists, the system generates one automatically:

**Go Apps:**
```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o app ./cmd/api

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /app/app .
EXPOSE ${PORT:-8080}
CMD ["./app"]
```

**Python Apps:**
```dockerfile
FROM python:3.11-slim
WORKDIR /app
RUN apt-get update && apt-get install -y gcc && rm -rf /var/lib/apt/lists/*
COPY requirements.txt* ./
RUN pip install --no-cache-dir -r requirements.txt
COPY . .
EXPOSE ${PORT:-8080}
CMD ["python", "main.py"]
```

### Custom Dockerfile

Specify your own Dockerfile in the build config:

```yaml
apps:
  - name: ml-service
    lang: python
    build:
      type: docker
      path: ./services/ml
      dockerfile: ./services/ml/Dockerfile  # Use this instead of auto-generation
      entrypoint: main.py
```

### How Docker Deployment Works

1. **Push code**: `git push ml-service-prod main`
2. **Hook extracts code** to release directory
3. **Checks for Dockerfile**:
   - If exists: uses it
   - If not: generates based on `lang`
4. **Builds image**: `ml-service:20250115-150405`
5. **Tags as latest**: `ml-service:latest`
6. **Systemd manages container**:
   ```ini
   [Service]
   ExecStart=/usr/bin/docker run --rm --name ml-service-production \
     --env-file /opt/gokku/apps/ml-service/production/shared/.env \
     -p ${PORT}:${PORT} \
     ml-service:latest
   ```
7. **Cleanup old images** (keeps last 5)

### Port Mapping

Docker apps use the same port strategy as systemd apps:

```bash
# Set PORT in .env
env-manager --app ml-service --env production set PORT=8080

# Container maps: -p 8080:8080
```

The container reads `PORT` from the environment and the systemd service maps it.

### Docker Commands

```bash
# View images
docker images ml-service

# View logs
docker logs ml-service-production -f

# Exec into container
docker exec -it ml-service-production sh

# Inspect container
docker inspect ml-service-production
```

### Rollback with Docker

Images are tagged with timestamps:

```bash
# Current
ml-service:latest → ml-service:20250115-150405

# Previous releases
ml-service:20250115-140305
ml-service:20250115-130200

# Rollback (manual for now)
docker stop ml-service-production
docker run --name ml-service-production ml-service:20250115-140305
```

### Mixed Deployments

You can mix systemd and Docker apps in the same project:

```yaml
apps:
  - name: api
    lang: go
    build:
      type: systemd  # Fast Go binary
      path: ./cmd/api
      binary_name: api
    
  - name: worker
    lang: go
    build:
      type: systemd  # Another Go binary
      path: ./cmd/worker
      binary_name: worker
    
  - name: vad
    lang: python
    build:
      type: docker   # Python in container
      path: ./services/vad
      entrypoint: main.py
```

---

## Using as Separate Repository

This system can be extracted to its own repository and used across multiple projects.

### Structure

```
deployment-system/          # Separate repo
├── gokku.yml.example      # Template config
├── config-loader.sh
├── deploy-server-setup-v2.sh
├── env-manager-v2.go
├── install.sh
└── README.md

your-project/              # Your Go project
├── cmd/
│   ├── api/
│   └── worker/
├── gokku.yml             # Project-specific config
└── ... (your code)
```

### Installation from Separate Repo

```bash
# One-line install (future)
curl -fsSL https://raw.githubusercontent.com/USER/deployment-system/main/install.sh | bash

# With custom config
curl -fsSL https://... | bash -s -- --config ./gokku.yml
```

### Benefits

1. **Reusable** - Use same deployer for all Go projects
2. **Versioned** - Update deployment system independently
3. **Shareable** - Teams can share deployment practices
4. **Maintainable** - One place to fix bugs/add features
5. **Clean** - Keep deployment separate from application code

---

## Quick Start

### Step 1: Create gokku.yml

Configure your project:

```yaml
project:
  name: my-project
  base_dir: /opt/my-project

apps:
  - name: api
    build_path: ./cmd/api
    binary_name: api
    
  - name: worker
    build_path: ./cmd/worker
    binary_name: worker

environments:
  - name: production
    branch: main
  - name: staging
    branch: staging

build:
  work_dir: .  # or apps/trunk for your structure
```

### Step 2: Setup Server

```bash
./deploy-server-setup.sh api production
```

### Step 3: Build env-manager

```bash
go build -o env-manager env-manager.go
# Copy to server
scp env-manager server:/usr/local/bin/
```

### Step 4: Deploy

```bash
git push api-production main
```

---

## Configuration Reference

### Project Section

```yaml
project:
  name: string        # Project name (used in logs)
  base_dir: string    # Base directory (default: /opt/gokku)
```

### Apps Section

```yaml
apps:
  - name: string           # App name (must be unique)
    build_path: string     # Path to main package (relative to work_dir)
    binary_name: string    # Output binary name (defaults to app name)
```

### Environments Section

```yaml
environments:
  - name: string                  # Environment name
    branch: string                # Git branch for this environment
    default_env_vars:             # Default variables (optional)
      KEY: value
```

### Build Section

```yaml
build:
  go_version_min: string    # Minimum Go version required
  goos: string             # Target OS (default: linux)
  goarch: string           # Target architecture (default: amd64)
  cgo_enabled: number      # CGO setting (default: 0)
  work_dir: string         # Build working directory (default: .)
```

### Deployment Section

```yaml
deployment:
  keep_releases: number           # Number of releases to keep (default: 5)
  health_check_timeout: number    # Seconds to wait for health check
  restart_policy: string          # Systemd restart policy (default: always)
  restart_delay: number           # Seconds between restarts (default: 5)
```

### User Section

```yaml
user:
  deploy_user: string      # User to run services (default: ubuntu)
  deploy_group: string     # Group for files (default: ubuntu)
```

---

## Environment Variables in Config

### System Variables

Set via `gokku.yml` or environment:

- `PAPIS_CONFIG` - Path to config file
- `PAPIS_PROJECT_NAME` - Override project name
- `PAPIS_BASE_DIR` - Override base directory
- `PAPIS_DEPLOY_USER` - Override deploy user

### Application Variables

Set via env-manager for each app/environment:

```bash
env-manager --app api --env production set \
  API_KEY=xxx \
  DATABASE_URL=postgres://... \
  PORT=8080
```

---

## Advantages

| Aspect | This System |
|--------|-------------|
| **Reusability** | ✅ Any Go project |
| **Portability** | ✅ Separate repo ready |
| **Configuration** | ✅ Edit YAML |
| **New App** | ✅ Add to config |
| **New Environment** | ✅ Add to config |
| **Documentation** | ✅ In config file |
| **Validation** | ✅ Automatic |
| **Errors** | ✅ Early detection |

---

## Future Enhancements

### Planned Features

1. **Binary Distribution**
   - Pre-compiled binaries for Linux/macOS
   - `curl -fsSL https://... | sh` installer
   - GitHub releases with checksums

2. **Config Validation**
   - JSON Schema for gokku.yml
   - Linter for common mistakes
   - Best practices checker

3. **Multi-Language Support**
   - Node.js applications
   - Python applications
   - Any compiled language

4. **Advanced Features**
   - Blue-green deployments
   - Canary releases
   - Load balancer integration
   - Health check endpoints
   - Metrics collection

5. **Cloud Integrations**
   - AWS Systems Manager
   - Parameter Store for secrets
   - CloudWatch Logs
   - Auto-scaling groups

---

## Testing

### Test on Fresh EC2

```bash
# 1. Create fresh Ubuntu EC2
# 2. Install Go
# 3. Copy files
scp -r infra/ ubuntu@ec2:/tmp/

# 4. SSH and test
ssh ubuntu@ec2
cd /tmp/infra

# 5. Create test config
cat > gokku.yml << EOF
project:
  name: test-project
  
apps:
  - name: test-app
    build_path: ./main.go
    binary_name: test-app

environments:
  - name: production
    branch: main
EOF

# 6. Run setup
./deploy-server-setup-v2.sh test-app production

# 7. Verify
sudo systemctl status test-app-production
```

---

## Contributing

When adding features:

1. Keep everything config-driven
2. No hard-coded values
3. Support any Go project structure
4. Document in gokku.yml.example
5. Add validation where possible

---

## License

Internal use / Open source (your choice)

