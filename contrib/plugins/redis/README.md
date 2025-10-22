# Gokku Redis Plugin

Redis plugin for Gokku that provides Redis database services with persistent storage and authentication.

## Features

- Redis server with authentication
- Persistent data storage
- Configurable memory management
- Automatic port assignment
- Comprehensive monitoring and management commands

## Installation

```bash
# Install the plugin
gokku plugins:add redis

# Create a Redis service (default version)
gokku services:create redis --name redis-cache

# Create with specific version
gokku services:create redis:7 --name redis-cache
gokku services:create redis:7-alpine --name redis-cache

# Link to an application
gokku services:link redis-cache -a myapp
```

## Commands

### Service Management
- `gokku redis:info <service>` - Show service information and statistics
- `gokku redis:logs <service>` - Show service logs
- `gokku redis:start <service>` - Start Redis service
- `gokku redis:stop <service>` - Stop Redis service
- `gokku redis:restart <service>` - Restart Redis service
- `gokku redis:config <service>` - Show Redis configuration

### Redis Operations
- `gokku redis:cli <service>` - Connect to Redis CLI
- `gokku redis:flushdb <service>` - Flush current database
- `gokku redis:flushall <service>` - Flush all databases
- `gokku redis:monitor <service>` - Monitor Redis commands in real-time

### Backup and Restore
- `gokku redis:backup-s3 <service> <bucket> [prefix]` - Backup RDB file to S3
- `gokku redis:restore-s3 <service> <bucket> <key>` - Restore RDB file from S3

## Configuration

The Redis service is configured with:

- **Authentication**: Password-protected access
- **Persistence**: Automatic data saving with multiple save points
- **Memory Management**: LRU eviction policy when memory limit is reached
- **Security**: Protected mode enabled with binding to all interfaces
- **Logging**: Notice level logging

## Data Storage

Redis data is persisted in `/opt/gokku/services/<service-name>/data/` and survives container restarts.

## Environment Variables

When linking to an application, the following environment variables are automatically set:

- `REDIS_HOST`: Redis server hostname
- `REDIS_PORT`: Redis server port
- `REDIS_PASSWORD`: Redis authentication password
- `REDIS_URL`: Complete Redis connection URL

## Examples

### Basic Usage
```bash
# Create and start Redis service
gokku services:create redis --name redis-cache
gokku redis:info redis-cache

# Connect to Redis CLI
gokku redis:cli redis-cache

# Monitor Redis activity
gokku redis:monitor redis-cache
```

### Application Integration
```bash
# Link Redis to your application
gokku services:link redis-cache -a myapp

# Check service status
gokku redis:info redis-cache

# View logs
gokku redis:logs redis-cache
```

### Backup to S3
```bash
# Backup Redis data to S3
gokku redis:backup-s3 redis-cache my-backup-bucket

# Backup with custom prefix
gokku redis:backup-s3 redis-cache my-backup-bucket redis-backups/production

# Restore from S3
gokku redis:restore-s3 redis-cache my-backup-bucket redis-backups/redis-cache/redis-cache_backup_20240101_120000.rdb
```

## Troubleshooting

### Service Not Starting
```bash
# Check service status
gokku redis:info redis-cache

# View logs for errors
gokku redis:logs redis-cache

# Restart service
gokku redis:restart redis-cache
```

### Connection Issues
```bash
# Verify service is running
gokku redis:info redis-cache

# Test Redis connection
gokku redis:cli redis-cache
```

### Data Persistence
Redis data is automatically persisted to disk. If you need to reset data:
```bash
# Flush current database
gokku redis:flushdb redis-cache

# Flush all databases (use with caution)
gokku redis:flushall redis-cache
```

## Security Notes

- Redis is configured with password authentication
- The service runs in protected mode
- Data is persisted locally on the Gokku server
- Access is restricted to the internal network

## Support

For issues and feature requests, please visit the [Gokku Redis Plugin repository](https://github.com/thadeu/gokku-redis).
