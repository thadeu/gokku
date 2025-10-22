# Let's Encrypt Plugin for Gokku

This plugin provides automatic SSL certificate management using Let's Encrypt for your Gokku applications.

## Features

- **Global Plugin Architecture**: Install once, use everywhere
- **Automatic Certificate Generation**: Create SSL certificates for your domains
- **Nginx Integration**: Automatic SSL configuration for nginx services
- **Auto-renewal Enabled by Default**: Set it and forget it - certificates renew automatically
- **Persistent Storage**: Certificates are stored persistently with local sync
- **Multiple Domains**: Support for unlimited domains

## Installation

```bash
# Install the Let's Encrypt plugin (run once, auto-renewal enabled by default)
gokku letsencrypt:install

# Create certificates for your domains
gokku letsencrypt:create example.com contact@example.com
gokku letsencrypt:create api.example.com

# Link to nginx service
gokku letsencrypt:link-nginx nginx-lb
```

## Commands

### Plugin Management
- `gokku letsencrypt:install` - Install Let's Encrypt plugin (auto-renewal enabled by default)
- `gokku letsencrypt:uninstall` - Uninstall Let's Encrypt plugin
- `gokku letsencrypt:status` - Show plugin status

### Certificate Management
- `gokku letsencrypt:create <domain> [email]` - Create SSL certificate
- `gokku letsencrypt:renew` - Renew all certificates
- `gokku letsencrypt:list` - List all certificates
- `gokku letsencrypt:info <domain>` - Show certificate information
- `gokku letsencrypt:logs` - Show certificate logs

### Nginx Integration
- `gokku letsencrypt:link-nginx <nginx-service>` - Link to nginx service
- `gokku letsencrypt:unlink-nginx <nginx-service>` - Unlink from nginx service

### Automation
- `gokku letsencrypt:auto-renew` - Setup auto-renewal
- `gokku letsencrypt:remove-auto-renew` - Remove auto-renewal

## Configuration

### Plugin Directory Structure
```
/opt/gokku/plugins/letsencrypt/
├── certs/                    # SSL certificates (global)
│   └── <domain>/             # Domain-specific certificates
│       ├── fullchain.pem     # Full certificate chain
│       └── privkey.pem       # Private key
├── accounts/                 # Let's Encrypt accounts
├── logs/                     # Certificate logs
├── plugin.conf               # Plugin configuration
├── nginx-services            # List of linked nginx services
└── renew-certificates.sh     # Renewal script
```

### Nginx Integration
When linked to nginx, the plugin automatically:
- Creates SSL server blocks for HTTPS
- Sets up HTTP to HTTPS redirects
- Links certificates to nginx SSL directory
- Configures security headers

### Auto-renewal
The plugin automatically sets up a scheduled job during installation that runs daily at 2:30 AM to check and renew certificates that are expiring within 30 days. No additional configuration needed!

## Examples

### Basic Setup
```bash
# Install plugin (once, auto-renewal included)
gokku letsencrypt:install

# Create certificate
gokku letsencrypt:create example.com contact@example.com

# Link to nginx
gokku letsencrypt:link-nginx nginx-lb
```

### Multiple Domains
```bash
# Create certificates for multiple domains
gokku letsencrypt:create api.example.com
gokku letsencrypt:create app.example.com
gokku letsencrypt:create admin.example.com

# List all certificates
gokku letsencrypt:list
```

### Certificate Management
```bash
# Check certificate status
gokku letsencrypt:info example.com

# Renew all certificates
gokku letsencrypt:renew

# Check plugin status
gokku letsencrypt:status
```

## Requirements

- Docker (for certbot container)
- Nginx service (for SSL configuration)
- Valid domain pointing to your server
- Ports 80 and 443 accessible

## Troubleshooting

### Certificate Creation Fails
- Ensure domain points to your server
- Check that ports 80 and 443 are accessible
- Verify email address is valid
- Check logs: `gokku letsencrypt:logs`

### Nginx Integration Issues
- Ensure nginx service is running
- Check nginx configuration: `gokku nginx:test nginx-lb`
- Verify SSL files are linked: `ls -la /opt/gokku/services/nginx-lb/ssl/`

### Auto-renewal Not Working
- Check scheduled job: `cat /etc/cron.d/gokku-letsencrypt`
- Test renewal script: `/opt/gokku/plugins/letsencrypt/renew-certificates.sh`
- Check renewal logs: `tail -f /var/log/gokku-letsencrypt.log`

## Security Notes

- Certificates are stored in `/opt/gokku/plugins/letsencrypt/certs/`
- Private keys are protected with appropriate permissions
- Auto-renewal runs as root via scheduled job
- SSL configurations include security headers
- Plugin is global - accessible by all services

## Support

For issues and feature requests, please visit the [Gokku Let's Encrypt Plugin repository](https://github.com/thadeu/gokku-letsencrypt).
