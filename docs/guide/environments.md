# Environments

Manage multiple deployment environments (production, staging, development) with Gokku.

## Overview

Gokku supports multiple environments per application:

- **Production** - Live application
- **Staging** - Pre-production testing
- **Development** - Development testing
- **Custom** - Any name you want

## Configuration

Define environments in `gokku.yml`:

```yaml
apps:
  - name: api
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

## Setup Environments

### Create Production Environment

```bash
ssh ubuntu@server "cd /opt/gokku && ./deploy-server-setup.sh api production"
```

Creates:
- `api` (Git repo)
- `/opt/gokku/apps/api/production/` (App directory)
- `api-production` (Systemd service or Docker container)

### Create Staging Environment

```bash
ssh ubuntu@server "cd /opt/gokku && ./deploy-server-setup.sh api staging"
```

Creates:
- `/opt/gokku/apps/api/staging/` (Separate from production)
- `api-staging` (Separate service)

## Git Remotes

Add remotes for each environment:

```bash
# Production
git remote add production ubuntu@server:api

# Staging
git remote add staging ubuntu@server:api
```

Same Git repository, different branches.

## Deploy to Environments

### Deploy to Production

```bash
git push production main
```

Deploys `main` branch to production.

### Deploy to Staging

```bash
git push staging staging
```

Deploys `staging` branch to staging.

### Deploy Same Commit

```bash
# Deploy feature branch to staging
git push staging feature/new-feature

# After testing, merge and deploy to production
git checkout main
git merge feature/new-feature
git push production main
```

## Environment Configuration

### Branch Mapping

Map branches to environments:

```yaml
environments:
  - name: production
    branch: main         # Push main to production
  
  - name: staging
    branch: staging      # Push staging to staging
  
  - name: develop
    branch: develop      # Push develop to develop
```

**Smart defaults:**
- `production` → `main`
- `staging` → `staging`
- `develop` → `develop`
- Custom → same as environment name

### Environment Variables

Different variables per environment:

```yaml
environments:
  - name: production
    default_env_vars:
      DATABASE_URL: postgres://prod-db/app
      LOG_LEVEL: error
      CACHE_TTL: 3600
      DEBUG: false
  
  - name: staging
    default_env_vars:
      DATABASE_URL: postgres://staging-db/app
      LOG_LEVEL: debug
      CACHE_TTL: 60
      DEBUG: true
```

Override with `gokku config`:

```bash
# Production
gokku config set DATABASE_URL=postgres://... --remote api-production

# Staging
gokku config set DATABASE_URL=postgres://... --remote api-staging
```

## Directory Structure

Each environment is isolated:

```
/opt/gokku/apps/api/
├── production/
│   ├── releases/
│   ├── current -> releases/5
│   └── .env
└── staging/
    ├── releases/
    ├── current -> releases/3
    └── .env
```

## Service Management

Each environment has its own service:

### Production

```bash
# Status
ssh ubuntu@server "sudo systemctl status api-production"

# Logs
ssh ubuntu@server "sudo journalctl -u api-production -f"

# Restart
ssh ubuntu@server "sudo systemctl restart api-production"
```

### Staging

```bash
# Status
ssh ubuntu@server "sudo systemctl status api-staging"

# Logs
ssh ubuntu@server "sudo journalctl -u api-staging -f"

# Restart
ssh ubuntu@server "sudo systemctl restart api-staging"
```

## Workflow Examples

### Basic Workflow

```bash
# 1. Develop locally
git checkout -b feature/new-feature
# ... make changes ...
git commit -am "Add new feature"

# 2. Deploy to staging
git push staging feature/new-feature

# 3. Test staging
curl https://staging.example.com/api/test

# 4. Merge to main
git checkout main
git merge feature/new-feature

# 5. Deploy to production
git push production main
```

### Hotfix Workflow

```bash
# 1. Create hotfix branch
git checkout -b hotfix/critical-bug main

# 2. Fix the bug
git commit -am "Fix critical bug"

# 3. Deploy directly to production
git push production hotfix/critical-bug

# 4. Merge back
git checkout main
git merge hotfix/critical-bug
git push origin main
```

### Feature Branch Workflow

```bash
# Developer 1: Feature A
git checkout -b feature/a
git push staging feature/a  # Test on staging

# Developer 2: Feature B
git checkout -b feature/b
git push staging feature/b  # Overwrites staging (OK for testing)

# When ready, merge to main
git checkout main
git merge feature/a
git merge feature/b
git push production main
```

## Multiple Servers

Different servers per environment:

```bash
# Production server
git remote add production ubuntu@prod-server:api

# Staging server (different server)
git remote add staging ubuntu@staging-server:api
```

Deploy:

```bash
# Deploy to production server
git push production main

