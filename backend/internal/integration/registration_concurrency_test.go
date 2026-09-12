package integration

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"gbevent/internal/constants"
	"gbevent/internal/model"
	"gbevent/internal/repository"
	"gbevent/internal/service"
	"gbevent/internal/util"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

// stack 按 cmd/server/main.go 的装配方式构建服务（真实 MySQL 连接）。
type stack struct {
	db       *gorm.DB
	activity *service.ActivityService
	reg      *service.RegistrationService
	checkin  *service.CheckInRecordService
}

func newStack(db *gorm.DB) *stack {
	logger := testLogger()
	activityRepo := repository.NewActivityRepository(db)
	regRepo := repository.NewRegistrationRepository(db)
	notifyRepo := repository.NewNotificationRepository(db)
	checkinRepo := repository.NewCheckInRecordRepository(db)

	notifySvc := service.NewNotificationService(notifyRepo, logger)
	access := service.NewActivityAccessPolicy()
	guard := service.NewSignupGuard(activityRepo)
	activitySvc := service.NewActivityService(activityRepo, checkinRepo, access, logger)
	regSvc := service.NewRegistrationService(db, regRepo, activitySvc, guard, notifySvc, logger)
	checkinSvc := service.NewCheckInRecordService(db, checkinRepo, regRepo, notifySvc, logger)
	return &stack{db: db, activity: activitySvc, reg: regSvc, checkin: checkinSvc}
}

func (s *stack) mustActivity(t *testing.T, organizerID uint64, status string, capacity int) *model.Activity {
	t.Helper()
	a := &model.Activity{
		Title:          "并发测试活动",
		ActivityType:   "lecture",
		StartTime:      time.Now().Add(48 * time.Hour),
		EndTime:        time.Now().Add(50 * time.Hour),
		Capacity:       capacity,
		SignupDeadline: time.Now().Add(24 * time.Hour),
		Status:         status,
		OrganizerID:    organizerID,
	}
	if err := s.db.Create(a).Error; err != nil {
		t.Fatalf("create activity fixture: %v", err)
	}
	return a
}

// signupWithVoucherRetry 报名；凭证号为 GB+日期+4 位随机数，高并发下可能撞号
// （voucher_no 唯一索引，MySQL 1062）。撞号属于既有已知缺陷，换号重试，
// 不影响本套件对名额/防重/回滚语义的断言。
func signupWithVoucherRetry(svc *service.RegistrationService, activityID, userID uint64, name, phone string) (*model.Registration, error) {
	var lastErr error
	for i := 0; i < 5; i++ {
		reg, err := svc.Create(activityID, userID, name, phone, "")
		if err == nil {
			return reg, nil
		}
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			lastErr = err
			continue
		}
		return nil, err
	}
	return nil, fmt.Errorf("voucher collision persists after retries: %w", lastErr)
}

// runConcurrent 并发执行 n 个任务（统一放行闸门制造真实竞争），收集全部结果。
func runConcurrent[T any](n int, fn func(i int) T) []T {
	start := make(chan struct{})
	results := make(chan T, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			results <- fn(i)
		}(i)
	}
	close(start)
	wg.Wait()
	close(results)
	out := make([]T, 0, n)
	for r := range results {
		out = append(out, r)
	}
	return out
}

type signupOutcome struct {
	userID uint64
	reg    *model.Registration
	err    error
}

func countRows(t *testing.T, db *gorm.DB, table, where string, args ...any) int64 {
	t.Helper()
	var n int64
	if err := db.Table(table).Where(where, args...).Count(&n).Error; err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return n
}

func requireAppErrorCode(t *testing.T, err error, wantCode int) {
	t.Helper()
	var appErr *util.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %T: %v", err, err)
	}
	if appErr.Code != wantCode {
		t.Fatalf("error code = %d, want %d (message=%q)", appErr.Code, wantCode, appErr.Message)
	}
}

