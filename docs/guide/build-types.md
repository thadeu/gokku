# Build Types

Gokku supports multiple build types to accommodate different application architectures and deployment scenarios.

## Overview

Gokku automatically detects and builds applications based on their structure and configuration files. The build process is optimized for different programming languages and frameworks.

## Supported Build Types

### Go Applications

**Detection:** Presence of `go.mod` file
**Build Process:**
- Installs dependencies with `go mod download`
- Builds with `go build -o app .`
- Supports Go modules and vendoring

**Configuration:**
```yaml
apps:
  my-go-app:
    build:
      type: go
      version: "1.21"
```

### Python Applications

**Detection:** Presence of `requirements.txt`, `pyproject.toml`, or `Pipfile`
**Build Process:**
- Creates virtual environment
- Installs dependencies
- Supports WSGI/ASGI applications

**Configuration:**
```yaml
apps:
  my-python-app:
    build:
      type: python
      version: "3.11"
      requirements: requirements.txt
```

### Node.js Applications

**Detection:** Presence of `package.json`
**Build Process:**
- Runs `npm install` or `yarn install`
- Executes build scripts if defined
- Supports static generation and API routes

**Configuration:**
```yaml
apps:
  my-node-app:
    build:
      type: nodejs
      version: "18"
      build_command: "npm run build"
```

### Docker Applications

**Detection:** Presence of `Dockerfile`
**Build Process:**
- Builds Docker image
- Uses multi-stage builds when available
- Supports custom Dockerfiles

**Configuration:**
```yaml
apps:
  my-docker-app:
    build:
      type: docker
      dockerfile: Dockerfile
      context: .
```

### Static Sites

**Detection:** Presence of static files (HTML, CSS, JS)
**Build Process:**
- Serves static files directly
- No build process required
- Supports SPA routing

**Configuration:**
```yaml
apps:
  my-static-site:
    build:
      type: static
      root: dist/
```

## Build Configuration

### Build Commands

You can customize build commands for each application:

```yaml
apps:
  my-app:
    build:
      pre_build:
        - "echo 'Starting build...'"
      build: "npm run build"
      post_build:
        - "echo 'Build completed'"
        - "./scripts/optimize.sh"
```

### Environment Variables

Set build-time environment variables:

```yaml
apps:
  my-app:
    build:
      env:
        NODE_ENV: production
        API_URL: https://api.example.com
```

### Caching

Gokku caches build artifacts to speed up subsequent deployments:

```yaml
apps:
  my-app:
    build:
      cache:
        - node_modules
        - .next/cache
```

## Advanced Configuration

### Custom Build Scripts

For complex build processes, use custom scripts:

```yaml
apps:
  my-app:
    build:
      type: custom
      script: "./build.sh"
```

### Build Hooks

Execute commands at different build stages:

```yaml
apps:
  my-app:
    build:
      hooks:
        before_install: "./scripts/setup.sh"
        after_build: "./scripts/test.sh"
        before_deploy: "./scripts/migrate.sh"
```

## Troubleshooting

### Common Issues

**Build fails with missing dependencies:**
- Ensure all required system packages are installed
- Check that language-specific package managers are available

**Build takes too long:**
- Enable build caching
- Use pre-built base images for Docker builds
- Optimize dependency installation

**Environment variables not available:**
- Check variable naming (use uppercase)
- Ensure variables are set before build starts

### Build Logs

View build logs with:
```bash
gokku logs my-app build
```

### Build Cache Management

Clear build cache when needed:
```bash
gokku build cache clear my-app
```
