# Go Application Example

Deploy a Go REST API with Gokku using systemd.

## Basic Setup

### Project Structure

```
my-go-api/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── handlers/
│   └── models/
├── go.mod
├── go.sum
└── gokku.yml
```

### gokku.yml

```yaml
project:
  name: my-go-api

apps:
  - name: api
    build:
      path: ./cmd/api
      binary_name: api
```

### main.go

```go
package main

import (
    "log"
    "net/http"
    "os"
)

func main() {
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello from Gokku!"))
    })

    log.Printf("Server starting on port %s", port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}
```

### Deploy

```bash
# Setup on server
ssh ubuntu@server "cd /opt/gokku && ./deploy-server-setup.sh api production"

# Add remote
git remote add production ubuntu@server:api

# Deploy
git push production main
```

## With Gin Framework

### main.go

```go
package main

import (
    "os"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()
    
    r.GET("/", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "message": "Hello from Gokku!",
        })
    })
    
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "status": "healthy",
        })
    })
    
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    
    r.Run(":" + port)
}
```

## Multiple Environments

### gokku.yml

```yaml
apps:
  - name: api
    build:
      path: ./cmd/api
    
    environments:
      - name: production
        branch: main
        default_env_vars:
          GIN_MODE: release
          LOG_LEVEL: info
      
      - name: staging
        branch: staging
        default_env_vars:
          GIN_MODE: debug
          LOG_LEVEL: debug
```

### Setup Both Environments

```bash
# Production
ssh ubuntu@server "cd /opt/gokku && ./deploy-server-setup.sh api production"
git remote add production ubuntu@server:api

# Staging
ssh ubuntu@server "cd /opt/gokku && ./deploy-server-setup.sh api staging"
git remote add staging ubuntu@server:api
```

### Deploy

```bash
# Deploy to staging
git push staging staging

# Deploy to production
git push production main
```

## With Database

### Environment Variables

```bash
# On server
cd /opt/gokku
gokku config set DATABASE_URL="postgres://user:pass@localhost/db" --app api --env production
```

### main.go

```go
package main

import (
    "database/sql"
    "log"
    "os"
    
    _ "github.com/lib/pq"
    "github.com/gin-gonic/gin"
)

func main() {
    dbURL := os.Getenv("DATABASE_URL")
    if dbURL == "" {
        log.Fatal("DATABASE_URL not set")
    }
    
    db, err := sql.Open("postgres", dbURL)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    if err := db.Ping(); err != nil {
        log.Fatal(err)
    }
    
    r := gin.Default()
    
    r.GET("/users", func(c *gin.Context) {
        rows, err := db.Query("SELECT id, name FROM users")
        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }
        defer rows.Close()
        
        // ... handle rows
    })
    
    r.Run(":" + os.Getenv("PORT"))
}
```

## With Redis

### gokku.yml

```yaml
apps:
  - name: api
    build:
      path: ./cmd/api
    
    environments:
      - name: production
        default_env_vars:
          REDIS_URL: redis://localhost:6379
```

### main.go

```go
package main

import (
    "context"
    "os"
    
    "github.com/gin-gonic/gin"
    "github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func main() {
    rdb := redis.NewClient(&redis.Options{
        Addr: os.Getenv("REDIS_URL"),
    })
    
    r := gin.Default()
    
    r.GET("/cache/:key", func(c *gin.Context) {
        key := c.Param("key")
        val, err := rdb.Get(ctx, key).Result()
        if err != nil {
            c.JSON(404, gin.H{"error": "not found"})
            return
        }
        c.JSON(200, gin.H{"value": val})
    })
    
    r.POST("/cache/:key", func(c *gin.Context) {
        key := c.Param("key")
        var body map[string]string
        c.BindJSON(&body)
        
        rdb.Set(ctx, key, body["value"], 0)
        c.JSON(200, gin.H{"status": "ok"})
    })
    
    r.Run(":" + os.Getenv("PORT"))
}
```

## With Mise (Tool Versions)

### .tool-versions

```
golang 1.25.0
```

### gokku.yml

```yaml
apps:
  - name: api
    build:
      path: ./cmd/api
      go_version: "1.25"  # Optional, mise takes precedence
```

Gokku will automatically use Go 1.25.0 from `.tool-versions`.

## WebSockets

### main.go

```go
package main

import (
    "net/http"
    "os"
    
    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

func wsHandler(c *gin.Context) {
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        return
    }
    defer conn.Close()
    
    for {
        messageType, message, err := conn.ReadMessage()
        if err != nil {
            break
        }
        
        // Echo message back
        conn.WriteMessage(messageType, message)
    }
}

func main() {
    r := gin.Default()
    
    r.GET("/ws", wsHandler)
    
    r.Run(":" + os.Getenv("PORT"))
}
```

## Monitoring

### Health Check Endpoint

```go
r.GET("/health", func(c *gin.Context) {
    c.JSON(200, gin.H{
        "status": "healthy",
        "version": os.Getenv("RELEASE_VERSION"),
    })
})
```

### Metrics

```go
import "github.com/gin-gonic/gin"
import "github.com/prometheus/client_golang/prometheus/promhttp"

func main() {
    r := gin.Default()
    
    r.GET("/metrics", gin.WrapH(promhttp.Handler()))
    
    r.Run(":" + os.Getenv("PORT"))
}
```

## Troubleshooting

### Check Logs

```bash
ssh ubuntu@server "sudo journalctl -u api-production -f"
```

### Check Status

```bash
ssh ubuntu@server "sudo systemctl status api-production"
```

### Restart Service

```bash
ssh ubuntu@server "sudo systemctl restart api-production"
```

### Check Environment Variables

```bash
ssh ubuntu@server "cat /opt/gokku/apps/api/production/.env"
```

## Complete Example

Full working example: [github.com/thadeu/gokku-examples/go-api](https://github.com/thadeu/gokku-examples/tree/main/go-api)

## Next Steps

- [Configuration](/guide/configuration) - Customize deployment
- [Environment Variables](/guide/env-vars) - Manage secrets
- [Multiple Apps](/examples/multi-app) - Deploy multiple services

