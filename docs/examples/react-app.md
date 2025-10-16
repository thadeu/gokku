# React Application with Procfile

Deploy a React application with Node.js backend and multiple processes using Procfile.

## Procfile vs gokku.yml

This example demonstrates **single-app with multiple processes** using Procfile (use case 1). When a `Procfile` is present in the project root:

- ✅ **Procfile takes priority** - Defines the processes that will run
- ✅ **Docker is required** - Procfile forces container usage
- ✅ **gokku.yml is optional** - Used only for environment configuration

For **multiple independent apps** (use case 2), use only `gokku.yml` without Procfile.

## Basic Setup

### Project Structure

```
my-react-app/
├── client/          # React frontend
│   ├── public/
│   ├── src/
│   ├── package.json
│   └── build/
├── server/          # Node.js API backend
│   ├── src/
│   ├── package.json
│   └── index.js
├── Procfile
└── gokku.yml
```

### Procfile

```bash
# Procfile
web: npm run start:prod
api: node server/index.js
worker: node server/worker.js
```

### gokku.yml

**Note**: When a `Procfile` is present, it takes priority over gokku.yml configurations. gokku.yml is used primarily for environment configuration.

```yaml
project:
  name: my-react-app

apps:
  - name: react-app
    lang: nodejs
    build:
      type: docker  # Automatically set when Procfile is detected
      path: .
    environments:
      - name: production
        branch: main
        default_env_vars:
          NODE_ENV: production
          PORT: 3000
          API_PORT: 3001
          DATABASE_URL: mongodb://localhost:27017/react_app
          REDIS_URL: redis://localhost:6379
          JWT_SECRET: "your-jwt-secret-here"
    deployment:
      post_deploy:
        - npm run db:migrate"
        - npm run cache:warm"
```

**Alternative without Procfile**: If not using Procfile, you can define individual processes in gokku.yml.

## Post-Deploy Commands

The `post_deploy` commands run automatically after successful deployment:

- **Database migrations**: `npm run db:migrate` (for schema updates)
- **Cache warming**: `npm run cache:warm` (pre-populate application caches)
- **Data seeding**: Populate initial data if needed
- **Custom tasks**: Any command needed after deployment

**Commands run in sequence** and **fail deployment** if any command fails. This ensures your app is fully ready before considering the deployment complete.

## Deployment

### 1. Initialize Git

```bash
cd my-react-app
git init
git add .
git commit -m "Initial commit"
```

### 2. Setup Remote

```bash
# Add remote (replace with your server)
git remote add production ubuntu@your-server:react-app
```

### 3. Deploy

```bash
git push production main
```

Gokku will automatically:
- Detect the Procfile
- Generate a Node.js Dockerfile
- Build the React app
- Create containers for web, api, and worker processes
- Start all processes
- **Run post-deploy commands** (database migrations, cache warming)

### 4. Check Status

```bash
# List all processes
gokku ps:list react-app production --remote react-app-production

# View web server logs
gokku ps:logs react-app production web -f --remote react-app-production
```

## Process Management

### Scaling Processes

```bash
# Web server (React app)
gokku ps:scale react-app production web=1 --remote react-app-production

# API server
gokku ps:scale react-app production api=1 --remote react-app-production

# Worker processes (scale based on load)
gokku ps:scale react-app production worker=2 --remote react-app-production
```

### Managing Individual Processes

```bash
# Restart API server
gokku ps:restart react-app production api --remote react-app-production

# Stop worker temporarily
gokku ps:stop react-app production worker --remote react-app-production

# Start worker again
gokku ps:start react-app production worker --remote react-app-production
```

## Next Steps

- [Procfile Guide](/guide/procfile) - Complete Procfile documentation
- [Environment Variables](/guide/environments) - Environment management
- [Docker Support](/guide/docker) - Container deployment
- [Configuration](/reference/configuration) - Advanced configuration
