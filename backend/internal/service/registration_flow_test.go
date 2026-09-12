package service

import (
	"strings"
	"testing"
	"time"

	"gbevent/internal/constants"
	"gbevent/internal/model"
)

// TestRegistrationCreateWritesNotification 报名成功：报名记录与通知同事务写入，
// 通知类型/标题/文案逐字核对。
func TestRegistrationCreateWritesNotification(t *testing.T) {
	s := newTestStack(t)
	a := s.mustActivity(t, 5, constants.ActivityStatusPublished, 10, time.Now().Add(24*time.Hour))

	reg, err := s.reg.Create(a.ID, 100, "张三", "13800000000", "备注")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if reg.VoucherNo == "" {
		t.Error("voucher_no should be generated")
	}
	if reg.Status != constants.RegistrationStatusRegistered {
		t.Errorf("status = %q, want registered", reg.Status)
	}
	if reg.ReviewStatus != constants.ReviewStatusPending {
		t.Errorf("review_status = %q, want pending", reg.ReviewStatus)
	}

	notifs := s.notifications(t, 100)
	if len(notifs) != 1 {
		t.Fatalf("notification count = %d, want 1", len(notifs))
	}
	n := notifs[0]
	if n.NotificationType != constants.NotificationSignupSuccess {
		t.Errorf("type = %q, want signup_success", n.NotificationType)
	}
	if n.Title != "报名成功" {
		t.Errorf("title = %q, want 报名成功", n.Title)
	}
	if n.Content != "您已成功报名活动，凭证号："+reg.VoucherNo {
		t.Errorf("content = %q, want 您已成功报名活动，凭证号：%s", n.Content, reg.VoucherNo)
	}
	if n.IsRead {
		t.Error("is_read should be false")
	}
}

// TestRegistrationDuplicateRollback 重复报名返回 40903，
// 且事务回滚：不产生第二条报名，也不产生第二条通知。
func TestRegistrationDuplicateRollback(t *testing.T) {
	s := newTestStack(t)
	a := s.mustActivity(t, 5, constants.ActivityStatusPublished, 10, time.Now().Add(24*time.Hour))

	if _, err := s.reg.Create(a.ID, 100, "张三", "13800000000", ""); err != nil {
		t.Fatalf("first create failed: %v", err)
	}
	_, err := s.reg.Create(a.ID, 100, "张三", "13800000000", "")
	requireAppError(t, err, constants.CodeDuplicateSignup, constants.MsgDuplicateSignup)

	if got := s.registrationCount(t, a.ID); got != 1 {
		t.Errorf("registration count = %d, want 1 (rolled back)", got)
	}
	if got := len(s.notifications(t, 100)); got != 1 {
		t.Errorf("notification count = %d, want 1 (rolled back)", got)
	}
}

// TestRegistrationFullAndEnded 名额满与活动结束/截止的报名拒绝，
// 且不落任何数据。
func TestRegistrationFullAndEnded(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)

	t.Run("activity full", func(t *testing.T) {
		s := newTestStack(t)
		a := s.mustActivity(t, 5, constants.ActivityStatusPublished, 1, future)
		if _, err := s.reg.Create(a.ID, 100, "张三", "13800000000", ""); err != nil {
			t.Fatalf("first create failed: %v", err)
		}
		_, err := s.reg.Create(a.ID, 101, "李四", "13900000000", "")
		requireAppError(t, err, constants.CodeActivityFull, constants.MsgActivityFull)
		if got := s.registrationCount(t, a.ID); got != 1 {
			t.Errorf("registration count = %d, want 1", got)
		}
		if got := len(s.notifications(t, 101)); got != 0 {
			t.Errorf("notification count = %d, want 0", got)
		}
	})

	t.Run("activity ended", func(t *testing.T) {
		s := newTestStack(t)
		a := s.mustActivity(t, 5, constants.ActivityStatusEnded, 10, future)
		_, err := s.reg.Create(a.ID, 100, "张三", "13800000000", "")
		requireAppError(t, err, constants.CodeActivityEnded, constants.MsgActivityEnded)
	})

	t.Run("deadline passed", func(t *testing.T) {
		s := newTestStack(t)
		a := s.mustActivity(t, 5, constants.ActivityStatusPublished, 10, time.Now().Add(-time.Hour))
		_, err := s.reg.Create(a.ID, 100, "张三", "13800000000", "")
		requireAppError(t, err, constants.CodeActivityEnded,
			"Activity[id="+itoa(a.ID)+"] signup deadline passed")
	})

	t.Run("activity draft", func(t *testing.T) {
		s := newTestStack(t)
		a := s.mustActivity(t, 5, constants.ActivityStatusDraft, 10, future)
		_, err := s.reg.Create(a.ID, 100, "张三", "13800000000", "")
		requireAppError(t, err, constants.CodeConflict,
			"Activity[id="+itoa(a.ID)+"] not published")
	})
}

