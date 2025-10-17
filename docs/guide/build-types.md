# Build Types

Gokku supports two main deployment strategies: **systemd** for compiled binaries and **docker** for containerized applications.

## Build Types

### systemd (Default)

**Use for:** Go, Rust, C/C++, and other compiled languages
**Process:** Compiles binary and runs it directly on the server
**Benefits:** Fast startup, low resource usage, direct system access

**Configuration:**
```yaml
apps:
  - name: my-go-app
    build:
      type: systemd
      path: ./cmd/api
      binary_name: api
```

**When Gokku detects:** `go.mod`, `Cargo.toml`, `Makefile`, or executable binary

### docker

**Use for:** Node.js, Python, Ruby, Java, and interpreted languages
**Process:** Builds Docker container and runs it with systemd
**Benefits:** Isolated environment, dependency management, multi-language support

**Configuration:**
```yaml
apps:
  - name: my-node-app
    build:
      type: docker
      path: .
```

**When Gokku detects:** `package.json`, `requirements.txt`, `Gemfile`, or `Dockerfile`

## Automatic Detection

Gokku automatically detects the appropriate build type based on your project files:

| Language | Files Detected | Default Build Type |
|----------|----------------|-------------------|
| Go | `go.mod` | systemd |
| Python | `requirements.txt`, `pyproject.toml` | docker |
| Node.js | `package.json` | docker |
| Ruby | `Gemfile` | docker |
| Rust | `Cargo.toml` | systemd |
| Java | `pom.xml`, `build.gradle` | docker |
| Other | `Dockerfile` | docker |

## Build Configuration

### Environment Variables

Set build-time and runtime environment variables:

```yaml
apps:
  - name: my-app
    build:
      type: docker
      env:
        NODE_ENV: production
        DATABASE_URL: postgres://localhost:5432/app
    environments:
      - name: production
        default_env_vars:
          PORT: 3000
          REDIS_URL: redis://localhost:6379
```

### Post-Deploy Commands

Execute commands after successful deployment:

```yaml
apps:
  - name: rails-app
    build:
      type: docker
      path: .
    environments:
      - name: production
        default_env_vars:
          RAILS_ENV: production
    post_deploy:
      - bundle exec rails db:migrate"
```

**Post-deploy commands:**
- Run after successful deployment
- Execute in the application directory
- Fail deployment if any command fails
- Useful for database migrations, asset compilation, cache warming

### Custom Dockerfile

For Docker builds, you can specify a custom Dockerfile:

```yaml
apps:
  - name: my-app
    build:
      type: docker
      dockerfile: ./Dockerfile.production
      path: .
```

## Deployment Process

### systemd Deployment

1. Extracts code to release directory
2. Installs dependencies
3. Builds binary
4. Updates symlink to new release
5. Restarts systemd service
6. **Runs post-deploy commands** (if configured)
7. Cleans up old releases

### Docker Deployment

1. Extracts code to release directory
2. Generates or uses existing Dockerfile
3. Builds Docker image
4. Creates/updates containers
5. Updates systemd service for containers
6. **Runs post-deploy commands** (if configured)
7. Cleans up old images and releases

## Troubleshooting

### systemd Issues

**Service fails to start:**
```bash
# Check service status
sudo systemctl status my-app-production

# View service logs
sudo journalctl -u my-app-production -f

# Check binary permissions
ls -la /opt/gokku/apps/my-app/production/current/my-app
```

**Binary not found:**
- Check build path configuration
- Verify Go module setup
- Check build logs for errors

### Docker Issues

**Container fails to start:**
```bash
# Check container logs
docker logs my-app-production

# Check container status
docker ps -a | grep my-app

# Verify Dockerfile
docker build -t test-build .
```

**Build fails:**
- Check Dockerfile syntax
- Verify base image availability
- Check disk space and permissions

### Post-Deploy Issues

**Commands fail:**
```bash
# Debug post-deploy commands
gokku run "cd /opt/gokku/apps/my-app/production/current && your-command" --remote my-app-production
```

**Database connection issues:**
- Verify DATABASE_URL environment variable
- Check database server status
- Test connection manually

## Performance Considerations

### systemd
- **Pros:** Fast startup, low memory overhead
- **Cons:** Requires system dependencies
- **Best for:** Microservices, APIs, background workers

### Docker
- **Pros:** Isolated environment, easy dependency management
- **Cons:** Higher memory usage, slower startup
- **Best for:** Web apps, complex dependency trees, multi-language stacks
