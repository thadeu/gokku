# Gokku Cron Plugin

Cron plugin for Gokku that provides scheduled task execution capabilities for automating Gokku commands and system maintenance tasks.

## Features

- Schedule Gokku commands and scripts
- Persistent job storage
- Comprehensive logging and monitoring
- Manual job execution
- Job management (add, remove, list)
- Automatic log rotation

## Installation

```bash
# Install the plugin
gokku plugins:add thadeu/gokku-cron

# Create a cron service
gokku services:create cron --name cron-scheduler
```

## Commands

### Service Management
- `gokku cron:info <service>` - Show service information and recent job executions
- `gokku cron:logs <service>` - Show service logs

### Job Management
- `gokku cron:schedule <service> <name> <schedule> <command>` - Schedule a new job
- `gokku cron:list <service>` - List all scheduled jobs
- `gokku cron:remove <service> <name>` - Remove a scheduled job
- `gokku cron:run <service> <name>` - Run a job immediately
- `gokku cron:logs-job <service> <name>` - Show logs for a specific job

## Schedule Format

The cron plugin uses standard cron syntax:

```
minute hour day month weekday
```

### Examples:
- `0 22 * * *` - Daily at 10:00 PM
- `0 */6 * * *` - Every 6 hours
- `30 2 * * 0` - Weekly on Sunday at 2:30 AM
- `*/15 * * * *` - Every 15 minutes

## Usage Examples

### Basic Scheduling
```bash
# Schedule a Redis backup daily at 10 PM
gokku cron:schedule cron-scheduler backup-redis "0 22 * * *" "gokku redis:backup-s3 redis-cache my-bucket"

# Schedule log cleanup every 6 hours
gokku cron:schedule cron-scheduler cleanup-logs "0 */6 * * *" "find /var/log -name '*.log' -mtime +7 -delete"

# Schedule weekly database maintenance
gokku cron:schedule cron-scheduler weekly-maintenance "30 2 * * 0" "gokku postgres:vacuum postgres-db"
```

### Job Management
```bash
# List all scheduled jobs
gokku cron:list cron-scheduler

# Run a job manually
gokku cron:run cron-scheduler backup-redis

# View job execution logs
gokku cron:logs-job cron-scheduler backup-redis

# Remove a job
gokku cron:remove cron-scheduler backup-redis
```

### Service Information
```bash
# Check service status and recent executions
gokku cron:info cron-scheduler

# View service logs
gokku cron:logs cron-scheduler
```

## Common Use Cases

### Database Backups
```bash
# Daily Redis backup at 10 PM
gokku cron:schedule cron-scheduler redis-backup "0 22 * * *" "gokku redis:backup-s3 redis-cache backup-bucket"

# Weekly PostgreSQL backup on Sunday at 2 AM
gokku cron:schedule cron-scheduler postgres-backup "0 2 * * 0" "gokku postgres:backup postgres-db"
```

### System Maintenance
```bash
# Clean up old logs daily at 3 AM
gokku cron:schedule cron-scheduler log-cleanup "0 3 * * *" "find /var/log -name '*.log' -mtime +30 -delete"

# Update system packages weekly
gokku cron:schedule cron-scheduler system-update "0 4 * * 1" "apt update && apt upgrade -y"
```

### Application Health Checks
```bash
# Check application health every 5 minutes
gokku cron:schedule cron-scheduler health-check "*/5 * * * *" "curl -f http://localhost:8080/health || exit 1"

# Restart application if unhealthy
gokku cron:schedule cron-scheduler restart-app "*/10 * * * *" "gokku ps:restart myapp"
```

## Log Management

- Job execution logs are stored in `/opt/gokku/services/<service-name>/logs/`
- Logs are automatically rotated (keeps last 100 executions per job)
- Each execution creates a timestamped log file
- Logs include command output, exit codes, and execution times

## Job File Format

Jobs are stored as `.job` files containing:
```bash
JOB_NAME="backup-redis"
JOB_COMMAND="gokku redis:backup-s3 redis-cache my-bucket"
SERVICE_NAME="cron-scheduler"
```

## Troubleshooting

### Service Not Starting
```bash
# Check service status
gokku cron:info cron-scheduler

# View service logs
gokku cron:logs cron-scheduler

# Restart service
gokku services:restart cron-scheduler
```

### Job Not Executing
```bash
# Check if job is scheduled
gokku cron:list cron-scheduler

# Run job manually to test
gokku cron:run cron-scheduler backup-redis

# Check job logs
gokku cron:logs-job cron-scheduler backup-redis
```

### Schedule Format Issues
- Use standard cron syntax: `minute hour day month weekday`
- Test schedules with manual execution first
- Check job logs for execution errors

## Security Notes

- Jobs run with the same permissions as the Gokku service
- Commands are executed in the cron container context
- Logs may contain sensitive information - secure appropriately
- Use environment variables for sensitive data in commands

## Best Practices

1. **Test Commands**: Always test commands manually before scheduling
2. **Use Absolute Paths**: Use full paths for commands and scripts
3. **Log Management**: Monitor log sizes and implement cleanup
4. **Error Handling**: Include error handling in your commands
5. **Documentation**: Document complex job purposes and schedules

## Support

For issues and feature requests, please visit the [Gokku Cron Plugin repository](https://github.com/thadeu/gokku-cron).