// TestRegistrationReviewFlow 审核的权限、状态机与通知文案。
func TestRegistrationReviewFlow(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)

	t.Run("approve writes notification", func(t *testing.T) {
		s := newTestStack(t)
		a := s.mustActivity(t, 5, constants.ActivityStatusPublished, 10, future)
		reg, err := s.reg.Create(a.ID, 100, "张三", "13800000000", "")
		if err != nil {
			t.Fatalf("create failed: %v", err)
		}
		got, err := s.reg.Review(reg.ID, 5, constants.RoleOrganizer, constants.ReviewStatusApproved)
		if err != nil {
			t.Fatalf("review failed: %v", err)
		}
		if got.ReviewStatus != constants.ReviewStatusApproved {
			t.Errorf("review_status = %q, want approved", got.ReviewStatus)
		}
		notifs := s.notifications(t, 100)
		if len(notifs) != 2 {
			t.Fatalf("notification count = %d, want 2", len(notifs))
		}
		n := notifs[1]
		if n.NotificationType != constants.NotificationReviewResult {
			t.Errorf("type = %q, want review_result", n.NotificationType)
		}
		if n.Title != "审核结果" {
			t.Errorf("title = %q, want 审核结果", n.Title)
		}
		if n.Content != "您的报名已通过审核，凭证号："+reg.VoucherNo {
			t.Errorf("content = %q", n.Content)
		}
	})

	t.Run("reject writes notification", func(t *testing.T) {
		s := newTestStack(t)
		a := s.mustActivity(t, 5, constants.ActivityStatusPublished, 10, future)
		reg, _ := s.reg.Create(a.ID, 100, "张三", "13800000000", "")
		if _, err := s.reg.Review(reg.ID, 5, constants.RoleOrganizer, constants.ReviewStatusRejected); err != nil {
			t.Fatalf("review failed: %v", err)
		}
		notifs := s.notifications(t, 100)
		if len(notifs) != 2 {
			t.Fatalf("notification count = %d, want 2", len(notifs))
		}
		// 文案为既有行为逐字保留："您的报名审核已" + "已拒绝"。
		if notifs[1].Content != "您的报名审核已已拒绝" {
			t.Errorf("content = %q, want 您的报名审核已已拒绝", notifs[1].Content)
		}
	})

	t.Run("other organizer forbidden", func(t *testing.T) {
		s := newTestStack(t)
		a := s.mustActivity(t, 5, constants.ActivityStatusPublished, 10, future)
		reg, _ := s.reg.Create(a.ID, 100, "张三", "13800000000", "")
		_, err := s.reg.Review(reg.ID, 6, constants.RoleOrganizer, constants.ReviewStatusApproved)
		requireAppError(t, err, constants.CodeForbidden,
			"Registration[id="+itoa(reg.ID)+"] review forbidden: organizer not match")
		// 审核状态与通知均未变化（事务回滚）。
		cur, _ := s.reg.Get(reg.ID)
		if cur.ReviewStatus != constants.ReviewStatusPending {
			t.Errorf("review_status changed to %q after forbidden review", cur.ReviewStatus)
		}
		if got := len(s.notifications(t, 100)); got != 1 {
			t.Errorf("notification count = %d, want 1", got)
		}
	})

	t.Run("admin can review", func(t *testing.T) {
		s := newTestStack(t)
		a := s.mustActivity(t, 5, constants.ActivityStatusPublished, 10, future)
		reg, _ := s.reg.Create(a.ID, 100, "张三", "13800000000", "")
		if _, err := s.reg.Review(reg.ID, 1, constants.RoleAdmin, constants.ReviewStatusApproved); err != nil {
			t.Fatalf("admin review failed: %v", err)
		}
	})

	t.Run("re-review conflicts and rolls back", func(t *testing.T) {
		s := newTestStack(t)
		a := s.mustActivity(t, 5, constants.ActivityStatusPublished, 10, future)
		reg, _ := s.reg.Create(a.ID, 100, "张三", "13800000000", "")
		if _, err := s.reg.Review(reg.ID, 5, constants.RoleOrganizer, constants.ReviewStatusApproved); err != nil {
			t.Fatalf("first review failed: %v", err)
		}
		_, err := s.reg.Review(reg.ID, 5, constants.RoleOrganizer, constants.ReviewStatusRejected)
		requireAppError(t, err, constants.CodeReviewConflict, constants.MsgReviewConflict)
		cur, _ := s.reg.Get(reg.ID)
		if cur.ReviewStatus != constants.ReviewStatusApproved {
			t.Errorf("review_status = %q, want approved (unchanged)", cur.ReviewStatus)
		}
		if got := len(s.notifications(t, 100)); got != 2 {
			t.Errorf("notification count = %d, want 2", got)
		}
	})

	t.Run("invalid review status", func(t *testing.T) {
		s := newTestStack(t)
		a := s.mustActivity(t, 5, constants.ActivityStatusPublished, 10, future)
		reg, _ := s.reg.Create(a.ID, 100, "张三", "13800000000", "")
		_, err := s.reg.Review(reg.ID, 5, constants.RoleOrganizer, "bogus")
		requireAppError(t, err, constants.CodeValidationFailed,
			"Registration[id="+itoa(reg.ID)+"] review invalid status=bogus")
	})
}