// TestConcurrentSignupNeverExceedsCapacity 20 个用户并发抢 5 个名额：
// 恰好 5 人成功，其余全部 40901；失败请求不留报名或通知。
func TestConcurrentSignupNeverExceedsCapacity(t *testing.T) {
	db := requireDB(t)
	resetTables(t, db)
	s := newStack(db)
	const capacity, workers = 5, 20
	a := s.mustActivity(t, 5, constants.ActivityStatusPublished, capacity)

	outcomes := runConcurrent(workers, func(i int) signupOutcome {
		uid := uint64(1000 + i)
		reg, err := signupWithVoucherRetry(s.reg, a.ID, uid, "用户", "13800000000")
		return signupOutcome{userID: uid, reg: reg, err: err}
	})

	var successes []signupOutcome
	vouchers := map[string]bool{}
	for _, o := range outcomes {
		if o.err == nil {
			if o.reg == nil || o.reg.VoucherNo == "" {
				t.Fatalf("user %d succeeded without voucher", o.userID)
			}
			if vouchers[o.reg.VoucherNo] {
				t.Fatalf("duplicate voucher issued: %s", o.reg.VoucherNo)
			}
			vouchers[o.reg.VoucherNo] = true
			successes = append(successes, o)
			continue
		}
		requireAppErrorCode(t, o.err, constants.CodeActivityFull)
	}
	if len(successes) != capacity {
		t.Fatalf("successes = %d, want %d (capacity must not be exceeded)", len(successes), capacity)
	}

	// 终态一致：报名行数 == 名额，通知行数 == 成功数，无孤儿数据。
	if got := countRows(t, db, "registrations", "activity_id = ?", a.ID); got != capacity {
		t.Errorf("registrations = %d, want %d", got, capacity)
	}
	if got := countRows(t, db, "notifications", "notification_type = ?", constants.NotificationSignupSuccess); got != capacity {
		t.Errorf("signup notifications = %d, want %d", got, capacity)
	}
	for _, o := range successes {
		if got := countRows(t, db, "notifications", "user_id = ?", o.userID); got != 1 {
			t.Errorf("user %d notifications = %d, want 1", o.userID, got)
		}
	}
	// 失败用户不留任何数据。
	for _, o := range outcomes {
		if o.err != nil {
			if got := countRows(t, db, "registrations", "activity_id = ? AND user_id = ?", a.ID, o.userID); got != 0 {
				t.Errorf("failed user %d left %d registrations", o.userID, got)
			}
			if got := countRows(t, db, "notifications", "user_id = ?", o.userID); got != 0 {
				t.Errorf("failed user %d left %d notifications", o.userID, got)
			}
		}
	}
}

// TestConcurrentSignupSameUserOnce 同一用户并发报名：恰好 1 次成功，其余 40903。
func TestConcurrentSignupSameUserOnce(t *testing.T) {
	db := requireDB(t)
	resetTables(t, db)
	s := newStack(db)
	const workers = 10
	a := s.mustActivity(t, 5, constants.ActivityStatusPublished, 10)

	outcomes := runConcurrent(workers, func(i int) signupOutcome {
		reg, err := signupWithVoucherRetry(s.reg, a.ID, 1000, "张三", "13800000000")
		return signupOutcome{userID: 1000, reg: reg, err: err}
	})

	success := 0
	for _, o := range outcomes {
		if o.err == nil {
			success++
			continue
		}
		requireAppErrorCode(t, o.err, constants.CodeDuplicateSignup)
	}
	if success != 1 {
		t.Fatalf("successes = %d, want 1 (same user must not register twice)", success)
	}
	if got := countRows(t, db, "registrations", "activity_id = ? AND user_id = ?", a.ID, 1000); got != 1 {
		t.Errorf("registrations = %d, want 1", got)
	}
	if got := countRows(t, db, "notifications", "user_id = ?", 1000); got != 1 {
		t.Errorf("notifications = %d, want 1", got)
	}
}

// TestConcurrentSignupAfterRejectFreesSlot 名额为 1：唯一报名被拒后空位释放，
// 10 人并发抢该空位，恰好 1 人成功。
func TestConcurrentSignupAfterRejectFreesSlot(t *testing.T) {
	db := requireDB(t)
	resetTables(t, db)
	s := newStack(db)
	const workers = 10
	a := s.mustActivity(t, 5, constants.ActivityStatusPublished, 1)

	reg, err := s.reg.Create(a.ID, 1000, "张三", "13800000000", "")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if _, err := s.reg.Review(reg.ID, 5, constants.RoleOrganizer, constants.ReviewStatusRejected); err != nil {
		t.Fatalf("reject failed: %v", err)
	}

	outcomes := runConcurrent(workers, func(i int) signupOutcome {
		uid := uint64(2000 + i)
		r, err := signupWithVoucherRetry(s.reg, a.ID, uid, "用户", "13900000000")
		return signupOutcome{userID: uid, reg: r, err: err}
	})

	success := 0
	for _, o := range outcomes {
		if o.err == nil {
			success++
			continue
		}
		requireAppErrorCode(t, o.err, constants.CodeActivityFull)
	}
	if success != 1 {
		t.Fatalf("successes = %d, want 1 (exactly one freed slot)", success)
	}
	// 被拒记录仍在，但有效报名（未取消且未被拒）只有胜出的 1 条。
	if got := countRows(t, db, "registrations", "activity_id = ?", a.ID); got != 2 {
		t.Errorf("total registrations = %d, want 2 (rejected + winner)", got)
	}
	if got := countRows(t, db, "registrations", "activity_id = ? AND status <> 'cancelled' AND review_status <> 'rejected'", a.ID); got != 1 {
		t.Errorf("valid registrations = %d, want 1", got)
	}
	// 通知：被拒用户 2 条（报名+审核结果），胜出者 1 条，失败者 0 条。
	if got := countRows(t, db, "notifications", "user_id = ?", 1000); got != 2 {
		t.Errorf("rejected user notifications = %d, want 2", got)
	}
	if got := countRows(t, db, "notifications", "notification_type = ?", constants.NotificationSignupSuccess); got != 2 {
		t.Errorf("signup notifications = %d, want 2", got)
	}
}

