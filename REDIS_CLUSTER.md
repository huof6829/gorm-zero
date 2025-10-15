# Redis Cluster 支持文档

## 概述

gorm-zero 现在**完全支持** Redis Cluster 模式，同时保持对单节点 Redis 的兼容。

## 功能对比

| 特性 | 单节点 Redis | Redis Cluster |
|------|-------------|---------------|
| 数据库选择 (DB) | ✅ 支持 (0-15) | ❌ 不支持 |
| 高可用性 | ❌ 单点故障 | ✅ 自动故障转移 |
| 水平扩展 | ❌ 不支持 | ✅ 支持分片 |
| 性能 | 🟡 单节点限制 | ✅ 多节点分布式 |
| 配置复杂度 | 🟢 简单 | 🟡 中等 |

## 使用方式

### 1. 单节点 Redis（默认）

适用于：开发环境、小规模应用、需要使用不同 DB 的场景

```go
import (
    "time"
    "github.com/SpectatorNan/gorm-zero/gormc"
)

redisConf := gormc.RedisConfig{
    Addr:     "127.0.0.1:6379",  // 单节点地址
    Password: "",
    DB:       0,                  // ✅ 支持 DB 选择 (0-15)
    PoolSize: 10,
}

cachedConn, err := gormc.NewConn(db, redisConf, time.Hour)
if err != nil {
    panic(err)
}
```

### 2. Redis Cluster

适用于：生产环境、大规模应用、高可用要求

```go
redisConf := gormc.RedisConfig{
    ClusterAddrs: []string{      // 设置集群地址
        "127.0.0.1:7000",
        "127.0.0.1:7001",
        "127.0.0.1:7002",
        "127.0.0.1:7003",
        "127.0.0.1:7004",
        "127.0.0.1:7005",
    },
    Password: "",
    // 注意：Cluster 不支持 DB 参数
    PoolSize: 20,                 // Cluster 建议更大的连接池
}

cachedConn, err := gormc.NewConn(db, redisConf, time.Hour)
if err != nil {
    panic(err)
}
```

## 配置说明

### RedisConfig 结构

```go
type RedisConfig struct {
    // 单节点模式配置
    Addr         string        // Redis 服务器地址 (单节点)
    Password     string        // Redis 密码
    DB           int           // Redis 数据库索引（仅单节点，Cluster 不支持）
    
    // Cluster 模式配置
    ClusterAddrs []string      // Redis Cluster 地址列表
    
    // 通用配置
    PoolSize     int           // 连接池大小
    MinIdleConns int           // 最小空闲连接数
    DialTimeout  time.Duration // 连接超时
    ReadTimeout  time.Duration // 读取超时
    WriteTimeout time.Duration // 写入超时
}
```

### 模式判断规则

```go
// 如果设置了 ClusterAddrs，使用 Cluster 模式
if len(conf.ClusterAddrs) > 0 {
    // Redis Cluster Mode
} else {
    // Single Node Mode (需要设置 Addr)
}
```

## 配置示例

### 开发环境（单节点）

```yaml
# config.yaml
Redis:
  Addr: "localhost:6379"
  Password: ""
  DB: 0
  PoolSize: 10
  MinIdleConns: 2
```

```go
redisConf := gormc.RedisConfig{
    Addr:         config.Redis.Addr,
    Password:     config.Redis.Password,
    DB:           config.Redis.DB,
    PoolSize:     config.Redis.PoolSize,
    MinIdleConns: config.Redis.MinIdleConns,
}
```

### 生产环境（Cluster）

```yaml
# config.yaml
Redis:
  ClusterAddrs:
    - "redis-node1.example.com:6379"
    - "redis-node2.example.com:6379"
    - "redis-node3.example.com:6379"
    - "redis-node4.example.com:6379"
    - "redis-node5.example.com:6379"
    - "redis-node6.example.com:6379"
  Password: "your_secure_password"
  PoolSize: 50
  MinIdleConns: 10
  DialTimeout: 10s
  ReadTimeout: 5s
  WriteTimeout: 5s
```

