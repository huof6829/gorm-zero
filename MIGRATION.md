# 迁移指南 (Migration Guide)

## 从 go-zero Redis 迁移到原生 Redis

本文档说明如何从使用 go-zero 的 Redis 实现迁移到原生 `github.com/redis/go-redis/v9`。

### 主要变化

1. **Redis 配置结构**
   - 旧版本使用 `cache.CacheConf`
   - 新版本使用 `gormc.RedisConfig`，并支持 DB 选择

2. **连接创建方式**
   - 旧版本：`NewConn(db, cacheConf, opts...)`
   - 新版本：`NewConn(db, redisConf, expiry)` 返回 `(CachedConn, error)`

3. **DB 支持**
   - 新版本支持通过 `RedisConfig.DB` 字段选择 Redis 数据库 (0-15)

### 迁移步骤

#### 第一步：更新配置结构

**旧版本：**
```go
import (
    "github.com/zeromicro/go-zero/core/stores/cache"
)

type Config struct {
    CacheConf cache.CacheConf
}
```

**新版本：**
```go
import (
    "time"
    "github.com/SpectatorNan/gorm-zero/gormc"
)

type Config struct {
    Redis struct {
        Addr         string
        Password     string
        DB           int
        PoolSize     int
        MinIdleConns int
    }
    CacheExpiry time.Duration // 缓存过期时间
}
```

#### 第二步：更新初始化代码

**旧版本：**
```go
import (
    "github.com/zeromicro/go-zero/core/stores/cache"
)

func NewServiceContext(c config.Config) *ServiceContext {
    db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    
    cachedConn := gormc.NewConn(db, c.CacheConf)
    
    return &ServiceContext{
        CachedConn: cachedConn,
    }
}
```

**新版本：**
```go
import (
    "time"
    "github.com/SpectatorNan/gorm-zero/gormc"
)

func NewServiceContext(c config.Config) *ServiceContext {
    db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    
    redisConf := gormc.RedisConfig{
        Addr:         c.Redis.Addr,
        Password:     c.Redis.Password,
        DB:           c.Redis.DB,
        PoolSize:     c.Redis.PoolSize,
        MinIdleConns: c.Redis.MinIdleConns,
        DialTimeout:  5 * time.Second,
        ReadTimeout:  3 * time.Second,
        WriteTimeout: 3 * time.Second,
    }
    
    cachedConn, err := gormc.NewConn(db, redisConf, c.CacheExpiry)
    if err != nil {
        panic(err)
    }
    
    return &ServiceContext{
        CachedConn: cachedConn,
    }
}
```

#### 第三步：更新配置文件

**旧版本 (YAML)：**
```yaml
CacheConf:
  - Host: 127.0.0.1:6379
    Pass: ""
    Type: node
```

**新版本 (YAML)：**
```yaml
Redis:
  Addr: "127.0.0.1:6379"
  Password: ""
  DB: 0
  PoolSize: 10
  MinIdleConns: 2
CacheExpiry: 3600000000000  # 1小时 (纳秒)
```

### 高级特性

#### 使用多个 Redis 数据库

新版本支持在同一个应用中使用不同的 Redis 数据库：

```go
// 用户缓存使用 DB 0
userRedisConf := gormc.RedisConfig{
    Addr: "127.0.0.1:6379",
    DB:   0,
}
userCache, _ := gormc.NewConn(db, userRedisConf, time.Hour)

// 订单缓存使用 DB 1
orderRedisConf := gormc.RedisConfig{
    Addr: "127.0.0.1:6379",
    DB:   1,
}
orderCache, _ := gormc.NewConn(db, orderRedisConf, 30*time.Minute)

// 会话缓存使用 DB 2
sessionRedisConf := gormc.RedisConfig{
    Addr: "127.0.0.1:6379",
    DB:   2,
}
sessionCache, _ := gormc.NewConn(db, sessionRedisConf, 15*time.Minute)
```

#### 连接池配置

```go
redisConf := gormc.RedisConfig{
    Addr:         "127.0.0.1:6379",
    Password:     "your_password",
    DB:           0,
    PoolSize:     50,              // 最大连接数
    MinIdleConns: 10,              // 最小空闲连接数
    DialTimeout:  10 * time.Second, // 连接超时
    ReadTimeout:  5 * time.Second,  // 读取超时
    WriteTimeout: 5 * time.Second,  // 写入超时
}
```

### API 变化

所有的 API 方法保持不变，只是初始化方式发生了变化：

- `QueryCtx` ✅ 不变
- `QueryWithExpireCtx` ✅ 不变
- `QueryNoCacheCtx` ✅ 不变
- `ExecCtx` ✅ 不变
- `ExecNoCacheCtx` ✅ 不变
- `SetCache` / `SetCacheCtx` ✅ 不变
- `GetCache` / `GetCacheCtx` ✅ 不变
- `DelCache` / `DelCacheCtx` ✅ 不变
- `Transact` / `TransactCtx` ✅ 不变

### 注意事项

1. **错误处理**：新版本的 `NewConn` 返回 `(CachedConn, error)`，需要处理错误
2. **过期时间**：现在需要在创建连接时指定默认过期时间
3. **连接测试**：新版本在初始化时会测试 Redis 连接，如果连接失败会返回错误
4. **依赖更新**：确保 `go.mod` 中包含 `github.com/redis/go-redis/v9`

### 常见问题

#### Q: 如何获取底层的 Redis 客户端？

```go
cache := cachedConn.cache
redisClient := cache.GetClient() // *redis.Client
```

#### Q: 如何关闭 Redis 连接？

```go
if err := cachedConn.cache.Close(); err != nil {
    log.Printf("Failed to close redis: %v", err)
}
```

#### Q: 如何检查 Redis 连接状态？

```go
ctx := context.Background()
if err := cache.GetClient().Ping(ctx).Err(); err != nil {
    log.Printf("Redis connection failed: %v", err)
}
```

### 总结

新版本使用原生 Redis 客户端提供了更多的灵活性和控制力，特别是：
- 支持 Redis DB 选择
- 更细粒度的连接池配置
- 更好的超时控制
- 与 go-redis 生态系统完全兼容

如有任何问题，请提交 issue。

