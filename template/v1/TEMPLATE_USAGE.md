# 模板使用说明 (Template Usage Guide)

## 生成的代码示例

使用更新后的模板，生成的代码将支持原生 Redis 和 DB 选择。

### 使用 goctl 生成代码

```bash
# 从数据库表生成 Model 代码
goctl model mysql -src=./schema.sql -dir=./model -cache --home ./template

# 或者从现有数据库生成
goctl model mysql datasource -url="user:password@tcp(127.0.0.1:3306)/database" -table="users" -dir="./model" -cache --home ./template
```

## 生成的代码接口

### 无缓存版本

```go
func NewUserModel(conn *gorm.DB) (UserModel, error) {
    // 无缓存实现
}
```

**使用方式：**
```go
import (
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
)

// 连接数据库
dsn := "user:password@tcp(127.0.0.1:3306)/database?charset=utf8mb4"
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
if err != nil {
    panic(err)
}

// 创建 Model（无缓存）
userModel, err := NewUserModel(db)
if err != nil {
    panic(err)
}
```

### 带缓存版本（原生 Redis + DB 选择 + Cluster 支持）

```go
func NewUserModel(
    conn *gorm.DB, 
    redisConf gormc.RedisConfig, 
    cacheExpiry time.Duration,
) (UserModel, error) {
    // 支持单节点 Redis（DB 选择）和 Redis Cluster
}
```

**使用方式（单节点）：**
```go
import (
    "time"
    "github.com/huof6829/gorm-zero/gormc"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
)

// 1. 连接数据库
dsn := "user:password@tcp(127.0.0.1:3306)/database?charset=utf8mb4"
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
if err != nil {
    panic(err)
}

// 2. 配置 Redis（支持 DB 选择）
redisConf := gormc.RedisConfig{
    Addr:         "127.0.0.1:6379",
    Password:     "",                  // Redis 密码
    DB:           0,                   // Redis 数据库索引 (0-15)
    PoolSize:     10,                  // 连接池大小
    MinIdleConns: 2,                   // 最小空闲连接数
    DialTimeout:  5 * time.Second,     // 连接超时
    ReadTimeout:  3 * time.Second,     // 读取超时
    WriteTimeout: 3 * time.Second,     // 写入超时
}

// 3. 创建 Model（带缓存，缓存1小时）
userModel, err := NewUserModel(db, redisConf, time.Hour)
if err != nil {
    panic(err)
}

// 4. 使用 Model
user, err := userModel.FindOne(ctx, 1)
if err != nil {
    // 处理错误
}
```

**使用方式（Redis Cluster）：**
```go
import (
    "time"
    "github.com/huof6829/gorm-zero/gormc"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
)

// 1. 连接数据库
dsn := "user:password@tcp(127.0.0.1:3306)/database?charset=utf8mb4"
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
if err != nil {
    panic(err)
}

// 2. 配置 Redis Cluster
redisConf := gormc.RedisConfig{
    ClusterAddrs: []string{
        "127.0.0.1:7000",
        "127.0.0.1:7001",
        "127.0.0.1:7002",
        "127.0.0.1:7003",
        "127.0.0.1:7004",
        "127.0.0.1:7005",
    },
    Password:     "",              // Cluster 密码
    PoolSize:     50,              // Cluster 建议更大的连接池
    MinIdleConns: 10,
    DialTimeout:  10 * time.Second,
    ReadTimeout:  5 * time.Second,
    WriteTimeout: 5 * time.Second,
}

// 3. 创建 Model（使用 Cluster，缓存1小时）
userModel, err := NewUserModel(db, redisConf, time.Hour)
if err != nil {
    panic(err)
}

// 4. 使用 Model（与单节点完全相同）
user, err := userModel.FindOne(ctx, 1)
if err != nil {
    // 处理错误
}
```

**注意：**
- Redis Cluster **不支持** DB 选择（DB 参数会被忽略）
- Cluster 建议使用更大的连接池（PoolSize: 50-100）
- 使用方式与单节点完全相同，只是配置不同

## 多数据库/Cluster 混合缓存示例

不同的 Model 可以使用不同的 Redis 配置（单节点 + Cluster 混合）：

```go
// 用户缓存 - 使用单节点 Redis DB 0，缓存 1 小时
userRedisConf := gormc.RedisConfig{
    Addr: "127.0.0.1:6379",
    DB:   0,
}
userModel, _ := NewUserModel(db, userRedisConf, time.Hour)

// 订单缓存 - 使用单节点 Redis DB 1，缓存 30 分钟
orderRedisConf := gormc.RedisConfig{
    Addr: "127.0.0.1:6379",
    DB:   1,
}
orderModel, _ := NewOrderModel(db, orderRedisConf, 30*time.Minute)

// 商品缓存 - 使用 Redis Cluster（高并发场景），缓存 2 小时
productRedisConf := gormc.RedisConfig{
    ClusterAddrs: []string{
        "127.0.0.1:7000",
        "127.0.0.1:7001",
        "127.0.0.1:7002",
    },
    PoolSize: 50, // Cluster 需要更大的连接池
}
productModel, _ := NewProductModel(db, productRedisConf, 2*time.Hour)
```