```go
redisConf := gormc.RedisConfig{
    ClusterAddrs: config.Redis.ClusterAddrs,
    Password:     config.Redis.Password,
    PoolSize:     config.Redis.PoolSize,
    MinIdleConns: config.Redis.MinIdleConns,
    DialTimeout:  config.Redis.DialTimeout,
    ReadTimeout:  config.Redis.ReadTimeout,
    WriteTimeout: config.Redis.WriteTimeout,
}
```

## 混合使用场景

在同一个应用中，可以同时使用单节点和 Cluster：

```go
// 用户缓存：使用单节点 DB 0（需要与遗留系统兼容）
userRedisConf := gormc.RedisConfig{
    Addr: "127.0.0.1:6379",
    DB:   0,
}
userCache, _ := gormc.NewConn(db, userRedisConf, time.Hour)

// 商品缓存：使用 Cluster（高并发，大数据量）
productRedisConf := gormc.RedisConfig{
    ClusterAddrs: []string{
        "127.0.0.1:7000",
        "127.0.0.1:7001",
        "127.0.0.1:7002",
    },
    PoolSize: 50,
}
productCache, _ := gormc.NewConn(db, productRedisConf, time.Hour)
```

## 搭建 Redis Cluster

### 使用 Docker Compose

```yaml
version: '3.8'

services:
  redis-node-1:
    image: redis:7-alpine
    command: redis-server --port 7000 --cluster-enabled yes --cluster-config-file nodes.conf
    ports:
      - "7000:7000"
  
  redis-node-2:
    image: redis:7-alpine
    command: redis-server --port 7001 --cluster-enabled yes --cluster-config-file nodes.conf
    ports:
      - "7001:7001"
  
  redis-node-3:
    image: redis:7-alpine
    command: redis-server --port 7002 --cluster-enabled yes --cluster-config-file nodes.conf
    ports:
      - "7002:7002"
  
  redis-node-4:
    image: redis:7-alpine
    command: redis-server --port 7003 --cluster-enabled yes --cluster-config-file nodes.conf
    ports:
      - "7003:7003"
  
  redis-node-5:
    image: redis:7-alpine
    command: redis-server --port 7004 --cluster-enabled yes --cluster-config-file nodes.conf
    ports:
      - "7004:7004"
  
  redis-node-6:
    image: redis:7-alpine
    command: redis-server --port 7005 --cluster-enabled yes --cluster-config-file nodes.conf
    ports:
      - "7005:7005"
```

### 初始化 Cluster

```bash
# 创建集群
redis-cli --cluster create \
  127.0.0.1:7000 \
  127.0.0.1:7001 \
  127.0.0.1:7002 \
  127.0.0.1:7003 \
  127.0.0.1:7004 \
  127.0.0.1:7005 \
  --cluster-replicas 1

# 检查集群状态
redis-cli -c -p 7000 cluster nodes
redis-cli -c -p 7000 cluster info
```

## 性能优化建议

### 单节点 Redis

```go
redisConf := gormc.RedisConfig{
    Addr:         "127.0.0.1:6379",
    PoolSize:     10,              // 适中的连接池
    MinIdleConns: 2,
    DialTimeout:  5 * time.Second,
    ReadTimeout:  3 * time.Second,
    WriteTimeout: 3 * time.Second,
}
```

### Redis Cluster（高并发）

```go
redisConf := gormc.RedisConfig{
    ClusterAddrs: clusterNodes,
    PoolSize:     100,             // 更大的连接池
    MinIdleConns: 20,              // 更多空闲连接
    DialTimeout:  10 * time.Second, // 更长的超时
    ReadTimeout:  5 * time.Second,
    WriteTimeout: 5 * time.Second,
}
```

## 故障处理

### 单节点故障

```go
cachedConn, err := gormc.NewConn(db, redisConf, time.Hour)
if err != nil {
    // 单节点连接失败，整个服务不可用
    log.Fatalf("Redis connection failed: %v", err)
}
```

### Cluster 节点故障

Redis Cluster 会自动处理节点故障：
- 从节点自动提升为主节点
- 请求自动路由到可用节点
- 只要大部分节点可用，集群仍可正常工作

