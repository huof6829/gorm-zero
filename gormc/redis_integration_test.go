package gormc_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/SpectatorNan/gorm-zero/gormc"
	"github.com/alicebob/miniredis/v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type TestUser struct {
	ID        int64     `gorm:"primaryKey"`
	Name      string    `gorm:"column:name"`
	Email     string    `gorm:"column:email"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (TestUser) TableName() string {
	return "users"
}

// setupTestEnv 设置测试环境（内存数据库 + 内存Redis）
func setupTestEnv(t *testing.T) (*gorm.DB, *miniredis.Miniredis, gormc.CachedConn) {
	// 创建内存 SQLite 数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// 自动迁移
	if err := db.AutoMigrate(&TestUser{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// 创建 miniredis 实例（内存 Redis）
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}

	// 配置 Redis
	redisConf := gormc.RedisConfig{
		Addr:         mr.Addr(),
		Password:     "",
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 2,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	// 创建缓存连接
	cachedConn, err := gormc.NewConn(db, redisConf, time.Minute)
	if err != nil {
		mr.Close()
		t.Fatalf("Failed to create cached conn: %v", err)
	}

	return db, mr, cachedConn
}

func TestRedisCache_BasicOperations(t *testing.T) {
	db, mr, cachedConn := setupTestEnv(t)
	defer mr.Close()

	ctx := context.Background()

	// 插入测试数据
	user := TestUser{
		ID:        1,
		Name:      "Test User",
		Email:     "test@example.com",
		CreatedAt: time.Now(),
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// 测试带缓存的查询
	key := "user:1"
	var result TestUser
	err := cachedConn.QueryCtx(ctx, &result, key, func(conn *gorm.DB) error {
		return conn.Where("id = ?", 1).First(&result).Error
	})

	if err != nil {
		t.Fatalf("QueryCtx failed: %v", err)
	}

	if result.Name != "Test User" {
		t.Errorf("Expected name 'Test User', got '%s'", result.Name)
	}

	// 验证缓存是否生效（再次查询）
	var cachedResult TestUser
	err = cachedConn.GetCacheCtx(ctx, key, &cachedResult)
	if err != nil {
		t.Fatalf("GetCacheCtx failed: %v", err)
	}

	if cachedResult.Name != "Test User" {
		t.Errorf("Expected cached name 'Test User', got '%s'", cachedResult.Name)
	}
}

func TestRedisCache_CustomExpiry(t *testing.T) {
	db, mr, cachedConn := setupTestEnv(t)
	defer mr.Close()

	ctx := context.Background()

	// 插入测试数据
	user := TestUser{
		ID:        2,
		Name:      "User With Custom Expiry",
		Email:     "user2@example.com",
		CreatedAt: time.Now(),
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// 测试自定义过期时间的查询（5秒）
	key := "user:2"
	var result TestUser
	err := cachedConn.QueryWithExpireCtx(ctx, &result, key, 5*time.Second, func(conn *gorm.DB) error {
		return conn.Where("id = ?", 2).First(&result).Error
	})

	if err != nil {
		t.Fatalf("QueryWithExpireCtx failed: %v", err)
	}

	if result.Name != "User With Custom Expiry" {
		t.Errorf("Expected name 'User With Custom Expiry', got '%s'", result.Name)
	}

	// 验证缓存存在
	var cachedResult TestUser
	err = cachedConn.GetCacheCtx(ctx, key, &cachedResult)
	if err != nil {
		t.Fatalf("GetCacheCtx failed: %v", err)
	}
}

func TestRedisCache_ExecWithCacheInvalidation(t *testing.T) {
	db, mr, cachedConn := setupTestEnv(t)
	defer mr.Close()

	ctx := context.Background()

	// 插入测试数据
	user := TestUser{
		ID:        3,
		Name:      "Original Name",
		Email:     "user3@example.com",
		CreatedAt: time.Now(),
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	key := "user:3"

	// 先查询并缓存
	var result TestUser
	err := cachedConn.QueryCtx(ctx, &result, key, func(conn *gorm.DB) error {
		return conn.Where("id = ?", 3).First(&result).Error
	})
	if err != nil {
		t.Fatalf("QueryCtx failed: %v", err)
	}

	// 更新数据并删除缓存
	err = cachedConn.ExecCtx(ctx, func(conn *gorm.DB) error {
		return conn.Model(&TestUser{}).Where("id = ?", 3).Update("name", "Updated Name").Error
	}, key)
	if err != nil {
		t.Fatalf("ExecCtx failed: %v", err)
	}

	// 验证缓存已被删除
	var cachedResult TestUser
	err = cachedConn.GetCacheCtx(ctx, key, &cachedResult)
	if err == nil {
		t.Error("Expected cache to be deleted, but it still exists")
	}

	// 再次查询应该从数据库获取新数据
	err = cachedConn.QueryCtx(ctx, &result, key, func(conn *gorm.DB) error {
		return conn.Where("id = ?", 3).First(&result).Error
	})
	if err != nil {
		t.Fatalf("QueryCtx after update failed: %v", err)
	}

	if result.Name != "Updated Name" {
		t.Errorf("Expected updated name 'Updated Name', got '%s'", result.Name)
	}
}

func TestRedisCache_QueryNoCache(t *testing.T) {
	db, mr, cachedConn := setupTestEnv(t)
	defer mr.Close()

	ctx := context.Background()

	// 插入测试数据
	user := TestUser{
		ID:        4,
		Name:      "No Cache User",
		Email:     "user4@example.com",
		CreatedAt: time.Now(),
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// 无缓存查询
	var result TestUser
	err := cachedConn.QueryNoCacheCtx(ctx, func(conn *gorm.DB) error {
		return conn.Where("id = ?", 4).First(&result).Error
	})
	if err != nil {
		t.Fatalf("QueryNoCacheCtx failed: %v", err)
	}

	if result.Name != "No Cache User" {
		t.Errorf("Expected name 'No Cache User', got '%s'", result.Name)
	}

	// 验证确实没有缓存
	key := "user:4"
	var cachedResult TestUser
	err = cachedConn.GetCacheCtx(ctx, key, &cachedResult)
	if err == nil {
		t.Error("Expected no cache, but cache exists")
	}
}

func TestRedisCache_ManualCacheOperations(t *testing.T) {
	_, mr, cachedConn := setupTestEnv(t)
	defer mr.Close()

	ctx := context.Background()

	// 测试手动设置缓存
	key := "manual:test"
	testData := TestUser{
		ID:    100,
		Name:  "Manual Cache Test",
		Email: "manual@example.com",
	}

	err := cachedConn.SetCacheCtx(ctx, key, testData)
	if err != nil {
		t.Fatalf("SetCacheCtx failed: %v", err)
	}

	// 测试手动获取缓存
	var result TestUser
	err = cachedConn.GetCacheCtx(ctx, key, &result)
	if err != nil {
		t.Fatalf("GetCacheCtx failed: %v", err)
	}

	if result.Name != "Manual Cache Test" {
		t.Errorf("Expected name 'Manual Cache Test', got '%s'", result.Name)
	}

	// 测试手动删除缓存
	err = cachedConn.DelCacheCtx(ctx, key)
	if err != nil {
		t.Fatalf("DelCacheCtx failed: %v", err)
	}

	// 验证缓存已删除
	err = cachedConn.GetCacheCtx(ctx, key, &result)
	if err == nil {
		t.Error("Expected cache to be deleted, but it still exists")
	}
}

func TestRedisCache_MultipleDBs(t *testing.T) {
	// 创建内存 SQLite 数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// 自动迁移
	if err := db.AutoMigrate(&TestUser{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// 创建两个 miniredis 实例模拟不同的 DB
	mr1, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis 1: %v", err)
	}
	defer mr1.Close()

	mr2, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis 2: %v", err)
	}
	defer mr2.Close()

	// 配置两个不同的 Redis 连接（使用不同的地址模拟不同的 DB）
	redis1Conf := gormc.RedisConfig{
		Addr: mr1.Addr(),
		DB:   0,
	}
	cache1, err := gormc.NewConn(db, redis1Conf, time.Hour)
	if err != nil {
		t.Fatalf("Failed to create cache1: %v", err)
	}

	redis2Conf := gormc.RedisConfig{
		Addr: mr2.Addr(),
		DB:   0,
	}
	cache2, err := gormc.NewConn(db, redis2Conf, time.Hour)
	if err != nil {
		t.Fatalf("Failed to create cache2: %v", err)
	}

	ctx := context.Background()

	// 在第一个缓存中设置数据
	key := "test:key"
	testData := TestUser{ID: 1, Name: "Test", Email: "test@example.com"}
	if err := cache1.SetCacheCtx(ctx, key, testData); err != nil {
		t.Fatalf("Failed to set cache1: %v", err)
	}

	// 验证第一个缓存中有数据
	var result1 TestUser
	if err := cache1.GetCacheCtx(ctx, key, &result1); err != nil {
		t.Fatalf("Failed to get from cache1: %v", err)
	}

	// 验证第二个缓存中没有数据（因为是不同的 Redis 实例）
	var result2 TestUser
	err = cache2.GetCacheCtx(ctx, key, &result2)
	if err == nil {
		t.Error("Expected cache2 to be empty, but it has data")
	}
}

func TestRedisCache_Transaction(t *testing.T) {
	db, mr, cachedConn := setupTestEnv(t)
	defer mr.Close()

	ctx := context.Background()

	// 测试事务
	err := cachedConn.TransactCtx(ctx, func(tx *gorm.DB) error {
		// 插入第一个用户
		user1 := TestUser{
			ID:        10,
			Name:      "User 10",
			Email:     "user10@example.com",
			CreatedAt: time.Now(),
		}
		if err := tx.Create(&user1).Error; err != nil {
			return err
		}

		// 插入第二个用户
		user2 := TestUser{
			ID:        11,
			Name:      "User 11",
			Email:     "user11@example.com",
			CreatedAt: time.Now(),
		}
		if err := tx.Create(&user2).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		t.Fatalf("Transaction failed: %v", err)
	}

	// 验证数据已插入
	var count int64
	db.Model(&TestUser{}).Where("id IN ?", []int{10, 11}).Count(&count)
	if count != 2 {
		t.Errorf("Expected 2 users, got %d", count)
	}
}

// 基准测试
func BenchmarkRedisCache_Query(b *testing.B) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		b.Fatalf("Failed to open database: %v", err)
	}

	if err := db.AutoMigrate(&TestUser{}); err != nil {
		b.Fatalf("Failed to migrate: %v", err)
	}

	mr, err := miniredis.Run()
	if err != nil {
		b.Fatalf("Failed to start miniredis: %v", err)
	}
	defer mr.Close()

	redisConf := gormc.RedisConfig{
		Addr: mr.Addr(),
		DB:   0,
	}

	cachedConn, err := gormc.NewConn(db, redisConf, time.Minute)
	if err != nil {
		b.Fatalf("Failed to create cached conn: %v", err)
	}

	// 插入测试数据
	user := TestUser{
		ID:        1,
		Name:      "Benchmark User",
		Email:     "benchmark@example.com",
		CreatedAt: time.Now(),
	}
	if err := db.Create(&user).Error; err != nil {
		b.Fatalf("Failed to create user: %v", err)
	}

	ctx := context.Background()
	key := "benchmark:user:1"

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var result TestUser
			_ = cachedConn.QueryCtx(ctx, &result, key, func(conn *gorm.DB) error {
				return conn.Where("id = ?", 1).First(&result).Error
			})
		}
	})
}

// 示例：完整的使用流程
func ExampleCachedConn_fullWorkflow() {
	// 1. 初始化数据库
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&TestUser{})

	// 2. 配置 Redis
	redisConf := gormc.RedisConfig{
		Addr:         "127.0.0.1:6379",
		Password:     "",
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 2,
	}

	// 3. 创建缓存连接
	cachedConn, _ := gormc.NewConn(db, redisConf, time.Hour)

	ctx := context.Background()

	// 4. 插入数据
	user := TestUser{ID: 1, Name: "John", Email: "john@example.com"}
	cachedConn.ExecNoCacheCtx(ctx, func(conn *gorm.DB) error {
		return conn.Create(&user).Error
	})

	// 5. 查询（带缓存）
	var result TestUser
	key := fmt.Sprintf("user:%d", 1)
	cachedConn.QueryCtx(ctx, &result, key, func(conn *gorm.DB) error {
		return conn.Where("id = ?", 1).First(&result).Error
	})

	// 6. 更新（删除缓存）
	cachedConn.ExecCtx(ctx, func(conn *gorm.DB) error {
		return conn.Model(&TestUser{}).Where("id = ?", 1).Update("name", "John Doe").Error
	}, key)

	fmt.Printf("User: %s\n", result.Name)
}

