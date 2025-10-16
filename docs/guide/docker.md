# Docker Support

Deploy applications using Docker containers with Gokku.

## Overview

Gokku supports two deployment modes:
- **systemd** - Native binary execution (best for Go apps)
- **docker** - Container-based execution (best for Python/Node.js)

This guide covers Docker-specific features.

## Enable Docker

Set `build.type: docker` in your app config:

```yaml
apps:
  - name: my-app
    lang: python
    build:
      type: docker
      path: ./app
      entrypoint: main.py
```

## Automatic Dockerfile Generation

If no Dockerfile exists, Gokku generates one based on your language:

### Python

```yaml
apps:
  - name: flask-app
    lang: python
    build:
      type: docker
      path: .
      entrypoint: app.py
```

Generated Dockerfile:
```dockerfile
FROM python:3.11-slim

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends gcc && rm -rf /var/lib/apt/lists/*

COPY requirements.txt* ./
RUN if [ -f requirements.txt ]; then pip install --no-cache-dir -r requirements.txt; fi

COPY . .

EXPOSE ${PORT:-8080}

CMD ["python", "app.py"]
```

### Node.js

```yaml
apps:
  - name: node-app
    lang: nodejs
    build:
      type: docker
      path: .
      entrypoint: index.js
```

Generated Dockerfile:
```dockerfile
FROM node:20-alpine

WORKDIR /app

COPY package*.json ./
RUN npm ci --only=production

COPY . .

EXPOSE ${PORT:-8080}

CMD ["node", "index.js"]
```

### Go (with Docker)

```yaml
apps:
  - name: go-app
    lang: go
    build:
      type: docker
      path: .
```

Generated Dockerfile:
```dockerfile
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o app .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/app .

EXPOSE ${PORT:-8080}

CMD ["./app"]
```

## Custom Dockerfile

Use your own Dockerfile:

```yaml
apps:
  - name: my-app
    build:
      type: docker
      dockerfile: ./Dockerfile
```

### Example: Python with System Dependencies

```dockerfile
FROM python:3.11-slim

# Install system packages
RUN apt-get update && apt-get install -y \
    gcc \
    ffmpeg \
    libpq-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

EXPOSE ${PORT:-8080}

CMD ["gunicorn", "--bind", "0.0.0.0:${PORT:-8080}", "app:app"]
```

### Example: Multi-Stage Build

```dockerfile
# Build stage
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

# Production stage
FROM node:20-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY --from=builder /app/dist ./dist

EXPOSE ${PORT:-8080}

CMD ["node", "dist/server.js"]
```

## Base Images

### Configure Global Base Images

```yaml
docker:
  base_images:
    go: "golang:1.25-alpine"
    python: "python:3.11-slim"
    nodejs: "node:20-alpine"
```

### Per-App Base Image

```yaml
apps:
  - name: ml-service
    lang: python
    build:
      type: docker
      base_image: "python:3.11"  # Full image, not slim
```

### From Private Registry

```yaml
docker:
  registry: "registry.example.com"

apps:
  - name: api
    build:
      type: docker
      base_image: "registry.example.com/python:3.11-custom"
```

Login on server first:
```bash
ssh ubuntu@server "docker login registry.example.com"
```

## Mise Integration

Gokku automatically integrates mise/asdf with Docker.

### Automatic Mise Dockerfile

If `.tool-versions` exists and no custom Dockerfile:

```
python 3.11
ffmpeg 8.0
```

Gokku generates:

```dockerfile
FROM python:3.11-slim

WORKDIR /app

# Install mise
RUN apt-get update && apt-get install -y curl git build-essential && rm -rf /var/lib/apt/lists/*
RUN curl https://mise.run | sh
ENV PATH="/root/.local/share/mise/shims:/root/.local/share/mise/bin:${PATH}"

# Copy .tool-versions and install tools
COPY .tool-versions .
RUN mise install

# Copy application
COPY . .

EXPOSE ${PORT:-8080}

CMD ["python", "app.py"]
```

### With Custom Plugins

```yaml
apps:
  - name: whisper
    lang: python
    build:
      type: docker
      mise:
        plugins:
          - name: whispercpp
            url: https://github.com/thadeu/asdf-whispercpp.git
```

`.tool-versions`:
```
python 3.11
ffmpeg 8.0
whispercpp 1.5.0
```

Generated Dockerfile includes plugin installation:

```dockerfile
# ... mise install ...
RUN MISE_EXPERIMENTAL=1 mise plugins install whispercpp https://github.com/...
RUN mise install
# ...
```

## Image Management

### Tagging

Gokku automatically tags images:

```
my-app:release-1
my-app:release-2
my-app:release-3
```

Each deployment creates a new tag.

### Keep Images

Configure how many images to keep:

```yaml
apps:
  - name: my-app
    deployment:
      keep_images: 10  # Keep last 10 images
```

Old images are automatically pruned after deployment.

### Manual Cleanup

```bash
# List images
ssh ubuntu@server "docker images | grep my-app"

# Remove specific image
ssh ubuntu@server "docker rmi my-app:release-5"

# Remove all unused images
ssh ubuntu@server "docker image prune -a"
```

## Port Mapping

Gokku uses fixed port mapping strategy.

