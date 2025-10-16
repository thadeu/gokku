# CLI Reference

Command-line tools and scripts reference.

## gokku (Client CLI)

Local CLI for managing deployments. Abstracts SSH commands for easier server management.

### Installation

```bash
curl -fsSL https://gokku-vm.com/install | bash
```

### Usage

```bash
gokku [command] [options]
```

### Global Options

- `--remote <git-remote>` - Specify git remote name (e.g., `api-production`, `vad-staging`)

The `--remote` flag uses your **git remote name**:
- Must be a configured git remote
- Gokku runs `git remote get-url <name>` to extract connection info
- Environment is parsed from the remote name suffix

**How it works:**

1. You add a git remote (standard git command):
   ```bash
   git remote add api-production ubuntu@server:api
   git remote add vad-staging ubuntu@server:vad
   ```

2. Gokku parses the remote URL to extract:
   - **SSH host**: `ubuntu@server`
   - **App name**: `api` (from `/repos/api.git`)
   - **Environment**: `production` (from remote name `api-production`)

3. You use the remote name directly:
   ```bash
   gokku config set PORT=8080 --remote api-production
   gokku logs --remote vad-staging -f
   ```

**Remote name format:**
- `<app>-<environment>` (recommended)
- Examples: `api-production`, `vad-staging`, `worker-dev`
- Environment is extracted from the last part after `-`
- If no `-`, defaults to `production`

### Commands

#### server - Manage Servers

```bash
# Add a server
gokku server add <name> <host>

# List servers
gokku server list

# Remove a server
gokku server remove <name>

# Set default server
gokku server set-default <name>
```

**Examples:**
```bash
gokku server add prod ubuntu@ec2-54-123-45-67.compute-1.amazonaws.com
gokku server list
gokku server set-default prod
```

#### config - Manage Environment Variables

Manage application environment variables without SSH commands.

```bash
# Set variable
gokku config set KEY=VALUE --remote <app>-<env>

# Set multiple variables
gokku config set KEY1=VALUE1 KEY2=VALUE2 --remote <app>-<env>

# Get variable
gokku config get KEY --remote <app>-<env>

# List all variables
gokku config list --remote <app>-<env>

# Unset variable
gokku config unset KEY --remote <app>-<env>
```

**Setup (one-time):**
```bash
# Add git remote (standard git command)
git remote add api-production ubuntu@server:api
```

**Examples:**
```bash
# Set variables
gokku config set PORT=8080 --remote api-production
gokku config set DATABASE_URL="postgres://localhost/db" --remote api-production
gokku config set LOG_LEVEL=info WORKERS=4 --remote api-production

# Get variable
gokku config get PORT --remote api-production

# List all
gokku config list --remote vad-production

# Unset
gokku config unset DEBUG --remote api-staging
```

**Before (with SSH):**
```bash
ssh ubuntu@server "cd /opt/gokku && ./env-manager --app api --env production set PORT=8080"
```

**Now (with gokku):**
```bash
# One-time setup
git remote add api-production ubuntu@server:api

# Use it
gokku config set PORT=8080 --remote api-production
```

#### run - Execute Commands

Run arbitrary commands on the server without manual SSH.

```bash
gokku run <command> --remote <app>-<env>
```

**Examples:**
```bash
# Check service status
gokku run "systemctl status api-production" --remote api-production

# View Docker containers
gokku run "docker ps" --remote vad-production

# Run Rails console
gokku run "bundle exec bin/console" --remote app-production

# Check disk usage
gokku run "df -h" --remote api-production

# View specific log file
gokku run "tail -f /var/log/app.log" --remote api-production

# Database migration
gokku run "python manage.py migrate" --remote django-production

# Clear cache
gokku run "redis-cli FLUSHALL" --remote api-production
```

**Before (with SSH):**
```bash
ssh ubuntu@server "docker ps"
```

