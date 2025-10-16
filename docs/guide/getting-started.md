# Getting Started

Get Gokku up and running in minutes.

## Prerequisites

- Linux server (Ubuntu 20.04+ recommended)
- SSH access to server
- Git installed locally and on server
- Go application to deploy (or Python/Node.js)

## Installation

### Step 1: Install on Server

SSH into your server and run:

```bash
curl -fsSL https://gokku-vm.com/install | bash -s -- --server
```

This installs:
- Gokku scripts
- mise (for runtime management)
- Required dependencies

### Step 2: Install CLI (Optional but Recommended)

Install the `gokku` CLI on your local machine:

```bash
curl -fsSL https://gokku-vm.com/install | bash
```

The CLI makes it easier to manage your deployments without SSH commands.

Verify installation:

```bash
gokku version
```

Add your server:

```bash
gokku server add production ubuntu@your-server
```

### Step 3: Create Configuration

In your project root, create `gokku.yml`:

```yaml
project:
  name: my-app

apps:
  - name: api
    build:
      path: ./cmd/api
      binary_name: api
```

### Step 4: Deploy (Auto-Setup)

**Using gokku CLI (Recommended):**

```bash
gokku deploy api production
```

**Or manual git push:**

```bash
# Add git remote
git remote add production ubuntu@your-server:api

# Push - setup happens automatically!
git push production main
```

Watch the magic happen:

```
-----> Deploying api to production...
-----> Checking if auto-setup is needed...
-----> First deploy detected, running auto-setup...
-----> Found gokku.yml, configuring from repository...
-----> Created .env file from gokku.yml configuration
-----> Created systemd service from gokku.yml
-----> Auto-setup complete!
-----> Extracting code...
-----> Building api...
-----> Build complete (5.2M)
-----> Deploying...
-----> Restarting api-production...
-----> Deploy successful!
```

### Step 5: Manage Your App

**Using gokku CLI:**

```bash
# View logs
gokku logs --remote api-production -f

# Check status
gokku status --remote api-production

# Configure environment
gokku config set PORT=8080 --remote api-production
gokku config set DATABASE_URL="postgres://..." --remote api-production

# Restart
gokku restart --remote api-production

# Run commands
gokku run "docker ps" --remote api-production
```

**Or use SSH directly:**

```bash
ssh ubuntu@your-server "sudo systemctl status api-production"
```

Your app is live! 🎉

## What Happened?

1. **Git push** triggered a post-receive hook
2. **Auto-setup detected** first deploy and configured everything
3. **Code extracted** to a new release directory
4. **Build executed** (compiled Go binary or built Docker image)
5. **Symlink updated** to new release (atomic deploy)
6. **Service restarted** automatically
7. **Old releases kept** for rollback

## Next Steps

- [Configuration](/guide/configuration) - Customize your deployment
- [Environments](/guide/environments) - Add staging environment
- [Environment Variables](/guide/env-vars) - Configure your app
- [Docker Support](/guide/docker) - Use Docker instead of systemd

## Common Issues

### SSH Permission Denied

Make sure your SSH key is added to the server:

```bash
ssh-copy-id ubuntu@your-server
```

### Build Failed

Check the deployment logs:

```bash
ssh ubuntu@your-server "sudo journalctl -u api-production -n 50"
```

### Port Already in Use

Set a different port using gokku CLI:

```bash
# Using gokku CLI
gokku config set PORT=8081 --remote api-production
```

## Troubleshooting

See [Troubleshooting](/reference/troubleshooting) for more help.

