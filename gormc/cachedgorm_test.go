package gormc

import (
	"context"
	"testing"
	"time"

	"github.com/zeromicro/go-zero/core/mathx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type mysqlcfg struct {
	Path         string // 服务器地址
	Port         string `json:",default=3306"`                                               // 端口
	Config       string `json:",default=charset%3Dutf8mb4%26parseTime%3Dtrue%26loc%3DLocal"` // 高级配置
	Dbname       string // 数据库名
	Username     string // 数据库用户名
	Password     string // 数据库密码
	MaxIdleConns int    `json:",default=10"` // 空闲中的最大连接数
	MaxOpenConns int    `json:",default=10"` // 打开到数据库的最大连接数
	LogMode      string `json:",default="`   // 是否开启Gorm全局日志
	LogZap       bool   // 是否通过zap写入日志文件
}

func (m *mysqlcfg) Dsn() string {
	return m.Username + ":" + m.Password + "@tcp(" + m.Path + ":" + m.Port + ")/" + m.Dbname + "?" + m.Config
}

// TestGormc_QueryWithExpire 需要真实的 MySQL 数据库
// 确保 MySQL 在 localhost:3306 运行
// 可以通过环境变量设置连接信息：
// MYSQL_USER (默认: root), MYSQL_PASSWORD (默认: ""), MYSQL_DB (默认: test)
func TestGormc_QueryWithExpire(t *testing.T) {
	// 从环境变量读取配置，或使用默认值
	username := "root"
	password := "root123456"
	dbname := "test"

	cfg := mysqlcfg{
		Path:     "localhost",
		Port:     "3306",
		Config:   "charset=utf8mb4&parseTime=true&loc=Local",
		Dbname:   dbname,
		Username: username,
		Password: password,
	}
	mcg := mysql.Config{
		DSN: cfg.Dsn(),
	}
	db, err := gorm.Open(mysql.New(mcg))
	if err != nil {
		t.Skipf("跳过测试：无法连接到 MySQL: %v\n提示：请修改测试中的 username、password、dbname 配置", err)
		return
	}
	t.Logf("成功连接到 MySQL: %s@%s:%s/%s", username, cfg.Path, cfg.Port, dbname)
	redisConf := RedisConfig{
		Addr:         "127.0.0.1:6379",
		Password:     "",
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 2,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}
	gormc, err := NewConn(db, redisConf, time.Second*5)
	if err != nil {
		t.Skipf("跳过测试：无法连接到 Redis: %v\n提示：请确保 Redis 在 127.0.0.1:6379 运行", err)
		return
	}
	t.Logf("成功连接到 Redis: %s DB:%d", redisConf.Addr, redisConf.DB)
	var str string
	err = gormc.QueryWithExpireCtx(context.Background(), &str, "any", time.Second*5, func(conn *gorm.DB) error {
		t.Logf("conn: %+v", conn)
		return nil
	})
	if err != nil {
		t.Error(err)
		return
	}

	err = gormc.QueryWithCallbackExpireCtx(context.Background(), &str, "any", func(conn *gorm.DB) error {
		return nil
	}, func(i interface{}) time.Duration {
		return time.Second * 5
	})

}

func TestUnstable(t *testing.T) {

	unstable := mathx.NewUnstable(0.1)
	t.Logf("unstable: %v", unstable.AroundDuration(5*time.Minute))
	t.Logf("unstable: %v", unstable.AroundDuration(5*time.Minute))
	t.Logf("unstable: %v", unstable.AroundDuration(5*time.Minute))
	t.Logf("unstable: %v", unstable.AroundDuration(5*time.Minute))
	t.Logf("unstable: %v", unstable.AroundDuration(5*time.Minute))

}