// TestConcurrentCheckInOnlyOnce 同一凭证并发签到：恰好 1 次成功，其余 40904，
// 签到记录与签到通知均不重复。
func TestConcurrentCheckInOnlyOnce(t *testing.T) {
	db := requireDB(t)
	resetTables(t, db)
	s := newStack(db)
	const workers = 10
	a := s.mustActivity(t, 5, constants.ActivityStatusPublished, 10)

	reg, err := s.reg.Create(a.ID, 1000, "张三", "13800000000", "")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if _, err := s.reg.Review(reg.ID, 5, constants.RoleOrganizer, constants.ReviewStatusApproved); err != nil {
		t.Fatalf("approve failed: %v", err)
	}

	type checkinOutcome struct{ err error }
	outcomes := runConcurrent(workers, func(i int) checkinOutcome {
		_, err := s.checkin.CheckInByVoucher(a.ID, 5, reg.VoucherNo)
		return checkinOutcome{err: err}
	})

	success := 0
	for _, o := range outcomes {
		if o.err == nil {
			success++
			continue
		}
		requireAppErrorCode(t, o.err, constants.CodeAlreadyCheckedIn)
	}
	if success != 1 {
		t.Fatalf("checkin successes = %d, want 1", success)
	}
	if got := countRows(t, db, "check_in_records", "activity_id = ?", a.ID); got != 1 {
		t.Errorf("checkin records = %d, want 1", got)
	}
	// 报名成功 + 审核结果 + 签到成功，各一条。
	if got := countRows(t, db, "notifications", "user_id = ?", 1000); got != 3 {
		t.Errorf("notifications = %d, want 3", got)
	}
	cur, err := s.reg.Get(reg.ID)
	if err != nil {
		t.Fatalf("get registration: %v", err)
	}
	if cur.Status != constants.RegistrationStatusCheckedIn {
		t.Errorf("status = %q, want checked_in", cur.Status)
	}
}

// TestOrderingAndRollbackOnRealDB 真实库上的状态顺序与回滚：
// 未审核不能签到、已签到不能拒绝、重复审核不新增数据。
func TestOrderingAndRollbackOnRealDB(t *testing.T) {
	db := requireDB(t)
	resetTables(t, db)
	s := newStack(db)
	a := s.mustActivity(t, 5, constants.ActivityStatusPublished, 10)

	reg, err := s.reg.Create(a.ID, 1000, "张三", "13800000000", "")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// 未审核签到：拒绝且事务回滚，不留签到记录与通知。
	if _, err := s.checkin.CheckInByVoucher(a.ID, 5, reg.VoucherNo); err == nil {
		t.Fatal("pending registration must not check in")
	} else {
		requireAppErrorCode(t, err, constants.CodeReviewConflict)
	}
	if got := countRows(t, db, "check_in_records", "activity_id = ?", a.ID); got != 0 {
		t.Errorf("checkin records = %d, want 0", got)
	}
	if got := countRows(t, db, "notifications", "user_id = ?", 1000); got != 1 {
		t.Errorf("notifications = %d, want 1 (signup only)", got)
	}

	// 审核通过后签到成功。
	if _, err := s.reg.Review(reg.ID, 5, constants.RoleOrganizer, constants.ReviewStatusApproved); err != nil {
		t.Fatalf("approve failed: %v", err)
	}
	if _, err := s.checkin.CheckInByVoucher(a.ID, 5, reg.VoucherNo); err != nil {
		t.Fatalf("checkin after approve failed: %v", err)
	}

	// 已签到不能再拒绝：重复审核原错误，状态与通知不变。
	if _, err := s.reg.Review(reg.ID, 5, constants.RoleOrganizer, constants.ReviewStatusRejected); err == nil {
		t.Fatal("checked-in registration must not be rejected")
	} else {
		requireAppErrorCode(t, err, constants.CodeReviewConflict)
	}
	cur, err := s.reg.Get(reg.ID)
	if err != nil {
		t.Fatalf("get registration: %v", err)
	}
	if cur.ReviewStatus != constants.ReviewStatusApproved || cur.Status != constants.RegistrationStatusCheckedIn {
		t.Errorf("registration = (%q, %q), want (approved, checked_in)", cur.ReviewStatus, cur.Status)
	}
	if got := countRows(t, db, "notifications", "user_id = ?", 1000); got != 3 {
		t.Errorf("notifications = %d, want 3 (no new rows from rejected review)", got)
	}
}
