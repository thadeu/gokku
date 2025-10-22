# Gokku Nginx Plugin

A Gokku plugin that provides nginx as a service for load balancing, reverse proxy, and static file serving.

## Features

- **Load Balancing**: Distribute traffic across multiple backend services
- **Reverse Proxy**: Route requests to different applications
- **SSL Termination**: Handle SSL certificates and HTTPS
- **Static File Serving**: Serve static assets efficiently
- **Health Checks**: Monitor backend service health
- **Configuration Management**: Easy nginx configuration updates

## Installation

```bash
gokku plugins:add thadeu/gokku-nginx
```

## Usage

### Create Nginx Service

```bash
# Create a new nginx service
gokku services:create nginx --name nginx-lb

# Link to an application
gokku services:link nginx-lb -a api-production
```

### Plugin Commands

```bash
# Show service information
gokku nginx:info nginx-lb

# View service logs
gokku nginx:logs nginx-lb

# Reload nginx configuration
gokku nginx:reload nginx-lb

# Check nginx status
gokku nginx:status nginx-lb

# Test nginx configuration
gokku nginx:test nginx-lb

# Show nginx configuration
gokku nginx:config nginx-lb
```

## Configuration

The nginx service creates a configuration directory at `/opt/gokku/services/<service-name>/` with:

- `nginx.conf` - Main nginx configuration
- `conf.d/` - Directory for site-specific configurations
- `ssl/` - SSL certificates directory

### Example Configuration

```nginx
# /opt/gokku/services/nginx-lb/conf.d/api.conf
upstream api_backend {
    server api-production:8080;
    server api-production-2:8080;
}

server {
    listen 80;
    server_name api.example.com;
    
    location / {
        proxy_pass http://api_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

## Environment Variables

Set nginx-specific environment variables:

```bash
# Set nginx worker processes
gokku config set NGINX_WORKER_PROCESSES=auto -a nginx-lb

# Set worker connections
gokku config set NGINX_WORKER_CONNECTIONS=1024 -a nginx-lb

# Set keepalive timeout
gokku config set NGINX_KEEPALIVE_TIMEOUT=65 -a nginx-lb
```

## SSL Configuration

To enable SSL, place your certificates in the service directory:

```bash
# Copy SSL certificates
cp server.crt /opt/gokku/services/nginx-lb/ssl/
cp server.key /opt/gokku/services/nginx-lb/ssl/

# Reload nginx to apply SSL configuration
gokku nginx:reload nginx-lb
```

## Load Balancing Methods

Configure different load balancing methods in your upstream blocks:

```nginx
# Round Robin (default)
upstream backend {
    server app1:8080;
    server app2:8080;
}

# Least Connections
upstream backend {
    least_conn;
    server app1:8080;
    server app2:8080;
}

# IP Hash
upstream backend {
    ip_hash;
    server app1:8080;
    server app2:8080;
}
```

## Health Checks

Configure health checks for your backends:

```nginx
upstream backend {
    server app1:8080 max_fails=3 fail_timeout=30s;
    server app2:8080 max_fails=3 fail_timeout=30s;
}
```

## Logging

Access logs are available through the plugin:

```bash
# View access logs
gokku nginx:logs nginx-lb

# View error logs
docker exec nginx-lb tail -f /var/log/nginx/error.log
```

## Troubleshooting

### Check Configuration

```bash
# Test nginx configuration
gokku nginx:test nginx-lb

# Show current configuration
gokku nginx:config nginx-lb
```

### Common Issues

1. **Port conflicts**: The plugin automatically assigns available ports
2. **Configuration errors**: Use `nginx:test` to validate configuration
3. **Backend connectivity**: Check if backend services are running
4. **SSL issues**: Verify certificate paths and permissions

## Examples

### Simple Reverse Proxy

```nginx
server {
    listen 80;
    server_name myapp.example.com;
    
    location / {
        proxy_pass http://myapp-production:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### Load Balancer with Health Checks

```nginx
upstream myapp {
    server myapp-1:8080 max_fails=3 fail_timeout=30s;
    server myapp-2:8080 max_fails=3 fail_timeout=30s;
}

server {
    listen 80;
    server_name myapp.example.com;
    
    location / {
        proxy_pass http://myapp;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### Static File Serving

```nginx
server {
    listen 80;
    server_name static.example.com;
    root /var/www/static;
    
    location / {
        try_files $uri $uri/ =404;
    }
}
```

## License

MIT License - see LICENSE file for details.