# Deploy to staging server
git push staging staging
```

## Port Management

Assign different ports per environment:

```yaml
environments:
  - name: production
    default_env_vars:
      PORT: 8080
  
  - name: staging
    default_env_vars:
      PORT: 8081
```

Or set manually:

```bash
gokku config set PORT=8080 --remote api-production
gokku config set PORT=8081 --remote api-staging
```

## Database Per Environment

### Separate Databases

```yaml
environments:
  - name: production
    default_env_vars:
      DATABASE_URL: postgres://localhost/app_production
  
  - name: staging
    default_env_vars:
      DATABASE_URL: postgres://localhost/app_staging
```

### Shared Database with Schemas

```yaml
environments:
  - name: production
    default_env_vars:
      DATABASE_URL: postgres://localhost/app?schema=production
  
  - name: staging
    default_env_vars:
      DATABASE_URL: postgres://localhost/app?schema=staging
```

## Environment Promotion

Test in staging, promote to production:

```bash
# 1. Deploy to staging
git push staging main

# 2. Run tests
./test-staging.sh

# 3. If OK, deploy to production
git push production main
```

## Comparing Environments

### View Configurations

```bash
# Production config
gokku config list --remote api-production

# Staging config
gokku config list --remote api-staging
```

### View Code Differences

```bash
# Show what's in staging but not in production
git log production/main..staging/staging

# Show file differences
git diff production/main staging/staging
```

### View Running Versions

```bash
# Production version
ssh ubuntu@server "ls -la /opt/gokku/apps/api/production/current"

# Staging version
ssh ubuntu@server "ls -la /opt/gokku/apps/api/staging/current"
```

## Environment-Specific Code

Sometimes you need different behavior per environment:

```go
package main

import "os"

func main() {
    env := os.Getenv("APP_ENV")
    
    if env == "production" {
        // Production behavior
        enableMonitoring()
    } else {
        // Staging/dev behavior
        enableDebugMode()
    }
}
```

Set `APP_ENV`:

```bash
gokku config set APP_ENV=production --remote api-production
gokku config set APP_ENV=staging --remote api-staging
```

## Custom Environments

Create any environment name:

```yaml
environments:
  - name: production
  - name: staging
  - name: demo        # Custom
  - name: load-test   # Custom
```

Setup:

```bash
ssh ubuntu@server "cd /opt/gokku && ./deploy-server-setup.sh api demo"
ssh ubuntu@server "cd /opt/gokku && ./deploy-server-setup.sh api load-test"
```

Deploy:

```bash
git remote add demo ubuntu@server:api
git push demo main
```

## Troubleshooting

### Wrong Environment Deployed

Check which branch is deployed:

```bash
# Production
ssh ubuntu@server "cd /opt/gokku/apps/api/production/current && git log -1"

# Staging
ssh ubuntu@server "cd /opt/gokku/apps/api/staging/current && git log -1"
```

### Environment Variables Not Applied

Restart service:

```bash
ssh ubuntu@server "sudo systemctl restart api-production"
ssh ubuntu@server "sudo systemctl restart api-staging"
```

### Port Conflict

Check which environment is using which port:

```bash
ssh ubuntu@server "sudo lsof -i :8080"
```

Change port:

```bash
gokku config set PORT=8081 --remote api-staging
ssh ubuntu@server "sudo systemctl restart api-staging"
```

## Best Practices

### 1. Always Test Staging First

```bash
# Deploy to staging
git push staging main

# Test
./test-staging.sh

# Only then deploy to production
git push production main
```

### 2. Keep Staging Similar to Production

Use same:
- Database schema
- Environment variables (different values)
- Server resources
- Dependencies

### 3. Automate Promotion

```bash
# deploy.sh
#!/bin/bash

# Deploy to staging
git push staging main

# Wait for tests
./wait-for-tests.sh

# If OK, deploy to production
if [ $? -eq 0 ]; then
    git push production main
fi
```

### 4. Document Differences

Keep a document of what's different:

```markdown
# Environment Differences

## Databases
- Production: prod-db.example.com
- Staging: staging-db.example.com

## Ports
- Production: 8080
- Staging: 8081

## Log Levels
- Production: error
- Staging: debug
```

### 5. Separate Secrets

Never share secrets between environments:

```bash
# Different API keys
./env-manager --app api --env production set API_KEY="prod-key"
./env-manager --app api --env staging set API_KEY="staging-key"
```

## Next Steps

- [Deployment](/guide/deployment) - Deploy to environments
- [Environment Variables](/guide/env-vars) - Manage env vars
- [Rollback](/guide/rollback) - Rollback environments
- [Configuration](/guide/configuration) - Configure environments

