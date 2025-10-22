# PostgreSQL Plugin for Gokku

Provides PostgreSQL database services with persistent storage.

## Installation

The PostgreSQL plugin is a global plugin that integrates with the services system.

## Usage

### Create a PostgreSQL Service

```bash
# Create with default version (latest)
gokku services:create postgres --name pg-0

# Create with specific version
gokku services:create postgres:14 --name pg-0
gokku services:create postgres:15 --name pg-0
gokku services:create postgres:16 --name pg-0
gokku services:create postgres:17 --name pg-0
gokku services:create postgres:14-alpine --name pg-0
```

Available versions: Any PostgreSQL Docker tag (14, 15, 16, 17, latest, alpine variants, etc.)

### Link Service to App

When you link a service to an app, Gokku automatically adds environment variables to your app's `.env` file:

```bash
gokku services:link pg-0 -a api-prod
```

This command:
1. Links the PostgreSQL service `pg-0` to the app `api-prod`
2. Automatically adds the following environment variables to `/opt/gokku/apps/api-prod/shared/.env`:
   - `DATABASE_URL` - Full PostgreSQL connection string (e.g., `postgres://postgres:password@localhost:5432/postgres`)
   - `POSTGRES_HOST` - PostgreSQL host (localhost)
   - `POSTGRES_PORT` - PostgreSQL port
   - `POSTGRES_USER` - PostgreSQL user
   - `POSTGRES_PASSWORD` - PostgreSQL password
   - `POSTGRES_DB` - PostgreSQL database name
3. Your app will have access to these variables on next restart/deploy

### Show Service Information

```bash
gokku postgres:info pg-0
```

### Connect to PostgreSQL

```bash
gokku postgres:psql pg-0
```

Once connected, you can run any SQL commands:
```sql
\l                    -- list databases
\c dbname            -- connect to database
CREATE DATABASE mydb; -- create database
\dt                  -- list tables
```

### View Logs

```bash
gokku postgres:logs pg-0
```

### Backup Database

```bash
gokku postgres:backup pg-0 > backup.sql
```

### Restore Database

```bash
gokku postgres:restore pg-0 backup.sql
```

### Unlink Service from App

```bash
gokku services:unlink pg-0 -a api-prod
```

### Destroy Service

```bash
gokku services:destroy pg-0
```

## Features

- Persistent data storage using Docker volumes
- Automatic password generation
- PostgreSQL 15 Alpine image for minimal footprint
- Auto-restart on container failure
- Full backup and restore capabilities
- Database management commands

## Data Persistence

Data is stored in a Docker volume named `{service-name}_data`. This ensures data persists across container restarts and recreations.

To manually inspect the volume:

```bash
docker volume inspect pg-0_data
```

## Connection Information

After linking a service to an app, the app will have access to the PostgreSQL credentials through environment variables. You can use the `DATABASE_URL` to connect to the database:

```ruby
# Ruby/Rails
database_url = ENV['DATABASE_URL']
```

```javascript
// Node.js
const databaseUrl = process.env.DATABASE_URL;
```

```python
# Python
import os
database_url = os.getenv('DATABASE_URL')
```

