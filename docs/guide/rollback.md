# Rollback

Quickly revert to a previous deployment when things go wrong.

## Overview

Gokku keeps multiple releases, allowing you to rollback instantly without redeploying.

## How It Works

### For Systemd Apps

Each deployment creates a new release directory:

```
/opt/gokku/apps/api/production/
├── releases/
│   ├── 1/        ← First deploy
│   ├── 2/        ← Second deploy
│   ├── 3/        ← Third deploy
│   ├── 4/        ← Fourth deploy
│   └── 5/        ← Latest deploy
└── current -> releases/5  ← Symlink to active release
```

The `current` symlink points to the active release.

### For Docker Apps

Each deployment creates a tagged image:

```
api:release-1
api:release-2
api:release-3
api:release-4
api:release-5  ← Latest
```

The container runs the latest tagged image.

## Configure Retention

Set how many releases/images to keep:

```yaml
apps:
  - name: api
    deployment:
      keep_releases: 10  # Keep last 10 releases (systemd)
      keep_images: 10    # Keep last 10 images (docker)
```

**Defaults:**
- `keep_releases`: 5
- `keep_images`: 5

## Manual Rollback

### Systemd Apps

#### List Available Releases

```bash
ssh ubuntu@server "ls -la /opt/gokku/apps/api/production/releases/"
```

Output:
```
1/  2024-01-15 10:00
2/  2024-01-15 11:30
3/  2024-01-15 14:20
4/  2024-01-16 09:15
5/  2024-01-16 10:45  ← Current
```

#### Rollback to Previous Release

```bash
# SSH to server
ssh ubuntu@server

# Navigate to app directory
cd /opt/gokku/apps/api/production

# Change symlink to previous release
rm current
ln -s releases/4 current

# Restart service
sudo systemctl restart api-production
```

#### Verify

```bash
sudo systemctl status api-production
```

### Docker Apps

#### List Available Images

```bash
ssh ubuntu@server "docker images | grep api"
```

Output:
```
api  release-5  1.2GB  10 minutes ago
api  release-4  1.2GB  2 hours ago
api  release-3  1.2GB  1 day ago
```

#### Rollback to Previous Image

```bash
# Stop current container
ssh ubuntu@server "docker stop api-production"
ssh ubuntu@server "docker rm api-production"

# Start with previous image
ssh ubuntu@server "docker run -d \
  --name api-production \
  --env-file /opt/gokku/apps/api/production/.env \
  -p 8080:8080 \
  api:release-4"
```

#### Verify

```bash
ssh ubuntu@server "docker logs api-production"
```

## Rollback Script (Planned)

Future `rollback.sh` script:

```bash
# Systemd
./rollback.sh api production 4

# Docker
./rollback.sh api production 4
```

This will:
1. Detect build type (systemd or docker)
2. Rollback to specified release
3. Restart service/container
4. Verify deployment

## Quick Rollback Workflow

### 1. Deploy New Version

```bash
git push production main
```

Output:
```
-----> Deploying api to production...
-----> Release 5 deployed
```

### 2. Notice Issue

```bash
# Check logs
ssh ubuntu@server "sudo journalctl -u api-production -f"

# Test endpoint
curl https://api.example.com/health
# Error!
```

### 3. Rollback

```bash
# Systemd
ssh ubuntu@server "cd /opt/gokku/apps/api/production && rm current && ln -s releases/4 current && sudo systemctl restart api-production"

# Docker
ssh ubuntu@server "docker stop api-production && docker rm api-production && docker run -d --name api-production --env-file /opt/gokku/apps/api/production/.env -p 8080:8080 api:release-4"
```

### 4. Verify

```bash
curl https://api.example.com/health
# OK!
```

### 5. Fix Issue Locally

```bash
# Fix the bug
git commit -am "fix: critical bug"

# Redeploy
git push production main
```

## Rollback Best Practices

### 1. Test Before Rollback

Check if issue is deployment-related:

```bash
# Check service status
ssh ubuntu@server "sudo systemctl status api-production"

# Check logs
ssh ubuntu@server "sudo journalctl -u api-production -n 100"

# Test endpoint
curl https://api.example.com/health
```

### 2. Document Rollback

Keep track of rollbacks:

```bash
# Note the issue
echo "2024-01-16 10:50 - Rolled back from release 5 to 4 due to database migration issue" >> rollback.log
```

### 3. Investigate Root Cause

After rollback, find and fix the issue:

```bash
# Get logs from failed deployment
ssh ubuntu@server "cat /opt/gokku/apps/api/production/releases/5/deploy.log"

# Review changes
git diff releases/4 releases/5
```

### 4. Don't Delete Failed Release

Keep it for debugging:

