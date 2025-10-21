# Environment Variables

Manage environment variables for your applications with Gokku.

## Overview

Gokku provides a unified CLI tool for managing environment variables, similar to Heroku's `config:set`, accessible both locally on the server and remotely from your machine.

## gokku config

The `gokku config` command manages environment variables.

### Usage Modes

**Remote execution (from local machine):**
```bash
gokku config <command> [args] -a <app-name>
```

**Local execution (on server):**
```bash
gokku config <command> [args] --app <app-name>
```

## Commands

### set - Set Variable

Set environment variable(s):

**From local machine:**
```bash
gokku config set PORT=8080 -a api-production
gokku config set DATABASE_URL="postgres://..." -a api-production
```

**On server:**
```bash
gokku config set PORT=8080 --app api
gokku config set DATABASE_URL="postgres://user:pass@localhost/db" --app api
```

**Multiple variables:**

```bash
gokku config set PORT=8080 -a api-production
gokku config set LOG_LEVEL=info -a api-production
gokku config set WORKERS=4 -a api-production
```

### get - Get Variable

Get a variable value:

**From local machine:**
```bash
gokku config get PORT -a api-production
```

**On server:**
```bash
gokku config get PORT --app api
```

Output:
```
PORT=8080
```

### list - List All Variables

List all environment variables:

**From local machine:**
```bash
gokku config list -a api-production
```

**On server:**
```bash
gokku config list --app api
```

Output:
```
DATABASE_URL=postgres://user:pass@localhost/db
LOG_LEVEL=info
PORT=8080
WORKERS=4
```

### unset - Delete Variable

Delete an environment variable:

**From local machine:**
```bash
gokku config unset PORT -a api-production
```

**On server:**
```bash
gokku config unset PORT --app api
```

## Storage

Environment variables are stored in:

```
/opt/gokku/apps/APP_NAME/shared/.env
```

Example: `/opt/gokku/apps/api/shared/.env`

### Format

Standard `.env` format:

```env
DATABASE_URL=postgres://localhost/db
LOG_LEVEL=info
PORT=8080
REDIS_URL=redis://localhost:6379
```

### How It Works

- Environment variables are stored in a single shared `.env` file per application
- Each release directory gets a symlink to this shared file
- Changes to environment variables affect all releases immediately
- No restart is needed for environment variable changes to take effect

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


## Apply Changes

Environment variables are automatically available to your application. No restart is required for environment variable changes to take effect.

### Restart Application (Optional)

If you want to restart the application after changing environment variables:

**From local machine:**
```bash
gokku restart -a api-production
```

**On server:**
```bash
gokku restart --app api
```

## Common Variables

### Server Configuration

```bash
# Port
gokku config set PORT=8080 -a api-production

# Workers/Threads
gokku config set WORKERS=4 -a api-production

# Timeout
gokku config set TIMEOUT=30 -a api-production
```

### Database

```bash
# PostgreSQL
gokku config set DATABASE_URL="postgres://user:pass@host:5432/dbname" -a api-production

# MySQL
gokku config set DATABASE_URL="mysql://user:pass@host:3306/dbname" -a api-production

# MongoDB
gokku config set MONGO_URL="mongodb://user:pass@host:27017/dbname" -a api-production
```

### Cache

```bash
# Redis
gokku config set REDIS_URL="redis://localhost:6379" -a api-production

# Memcached
gokku config set MEMCACHE_SERVERS="localhost:11211" -a api-production
```

### AWS

```bash
gokku config set AWS_ACCESS_KEY_ID="AKIA..." -a api-production
gokku config set AWS_SECRET_ACCESS_KEY="secret..." -a api-production
gokku config set AWS_REGION="us-east-1" -a api-production
gokku config set S3_BUCKET="my-bucket" -a api-production
```

### API Keys

```bash
gokku config set STRIPE_API_KEY="sk_live_..." -a api-production
gokku config set SENDGRID_API_KEY="SG...." -a api-production
gokku config set OPENAI_API_KEY="sk-..." -a api-production
```

### Application

```bash
# Environment
gokku config set APP_ENV=production -a api-production

# Debug mode
gokku config set DEBUG=false -a api-production

# Log level
gokku config set LOG_LEVEL=info -a api-production

# Secret key
gokku config set SECRET_KEY="random-secret-key" -a api-production
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
gokku config set SECRET_KEY="$NEW_SECRET" -a api-production

# Restart app
gokku restart api production -a api-production
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
gokku config set -a production DATABASE_URL="postgres://prod-db/app"
gokku config set -a production LOG_LEVEL=info
gokku config set -a production DEBUG=false

# Staging - Staging database
gokku config set -a staging DATABASE_URL="postgres://staging-db/app"
gokku config set -a staging LOG_LEVEL=debug
gokku config set -a staging DEBUG=true
```

## Next Steps

- [Configuration](/guide/configuration) - Configure apps
- [Deployment](/guide/deployment) - Deploy your app
- [Troubleshooting](/reference/troubleshooting) - Common issues

