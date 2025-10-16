# Procfile Support

Gokku supports **Procfile-based deployment** for multi-process applications, providing the same experience as Dokku!

## What is a Procfile?

A Procfile declares the **process types** that make up your application. Each line defines a process name and its command:

```bash
# Procfile
web: bundle exec rails server -p $PORT
worker: bundle exec sidekiq -c 5
scheduler: bundle exec whenever
```

## Procfile vs gokku.yml

### When to use Procfile (Single App, Multiple Processes)

Use **Procfile** para aplicações **single-app com múltiplos processos** (caso de uso 1):

- ✅ **Rails apps** com web, worker, scheduler
- ✅ **Node.js apps** com web, api, worker
- ✅ **Qualquer app monolítica** que precisa de múltiplos processos
- ✅ **Migração fácil do Dokku**

**Características:**
- Procfile define os processos executados
- Docker é automaticamente usado
- gokku.yml é opcional (apenas configurações de ambiente)
- **Procfile tem prioridade** sobre gokku.yml

### When to use gokku.yml (Multiple Independent Apps)

Use **gokku.yml** para aplicações **multi-app independentes** (caso de uso 2):

- ✅ **Microserviços** com deploy independente
- ✅ **Apps Go** com cmd/api, cmd/worker, cmd/scheduler
- ✅ **Deploy granular** - atualize apenas um serviço
- ✅ **Flexibilidade total** de configuração por app

**Características:**
- Cada app é independente
- Pode usar systemd ou docker
- Deploy separado por app
- Controle total de cada processo

## Process Types

### web
- **Automatically scaled to 1 instance**
- Gets the main `PORT` environment variable
- Receives HTTP traffic
- Example: Rails server, Express app, Django server

### Other Processes
- **Background workers, schedulers, etc.**
- Can be scaled independently
- Run without exposed ports
- Example: Sidekiq workers, cron jobs, background processors

## Configuration

### 1. Create Procfile

In your project root:

```bash
# Rails app
web: bundle exec rails server -p $PORT
worker: bundle exec sidekiq
scheduler: bundle exec whenever

# Node.js app
web: npm start
api: node api.js
worker: node worker.js

# Python app
web: python app.py
worker: python worker.py
```

### 2. Configure gokku.yml

```yaml
apps:
  - name: my-app
    lang: ruby  # or nodejs, python, etc.
    build:
      type: docker  # Required for Procfile
      path: .
    environments:
      - name: production
        branch: main
        default_env_vars:
          RAILS_ENV: production
          PORT: 3000
          REDIS_URL: redis://localhost:6379
```

### 3. Deploy

```bash
git push production main
```

Gokku automatically:
- Detects the Procfile
- Generates appropriate Dockerfile
- Creates individual containers for each process
- Sets up systemd services
- Starts all processes

## Process Management

### List Processes

```bash
gokku ps:list my-app production --remote my-app-production
```

Output:
```
=== Procfile Processes: my-app (production) ===

web:
  Systemd: active
  Container: running

worker:
  Systemd: active
  Container: running

scheduler:
  Systemd: active
  Container: running
```

### View Logs

```bash
# All processes
gokku ps:logs my-app production -f --remote my-app-production

# Specific process
gokku ps:logs my-app production web -f --remote my-app-production
```

### Scale Processes

```bash
# Scale worker to 3 instances (Dokku-style)
gokku ps:scale my-app production worker=3 --remote my-app-production

# Scale down scheduler
gokku ps:scale my-app production scheduler=0 --remote my-app-production
```

### Start/Stop/Restart

```bash
# Restart all processes
gokku ps:restart my-app production --remote my-app-production

# Restart specific process
gokku ps:restart my-app production worker --remote my-app-production

# Stop specific process
gokku ps:stop my-app production scheduler --remote my-app-production

# Start stopped process
gokku ps:start my-app production scheduler --remote my-app-production
```

## Automatic Setup

### Dockerfile Generation

Gokku automatically generates a Dockerfile based on your language:

