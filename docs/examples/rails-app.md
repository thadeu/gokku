# Rails Application with Procfile

Deploy a Ruby on Rails application with multiple processes using Procfile.

## Procfile vs gokku.yml

This example demonstrates **single-app with multiple processes** using Procfile (use case 1). When a `Procfile` is present in the project root:

- ✅ **Procfile takes priority** - Defines the processes that will run
- ✅ **Docker is required** - Procfile forces container usage
- ✅ **gokku.yml is optional** - Used only for environment configuration

For **multiple independent apps** (use case 2), use only `gokku.yml` without Procfile.

## Basic Setup

### Project Structure

```
my-rails-app/
├── app/
├── config/
├── db/
├── Gemfile
├── Gemfile.lock
├── Procfile
├── gokku.yml
```

### Procfile

```bash
# Procfile
web: bundle exec rails server -p $PORT -e $RAILS_ENV
worker: bundle exec sidekiq -e $RAILS_ENV
scheduler: bundle exec whenever --update-crontab && cron -f
```

### gokku.yml

**Note**: When a `Procfile` is present, it takes priority over gokku.yml configurations. gokku.yml is used primarily for environment configuration.

```yaml
project:
  name: my-rails-app

apps:
  - name: rails-app
    lang: ruby
    build:
      type: docker  # Automatically set when Procfile is detected
      path: .
    environments:
      - name: production
        branch: main
        default_env_vars:
          RAILS_ENV: production
          PORT: 3000
          DATABASE_URL: postgres://user:pass@localhost:5432/rails_app
          REDIS_URL: redis://localhost:6379
          SECRET_KEY_BASE: "your-secret-key-here"
    deployment:
      post_deploy:
        - "cd /opt/gokku/apps/rails-app/production/current && bundle exec rails db:migrate"
        - "cd /opt/gokku/apps/rails-app/production/current && bundle exec rails assets:precompile"
```

**Alternative without Procfile**: If not using Procfile, you can define individual processes in gokku.yml.

## Post-Deploy Commands

The `post_deploy` commands run automatically after successful deployment:

- **Database migrations**: `bundle exec rails db:migrate`
- **Asset compilation**: `bundle exec rails assets:precompile`
- **Cache warming**: Pre-populate caches for faster responses
- **Custom tasks**: Any command needed after deployment

**Commands run in sequence** and **fail deployment** if any command fails. This ensures your app is fully ready before considering the deployment complete.

## Deployment

### 1. Initialize Git

```bash
cd my-rails-app
git init
git add .
git commit -m "Initial commit"
```

### 2. Setup Remote

```bash
# Add remote (replace with your server)
git remote add production ubuntu@your-server:rails-app
```

### 3. Deploy

```bash
git push production main
```

Gokku will automatically:
- Detect the Procfile
- Generate a Ruby Dockerfile
- Create containers for web, worker, and scheduler processes
- Start all processes
- **Run post-deploy commands** (database migrations, asset compilation)

### 4. Check Status

```bash
# List all processes
gokku ps:list rails-app production --remote rails-app-production

# View logs
gokku ps:logs rails-app production web -f --remote rails-app-production
```

## Process Management

### Scaling Processes

```bash
# Scale web to 1 instance (default)
gokku ps:scale rails-app production web=1 --remote rails-app-production

# Scale workers based on load
gokku ps:scale rails-app production worker=3 --remote rails-app-production

# Stop scheduler if not needed
gokku ps:scale rails-app production scheduler=0 --remote rails-app-production
```

### Managing Individual Processes

```bash
# Restart web process
gokku ps:restart rails-app production web --remote rails-app-production

# Restart worker process
gokku ps:restart rails-app production worker --remote rails-app-production

# Stop worker temporarily
gokku ps:stop rails-app production worker --remote rails-app-production

# Start worker again
gokku ps:start rails-app production worker --remote rails-app-production
```

## Environment Variables

### Database Configuration

```bash
# Set database URL
gokku config set DATABASE_URL="postgres://user:password@localhost:5432/rails_app_production" --remote rails-app-production

# Set Redis URL for Sidekiq
gokku config set REDIS_URL="redis://localhost:6379" --remote rails-app-production
```

### Rails Configuration

```bash
# Set Rails secret key base
gokku config set SECRET_KEY_BASE="$(bundle exec rails secret)" --remote rails-app-production

# Set environment
gokku config set RAILS_ENV=production --remote rails-app-production

# Set log level
gokku config set RAILS_LOG_LEVEL=info --remote rails-app-production
```

### Application-Specific Variables

```bash
# Set SMTP configuration
gokku config set SMTP_HOST="smtp.gmail.com" --remote rails-app-production
gokku config set SMTP_PORT="587" --remote rails-app-production

# Set API keys
gokku config set STRIPE_SECRET_KEY="sk_live_..." --remote rails-app-production
gokku config set AWS_ACCESS_KEY_ID="..." --remote rails-app-production
```

## Monitoring

### Process Status

```bash
# Check all processes
gokku ps:list rails-app production --remote rails-app-production

# Output:
# === Procfile Processes: rails-app (production) ===
#
# web:
#   Systemd: active
#   Container: running
#
# worker:
#   Systemd: active
#   Container: running
#
# scheduler:
#   Systemd: active
#   Container: running
```

### Application Logs

```bash
# Rails application logs
gokku ps:logs rails-app production web -f --remote rails-app-production

# Sidekiq worker logs
gokku ps:logs rails-app production worker -f --remote rails-app-production

# Scheduler logs
gokku ps:logs rails-app production scheduler -f --remote rails-app-production
```

## Next Steps

- [Procfile Guide](/guide/procfile) - Complete Procfile documentation
- [Environment Variables](/guide/environments) - Environment management
- [Docker Support](/guide/docker) - Container deployment
- [Configuration](/reference/configuration) - Advanced configuration
