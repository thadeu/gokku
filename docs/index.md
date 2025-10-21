---
layout: home

hero:
  name: Gokku
  text: Lightweight Git-Push Deployment
  tagline: Deploy Go, Python, and Node.js applications with zero hassle. Like Dokku, but focused on simplicity.
  actions:
    - theme: brand
      text: Get Started
      link: /guide/getting-started
    - theme: alt
      text: View on GitHub
      link: https://github.com/thadeu/gokku

features:
  - icon: 🚀
    title: Git Push to Deploy
    details: Deploy your applications with a simple git push. No complex CI/CD pipelines required.
  
  - icon: 🔧
    title: Multi-Language Support
    details: Native support for Go, Python, Node.js with automatic runtime detection.
  
  - icon: 🐳
    title: Docker & Systemd
    details: Choose between Docker containers or systemd services. Mix and match per application.
  
  - icon: ⚙️
    title: Zero Configuration
    details: Sensible defaults for everything. Start with minimal config and customize as needed.
  
  
  - icon: 🔄
    title: Easy Rollback
    details: Keep multiple releases. Rollback to any previous version instantly.
---

## Quick Start

Install Gokku on your server:

```bash
curl -fsSL https://gokku-vm.com/install | bash -s -- --server
```

Install Gokku on your client:

```bash
curl -fsSL https://gokku-vm.com/install | bash -s -- --client
```

Check version

```bash
gokku version
```

Create a minimal `gokku.yml`:

```yaml
project:
  name: my-app

apps:
  - name: api
    build:
      path: ./cmd/api
      binary_name: api
    ports:
      - 80:3000
```

Deploy your app:

```bash
git remote add production user@server:api 

git push production main
```

That's it! Your app is live. 🎉

## Why Gokku?

**Gokku** = Go + Dokku. A lightweight deployment system that:

- ✅ Works great with Go applications (but supports Python, Node.js too)
- ✅ Simple git-push workflow like Heroku/Dokku
- ✅ No Docker dependency (but Docker is supported)
- ✅ Per-app configuration and environments
- ✅ Open source and hackable

Perfect for:
- Side projects and startups
- Internal tools
- Microservices
- Development/staging environments
- Learning deployment workflows

## What's Different?

| Feature | Heroku | Dokku | Gokku |
|---------|--------|-------|-------|
| Cost | 💰 Paid | ✅ Free | ✅ Free |
| Docker Required | ✅ Yes | ✅ Yes | ⚠️ Optional |
| Go-First | ❌ No | ❌ No | ✅ Yes |
| Config File | ❌ No | ❌ No | ✅ Yes |
| Systemd Option | ❌ No | ❌ No | ✅ Yes |

## Community

- 📖 [Documentation](/)
- 🐛 [Report Issues](https://github.com/thadeu/gokku/issues)
- 💬 [Discussions](https://github.com/thadeu/gokku/discussions)

