// Package integration 基于项目真实数据库（MySQL）的报名并发与回滚测试。
//
// 运行方式（数据库名必须含 "test"，测试会清空并删除该库）：
//
//	GBEVENT_TEST_MYSQL_DSN="root:gbevent_root@tcp(127.0.0.1:57506)/gbevent_testdb?charset=utf8mb4&parseTime=True&loc=Local" \
//	  go test ./internal/integration/ -v
//
// 未设置环境变量时使用上述默认 DSN（对应 docker-compose 的端口与 root 密码）；
// MySQL 不可达时全部用例跳过，不影响其他环境的 go test ./...。
package integration

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"

	"gbevent/internal/model"

	"github.com/go-sql-driver/mysql"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// 默认 DSN 对应 docker-compose 暴露的端口（57506）与 root 密码（gbevent_root）。
const defaultDSN = "root:gbevent_root@tcp(127.0.0.1:57506)/gbevent_testdb?charset=utf8mb4&parseTime=True&loc=Local"

var (
	testDB    *gorm.DB
	testDBCfg *mysql.Config
)

// allTables 全部业务表，测试间通过 TRUNCATE 隔离（同时重置自增，保证连续运行结果稳定）。
var allTables = []string{
	"check_in_records", "notifications", "registrations", "activities",
	"comments", "favorites", "audit_logs", "users",
}

func TestMain(m *testing.M) {
	dsn := os.Getenv("GBEVENT_TEST_MYSQL_DSN")
	if dsn == "" {
		dsn = defaultDSN
	}
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		fmt.Println("SKIP: invalid GBEVENT_TEST_MYSQL_DSN:", err)
		os.Exit(0)
	}
	// 安全护栏：本套件会清空并删除目标库，仅允许名称含 test 的数据库。
	if !strings.Contains(strings.ToLower(cfg.DBName), "test") {
		fmt.Printf("SKIP: database %q must contain \"test\" in its name\n", cfg.DBName)
		os.Exit(0)
	}
	testDBCfg = cfg

	// 先连到服务器（不指定库）创建测试库。
	serverCfg := *cfg
	serverCfg.DBName = ""
	admin, err := gorm.Open(gormmysql.Open(serverCfg.FormatDSN()), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		fmt.Println("SKIP: mysql unreachable:", err)
		os.Exit(0)
	}
	if err := admin.Exec("CREATE DATABASE IF NOT EXISTS `" + cfg.DBName + "` CHARACTER SET utf8mb4").Error; err != nil {
		fmt.Println("SKIP: cannot create test database:", err)
		os.Exit(0)
	}

	db, err := gorm.Open(gormmysql.Open(cfg.FormatDSN()), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		fmt.Println("SKIP: mysql open failed:", err)
		os.Exit(0)
	}
	sqlDB, err := db.DB()
	if err != nil {
		fmt.Println("SKIP: mysql handle failed:", err)
		os.Exit(0)
	}
	if err := sqlDB.Ping(); err != nil {
		fmt.Println("SKIP: mysql ping failed:", err)
		os.Exit(0)
	}
	sqlDB.SetMaxOpenConns(32)
	sqlDB.SetMaxIdleConns(8)

	if err := db.AutoMigrate(
		&model.User{}, &model.Activity{}, &model.Registration{}, &model.CheckInRecord{},
		&model.Comment{}, &model.Favorite{}, &model.Notification{}, &model.AuditLog{},
	); err != nil {
		fmt.Println("SKIP: migrate failed:", err)
		os.Exit(0)
	}
	testDB = db

	code := m.Run()

	// 结束后清理：删除整个测试库。
	if err := admin.Exec("DROP DATABASE IF EXISTS `" + cfg.DBName + "`").Error; err != nil {
		fmt.Println("cleanup: drop test database failed:", err)
	}
	os.Exit(code)
}

// requireDB 在 MySQL 不可用时跳过当前用例。
func requireDB(t *testing.T) *gorm.DB {
	t.Helper()
	if testDB == nil {
		t.Skip("mysql not available, skipping integration test")
	}
	return testDB
}

// resetTables 清空全部业务表并重置自增，保证用例间数据隔离、连续运行结果一致。
func resetTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	for _, table := range allTables {
		if err := db.Exec("TRUNCATE TABLE " + table).Error; err != nil {
			t.Fatalf("truncate %s: %v", table, err)
		}
	}
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