// TestRegistrationCancelFlow 取消报名的权限与状态机。
func TestRegistrationCancelFlow(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)

	t.Run("owner cancels", func(t *testing.T) {
		s := newTestStack(t)
		a := s.mustActivity(t, 5, constants.ActivityStatusPublished, 10, future)
		reg, _ := s.reg.Create(a.ID, 100, "张三", "13800000000", "")
		got, err := s.reg.Cancel(reg.ID, 100, constants.RoleUser)
		if err != nil {
			t.Fatalf("cancel failed: %v", err)
		}
		if got.Status != constants.RegistrationStatusCancelled {
			t.Errorf("status = %q, want cancelled", got.Status)
		}
	})

	t.Run("cancel twice conflicts", func(t *testing.T) {
		s := newTestStack(t)
		a := s.mustActivity(t, 5, constants.ActivityStatusPublished, 10, future)
		reg, _ := s.reg.Create(a.ID, 100, "张三", "13800000000", "")
		if _, err := s.reg.Cancel(reg.ID, 100, constants.RoleUser); err != nil {
			t.Fatalf("first cancel failed: %v", err)
		}
		_, err := s.reg.Cancel(reg.ID, 100, constants.RoleUser)
		requireAppError(t, err, constants.CodeCancelConflict, constants.MsgCancelConflict)
	})

	t.Run("other user forbidden", func(t *testing.T) {
		s := newTestStack(t)
		a := s.mustActivity(t, 5, constants.ActivityStatusPublished, 10, future)
		reg, _ := s.reg.Create(a.ID, 100, "张三", "13800000000", "")
		_, err := s.reg.Cancel(reg.ID, 101, constants.RoleUser)
		requireAppError(t, err, constants.CodeForbidden,
			"Registration[id="+itoa(reg.ID)+"] cancel forbidden: not owner")
	})

	t.Run("admin cancels others", func(t *testing.T) {
		s := newTestStack(t)
		a := s.mustActivity(t, 5, constants.ActivityStatusPublished, 10, future)
		reg, _ := s.reg.Create(a.ID, 100, "张三", "13800000000", "")
		if _, err := s.reg.Cancel(reg.ID, 1, constants.RoleAdmin); err != nil {
			t.Fatalf("admin cancel failed: %v", err)
		}
	})
}

