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
        default_env_vars:
          LOG_LEVEL: info
          WORKERS: 4
      
      - name: staging
        default_env_vars:
          LOG_LEVEL: debug
          WORKERS: 2
```

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
gokku config set DATABASE_URL=postgres://... -a api-production

# Staging
gokku config set DATABASE_URL=postgres://... -a api-staging
```

## Next Steps

- [Deployment](/guide/deployment) - Deploy to environments
- [Environment Variables](/guide/env-vars) - Manage env vars
- [Rollback](/guide/rollback) - Rollback environments
- [Configuration](/guide/configuration) - Configure environments

