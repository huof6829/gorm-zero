package gormc_test

import (
	"fmt"
	"time"

	"github.com/SpectatorNan/gorm-zero/gormc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ExampleNewConn_singleNode 展示单节点 Redis 配置
func ExampleNewConn_singleNode() {
	// 数据库连接
	dsn := "user:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	// 单节点 Redis 配置
	redisConf := gormc.RedisConfig{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0, // 单节点支持 DB 选择
		PoolSize: 10,
	}

	// 创建缓存连接
	cachedConn, err := gormc.NewConn(db, redisConf, time.Hour)
	if err != nil {
		panic(err)
	}

	_ = cachedConn
	fmt.Println("Single node Redis connected")
	// Output: Single node Redis connected
}

// ExampleNewConn_cluster 展示 Redis Cluster 配置
func ExampleNewConn_cluster() {
	// 数据库连接
	dsn := "user:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	// Redis Cluster 配置
	redisConf := gormc.RedisConfig{
		ClusterAddrs: []string{
			"127.0.0.1:7000",
			"127.0.0.1:7001",
			"127.0.0.1:7002",
			"127.0.0.1:7003",
			"127.0.0.1:7004",
			"127.0.0.1:7005",
		},
		Password: "",
		// 注意：Cluster 模式不支持 DB 选择
		PoolSize: 10,
	}

	// 创建缓存连接
	cachedConn, err := gormc.NewConn(db, redisConf, time.Hour)
	if err != nil {
		panic(err)
	}

	_ = cachedConn
	fmt.Println("Redis Cluster connected")
	// Output: Redis Cluster connected
}

// ExampleNewConn_clusterWithPassword 展示带密码的 Redis Cluster 配置
func ExampleNewConn_clusterWithPassword() {
	dsn := "user:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	// 带密码的 Redis Cluster 配置
	redisConf := gormc.RedisConfig{
		ClusterAddrs: []string{
			"node1.redis.example.com:6379",
			"node2.redis.example.com:6379",
			"node3.redis.example.com:6379",
		},
		Password:     "your_redis_password",
		PoolSize:     20,
		MinIdleConns: 5,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	cachedConn, err := gormc.NewConn(db, redisConf, 30*time.Minute)
	if err != nil {
		panic(err)
	}

	_ = cachedConn
}

// ExampleNewConn_multipleModels 展示不同 Model 使用不同的 Redis 配置
func ExampleNewConn_multipleModels() {
	dsn := "user:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	// 用户服务：使用单节点 Redis DB 0
	userRedisConf := gormc.RedisConfig{
		Addr: "127.0.0.1:6379",
		DB:   0,
	}
	userCache, _ := gormc.NewConn(db, userRedisConf, time.Hour)

	// 订单服务：使用单节点 Redis DB 1
	orderRedisConf := gormc.RedisConfig{
		Addr: "127.0.0.1:6379",
		DB:   1,
	}
	orderCache, _ := gormc.NewConn(db, orderRedisConf, 30*time.Minute)

	// 商品服务：使用 Redis Cluster（高并发场景）
	productRedisConf := gormc.RedisConfig{
		ClusterAddrs: []string{
			"127.0.0.1:7000",
			"127.0.0.1:7001",
			"127.0.0.1:7002",
		},
		PoolSize: 50, // 更大的连接池
	}
	productCache, _ := gormc.NewConn(db, productRedisConf, 2*time.Hour)

	_, _, _ = userCache, orderCache, productCache
}