```go
cachedConn, err := gormc.NewConn(db, redisConf, time.Hour)
if err != nil {
    // Cluster 整体不可用（超过半数节点故障）
    log.Fatalf("Redis Cluster connection failed: %v", err)
}
```

## 监控和调试

### 获取底层客户端

```go
cache := cachedConn.cache
client := cache.GetClient() // 返回 redis.Cmdable

// 类型断言以使用特定功能
switch c := client.(type) {
case *redis.Client:
    // 单节点客户端
    stats := c.PoolStats()
    fmt.Printf("Single node pool stats: %+v\n", stats)
    
case *redis.ClusterClient:
    // Cluster 客户端
    stats := c.PoolStats()
    fmt.Printf("Cluster pool stats: %+v\n", stats)
    
    // 获取 Cluster 信息
    clusterInfo, _ := c.ClusterInfo(context.Background()).Result()
    fmt.Printf("Cluster info: %s\n", clusterInfo)
}
```

### 健康检查

```go
func HealthCheck(cache *gormc.RedisCache) error {
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()
    
    client := cache.GetClient()
    return client.Ping(ctx).Err()
}
```

## 迁移指南

### 从单节点迁移到 Cluster

#### 1. 准备 Cluster 环境

搭建至少 6 个节点的 Redis Cluster（3主3从）

#### 2. 更新配置

```go
// 之前（单节点）
redisConf := gormc.RedisConfig{
    Addr: "127.0.0.1:6379",
    DB:   0,  // ⚠️ Cluster 不支持 DB
}

// 之后（Cluster）
redisConf := gormc.RedisConfig{
    ClusterAddrs: []string{
        "node1:6379",
        "node2:6379",
        "node3:6379",
        "node4:6379",
        "node5:6379",
        "node6:6379",
    },
    // 移除 DB 配置
}
```

#### 3. 数据迁移

如果之前使用了多个 DB，需要在应用层区分不同的 key 前缀：

```go
// 之前使用 DB 0
userKey := "user:123"

// 之前使用 DB 1
orderKey := "order:456"

// Cluster 中需要使用前缀区分
userKey := "db0:user:123"
orderKey := "db1:order:456"
```

## 常见问题

### Q: 如何选择单节点还是 Cluster？

**使用单节点如果：**
- 数据量 < 10GB
- QPS < 10,000
- 可以接受短暂的不可用
- 需要使用不同的 DB

**使用 Cluster 如果：**
- 数据量 > 10GB
- QPS > 10,000
- 需要高可用性
- 需要水平扩展

### Q: Cluster 模式下如何实现类似 DB 的隔离？

使用 key 前缀：

```go
// DB 0 的效果
userCache := "app:user:" + userID

// DB 1 的效果  
orderCache := "app:order:" + orderID

// DB 2 的效果
productCache := "app:product:" + productID
```

### Q: 可以动态切换单节点和 Cluster 吗？

可以，只需修改配置并重启应用：

```go
func NewRedisConfig() gormc.RedisConfig {
    if os.Getenv("REDIS_MODE") == "cluster" {
        return gormc.RedisConfig{
            ClusterAddrs: strings.Split(os.Getenv("REDIS_CLUSTER_ADDRS"), ","),
        }
    }
    return gormc.RedisConfig{
        Addr: os.Getenv("REDIS_ADDR"),
        DB:   0,
    }
}
```

### Q: Cluster 的连接池如何配置？

```go
// Cluster 需要更大的连接池，因为需要连接多个节点
redisConf := gormc.RedisConfig{
    ClusterAddrs: []string{...},
    PoolSize:     节点数 * 10,    // 例如：6个节点 * 10 = 60
    MinIdleConns: 节点数 * 2,     // 例如：6个节点 * 2 = 12
}
```

## 参考资源

- [Redis Cluster 官方文档](https://redis.io/docs/management/scaling/)
- [go-redis Cluster 文档](https://redis.uptrace.dev/guide/go-redis-cluster.html)
- [Redis Cluster 教程](https://redis.io/docs/manual/scaling/)

## 更多信息

- 项目主页：https://github.com/huof6829/gorm-zero
- 基础文档：[README.md](./README.md)
- 迁移指南：[MIGRATION.md](./MIGRATION.md)