**Now (with gokku):**
```bash
gokku run "docker ps" --remote api-production
```

#### logs - View Logs

View application logs (systemd or Docker).

```bash
# View logs
gokku logs <app> <env>

# Follow logs
gokku logs <app> <env> -f

# With --remote flag
gokku logs --remote <app>-<env>
gokku logs --remote <app>-<env> -f
```

**Examples:**
```bash
# View last 100 lines
gokku logs api production

# Follow logs in real-time
gokku logs api production -f

# Using --remote
gokku logs --remote vad-staging -f
```

Automatically detects systemd or Docker and shows appropriate logs.

#### status - Check Status

Check service or container status.

```bash
# Specific app
gokku status <app> <env>

# With --remote
gokku status --remote <app>-<env>

# All services
gokku status
```

**Examples:**
```bash
gokku status api production
gokku status --remote vad-production
gokku status  # All services
```

#### restart - Restart Service

Restart a service or container.

```bash
gokku restart <app> <env>

# With --remote
gokku restart --remote <app>-<env>
```

**Examples:**
```bash
gokku restart api production
gokku restart --remote vad-staging
```

Works with both systemd services and Docker containers.

#### deploy - Deploy Application

Deploy application via git push.

```bash
gokku deploy <app> <env>

# With --remote
gokku deploy --remote <app>-<env>
```

**Examples:**
```bash
gokku deploy api production
gokku deploy --remote vad-staging
```

Automatically:
1. Adds git remote if not exists
2. Determines branch based on environment
3. Pushes code
4. Triggers deployment

#### rollback - Rollback Deployment

Rollback to previous release.

```bash
# Rollback to previous release
gokku rollback <app> <env>

# Rollback to specific release
gokku rollback <app> <env> <release-id>

# With --remote
gokku rollback --remote <app>-<env>
gokku rollback --remote <app>-<env> <release-id>
```

**Examples:**
```bash
gokku rollback api production
gokku rollback api production 5
gokku rollback --remote vad-production
```

#### apps - List Applications

List all deployed applications on server.

```bash
gokku apps
```

#### ssh - SSH to Server

Direct SSH to server.

```bash
gokku ssh [command]
```

**Examples:**
```bash
# Interactive shell
gokku ssh

# Run command
gokku ssh "uptime"
```

### Configuration File

Location: `~/.gokku/config.yml`

**Example:**
```yaml
servers:
  - name: production
    host: ubuntu@ec2-54-123-45-67.compute-1.amazonaws.com
    base_dir: /opt/gokku
    default: true
  
  - name: staging
    host: ubuntu@ec2-54-234-56-78.compute-1.amazonaws.com
    base_dir: /opt/gokku
```

### Complete Workflow

#### 1. Add Git Remote (One-Time Setup)

```bash
# Add remote using standard git command
git remote add api-production ubuntu@server:api
git remote add vad-staging ubuntu@server:vad
```

Gokku will automatically extract:
- SSH host: `ubuntu@server`
- App name: `api`, `vad`
- Environment: `production`, `staging`

#### 2. Configure Application

```bash
gokku config set PORT=8080 --remote api-production
gokku config set DATABASE_URL="postgres://..." --remote api-production
gokku config set LOG_LEVEL=info --remote api-production
```

#### 3. Deploy

```bash
gokku deploy --remote api-production
```

Or traditional:
```bash
git push api-production main
```

#### 4. Check Status

```bash
gokku status --remote api-production
```

#### 5. View Logs

```bash
gokku logs --remote api-production -f
```

#### 6. Run Commands

```bash
gokku run "docker ps" --remote api-production
gokku run "df -h" --remote api-production
```

### Comparison: Before vs After

**Before (Manual SSH):**
```bash
# Set environment variable
ssh ubuntu@server "cd /opt/gokku && ./env-manager --app api --env production set PORT=8080"

# View logs
ssh ubuntu@server "sudo journalctl -u api-production -f"

# Restart service
ssh ubuntu@server "sudo systemctl restart api-production"

# Run command
ssh ubuntu@server "docker ps"
```

