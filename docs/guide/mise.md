# Mise Integration

Gokku has first-class support for [mise](https://mise.jdx.dev/) (and asdf), allowing you to manage tool versions seamlessly.

## What is Mise?

Mise (formerly rtx) is a fast, polyglot tool version manager. It reads `.tool-versions` files and installs the right versions of:

- Programming languages (Go, Python, Node.js, Ruby, etc.)
- CLI tools (ffmpeg, terraform, kubectl, etc.)
- Custom tools via plugins

## How Gokku Uses Mise

If Gokku detects a `.tool-versions` file in your project, it automatically:

1. Installs mise on the server (if needed)
2. Installs configured plugins
3. Runs `mise install`
4. Activates tools during build/runtime

**No configuration required!**

## Basic Usage

### Simple Example

Create `.tool-versions` in your project root:

```
golang 1.25.0
```

Gokku will:
- Install Go 1.25.0 via mise
- Use it to compile your app
- Keep this version per release

### Multiple Tools

```
golang 1.25.0
python 3.11
nodejs 20
ffmpeg 8.0
```

Perfect for projects needing multiple runtimes or tools.

## Custom Plugins

Some tools need custom plugins. Configure them in `gokku.yml`:

```yaml
apps:
  - name: whisper
    lang: python
    build:
      type: docker
      path: ./apps/whisper
      mise:
        plugins:
          - name: whispercpp
            url: https://github.com/thadeu/asdf-whispercpp.git
```

Then in `.tool-versions`:

```
python 3.11
ffmpeg 8.0
whispercpp 1.5.0
```

Gokku will:
1. Install the whispercpp plugin
2. Run `mise install`
3. All tools available during build

## Build Types

### Systemd (Go Apps)

```yaml
apps:
  - name: api
    build:
      type: systemd
      path: ./cmd/api
```

`.tool-versions`:
```
golang 1.25.0
```

Gokku uses Go 1.25.0 to compile, then deploys the binary.

### Docker (Any Language)

```yaml
apps:
  - name: ml-service
    lang: python
    build:
      type: docker
      path: ./services/ml
```

`.tool-versions`:
```
python 3.11
ffmpeg 8.0
```

Gokku generates a Dockerfile with mise:

```dockerfile
FROM ubuntu:22.04

# Install mise
RUN curl https://mise.run | sh
ENV PATH="/root/.local/share/mise/shims:${PATH}"

# Copy .tool-versions
COPY .tool-versions .

# Install tools
RUN mise install

# Copy app
COPY . .

CMD ["python", "main.py"]
```

## Per-App Versions

Different apps can use different versions:

**apps/api/.tool-versions:**
```
golang 1.25.0
```

**apps/worker/.tool-versions:**
```
golang 1.24.0
```

**apps/ml/.tool-versions:**
```
python 3.11
```

Each app gets its own isolated environment!

## Available Tools

Mise supports hundreds of tools. Common ones:

**Languages:**
- golang, python, nodejs, ruby, rust, elixir, java, kotlin, etc.

**Databases:**
- postgres, mysql, redis, mongodb

**Tools:**
- ffmpeg, terraform, kubectl, helm, awscli, gh

**Custom:**
- Any tool with an asdf plugin
- Your own custom plugins

See [mise plugins registry](https://mise.jdx.dev/plugins.html).

## Best Practices

### 1. Pin Exact Versions

✅ Good:
```
golang 1.25.0
python 3.11.5
```

❌ Bad:
```
golang latest
python 3
```

### 2. Keep .tool-versions in Git

Always commit `.tool-versions`:

```bash
git add .tool-versions
git commit -m "Pin tool versions"
```

### 3. Test Locally

Use mise locally too:

```bash
# Install mise locally
curl https://mise.run | sh

# Install tools
mise install

# Verify
go version
python --version
```

### 4. Document Custom Plugins

If using custom plugins, document them in README:

```markdown
## Setup

Install custom plugins:

```bash
mise plugins install whispercpp https://github.com/...
mise install
```
```

## Troubleshooting

### Tool Install Failed

Check mise logs:

```bash
ssh server "mise doctor"
```

### Wrong Version Used

Verify mise is activated:

```bash
ssh server "cd /opt/gokku/apps/api/production/current && mise list"
```

### Plugin Not Found

Install plugin manually:

```bash
ssh server "mise plugins install PLUGIN_NAME URL"
```

## Advanced

### Custom Install Scripts

Some tools need setup after install. Use mise hooks:

```bash
# .mise/hooks/post_install
#!/bin/bash
# Run after mise install
```

### Caching

Mise caches tool installations. Shared across releases for speed.

### Environment Variables

Mise can manage env vars too:

```bash
mise env set NODE_ENV=production
```

But Gokku's env-manager is recommended for app config.

## Examples

See [Examples](/examples/) for real-world usage:
- [Go app with mise](/examples/go-app#with-mise)
- [Python app with ffmpeg](/examples/python-app#with-ffmpeg)
- [Multi-tool project](/examples/multi-app#different-versions)

## Learn More

- [mise documentation](https://mise.jdx.dev/)
- [asdf plugins](https://github.com/asdf-vm/asdf-plugins)
- [Gokku configuration](/reference/configuration#mise)