### How It Works

Each app gets assigned a port:
- First app: 8080
- Second app: 8081
- Third app: 8082
- etc.

```yaml
apps:
  - name: api
    environments:
      - name: production
        default_env_vars:
          PORT: 8080  # Container port
```

Docker maps: `8080:8080`

### Custom Ports

Set via environment variables:

```bash
gokku config set PORT=8081 --app api --env production --remote api-production
```

Then redeploy.

## Container Management

### View Running Containers

```bash
ssh ubuntu@server "docker ps | grep my-app"
```

### View Logs

```bash
ssh ubuntu@server "docker logs -f my-app-production"
```

### Restart Container

```bash
ssh ubuntu@server "docker restart my-app-production"
```

### Stop Container

```bash
ssh ubuntu@server "docker stop my-app-production"
```

### Start Container

```bash
ssh ubuntu@server "docker start my-app-production"
```

### Access Container Shell

```bash
ssh ubuntu@server "docker exec -it my-app-production /bin/sh"
```

### Inspect Container

```bash
ssh ubuntu@server "docker inspect my-app-production"
```

## Environment Variables

All environment variables are passed to the container:

```bash
# Set variable
gokku config set DATABASE_URL=postgres://... --app api --env production --remote api-production

# Restart container to pick up changes
ssh ubuntu@server "docker restart api-production"
```

## Health Checks

Add health check to Dockerfile:

```dockerfile
FROM python:3.11-slim

WORKDIR /app
COPY . .

EXPOSE ${PORT:-8080}

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD python -c "import urllib.request; urllib.request.urlopen('http://localhost:${PORT:-8080}/health')" || exit 1

CMD ["python", "app.py"]
```

Check health status:

```bash
ssh ubuntu@server "docker inspect --format='{{.State.Health.Status}}' api-production"
```

## Volumes (Limited Support)

Gokku doesn't currently support Docker volumes in config. Workarounds:

### Option 1: External Database

Use managed database instead of volume:

```yaml
environments:
  - name: production
    default_env_vars:
      DATABASE_URL: postgres://external-db:5432/mydb
```

### Option 2: Host Bind Mount (Manual)

Modify the hook template (advanced users):

```bash
docker run -d \
  --name $APP_NAME-$ENV \
  -v /data/uploads:/app/uploads \
  -p $PORT:$PORT \
  $IMAGE_NAME
```

### Option 3: Separate Data Container

Run database/redis outside Gokku:

```bash
# On server
docker run -d \
  --name postgres \
  -v /data/postgres:/var/lib/postgresql/data \
  -p 5432:5432 \
  postgres:15
```

## Networking

### Container to Container

Containers can communicate via host network:

```yaml
apps:
  - name: api
    environments:
      - name: production
        default_env_vars:
          REDIS_URL: redis://172.17.0.1:6379
```

Use Docker host IP (`172.17.0.1`) or `host.docker.internal`.

### External Services

Connect to external services normally:

```yaml
environments:
  - name: production
    default_env_vars:
      DATABASE_URL: postgres://external-db.example.com:5432/db
```

## Comparison: Docker vs Systemd

| Feature | Docker | Systemd |
|---------|--------|---------|
| **Deployment Speed** | Slower (build image) | Faster (compile binary) |
| **Resource Usage** | Higher | Lower |
| **Isolation** | Full container | Process only |
| **Dependencies** | In container | System-wide or mise |
| **Best For** | Python/Node/Complex | Go apps |
| **Rollback** | Image tags | Release directories |

## When to Use Docker

✅ **Use Docker if:**
- Python/Node.js/Ruby app
- Complex system dependencies
- Need full isolation
- Multiple conflicting dependencies
- Different versions of same tool

❌ **Avoid Docker if:**
- Simple Go app
- Want fastest deploys
- Limited server resources
- No complex dependencies

## Troubleshooting

### Build Failed

Check deploy logs:

```bash
ssh ubuntu@server "cat /opt/gokku/apps/my-app/production/deploy.log"
```

### Container Won't Start

Check Docker logs:

```bash
ssh ubuntu@server "docker logs my-app-production"
```

### Out of Disk Space

Clean up old images:

```bash
ssh ubuntu@server "docker system prune -a"
```

### Image Pull Failed

Check registry login:

```bash
ssh ubuntu@server "docker login registry.example.com"
```

## Advanced

### Custom Docker Run Options

To add custom Docker run options, modify the hook template:

`/opt/gokku/hooks/post-receive-docker.template`

Add options to the `docker run` command.

### Docker Compose (Not Supported)

Gokku uses `docker run`, not `docker-compose`. For multi-container apps:

1. Deploy each service separately
2. Or use a process manager inside one container (supervisord)

### Private Registry

```yaml
docker:
  registry: "registry.example.com"

apps:
  - name: api
    build:
      base_image: "registry.example.com/python:3.11"
```

Login on server:
```bash
ssh ubuntu@server "docker login registry.example.com"
```

## Next Steps

- [Mise Integration](/guide/mise) - Auto-install dependencies
- [Environment Variables](/guide/env-vars) - Configure containers
- [Rollback](/guide/rollback) - Rollback to previous images
- [Examples](/examples/docker-app) - Real-world examples