// TestRegistrationOfflineCreate 线下补录：直接通过审核、受名额约束、不做防重。
func TestRegistrationOfflineCreate(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)
	s := newTestStack(t)
	a := s.mustActivity(t, 5, constants.ActivityStatusPublished, 1, future)

	reg, err := s.reg.OfflineCreate(a.ID, 5, "现场用户", "13700000000", "")
	if err != nil {
		t.Fatalf("offline create failed: %v", err)
	}
	if reg.ReviewStatus != constants.ReviewStatusApproved {
		t.Errorf("review_status = %q, want approved", reg.ReviewStatus)
	}
	if reg.UserID != 5 {
		t.Errorf("user_id = %d, want operator 5", reg.UserID)
	}

	// 名额满后补录同样被拒绝。
	_, err = s.reg.OfflineCreate(a.ID, 5, "现场用户2", "13700000001", "")
	requireAppError(t, err, constants.CodeActivityFull, constants.MsgActivityFull)
}

// TestCheckInFlow 签到状态机、通知写入与重复签到回滚。
func TestCheckInFlow(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)

	t.Run("voucher checkin writes notification", func(t *testing.T) {
		s := newTestStack(t)
		a := s.mustActivity(t, 5, constants.ActivityStatusPublished, 10, future)
		reg, _ := s.reg.Create(a.ID, 100, "张三", "13800000000", "")

		rec, err := s.checkin.CheckInByVoucher(a.ID, 5, reg.VoucherNo)
		if err != nil {
			t.Fatalf("checkin failed: %v", err)
		}
		if rec.CheckInMethod != constants.CheckInMethodVoucher {
			t.Errorf("method = %q, want voucher", rec.CheckInMethod)
		}
		cur, _ := s.reg.Get(reg.ID)
		if cur.Status != constants.RegistrationStatusCheckedIn {
			t.Errorf("status = %q, want checked_in", cur.Status)
		}
		notifs := s.notifications(t, 100)
		if len(notifs) != 2 {
			t.Fatalf("notification count = %d, want 2", len(notifs))
		}
		n := notifs[1]
		if n.NotificationType != constants.NotificationCheckinSuccess {
			t.Errorf("type = %q, want checkin_success", n.NotificationType)
		}
		if n.Title != "签到成功" || n.Content != "您已完成签到，凭证号："+reg.VoucherNo {
			t.Errorf("notification = %q / %q", n.Title, n.Content)
		}
	})

	t.Run("double checkin conflicts and rolls back", func(t *testing.T) {
		s := newTestStack(t)
		a := s.mustActivity(t, 5, constants.ActivityStatusPublished, 10, future)
		reg, _ := s.reg.Create(a.ID, 100, "张三", "13800000000", "")
		if _, err := s.checkin.CheckInByVoucher(a.ID, 5, reg.VoucherNo); err != nil {
			t.Fatalf("first checkin failed: %v", err)
		}
		_, err := s.checkin.CheckInByVoucher(a.ID, 5, reg.VoucherNo)
		requireAppError(t, err, constants.CodeAlreadyCheckedIn, constants.MsgAlreadyCheckedIn)
		// 只有一条签到记录与一条签到通知。
		var cnt int64
		if err := s.db.Model(&model.CheckInRecord{}).Where("activity_id = ?", a.ID).Count(&cnt).Error; err != nil {
			t.Fatalf("count checkins: %v", err)
		}
		if cnt != 1 {
			t.Errorf("checkin record count = %d, want 1", cnt)
		}
		if got := len(s.notifications(t, 100)); got != 2 {
			t.Errorf("notification count = %d, want 2", got)
		}
	})

	t.Run("cancelled registration cannot checkin", func(t *testing.T) {
		s := newTestStack(t)
		a := s.mustActivity(t, 5, constants.ActivityStatusPublished, 10, future)
		reg, _ := s.reg.Create(a.ID, 100, "张三", "13800000000", "")
		if _, err := s.reg.Cancel(reg.ID, 100, constants.RoleUser); err != nil {
			t.Fatalf("cancel failed: %v", err)
		}
		_, err := s.checkin.CheckInByVoucher(a.ID, 5, reg.VoucherNo)
		requireAppError(t, err, constants.CodeCancelConflict,
			"Registration[id="+itoa(reg.ID)+"] cannot checkin: status=cancelled")
	})

	t.Run("voucher of other activity rejected", func(t *testing.T) {
		s := newTestStack(t)
		a := s.mustActivity(t, 5, constants.ActivityStatusPublished, 10, future)
		b := s.mustActivity(t, 5, constants.ActivityStatusPublished, 10, future)
		reg, _ := s.reg.Create(a.ID, 100, "张三", "13800000000", "")
		_, err := s.checkin.CheckInByVoucher(b.ID, 5, reg.VoucherNo)
		requireAppError(t, err, constants.CodeInvalidVoucher,
			"CheckInRecord[activity_id="+itoa(b.ID)+"] voucher not match this activity")
	})

	t.Run("scan checkin with registration id", func(t *testing.T) {
		s := newTestStack(t)
		a := s.mustActivity(t, 5, constants.ActivityStatusPublished, 10, future)
		reg, _ := s.reg.Create(a.ID, 100, "张三", "13800000000", "")
		rec, err := s.checkin.CheckInByScan(a.ID, 5, itoa(reg.ID))
		if err != nil {
			t.Fatalf("scan checkin failed: %v", err)
		}
		if rec.CheckInMethod != constants.CheckInMethodScan {
			t.Errorf("method = %q, want scan", rec.CheckInMethod)
		}
	})
}

// TestExportCSVPermissions 导出名单的权限与内容。
func TestExportCSVPermissions(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)
	s := newTestStack(t)
	a := s.mustActivity(t, 5, constants.ActivityStatusPublished, 10, future)
	reg, _ := s.reg.Create(a.ID, 100, "张三", "13800000000", "")

	filename, content, err := s.reg.ExportCSV(a.ID, 5, constants.RoleOrganizer)
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}
	if !strings.HasPrefix(filename, "registrations_") || !strings.HasSuffix(filename, ".csv") {
		t.Errorf("filename = %q", filename)
	}
	if !strings.Contains(content, "ID,活动ID,报名人") {
		t.Errorf("csv header missing: %q", content)
	}
	if !strings.Contains(content, reg.VoucherNo) {
		t.Errorf("csv should contain voucher %s", reg.VoucherNo)
	}

	if _, _, err := s.reg.ExportCSV(a.ID, 1, constants.RoleAdmin); err != nil {
		t.Errorf("admin export failed: %v", err)
	}
	_, _, err = s.reg.ExportCSV(a.ID, 6, constants.RoleOrganizer)
	requireAppError(t, err, constants.CodeForbidden,
		"Registration[activity_id="+itoa(a.ID)+"] export forbidden: organizer not match")
}
