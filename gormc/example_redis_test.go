package gormc_test

import (
	"fmt"
	"time"

	"github.com/huof6829/gorm-zero/gormc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// User represents a user model
type User struct {
	ID   int64  `gorm:"primaryKey"`
	Name string `gorm:"column:name"`
}

func ExampleNewConn() {
	// Initialize database connection
	dsn := "user:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	// Configure Redis with DB selection
	redisConf := gormc.RedisConfig{
		Addr:         "127.0.0.1:6379",
		Password:     "",              // no password
		DB:           0,               // use database 0
		PoolSize:     10,              // connection pool size
		MinIdleConns: 2,               // minimum idle connections
		DialTimeout:  5 * time.Second, // dial timeout
		ReadTimeout:  3 * time.Second, // read timeout
		WriteTimeout: 3 * time.Second, // write timeout
	}

	// Create cached connection with 1 hour expiry
	cachedConn, err := gormc.NewConn(db, redisConf, time.Hour)
	if err != nil {
		panic(err)
	}

	// Use the cached connection
	var user User
	key := fmt.Sprintf("user:%d", 1)
	err = cachedConn.QueryCtx(nil, &user, key, func(conn *gorm.DB) error {
		return conn.Where("id = ?", 1).First(&user).Error
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("User: %+v\n", user)
}

// ExampleNewConn_multipleDB shows how to use different Redis databases
func ExampleNewConn_multipleDB() {
	dsn := "user:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	// Use Redis DB 0 for user cache
	userRedisConf := gormc.RedisConfig{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0, // database 0
	}
	userCache, err := gormc.NewConn(db, userRedisConf, time.Hour)
	if err != nil {
		panic(err)
	}

	// Use Redis DB 1 for order cache
	orderRedisConf := gormc.RedisConfig{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       1, // database 1
	}
	orderCache, err := gormc.NewConn(db, orderRedisConf, 30*time.Minute)
	if err != nil {
		panic(err)
	}

	_, _ = userCache, orderCache
	// Now you can use userCache and orderCache with different Redis databases
}