```bash
# Don't do this immediately
rm -rf /opt/gokku/apps/api/production/releases/5
```

### 5. Test Staging First

Deploy to staging before production:

```bash
# Deploy to staging
git push staging main

# Test thoroughly
curl https://staging.example.com/health

# Then deploy to production
git push production main
```

## Database Migrations

### Problem

Rolling back code doesn't rollback database:

```
Release 5: Add column 'email' to users
Release 4: Code doesn't know about 'email' column
```

Rollback to Release 4 → Code expects old schema!

### Solutions

#### 1. Backward Compatible Migrations

Always write backward-compatible migrations:

```sql
-- Good: Add column with default
ALTER TABLE users ADD COLUMN email VARCHAR(255) DEFAULT '';

-- Bad: Add required column
ALTER TABLE users ADD COLUMN email VARCHAR(255) NOT NULL;
```

#### 2. Two-Phase Migrations

**Phase 1:** Add column (deploy Release 5)
```sql
ALTER TABLE users ADD COLUMN email VARCHAR(255);
```

**Phase 2:** Make required (deploy Release 6)
```sql
ALTER TABLE users ALTER COLUMN email SET NOT NULL;
```

Now rollback from Release 6 to Release 5 is safe.

#### 3. Separate Migration Rollback

If you must rollback:

```bash
# 1. Rollback code
ssh ubuntu@server "cd /opt/gokku/apps/api/production && rm current && ln -s releases/4 current"

# 2. Rollback database manually
ssh ubuntu@server "psql -d mydb -c \"ALTER TABLE users DROP COLUMN email;\""

# 3. Restart
ssh ubuntu@server "sudo systemctl restart api-production"
```

## Environment Variables

### After Rollback

Check if env vars are compatible:

```bash
# List current vars
gokku config list -a api-production

# If needed, restore backup
scp backup.env ubuntu@server:/opt/gokku/apps/api/production/.env
ssh ubuntu@server "sudo systemctl restart api-production"
```

## Partial Rollback

### Rollback One App in Monorepo

If deploying multiple apps:

```bash
# Rollback only API
ssh ubuntu@server "cd /opt/gokku/apps/api/production && rm current && ln -s releases/4 current && sudo systemctl restart api-production"

# Keep worker on latest
# (no changes)
```

### Rollback Environment Only

Rollback staging but not production:

```bash
# Rollback staging
ssh ubuntu@server "cd /opt/gokku/apps/api/staging && rm current && ln -s releases/3 current && sudo systemctl restart api-staging"

# Production stays on latest
```

## Automated Rollback (Future)

Planned auto-rollback on failure:

```yaml
apps:
  - name: api
    deployment:
      auto_rollback: true
      health_check: /health
```

If deployment fails health check, automatically rollback.

## Troubleshooting

### Symlink Error (Systemd)

```
ln: failed to create symbolic link 'current': File exists
```

**Fix:**
```bash
ssh ubuntu@server "cd /opt/gokku/apps/api/production && rm -f current && ln -s releases/4 current"
```

### Container Won't Start (Docker)

```
docker: Error response from daemon: Conflict
```

**Fix:**
```bash
# Remove old container first
ssh ubuntu@server "docker rm -f api-production"

# Then start with old image
ssh ubuntu@server "docker run -d --name api-production ... api:release-4"
```

### Release Not Found

```
ls: cannot access 'releases/4': No such file or directory
```

**Fix:**

Release was cleaned up. Check available releases:

```bash
ssh ubuntu@server "ls /opt/gokku/apps/api/production/releases/"
```

Rollback to an existing one.

### Service Won't Restart

```bash
# Check status
ssh ubuntu@server "sudo systemctl status api-production"

# Check logs
ssh ubuntu@server "sudo journalctl -u api-production -n 50"

# Manual restart
ssh ubuntu@server "sudo systemctl daemon-reload && sudo systemctl restart api-production"
```

## Monitoring Rollbacks

### Track Rollbacks

Keep a log:

```bash
# Create rollback log
ssh ubuntu@server "echo \"$(date) - Rollback api-production: release-5 -> release-4\" >> /opt/gokku/rollback.log"

# View history
ssh ubuntu@server "cat /opt/gokku/rollback.log"
```

### Alert on Rollback

Send notification when rollback happens (custom script):

```bash
#!/bin/bash
# rollback-notify.sh

# Send to Slack, email, etc.
curl -X POST "https://slack.com/api/chat.postMessage" \
  -d "text=Rollback performed: $APP_NAME from $OLD_RELEASE to $NEW_RELEASE"
```

## Next Steps

- [Deployment](/guide/deployment) - Deployment strategies
- [Environment Variables](/guide/env-vars) - Manage configs
- [Troubleshooting](/reference/troubleshooting) - Debug issues