**Ruby:**
```dockerfile
FROM ruby:3.2-alpine
WORKDIR /app
COPY Gemfile* ./
RUN bundle install
COPY . .
EXPOSE ${PORT:-3000}
# Commands overridden by Procfile processes
```

**Node.js:**
```dockerfile
FROM node:20-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY . .
EXPOSE ${PORT:-3000}
```

**Python:**
```dockerfile
FROM python:3.11-slim
WORKDIR /app
RUN apt-get update && apt-get install -y gcc && rm -rf /var/lib/apt/lists/*
COPY requirements*.txt ./
RUN pip install --no-cache-dir -r requirements.txt
COPY . .
EXPOSE ${PORT:-8000}
```

### Custom Dockerfile

You can provide your own Dockerfile:

```yaml
apps:
  - name: my-app
    build:
      type: docker
      dockerfile: ./Dockerfile  # Your custom Dockerfile
      path: .
```

## Environment Variables

All processes share the same environment variables:

```bash
# Set for all processes
gokku config set DATABASE_URL=postgres://... --remote my-app-production

# Process-specific (use in Procfile commands)
gokku config set WORKER_COUNT=5 --remote my-app-production
```

## Process Isolation

Each process runs in its own Docker container:
- Independent scaling
- Isolated failures
- Individual logging
- Separate resource usage

## Migration from Dokku

### Same Procfile Format
```bash
# This works in both Dokku and Gokku
web: bundle exec rails server -p $PORT
worker: bundle exec sidekiq
```

### Same Commands
```bash
# Dokku
dokku ps:scale myapp worker=2

# Gokku (same syntax!)
gokku ps:scale myapp production worker=2
```

### Drop-in Replacement
- Same Procfile format
- Same scaling syntax
- Same process management
- Better systemd integration

## Examples

### Rails Application

**Procfile:**
```
web: bundle exec rails server -p $PORT -e $RAILS_ENV
worker: bundle exec sidekiq -e $RAILS_ENV
scheduler: bundle exec whenever --update-crontab && cron -f
```

**gokku.yml:**
```yaml
apps:
  - name: rails-app
    lang: ruby
    build:
      type: docker
      path: .
    environments:
      - name: production
        default_env_vars:
          RAILS_ENV: production
          PORT: 3000
```

### Node.js API with Workers

**Procfile:**
```
web: npm start
api: node api.js
worker: node worker.js
```

**gokku.yml:**
```yaml
apps:
  - name: node-app
    lang: nodejs
    build:
      type: docker
      path: .
    environments:
      - name: production
        default_env_vars:
          NODE_ENV: production
          PORT: 3000
```

## Best Practices

### Process Naming
- Use descriptive names: `web`, `worker`, `api`, `scheduler`
- Keep names short and clear
- Follow Heroku/Dokku conventions

### Scaling
- `web` process: usually 1 instance (load balancer handles scaling)
- Worker processes: scale based on queue size
- Background processes: 1 instance unless specifically needed

### Environment Variables
- Put shared config in `gokku.yml` default_env_vars
- Use `gokku config` for secrets and environment-specific values
- Access via `$VAR_NAME` in Procfile commands

### Monitoring
```bash
# Check all processes
gokku ps:list my-app production --remote my-app-production

# Monitor logs
gokku ps:logs my-app production worker -f --remote my-app-production

# Check status
gokku status --remote my-app-production
```

## Troubleshooting

### Process Won't Start
```bash
# Check logs
gokku ps:logs my-app production web --remote my-app-production

# Check systemd status
ssh ubuntu@server "sudo systemctl status my-app-production-web"
```

### Container Issues
```bash
# Check Docker logs
ssh ubuntu@server "docker logs my-app-production-web"

# Check container status
ssh ubuntu@server "docker ps -a | grep my-app"
```

### Environment Variables
```bash
# List all variables
gokku config list --remote my-app-production

# Check specific variable
gokku config get DATABASE_URL --remote my-app-production
```

## Next Steps

- [Getting Started](/guide/getting-started) - Basic deployment
- [Configuration](/guide/configuration) - gokku.yml setup
- [Docker Support](/guide/docker) - Container deployment
- [Examples](/examples/) - Real-world configurations
