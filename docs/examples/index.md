# Examples

Real-world examples of Gokku deployments for different use cases.

## Quick Navigation

- [Go Application](/examples/go-app) - Deploy a Go REST API
- [Python Application](/examples/python-app) - Deploy a Python Flask app
- [Rails Application](/examples/rails-app) - Deploy Ruby on Rails with Procfile
- [React Application](/examples/react-app) - Deploy React with Node.js API
- [Docker Application](/examples/docker-app) - Use Docker for deployment
- [Multi-App Project](/examples/multi-app) - Deploy multiple apps from one repo

## Example Projects

### Go REST API

Simple Go API with systemd deployment:

```yaml
apps:
  - name: api
    build:
      path: ./cmd/api
      binary_name: api
```

[View full example →](/examples/go-app)

### Python Flask App

Python web app with Docker:

```yaml
apps:
  - name: flask-app
    lang: python
    build:
      type: docker
      path: ./app
      entrypoint: app.py
```

[View full example →](/examples/python-app)

### Docker Deployment

Using custom Dockerfile:

```yaml
apps:
  - name: service
    lang: python
    build:
      type: docker
      path: ./services/ml
      dockerfile: ./services/ml/Dockerfile
```

[View full example →](/examples/docker-app)

### Ruby on Rails with Procfile

Full-stack Rails app with web, worker, and scheduler processes:

```yaml
apps:
  - name: rails-app
    lang: ruby
    build:
      type: docker
      path: .
    environments:
      - name: production
        default_env_vars:
          RAILS_ENV: production
          PORT: 3000
```

```bash
# Procfile
web: bundle exec rails server -p $PORT -e $RAILS_ENV
worker: bundle exec sidekiq -e $RAILS_ENV
scheduler: bundle exec whenever --update-crontab && cron -f
```

[View full example →](/examples/rails-app)

### React with Node.js API

React frontend with Express API backend:

```yaml
apps:
  - name: react-app
    lang: nodejs
    build:
      type: docker
      path: .
    environments:
      - name: production
        default_env_vars:
          NODE_ENV: production
          PORT: 3000
          API_PORT: 3001
```

```bash
# Procfile
web: npm run start:prod
api: node server/index.js
worker: node server/worker.js
```

[View full example →](/examples/react-app)

### Monorepo with Multiple Apps

Deploy multiple services from one repository:

```yaml
apps:
  - name: api
    build:
      path: ./cmd/api
  
  - name: worker
    build:
      path: ./cmd/worker
  
  - name: ml-service
    lang: python
    build:
      type: docker
      path: ./services/ml
```

[View full example →](/examples/multi-app)

## By Use Case

### Microservices

Deploy multiple independent services:
- [Multi-App Project](/examples/multi-app)

### With Database

Connect to PostgreSQL/MySQL:
- [Go Application](/examples/go-app#with-database)

### With Cache

Use Redis for caching:
- [Go Application](/examples/go-app#with-redis)

### Background Jobs

Deploy workers and cron jobs:
- [Multi-App Project](/examples/multi-app#background-workers)
- [Rails Application](/examples/rails-app#process-management)
- [React Application](/examples/react-app#process-management)

### Machine Learning

Deploy ML services with dependencies:
- [Python Application](/examples/python-app#machine-learning)

### Real-time

WebSocket and real-time apps:
- [Go Application](/examples/go-app#websockets)

### Procfile Applications

Multi-process applications with Dokku-style process management:
- [Rails Application](/examples/rails-app) - Web, worker, scheduler
- [React Application](/examples/react-app) - Web, API, worker

## By Technology

### Languages

- **Go**: [Go Application](/examples/go-app)
- **Python**: [Python Application](/examples/python-app)
- **Ruby**: [Rails Application](/examples/rails-app)
- **Node.js**: [React Application](/examples/react-app)

### Frameworks

- **Gin/Echo**: [Go Application](/examples/go-app)
- **Flask/FastAPI**: [Python Application](/examples/python-app)
- **Rails**: [Rails Application](/examples/rails-app)
- **React + Express**: [React Application](/examples/react-app)

### Deployment

- **Systemd**: [Go Application](/examples/go-app)
- **Docker**: [Docker Application](/examples/docker-app)
- **Procfile**: [Rails Application](/examples/rails-app), [React Application](/examples/react-app)

## Community Examples

Share your examples! Submit a PR:

[github.com/thadeu/gokku/tree/main/examples](https://github.com/thadeu/gokku/tree/main/examples)

## Need Help?

- [GitHub Discussions](https://github.com/thadeu/gokku/discussions)
- [GitHub Issues](https://github.com/thadeu/gokku/issues)

