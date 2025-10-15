# Environment Variables

Manage environment variables for your applications with Gokku.

## Overview

Gokku provides a unified CLI tool for managing environment variables, similar to Heroku's `config:set`, accessible both locally on the server and remotely from your machine.

## gokku config

The `gokku config` command manages environment variables.

### Usage Modes

**Local execution (on server):**
```bash
gokku config <command> [args] --app <app> [--env <env>]
gokku config <command> [args] -a <app> [-e <env>]  # shorthand
```

**Remote execution (from local machine):**
```bash
gokku config <command> [args] --remote <git-remote>
```

::: tip Default Environment
If `--env` is not specified, Gokku uses `default` as the environment name.
:::

## Commands

### set - Set Variable

Set environment variable(s):

**On server:**
```bash
# Using default environment
gokku config set PORT=8080 -a api

# Explicit environment
gokku config set PORT=8080 -a api -e production
gokku config set DATABASE_URL="postgres://user:pass@localhost/db" -a api -e production
```

**From local machine:**
```bash
gokku config set PORT=8080 --remote api-production
gokku config set DATABASE_URL="postgres://..." --remote api-production
```

**Multiple variables:**

```bash
gokku config set PORT=8080 -a api
gokku config set LOG_LEVEL=info -a api
gokku config set WORKERS=4 -a api
```

### get - Get Variable

Get a variable value:

**On server:**
```bash
gokku config get PORT -a api
gokku config get PORT -a api -e production
```

**From local machine:**
```bash
gokku config get PORT --remote api-production
```

Output:
```
8080
```

### list - List All Variables

List all environment variables:

**On server:**
```bash
gokku config list -a api
gokku config list -a api -e production
```

**From local machine:**
```bash
gokku config list --remote api-production
```

Output:
```
PORT=8080
LOG_LEVEL=info
WORKERS=4
DATABASE_URL=postgres://user:pass@localhost/db
```

### unset - Delete Variable

Delete an environment variable:

**On server:**
```bash
gokku config unset PORT -a api
gokku config unset PORT -a api -e production
```

**From local machine:**
```bash
gokku config unset PORT --remote api-production
```

## Storage

Environment variables are stored in:

```
/opt/gokku/apps/APP_NAME/ENVIRONMENT/.env
```

Example: `/opt/gokku/apps/api/production/.env`

### Format

Standard `.env` format:

```env
PORT=8080
LOG_LEVEL=info
DATABASE_URL=postgres://localhost/db
REDIS_URL=redis://localhost:6379
```

### Access from Application

**Go:**
```go
import "os"

port := os.Getenv("PORT")
dbURL := os.Getenv("DATABASE_URL")
```

**Python:**
```python
import os

port = os.getenv("PORT", "8080")
db_url = os.getenv("DATABASE_URL")
```

**Node.js:**
```javascript
const port = process.env.PORT || 8080;
const dbURL = process.env.DATABASE_URL;
```

## Default Variables

Set default values in `gokku.yml`:

```yaml
apps:
  - name: api
    environments:
      - name: production
        default_env_vars:
          PORT: 8080
          LOG_LEVEL: info
          WORKERS: 4
      
      - name: staging
        default_env_vars:
          PORT: 8080
          LOG_LEVEL: debug
          WORKERS: 2
```

These are created when you run `scripts/deploy-server-setup.sh`.

**Override defaults:**

```bash
gokku config set LOG_LEVEL=debug -a api -e production
```

## Remote Management

With `gokku config --remote`, you don't need SSH commands:

```bash
# Set variable
gokku config set PORT=8081 --remote api-production

# Get variable
gokku config get PORT --remote api-production

# List all
gokku config list --remote api-production
```

## Apply Changes

After changing environment variables, restart the application:

### Using gokku CLI

**From local machine:**
```bash
gokku restart api production --remote api-production
```

**On server:**
```bash
gokku restart api production
```

### Manual Restart

**Systemd:**
```bash
sudo systemctl restart api-production
```

**Docker:**
```bash
docker restart api-production
```

### Or Redeploy

```bash
git push production main
```

## Common Variables

### Server Configuration

```bash
# Port
gokku config set PORT=8080 -a api -e production

# Workers/Threads
gokku config set WORKERS=4 -a api -e production

# Timeout
gokku config set TIMEOUT=30 -a api -e production
```

### Database

```bash
# PostgreSQL
gokku config set DATABASE_URL="postgres://user:pass@host:5432/dbname" -a api -e production

# MySQL
gokku config set DATABASE_URL="mysql://user:pass@host:3306/dbname" -a api -e production

# MongoDB
gokku config set MONGO_URL="mongodb://user:pass@host:27017/dbname" -a api -e production
```

### Cache

```bash
# Redis
gokku config set REDIS_URL="redis://localhost:6379" -a api -e production

# Memcached
gokku config set MEMCACHE_SERVERS="localhost:11211" -a api -e production
```

### AWS

```bash
gokku config set AWS_ACCESS_KEY_ID="AKIA..." -a api -e production
gokku config set AWS_SECRET_ACCESS_KEY="secret..." -a api -e production
gokku config set AWS_REGION="us-east-1" -a api -e production
gokku config set S3_BUCKET="my-bucket" -a api -e production
```

### API Keys