**After (gokku CLI):**
```bash
# Set environment variable
gokku config set PORT=8080 --remote api-production

# View logs
gokku logs --remote api-production -f

# Restart service
gokku restart --remote api-production

# Run command
gokku run "docker ps" --remote api-production
```

✅ **Benefits:**
- No manual SSH commands
- Cleaner syntax
- Automatic app/env parsing
- Works with both systemd and Docker
- Consistent interface

## Server Scripts

Scripts located in `/opt/gokku/scripts/` on the server.

### deploy-server-setup.sh

Setup a new application on the server.

**Usage:**
```bash
./deploy-server-setup.sh APP_NAME ENVIRONMENT
```

**Arguments:**
- `APP_NAME` - Application name from `gokku.yml`
- `ENVIRONMENT` - Environment name (e.g., `production`, `staging`)

**Example:**
```bash
./deploy-server-setup.sh api production
```

**What it does:**
1. Creates Git repository at `/opt/gokku/repos/APP_NAME.git`
2. Creates app directory at `/opt/gokku/apps/APP_NAME/ENVIRONMENT/`
3. Sets up Git post-receive hook
4. Creates systemd service (if `build.type: systemd`)
5. Creates environment file

### env-manager

Manage environment variables for applications.

**Usage:**
```bash
./env-manager --app APP_NAME --env ENVIRONMENT COMMAND [KEY[=VALUE]]
```

**Arguments:**
- `--app, -a` - Application name
- `--env, -e` - Environment name

**Commands:**

#### set

Set an environment variable:

```bash
./env-manager --app api --env production set PORT=8080
./env-manager --app api --env production set DATABASE_URL="postgres://..."
```

#### get

Get a variable value:

```bash
./env-manager --app api --env production get PORT
```

#### list

List all variables:

```bash
./env-manager --app api --env production list
```

Output:
```
PORT=8080
DATABASE_URL=postgres://...
LOG_LEVEL=info
```

#### del

Delete a variable:

```bash
./env-manager --app api --env production del PORT
```

**Examples:**

```bash
# Set multiple variables
./env-manager -a api -e production set PORT=8080
./env-manager -a api -e production set LOG_LEVEL=info
./env-manager -a api -e production set DATABASE_URL="postgres://localhost/db"

# List all
./env-manager -a api -e production list

# Delete
./env-manager -a api -e production del LOG_LEVEL
```

## Git Commands

### Deploy

Push to deploy:

```bash
git push REMOTE BRANCH
```

**Example:**
```bash
git push production main
git push staging develop
```

### Add Remote

```bash
git remote add REMOTE_NAME USER@SERVER:APP_NAME
```

**Example:**
```bash
git remote add production ubuntu@server:api
git remote add staging ubuntu@server:api
```

### List Remotes

```bash
git remote -v
```

### Remove Remote

```bash
git remote remove REMOTE_NAME
```

## Systemd Commands

Manage services on the server.

### Status

Check service status:

```bash
sudo systemctl status APP_NAME-ENVIRONMENT
```

**Example:**
```bash
sudo systemctl status api-production
```

### Start

Start service:

```bash
sudo systemctl start APP_NAME-ENVIRONMENT
```

### Stop

Stop service:

```bash
sudo systemctl stop APP_NAME-ENVIRONMENT
```

### Restart

Restart service:

```bash
sudo systemctl restart APP_NAME-ENVIRONMENT
```

### Enable

Enable service to start on boot:

```bash
sudo systemctl enable APP_NAME-ENVIRONMENT
```

### Disable

Disable service:

```bash
sudo systemctl disable APP_NAME-ENVIRONMENT
```

### Logs

View service logs:

```bash
sudo journalctl -u APP_NAME-ENVIRONMENT
```

