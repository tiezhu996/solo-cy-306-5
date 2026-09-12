package service

import (
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"gbevent/internal/model"
	"gbevent/internal/repository"
	"gbevent/internal/util"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	gormlogger "gorm.io/gorm/logger"
)

// newTestDB 打开独立的 SQLite 内存库并迁移全部模型。
// SQLite 不支持 FOR UPDATE 行锁，这里通过 ClauseBuilders 覆盖剥离 LOCKING 子句；
// 仅影响测试环境，生产 MySQL 路径不变。
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := "file:" + strings.NewReplacer("/", "_", " ", "_").Replace(t.Name()) + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
		ClauseBuilders: map[string]clause.ClauseBuilder{
			"LOCKING": func(c clause.Clause, builder clause.Builder) {},
		},
	})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(
		&model.User{}, &model.Activity{}, &model.Registration{}, &model.CheckInRecord{},
		&model.Comment{}, &model.Favorite{}, &model.Notification{}, &model.AuditLog{},
	); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	return db
}

// testStack 按 main.go 的装配方式构建整套服务，用于回归测试。
type testStack struct {
	db       *gorm.DB
	activity *ActivityService
	reg      *RegistrationService
	checkin  *CheckInRecordService
	notify   *NotificationService
	guard    *SignupGuard
	access   *ActivityAccessPolicy
}

func newTestStack(t *testing.T) *testStack {
	t.Helper()
	db := newTestDB(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	activityRepo := repository.NewActivityRepository(db)
	regRepo := repository.NewRegistrationRepository(db)
	notifyRepo := repository.NewNotificationRepository(db)
	checkinRepo := repository.NewCheckInRecordRepository(db)

	notifySvc := NewNotificationService(notifyRepo, logger)
	access := NewActivityAccessPolicy()
	guard := NewSignupGuard(activityRepo)
	activitySvc := NewActivityService(activityRepo, checkinRepo, access, logger)
	regSvc := NewRegistrationService(db, regRepo, activitySvc, guard, notifySvc, logger)
	checkinSvc := NewCheckInRecordService(db, checkinRepo, regRepo, notifySvc, logger)

	return &testStack{db: db, activity: activitySvc, reg: regSvc, checkin: checkinSvc, notify: notifySvc, guard: guard, access: access}
}

// mustActivity 直接落库一个活动 fixture（绕过 service 校验，可构造任意状态）。
func (s *testStack) mustActivity(t *testing.T, organizerID uint64, status string, capacity int, signupDeadline time.Time) *model.Activity {
	t.Helper()
	a := &model.Activity{
		Title:          "测试活动",
		ActivityType:   "lecture",
		StartTime:      time.Now().Add(24 * time.Hour),
		EndTime:        time.Now().Add(26 * time.Hour),
		Capacity:       capacity,
		SignupDeadline: signupDeadline,
		Status:         status,
		OrganizerID:    organizerID,
	}
	if err := s.db.Create(a).Error; err != nil {
		t.Fatalf("create activity fixture: %v", err)
	}
	return a
}

// mustRegistration 直接落库一个报名 fixture。
func (s *testStack) mustRegistration(t *testing.T, activityID, userID uint64, status, reviewStatus string) *model.Registration {
	t.Helper()
	reg := &model.Registration{
		ActivityID:   activityID,
		UserID:       userID,
		Name:         "张三",
		Phone:        "13800000000",
		VoucherNo:    util.GenerateVoucherNo(),
		Status:       status,
		ReviewStatus: reviewStatus,
	}
	if err := s.db.Create(reg).Error; err != nil {
		t.Fatalf("create registration fixture: %v", err)
	}
	return reg
}

// notifications 返回某用户的全部通知（按 id 升序）。
func (s *testStack) notifications(t *testing.T, userID uint64) []model.Notification {
	t.Helper()
	var list []model.Notification
	if err := s.db.Where("user_id = ?", userID).Order("id ASC").Find(&list).Error; err != nil {
		t.Fatalf("list notifications: %v", err)
	}
	return list
}

// registrationCount 返回某活动的报名行数（含已取消）。
func (s *testStack) registrationCount(t *testing.T, activityID uint64) int64 {
	t.Helper()
	var n int64
	if err := s.db.Model(&model.Registration{}).Where("activity_id = ?", activityID).Count(&n).Error; err != nil {
		t.Fatalf("count registrations: %v", err)
	}
	return n
}

// requireAppError 断言错误为指定错误码且文案与期望完全一致的 AppError。
func requireAppError(t *testing.T, err error, code int, wantMessage string) {
	t.Helper()
	var appErr *util.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %T: %v", err, err)
	}
	if appErr.Code != code {
		t.Errorf("error code = %d, want %d (message=%q)", appErr.Code, code, appErr.Message)
	}
	if wantMessage != "" && appErr.Message != wantMessage {
		t.Errorf("error message = %q, want %q", appErr.Message, wantMessage)
	}
}
