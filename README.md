# gorm-zero
A go-zero gorm extension with native Redis support. If you use go-zero, and you want to use GORM with native Redis client. You can use this extension.

## Features
- Native Redis support (github.com/redis/go-redis/v9)
- Support for Redis DB selection
- Connection pool configuration
- Custom cache expiration
- Compatible with GORM v2

## Installation

- Add the dependency
```shell
go get github.com/SpectatorNan/gorm-zero
```
- Replace `template/model` in your project with `gorm-zero/template/v1/model`
- Generate
```shell
goctl model mysql -src={patterns} -dir={dir} -cache --home ./template
```

## Basic Usage
Currently we support two databases: MySQL and PostgreSQL. For example:

### MySQL
* Config
```go
import (
    "github.com/SpectatorNan/gorm-zero/gormc/config/mysql"
)
type Config struct {
    Mysql mysql.Mysql
    ...
}
```
* Initialization
```go
import (
    "github.com/SpectatorNan/gorm-zero/gormc/config/mysql"
)
func NewServiceContext(c config.Config) *ServiceContext {
    db, err := mysql.Connect(c.Mysql)
    if err != nil {
        log.Fatal(err)
    }
    ...
}
```

### PostgreSQL
* Config
```go
import (
    "github.com/SpectatorNan/gorm-zero/gormc/config/pg"
)
type Config struct {
    PgSql pg.PgSql
    ...
}
```

* Initialization
```go
import (
    "github.com/SpectatorNan/gorm-zero/gormc/config/pg"
)
func NewServiceContext(c config.Config) *ServiceContext {
    db, err := pg.Connect(c.PgSql)
    if err != nil {
        log.Fatal(err)
    }
    ...
}
```

## Redis Configuration

### Basic Configuration with DB Selection
```go
import (
    "time"
    "github.com/SpectatorNan/gorm-zero/gormc"
    "gorm.io/gorm"
)

// Configure Redis with DB selection
redisConf := gormc.RedisConfig{
    Addr:         "127.0.0.1:6379",    // Redis server address
    Password:     "",                  // Redis password (empty if no auth)
    DB:           0,                   // Redis database index (0-15)
    PoolSize:     10,                  // Connection pool size
    MinIdleConns: 2,                   // Minimum idle connections
    DialTimeout:  5 * time.Second,     // Connection timeout
    ReadTimeout:  3 * time.Second,     // Read timeout
    WriteTimeout: 3 * time.Second,     // Write timeout
}

// Create cached connection
cachedConn, err := gormc.NewConn(db, redisConf, time.Hour)
if err != nil {
    panic(err)
}
```

### Using Multiple Redis Databases
```go
// Use Redis DB 0 for user cache
userRedisConf := gormc.RedisConfig{
    Addr: "127.0.0.1:6379",
    DB:   0, // database 0
}
userCache, _ := gormc.NewConn(db, userRedisConf, time.Hour)

// Use Redis DB 1 for order cache
orderRedisConf := gormc.RedisConfig{
    Addr: "127.0.0.1:6379",
    DB:   1, // database 1
}
orderCache, _ := gormc.NewConn(db, orderRedisConf, 30*time.Minute)
```

## Quick Start

### Query with cache and custom expire duration
```go
gormzeroUsersIdKey := fmt.Sprintf("%s%v", cacheGormzeroUsersIdExpirePrefix, id)
var resp Users
err := m.QueryWithExpireCtx(ctx, &resp, gormzeroUsersIdKey, expire, func(conn *gorm.DB) error {
    return conn.Model(&Users{}).Where("`id` = ?", id).First(&resp).Error
})
switch err {
    case nil:
        return &resp, nil
    case gormc.ErrNotFound:
        return nil, ErrNotFound
    default:
        return nil, err
}
```

### Query with cache and default expire duration
```go
gormzeroUsersIdKey := fmt.Sprintf("%s%v", cacheGormzeroUsersIdPrefix, id)
var resp Users
err := m.QueryCtx(ctx, &resp, gormzeroUsersIdKey, func(conn *gorm.DB) error {
    return conn.Model(&Users{}).Where("`id` = ?", id).First(&resp).Error
})
switch err {
    case nil:
        return &resp, nil
    case gormc.ErrNotFound:
        return nil, ErrNotFound
    default:
        return nil, err
}
```

### Execute with cache invalidation
```go
err := m.ExecCtx(ctx, func(conn *gorm.DB) error {
    return conn.Model(&Users{}).Where("id = ?", id).Update("name", "new name").Error
}, gormzeroUsersIdKey)
```

### Query without cache
```go
var resp Users
err := m.QueryNoCacheCtx(ctx, func(conn *gorm.DB) error {
    return conn.Model(&Users{}).Where("`id` = ?", id).First(&resp).Error
})
```

## API Reference

### CachedConn Methods
- `QueryCtx` - Query with cache and default expiration
- `QueryWithExpireCtx` - Query with cache and custom expiration
- `QueryNoCacheCtx` - Query without cache
- `ExecCtx` - Execute with cache invalidation
- `ExecNoCacheCtx` - Execute without affecting cache
- `SetCache` / `SetCacheCtx` - Manually set cache
- `GetCache` / `GetCacheCtx` - Manually get cache
- `DelCache` / `DelCacheCtx` - Manually delete cache
- `Transact` / `TransactCtx` - Execute in transaction

## Examples
- go zero model example link: [gorm-zero-example](https://github.com/SpectatorNan/gorm-zero-example)
