// Package integration 基于项目真实数据库（MySQL）的报名并发与回滚测试。
//
// 本包是测试门禁：默认运行时 MySQL 环境缺失（不可达、库名不含 test、迁移失败等）
// 每条用例都会以 FAIL 结束并给出明确提示——绿色构建 ⇔ 并发与回滚用例真实执行过。
// 显式跳过请使用 go test -short（输出可见的 SKIP 记录）。
//
// 运行方式（数据库名必须含 "test"，测试会清空并删除该库）：
//
//	GBEVENT_TEST_MYSQL_DSN="root:gbevent_root@tcp(127.0.0.1:57506)/gbevent_testdb?charset=utf8mb4&parseTime=True&loc=Local" \
//	  go test ./internal/integration/ -v
//
// 未设置环境变量时使用上述默认 DSN（对应 docker-compose 的端口与 root 密码）。
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
	// setupErr 记录环境准备失败原因；非 nil 时每条用例以 FAIL 结束（测试门禁）。
	setupErr error
)

// allTables 全部业务表，测试间通过 TRUNCATE 隔离（同时重置自增，保证连续运行结果稳定）。
var allTables = []string{
	"check_in_records", "notifications", "registrations", "activities",
	"comments", "favorites", "audit_logs", "users",
}

func TestMain(m *testing.M) {
	admin, db := setupMySQL()
	if db != nil {
		testDB = db
	}
	code := m.Run()
	// 结束后清理：删除整个测试库（仅在环境就绪时）。
	if admin != nil && testDBCfg != nil {
		if err := admin.Exec("DROP DATABASE IF EXISTS `" + testDBCfg.DBName + "`").Error; err != nil {
			fmt.Println("cleanup: drop test database failed:", err)
		}
	}
	os.Exit(code)
}

// setupMySQL 连接真实 MySQL 并准备测试库；失败时把原因写入 setupErr，由用例统一 FAIL。
func setupMySQL() (admin *gorm.DB, db *gorm.DB) {
	dsn := os.Getenv("GBEVENT_TEST_MYSQL_DSN")
	if dsn == "" {
		dsn = defaultDSN
	}
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		setupErr = fmt.Errorf("解析 GBEVENT_TEST_MYSQL_DSN 失败: %w", err)
		return nil, nil
	}
	// 安全护栏：本套件会清空并删除目标库，仅允许名称含 test 的数据库。
	if !strings.Contains(strings.ToLower(cfg.DBName), "test") {
		setupErr = fmt.Errorf("数据库名 %q 必须包含 \"test\"（本套件会清空并删除该库）", cfg.DBName)
		return nil, nil
	}
	testDBCfg = cfg

	silent := &gorm.Config{Logger: gormlogger.Default.LogMode(gormlogger.Silent)}
	// 先连到服务器（不指定库）创建测试库。
	serverCfg := *cfg
	serverCfg.DBName = ""
	admin, err = gorm.Open(gormmysql.Open(serverCfg.FormatDSN()), silent)
	if err != nil {
		setupErr = fmt.Errorf("连接 MySQL 失败: %w", err)
		return nil, nil
	}
	if err := admin.Exec("CREATE DATABASE IF NOT EXISTS `" + cfg.DBName + "` CHARACTER SET utf8mb4").Error; err != nil {
		setupErr = fmt.Errorf("创建测试库 %s 失败: %w", cfg.DBName, err)
		return nil, nil
	}

	db, err = gorm.Open(gormmysql.Open(cfg.FormatDSN()), silent)
	if err != nil {
		setupErr = fmt.Errorf("打开测试库连接失败: %w", err)
		return nil, nil
	}
	sqlDB, err := db.DB()
	if err != nil {
		setupErr = fmt.Errorf("获取连接句柄失败: %w", err)
		return nil, nil
	}
	if err := sqlDB.Ping(); err != nil {
		setupErr = fmt.Errorf("MySQL 不可达: %w", err)
		return nil, nil
	}
	sqlDB.SetMaxOpenConns(32)
	sqlDB.SetMaxIdleConns(8)

	if err := db.AutoMigrate(
		&model.User{}, &model.Activity{}, &model.Registration{}, &model.CheckInRecord{},
		&model.Comment{}, &model.Favorite{}, &model.Notification{}, &model.AuditLog{},
	); err != nil {
		setupErr = fmt.Errorf("迁移测试库失败: %w", err)
		return nil, nil
	}
	return admin, db
}

// requireDB 测试门禁：环境未就绪时以 FAIL 结束并给出明确提示，
// 保证"全部通过"一定意味着并发与回滚用例真实执行过。
// 显式跳过请使用 go test -short（输出可见的 SKIP）。
func requireDB(t *testing.T) *gorm.DB {
	t.Helper()
	if testing.Short() {
		t.Skip("short 模式：跳过真实数据库集成测试")
	}
	if setupErr != nil {
		t.Fatalf("真实数据库集成测试未执行（测试门禁）：环境不可用。\n"+
			"原因: %v\n"+
			"请启动项目 MySQL（docker compose up -d db），或设置 GBEVENT_TEST_MYSQL_DSN 指向测试库"+
			"（库名须含 test，例如 %s）。详见 README「本地开发-测试」一节。", setupErr, defaultDSN)
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