```bash
gokku config set STRIPE_API_KEY="sk_live_..." -a api -e production
gokku config set SENDGRID_API_KEY="SG...." -a api -e production
gokku config set OPENAI_API_KEY="sk-..." -a api -e production
```

### Application

```bash
# Environment
gokku config set APP_ENV=production -a api -e production

# Debug mode
gokku config set DEBUG=false -a api -e production

# Log level
gokku config set LOG_LEVEL=info -a api -e production

# Secret key
gokku config set SECRET_KEY="random-secret-key" -a api -e production
```

## Security Best Practices

### 1. Never Commit Secrets

❌ **Bad:**
```yaml
# gokku.yml
environments:
  - name: production
    default_env_vars:
      DATABASE_URL: "postgres://user:password@localhost/db"  # NO!
```

✅ **Good:**
```bash
# Use gokku config instead
gokku config set DATABASE_URL="postgres://..." -a api -e production
```

### 2. Use Strong Secrets

```bash
# Generate random secret
SECRET=$(openssl rand -hex 32)
gokku config set SECRET_KEY="$SECRET" -a api -e production
```

### 3. Different Secrets Per Environment

```bash
# Production
gokku config set SECRET_KEY="prod-secret" -a api -e production

# Staging
gokku config set SECRET_KEY="staging-secret" -a api -e staging
```

### 4. Rotate Secrets Regularly

```bash
# Generate new secret
NEW_SECRET=$(openssl rand -hex 32)

# Update
gokku config set SECRET_KEY="$NEW_SECRET" --remote api-production

# Restart app
gokku restart api production --remote api-production
```

### 5. Limit Access

```bash
# Set proper permissions on .env file
ssh ubuntu@server "chmod 600 /opt/gokku/apps/api/production/.env"
```

## Multiple Environments

Different variables for different environments:

```bash
# Production - Real database
gokku config set DATABASE_URL="postgres://prod-db/app"
gokku config set LOG_LEVEL=info
gokku config set DEBUG=false

# Staging - Staging database
gokku config set DATABASE_URL="postgres://staging-db/app"
gokku config set LOG_LEVEL=debug
gokku config set DEBUG=true
```

## Environment-Specific Configs

### Development vs Production

```bash
# Production
gokku config set NODE_ENV=production
gokku config set LOG_LEVEL=error
gokku config set CACHE_TTL=3600

# Staging/Development
gokku config set NODE_ENV=development
gokku config set LOG_LEVEL=debug
gokku config set CACHE_TTL=60
```

## Troubleshooting

### Variable Not Found

Check if variable exists:

```bash
ssh ubuntu@server "cd /opt/gokku && gokku config get PORT"
```

If empty, set it:

```bash
ssh ubuntu@server "cd /opt/gokku && gokku config set PORT=8080"
```

### Changes Not Applied

Restart the application:

```bash
# Using gokku CLI
gokku restart api production --remote api-production

# Or manually on server
sudo systemctl restart api-production  # Systemd
docker restart api-production          # Docker
```

### .env File Missing

Re-run setup:

```bash
ssh ubuntu@server "cd /opt/gokku && ./scripts/deploy-server-setup.sh api production"
```

### Permission Denied

Fix permissions:

```bash
ssh ubuntu@server "sudo chown ubuntu:ubuntu /opt/gokku/apps/api/production/.env"
ssh ubuntu@server "chmod 600 /opt/gokku/apps/api/production/.env"
```

## Advanced Usage

### Bulk Set

Set multiple variables at once:

```bash
ssh ubuntu@server "cd /opt/gokku && cat << 'EOF' > /opt/gokku/apps/api/production/.env
PORT=8080
LOG_LEVEL=info
WORKERS=4
DATABASE_URL=postgres://localhost/db
REDIS_URL=redis://localhost:6379
EOF"
```

### Export from Local .env

```bash
# Read local .env and set on server
while IFS= read -r line; do
  if [[ $line && $line != \#* ]]; then
    ssh ubuntu@server "cd /opt/gokku && gokku config set $line"
  fi
done < .env.production
```

### Backup Variables

```bash
# Backup
ssh ubuntu@server "cat /opt/gokku/apps/api/production/.env" > backup.env

# Restore
scp backup.env ubuntu@server:/opt/gokku/apps/api/production/.env
gokku restart api production --remote api-production
```

### Template Variables

Use variable substitution in your app:

```bash
# Set template
gokku config set DATABASE_HOST=localhost
gokku config set DATABASE_NAME=mydb

# In app, construct URL
# DATABASE_URL = f"postgres://{os.getenv('DATABASE_HOST')}/{os.getenv('DATABASE_NAME')}"
```

## Additional Commands

The `gokku` CLI provides additional management commands:

### Restart Application

```bash
# From local machine
gokku restart api production --remote api-production

# On server
gokku restart api production
```

### View Logs

```bash
# From local machine
gokku logs api production --remote api-production

# On server  
gokku logs api production
```

### Check Status

```bash
# From local machine
gokku status api production --remote api-production

# On server
gokku status api production
```

## Next Steps

- [Configuration](/guide/configuration) - Configure apps
- [Deployment](/guide/deployment) - Deploy your app
- [Troubleshooting](/reference/troubleshooting) - Common issues