**Options:**
- `-f` - Follow (tail) logs
- `-n 100` - Show last 100 lines
- `--since "1 hour ago"` - Show logs from last hour

**Examples:**
```bash
# Follow logs
sudo journalctl -u api-production -f

# Last 50 lines
sudo journalctl -u api-production -n 50

# Logs from last hour
sudo journalctl -u api-production --since "1 hour ago"
```

## Docker Commands

Manage Docker containers (when `build.type: docker`).

### List Containers

```bash
docker ps | grep APP_NAME
```

### Logs

View container logs:

```bash
docker logs APP_NAME-ENVIRONMENT
```

**Options:**
- `-f` - Follow (tail) logs
- `--tail 100` - Show last 100 lines

**Example:**
```bash
docker logs -f --tail 50 api-production
```

### Stop Container

```bash
docker stop APP_NAME-ENVIRONMENT
```

### Start Container

```bash
docker start APP_NAME-ENVIRONMENT
```

### Restart Container

```bash
docker restart APP_NAME-ENVIRONMENT
```

### Inspect Container

```bash
docker inspect APP_NAME-ENVIRONMENT
```

### Execute Command in Container

```bash
docker exec -it APP_NAME-ENVIRONMENT COMMAND
```

**Example:**
```bash
# Shell access
docker exec -it api-production /bin/sh

# Run command
docker exec api-production python manage.py migrate
```

### List Images

```bash
docker images | grep APP_NAME
```

### Remove Old Images

```bash
docker rmi APP_NAME:TAG
```

## SSH Commands

### Connect to Server

```bash
ssh USER@SERVER
```

### Run Command on Server

```bash
ssh USER@SERVER "COMMAND"
```

**Examples:**
```bash
# Check service status
ssh ubuntu@server "sudo systemctl status api-production"

# View logs
ssh ubuntu@server "sudo journalctl -u api-production -n 50"

# List containers
ssh ubuntu@server "docker ps"
```

### Copy Files

```bash
# From local to server
scp FILE USER@SERVER:/path/to/destination

# From server to local
scp USER@SERVER:/path/to/file LOCAL_PATH
```

## Common Workflows

### Deploy New App

```bash
# 1. Setup on server
ssh ubuntu@server "cd /opt/gokku && ./deploy-server-setup.sh api production"

# 2. Add Git remote locally
git remote add production ubuntu@server:api

# 3. Deploy
git push production main
```

### Update Environment Variable

```bash
# 1. Set variable
ssh ubuntu@server "cd /opt/gokku && ./env-manager --app api --env production set PORT=8081"

# 2. Restart service
ssh ubuntu@server "sudo systemctl restart api-production"
```

### View Logs

```bash
# Systemd
ssh ubuntu@server "sudo journalctl -u api-production -f"

# Docker
ssh ubuntu@server "docker logs -f api-production"
```

### Rollback (Manual)

```bash
# 1. SSH to server
ssh ubuntu@server

# 2. Navigate to app directory
cd /opt/gokku/apps/api/production

# 3. List releases
ls -la releases/

# 4. Change symlink
rm current
ln -s releases/2 current

# 5. Restart
sudo systemctl restart api-production
```

## Troubleshooting Commands

### Check Disk Space

```bash
ssh ubuntu@server "df -h"
```

### Check Memory Usage

```bash
ssh ubuntu@server "free -h"
```

### Check Ports

```bash
ssh ubuntu@server "sudo lsof -i :8080"
```

### Check Processes

```bash
ssh ubuntu@server "ps aux | grep api"
```

### Test HTTP Endpoint

```bash
ssh ubuntu@server "curl http://localhost:8080/health"
```

## Next Steps

- [Configuration Reference](/reference/configuration) - Config file docs
- [Troubleshooting](/reference/troubleshooting) - Common issues
- [Examples](/examples/) - Real-world usage

