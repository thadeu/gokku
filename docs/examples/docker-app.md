# Docker Application Example

Deploy any application using Docker with Gokku.

## Basic Docker Deployment

### With Custom Dockerfile

```yaml
apps:
  - name: my-app
    lang: python
    build:
      path: ./app
      dockerfile: ./app/Dockerfile
```

Gokku will use your existing Dockerfile.

### Auto-Generated Dockerfile

```yaml
apps:
  - name: my-app
    lang: python
    build:
      path: ./app
      entrypoint: main.py
```

Gokku generates a Dockerfile automatically based on language.

## Language-Specific Examples

### Python

```yaml
apps:
  - name: python-app
    lang: python
    build:
      path: .
      entrypoint: app.py
      base_image: python:3.11-slim
```

Generated Dockerfile:
```dockerfile
FROM python:3.11-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY . .
EXPOSE ${PORT:-8080}
CMD ["python", "app.py"]
```

### Node.js

```yaml
apps:
  - name: nodejs-app
    lang: nodejs
    build:
      path: .
      entrypoint: index.js
      base_image: node:20-alpine
```

Generated Dockerfile:
```dockerfile
FROM node:20-alpine
WORKDIR /app
COPY package*.json .
RUN npm ci --only=production
COPY . .
EXPOSE ${PORT:-8080}
CMD ["node", "index.js"]
```

### Go (Docker Build)

```yaml
apps:
  - name: go-app
    lang: go
    build:
      path: ./cmd/api
      base_image: golang:1.25-alpine
```

## Custom Dockerfile Examples

### Multi-Stage Build (Go)

```dockerfile
# Build stage
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.* .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o app ./cmd/api

# Runtime stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/app .
EXPOSE ${PORT:-8080}
CMD ["./app"]
```

### Multi-Stage Build (Node.js)

```dockerfile
# Build stage
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json .
RUN npm ci
COPY . .
RUN npm run build

# Runtime stage
FROM node:20-alpine
WORKDIR /app
COPY package*.json .
RUN npm ci --only=production
COPY --from=builder /app/dist ./dist
EXPOSE ${PORT:-8080}
CMD ["node", "dist/index.js"]
```

### With System Dependencies

```dockerfile
FROM python:3.11-slim

# Install system dependencies
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

CMD ["python", "app.py"]
```

## With Docker Compose (Not Supported Yet)

Gokku currently uses `docker run`, not `docker-compose`. For services requiring multiple containers:

### Option 1: Deploy Separately

```yaml
apps:
  - name: web
    lang: python
    build:
      path: ./web
  
  - name: worker
    lang: python
    build:
      path: ./worker
```

### Option 2: Single Container with Multiple Processes

Use a process manager like `supervisord`:

```dockerfile
FROM python:3.11-slim

RUN apt-get update && apt-get install -y supervisor

COPY supervisord.conf /etc/supervisor/conf.d/supervisord.conf
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

CMD ["/usr/bin/supervisord"]
```

## Environment Variables

Set via gokku config:

```bash
gokku config set DATABASE_URL=postgres://... -a my-app-production
```

Access in container:

```python
import os
db_url = os.getenv('DATABASE_URL')
```

## Port Mapping

Gokku automatically maps ports:

```yaml
apps:
  - name: my-app
    environments:
      - name: production
        default_env_vars:
          PORT: 8080  # Container port
```

External port is managed by Gokku (incremental: 8080, 8081, 8082...).

## Volumes (Persistent Data)

Currently, Gokku doesn't support Docker volumes directly. Workarounds:

### Option 1: Host Path Mount

Modify the Docker run command in the hook (advanced).

### Option 2: External Storage

Use S3, PostgreSQL, Redis for persistent data.

### Option 3: Database Container

Run database outside Gokku:

```bash
# On server
docker run -d \
  --name postgres \
  -e POSTGRES_PASSWORD=secret \
  -v /data/postgres:/var/lib/postgresql/data \
  -p 5432:5432 \
  postgres:15
```

Then connect from your app:

```yaml
apps:
  - name: my-app
    environments:
      - name: production
        default_env_vars:
          DATABASE_URL: postgres://postgres:secret@172.17.0.1:5432/mydb
```

## Health Checks

Add health check to your Dockerfile:

```dockerfile
FROM python:3.11-slim

WORKDIR /app
COPY . .

EXPOSE ${PORT:-8080}

HEALTHCHECK --interval=30s --timeout=3s \
  CMD curl -f http://localhost:${PORT:-8080}/health || exit 1

CMD ["python", "app.py"]
```

## Image Tagging

Gokku automatically tags images with release number:

```
my-app:release-1
my-app:release-2
my-app:release-3
```

For rollback:

```bash
ssh ubuntu@server "cd /opt/gokku && ./rollback.sh my-app production 2"
```

This switches back to `my-app:release-2`.

## Registry Support (Private Images)

```yaml
docker:
  registry: "registry.example.com"

apps:
  - name: my-app
    build:
      base_image: registry.example.com/python:3.11
```

Login on server first:

```bash
ssh ubuntu@server "docker login registry.example.com"
```

## Debugging

### View Container Logs

```bash
ssh ubuntu@server "docker logs -f my-app-production"
```

### Inspect Container

```bash
ssh ubuntu@server "docker inspect my-app-production"
```

### Access Container Shell

```bash
ssh ubuntu@server "docker exec -it my-app-production /bin/sh"
```

### View Build Logs

```bash
ssh ubuntu@server "cat /opt/gokku/apps/my-app/production/deploy.log"
```

## Docker Benefits

Docker provides several advantages for application deployment:

- **Isolation**: Each app runs in its own container
- **Consistency**: Same environment in dev and production
- **Dependencies**: All dependencies bundled in the image
- **Portability**: Easy to move between servers
- **Zero-Downtime**: Built-in blue-green deployment support

## Complete Examples

Full working examples:

- [Python with Docker](https://github.com/thadeu/gokku-examples/tree/main/python-docker)
- [Node.js with Docker](https://github.com/thadeu/gokku-examples/tree/main/nodejs-docker)
- [Custom Dockerfile](https://github.com/thadeu/gokku-examples/tree/main/custom-dockerfile)

## Next Steps

- [Configuration](/guide/configuration) - Customize Docker settings
- [Environment Variables](/guide/env-vars) - Manage secrets