## 配置文件示例

### YAML 配置

```yaml
# config.yaml
Database:
  Mysql:
    Host: "127.0.0.1"
    Port: 3306
    Username: "root"
    Password: "password"
    Database: "mydb"

Redis:
  Addr: "127.0.0.1:6379"
  Password: ""
  DB: 0
  PoolSize: 10
  MinIdleConns: 2

Cache:
  Expiry: 3600  # 秒
```

### 在 Go 中使用配置

```go
type Config struct {
    Database struct {
        Mysql struct {
            Host     string
            Port     int
            Username string
            Password string
            Database string
        }
    }
    Redis struct {
        Addr         string
        Password     string
        DB           int
        PoolSize     int
        MinIdleConns int
    }
    Cache struct {
        Expiry int // 秒
    }
}

// 初始化
func InitModels(c Config) error {
    // 数据库连接
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True",
        c.Database.Mysql.Username,
        c.Database.Mysql.Password,
        c.Database.Mysql.Host,
        c.Database.Mysql.Port,
        c.Database.Mysql.Database,
    )
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        return err
    }

    // Redis 配置
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

    // 缓存过期时间
    cacheExpiry := time.Duration(c.Cache.Expiry) * time.Second

    // 创建 Models
    userModel, err := NewUserModel(db, redisConf, cacheExpiry)
    if err != nil {
        return err
    }
    
    // ... 其他 Models
    
    return nil
}
```

## RedisConfig 完整参数说明

```go
type RedisConfig struct {
    // 单节点模式配置
    Addr         string        // Redis 服务器地址（单节点），格式：host:port
    Password     string        // Redis 密码（如果有）
    DB           int           // Redis 数据库索引（0-15，仅单节点）
    
    // Cluster 模式配置
    ClusterAddrs []string      // Redis Cluster 地址列表（设置后使用 Cluster 模式）
    
    // 通用配置
    PoolSize     int           // 连接池最大连接数（默认：10，Cluster 建议 50+）
    MinIdleConns int           // 最小空闲连接数（默认：2）
    DialTimeout  time.Duration // 连接超时（默认：5秒）
    ReadTimeout  time.Duration // 读取超时（默认：3秒）
    WriteTimeout time.Duration // 写入超时（默认：3秒）
}
```

**模式选择规则：**
- 如果设置了 `ClusterAddrs`，则使用 **Redis Cluster 模式**
- 否则使用 **单节点模式**（需要设置 `Addr`）

### 默认值

如果某些字段为零值，会使用以下默认值：
- `DialTimeout`: 5秒
- `ReadTimeout`: 3秒
- `WriteTimeout`: 3秒
- `PoolSize`: 10
- `MinIdleConns`: 2

**示例（最小配置）：**
```go
redisConf := gormc.RedisConfig{
    Addr: "127.0.0.1:6379",
    DB:   0,
}
// 其他参数将使用默认值
```

## 错误处理

生成的 Model 创建函数现在返回错误，请务必处理：

```go
// ✅ 正确的做法
userModel, err := NewUserModel(db, redisConf, time.Hour)
if err != nil {
    // 处理错误，可能是：
    // 1. Redis 连接失败
    // 2. Redis 认证失败
    // 3. 其他网络问题
    log.Fatalf("Failed to create user model: %v", err)
}

// ❌ 错误的做法（忽略错误）
userModel, _ := NewUserModel(db, redisConf, time.Hour)
```

## 迁移说明

如果你从旧版本的模板迁移，请参考项目根目录的 `MIGRATION.md` 文件。

主要变化：
1. `cache.CacheConf` → `gormc.RedisConfig`
2. 新增 `cacheExpiry time.Duration` 参数
3. Model 创建函数现在返回 `(Model, error)`
4. 支持 Redis DB 选择

## 常见问题

### Q: 如何在开发环境禁用缓存？

**方法1：** 生成不带缓存的 Model
```bash
goctl model mysql -src=./schema.sql -dir=./model --home ./template
# 注意：不加 -cache 参数
```

**方法2：** 使用极短的过期时间
```go
// 缓存只保持1秒（几乎等于不缓存）
userModel, _ := NewUserModel(db, redisConf, time.Second)
```

### Q: 如何手动刷新缓存？

生成的 Model 提供了缓存操作方法：
```go
// 删除缓存
err := userModel.DelCache(ctx, cacheKey)

// 手动设置缓存
err := userModel.SetCache(ctx, cacheKey, userData)
```

### Q: 不同环境使用不同的 Redis DB？

```go
var redisDB int

switch env {
case "dev":
    redisDB = 0
case "test":
    redisDB = 1
case "prod":
    redisDB = 2
}

redisConf := gormc.RedisConfig{
    Addr: "127.0.0.1:6379",
    DB:   redisDB,
}
```

## 更多信息

- 项目主页：https://github.com/huof6829/gorm-zero
- 迁移指南：[MIGRATION.md](../../MIGRATION.md)
- 更新日志：[CHANGELOG.md](../../CHANGELOG.md)

