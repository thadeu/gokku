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
apps:
  api:
    build:
      path: ./cmd/api
      binary_name: api
      go_version: "1.25"
      goos: linux
      goarch: amd64
      cgo_enabled: 0
```

### Step 4: Create Application

Add a git remote for your application:

```bash
# Add git remote
git remote add production ubuntu@your-server:api
```

The application will be automatically created on first deployment.

### Step 5: Deploy

Now deploy your application:

**Using gokku CLI:**

```bash
gokku deploy -a production
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
-----> Auto-setup complete!
-----> Extracting code...
-----> Building Go application...
-----> Building Docker image...
-----> Build complete
-----> Deploying with blue-green deployment...
-----> Starting green container...
-----> Health check passed
-----> Switching traffic to green
-----> Deploy successful!
```

### Step 6: Manage Your App

**Using gokku CLI:**

```bash
# View logs
gokku logs -a api-production -f

# Check status
gokku status -a api-production

# Configure environment
gokku config set PORT=8080 -a api-production
gokku config set DATABASE_URL="postgres://..." -a api-production

# Restart
gokku restart -a api-production

# Run commands
gokku run "docker ps" -a api-production
```

**Or use SSH directly:**

```bash
ssh ubuntu@your-server "docker ps | grep api"
```

Your app is live! 🎉

## What Happened?

1. **Git push** triggered the post-receive hook
2. **Auto-setup detected** first deploy and configured from gokku.yml
3. **Code extracted** to a new release directory
4. **Docker image built** from your application code
5. **Blue-green deployment** started new container
6. **Health checks performed** on new container
7. **Traffic switched** to new container (zero downtime)
8. **Old container stopped** and cleaned up

## Next Steps

- [Configuration](/guide/configuration) - Customize your deployment
- [Environments](/guide/environments) - Add staging environment
- [Environment Variables](/guide/env-vars) - Configure your app
- [Blue-Green Deployment](/guide/blue-green-deployment) - Zero-downtime deployments

## Common Issues

### SSH Permission Denied

Make sure your SSH key is added to the server:

```bash
ssh-copy-id ubuntu@your-server
```

### Build Failed

Check the deployment logs:

```bash
# Using CLI
gokku logs -a api-production

# Or directly
ssh ubuntu@your-server "docker logs api-blue"
```

### Port Already in Use

Set a different port using gokku CLI:

```bash
# Using gokku CLI
gokku config set PORT=8081 -a api-production
```

## Troubleshooting

See [Troubleshooting](/reference/troubleshooting) for more help.

