# 更新日志 (Changelog)

## [未发布版本] - Native Redis 支持

### 重大变更 (Breaking Changes)

#### 1. 替换为原生 Redis 客户端
- **之前**: 使用 `github.com/zeromicro/go-zero/core/stores/redis` 和 `cache`
- **现在**: 使用 `github.com/redis/go-redis/v9` 原生客户端

#### 2. API 变更

##### NewConn 函数签名变更
```go
// 旧版本
func NewConn(db *gorm.DB, c cache.CacheConf, opts ...cache.Option) CachedConn

// 新版本
func NewConn(db *gorm.DB, redisConf RedisConfig, expiry time.Duration) (CachedConn, error)
```

##### 移除的函数
- `NewNodeConn` - 不再需要，统一使用 `NewConn`

#### 3. 配置结构变更

##### 新的 RedisConfig 结构
```go
type RedisConfig struct {
    Addr         string        // Redis 服务器地址
    Password     string        // Redis 密码
    DB           int           // Redis 数据库索引 (0-15)
    PoolSize     int           // 连接池大小
    MinIdleConns int           // 最小空闲连接数
    DialTimeout  time.Duration // 拨号超时
    ReadTimeout  time.Duration // 读取超时
    WriteTimeout time.Duration // 写入超时
}
```

### 新增功能 (New Features)

#### 1. ✨ Redis 数据库选择支持
现在可以通过 `RedisConfig.DB` 字段选择不同的 Redis 数据库（0-15）

```go
redisConf := gormc.RedisConfig{
    Addr: "127.0.0.1:6379",
    DB:   1, // 选择数据库 1
}
```

#### 2. ✨ 更灵活的连接池配置
提供了完整的连接池配置选项：
- `PoolSize`: 最大连接数
- `MinIdleConns`: 最小空闲连接数
- `DialTimeout`: 连接超时
- `ReadTimeout`: 读取超时
- `WriteTimeout`: 写入超时

#### 3. ✨ 连接测试
在初始化时自动测试 Redis 连接，确保配置正确

#### 4. ✨ 获取底层 Redis 客户端
```go
cache := cachedConn.cache
redisClient := cache.GetClient() // 返回 *redis.Client
```

### 优化改进 (Improvements)

#### 1. 🚀 更好的性能
- 原生 Redis 客户端性能更优
- 减少了中间层的开销

#### 2. 🔧 更好的错误处理
- `NewConn` 返回错误，便于处理连接失败的情况
- 初始化时进行连接测试

#### 3. 📝 更完善的文档
- 添加了详细的使用示例
- 添加了迁移指南
- 添加了完整的集成测试

### 不变的功能 (Unchanged)

以下 API 保持不变，可以无缝使用：
- ✅ `QueryCtx` - 查询并缓存（使用默认过期时间）
- ✅ `QueryWithExpireCtx` - 查询并缓存（自定义过期时间）
- ✅ `QueryNoCacheCtx` - 不使用缓存查询
- ✅ `ExecCtx` - 执行并删除缓存
- ✅ `ExecNoCacheCtx` - 执行不影响缓存
- ✅ `SetCache` / `SetCacheCtx` - 手动设置缓存
- ✅ `GetCache` / `GetCacheCtx` - 手动获取缓存
- ✅ `DelCache` / `DelCacheCtx` - 手动删除缓存
- ✅ `Transact` / `TransactCtx` - 事务执行

### 测试覆盖 (Test Coverage)

新增了完整的测试套件：
- ✅ 基础操作测试
- ✅ 自定义过期时间测试
- ✅ 缓存失效测试
- ✅ 无缓存查询测试
- ✅ 手动缓存操作测试
- ✅ 多数据库支持测试
- ✅ 事务测试
- ✅ 性能基准测试

### 迁移指南

请参阅 [MIGRATION.md](./MIGRATION.md) 了解详细的迁移步骤。

### 依赖变更 (Dependencies)

#### 新增依赖
- `github.com/redis/go-redis/v9` - 原生 Redis 客户端
- `github.com/alicebob/miniredis/v2` - 测试用内存 Redis (开发依赖)

#### 移除依赖
- `github.com/zeromicro/go-zero/core/stores/redis` - 不再需要
- `github.com/zeromicro/go-zero/core/stores/cache` - 不再需要

#### 保留依赖
- `github.com/zeromicro/go-zero/core/mathx` - 用于不稳定过期时间
- 其他 GORM 和数据库驱动依赖保持不变

### 使用示例

#### 基础使用
```go
import (
    "time"
    "github.com/huof6829/gorm-zero/gormc"
)

// 配置 Redis
redisConf := gormc.RedisConfig{
    Addr:         "127.0.0.1:6379",
    Password:     "",
    DB:           0,
    PoolSize:     10,
    MinIdleConns: 2,
}

// 创建缓存连接
cachedConn, err := gormc.NewConn(db, redisConf, time.Hour)
if err != nil {
    panic(err)
}

// 使用
var user User
err = cachedConn.QueryCtx(ctx, &user, "user:1", func(conn *gorm.DB) error {
    return conn.Where("id = ?", 1).First(&user).Error
})
```

#### 多数据库使用
```go
// 用户缓存 - DB 0
userCache, _ := gormc.NewConn(db, gormc.RedisConfig{
    Addr: "127.0.0.1:6379",
    DB:   0,
}, time.Hour)

// 订单缓存 - DB 1
orderCache, _ := gormc.NewConn(db, gormc.RedisConfig{
    Addr: "127.0.0.1:6379",
    DB:   1,
}, 30*time.Minute)
```

### 已知问题 (Known Issues)

无

### 后续计划 (Roadmap)

- [ ] 支持 Redis Cluster
- [ ] 支持 Redis Sentinel
- [ ] 添加更多的缓存策略
- [ ] 性能优化和监控

### 贡献者 (Contributors)

感谢所有为这个版本做出贡献的开发者！

---

## 之前的版本

[保留之前的更新日志...]

