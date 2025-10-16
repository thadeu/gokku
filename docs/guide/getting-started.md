# Getting Started

Get Gokku up and running in minutes.

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

### Step 4: Create Application

First, create the application on the server:

**Using gokku CLI:**

```bash
# Add git remote
git remote add production ubuntu@your-server:/opt/gokku/repos/api.git

# Create the app (sets up repository and hooks)
gokku apps create api --remote production
```

### Step 5: Deploy

Now deploy your application:

**Using gokku CLI:**

```bash
gokku deploy --remote production
```

**Or manual git push:**

```bash
# Push - deployment happens automatically!
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

### Step 6: Manage Your App

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

1. **`gokku apps create`** set up the git repository and hooks on server
2. **Git push** triggered the post-receive hook
3. **Auto-setup detected** first deploy and configured systemd service
4. **Code extracted** to a new release directory
5. **Build executed** (compiled Go binary or built Docker image)
6. **Symlink updated** to new release (atomic deploy)
7. **Service restarted** automatically
8. **Old releases kept** for rollback

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

