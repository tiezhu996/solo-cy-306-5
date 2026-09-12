package service

import (
	"errors"
	"testing"
	"time"

	"gbevent/internal/constants"
	"gbevent/internal/model"
	"gbevent/internal/repository"
)

// TestActivityPublishFlow 发布动作的状态机与权限矩阵。
func TestActivityPublishFlow(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)

	t.Run("draft to published by owner", func(t *testing.T) {
		s := newTestStack(t)
		a := s.mustActivity(t, 5, constants.ActivityStatusDraft, 0, future)
		got, err := s.activity.Publish(a.ID, 5, constants.RoleOrganizer)
		if err != nil {
			t.Fatalf("publish failed: %v", err)
		}
		if got.Status != constants.ActivityStatusPublished {
			t.Errorf("status = %q, want published", got.Status)
		}
	})

	t.Run("publish again conflicts", func(t *testing.T) {
		s := newTestStack(t)
		a := s.mustActivity(t, 5, constants.ActivityStatusPublished, 0, future)
		_, err := s.activity.Publish(a.ID, 5, constants.RoleOrganizer)
		requireAppError(t, err, constants.CodeConflict,
			"Activity[id="+itoa(a.ID)+"] publish conflict: status=published")
	})

	t.Run("publish ended conflicts", func(t *testing.T) {
		s := newTestStack(t)
		a := s.mustActivity(t, 5, constants.ActivityStatusEnded, 0, future)
		_, err := s.activity.Publish(a.ID, 5, constants.RoleOrganizer)
		requireAppError(t, err, constants.CodeConflict,
			"Activity[id="+itoa(a.ID)+"] publish conflict: status=ended")
	})

	t.Run("other organizer forbidden", func(t *testing.T) {
		s := newTestStack(t)
		a := s.mustActivity(t, 5, constants.ActivityStatusDraft, 0, future)
		_, err := s.activity.Publish(a.ID, 6, constants.RoleOrganizer)
		requireAppError(t, err, constants.CodeForbidden,
			"Activity[id="+itoa(a.ID)+"] publish forbidden: organizer not match")
		// 状态未被改变。
		cur, _, _ := s.activity.Get(a.ID)
		if cur.Status != constants.ActivityStatusDraft {
			t.Errorf("status changed to %q after forbidden publish", cur.Status)
		}
	})

	t.Run("admin can publish others", func(t *testing.T) {
		s := newTestStack(t)
		a := s.mustActivity(t, 5, constants.ActivityStatusDraft, 0, future)
		got, err := s.activity.Publish(a.ID, 1, constants.RoleAdmin)
		if err != nil {
			t.Fatalf("admin publish failed: %v", err)
		}
		if got.Status != constants.ActivityStatusPublished {
			t.Errorf("status = %q, want published", got.Status)
		}
	})

	t.Run("publish missing activity", func(t *testing.T) {
		s := newTestStack(t)
		_, err := s.activity.Publish(999, 1, constants.RoleAdmin)
		if !errors.Is(err, repository.ErrNotFound) {
			t.Fatalf("expected wrapped ErrNotFound, got %v", err)
		}
	})
}

// TestActivityEndFlow 结束动作的状态机与权限矩阵。
func TestActivityEndFlow(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)

	t.Run("published to ended by owner", func(t *testing.T) {
		s := newTestStack(t)
		a := s.mustActivity(t, 5, constants.ActivityStatusPublished, 0, future)
		got, err := s.activity.End(a.ID, 5, constants.RoleOrganizer)
		if err != nil {
			t.Fatalf("end failed: %v", err)
		}
		if got.Status != constants.ActivityStatusEnded {
			t.Errorf("status = %q, want ended", got.Status)
		}
	})

	t.Run("end draft conflicts", func(t *testing.T) {
		s := newTestStack(t)
		a := s.mustActivity(t, 5, constants.ActivityStatusDraft, 0, future)
		_, err := s.activity.End(a.ID, 5, constants.RoleOrganizer)
		requireAppError(t, err, constants.CodeConflict,
			"Activity[id="+itoa(a.ID)+"] end conflict: status=draft")
	})

	t.Run("other organizer forbidden", func(t *testing.T) {
		s := newTestStack(t)
		a := s.mustActivity(t, 5, constants.ActivityStatusPublished, 0, future)
		_, err := s.activity.End(a.ID, 6, constants.RoleOrganizer)
		requireAppError(t, err, constants.CodeForbidden,
			"Activity[id="+itoa(a.ID)+"] end forbidden: organizer not match")
	})

	t.Run("admin can end others", func(t *testing.T) {
		s := newTestStack(t)
		a := s.mustActivity(t, 5, constants.ActivityStatusPublished, 0, future)
		if _, err := s.activity.End(a.ID, 1, constants.RoleAdmin); err != nil {
			t.Fatalf("admin end failed: %v", err)
		}
	})
}

// TestActivityUpdateFlow 更新动作的权限与字段校验。
func TestActivityUpdateFlow(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)

	t.Run("owner updates fields", func(t *testing.T) {
		s := newTestStack(t)
		a := s.mustActivity(t, 5, constants.ActivityStatusDraft, 10, future)
		got, err := s.activity.Update(a.ID, 5, constants.RoleOrganizer, map[string]any{
			"title": "新标题", "capacity": 20, "location": "线上",
		})
		if err != nil {
			t.Fatalf("update failed: %v", err)
		}
		if got.Title != "新标题" || got.Capacity != 20 || got.Location != "线上" {
			t.Errorf("fields not applied: %+v", got)
		}
	})

	t.Run("other organizer forbidden", func(t *testing.T) {
		s := newTestStack(t)
		a := s.mustActivity(t, 5, constants.ActivityStatusDraft, 10, future)
		_, err := s.activity.Update(a.ID, 6, constants.RoleOrganizer, map[string]any{"title": "越权"})
		requireAppError(t, err, constants.CodeForbidden,
			"Activity[id="+itoa(a.ID)+"] update forbidden: organizer not match")
	})

	t.Run("admin updates others", func(t *testing.T) {
		s := newTestStack(t)
		a := s.mustActivity(t, 5, constants.ActivityStatusDraft, 10, future)
		if _, err := s.activity.Update(a.ID, 1, constants.RoleAdmin, map[string]any{"title": "管理员改"}); err != nil {
			t.Fatalf("admin update failed: %v", err)
		}
	})

	t.Run("invalid activity type", func(t *testing.T) {
		s := newTestStack(t)
		a := s.mustActivity(t, 5, constants.ActivityStatusDraft, 10, future)
		_, err := s.activity.Update(a.ID, 5, constants.RoleOrganizer, map[string]any{"activity_type": "bogus"})
		requireAppError(t, err, constants.CodeValidationFailed,
			"Activity[id="+itoa(a.ID)+"] update: invalid activity_type=bogus")
	})
}

// TestActivityDeleteFlow 删除动作的权限矩阵。
func TestActivityDeleteFlow(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)

	t.Run("other organizer forbidden", func(t *testing.T) {
		s := newTestStack(t)
		a := s.mustActivity(t, 5, constants.ActivityStatusDraft, 0, future)
		err := s.activity.Delete(a.ID, 6, constants.RoleOrganizer)
		requireAppError(t, err, constants.CodeForbidden,
			"Activity[id="+itoa(a.ID)+"] delete forbidden: organizer not match")
	})

	t.Run("owner deletes", func(t *testing.T) {
		s := newTestStack(t)
		a := s.mustActivity(t, 5, constants.ActivityStatusDraft, 0, future)
		if err := s.activity.Delete(a.ID, 5, constants.RoleOrganizer); err != nil {
			t.Fatalf("delete failed: %v", err)
		}
		if _, _, err := s.activity.Get(a.ID); !errors.Is(err, repository.ErrNotFound) {
			t.Errorf("expected not found after delete, got %v", err)
		}
	})

	t.Run("admin deletes others", func(t *testing.T) {
		s := newTestStack(t)
		a := s.mustActivity(t, 5, constants.ActivityStatusDraft, 0, future)
		if err := s.activity.Delete(a.ID, 1, constants.RoleAdmin); err != nil {
			t.Fatalf("admin delete failed: %v", err)
		}
	})
}

// TestActivityStats 统计字段与权限。
func TestActivityStats(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)
	s := newTestStack(t)
	a := s.mustActivity(t, 5, constants.ActivityStatusPublished, 10, future)
	// 2 个有效报名（1 已签到）+ 1 个已取消（不计入）。
	s.mustRegistration(t, a.ID, 100, constants.RegistrationStatusRegistered, constants.ReviewStatusApproved)
	checked := s.mustRegistration(t, a.ID, 101, constants.RegistrationStatusCheckedIn, constants.ReviewStatusApproved)
	s.mustRegistration(t, a.ID, 102, constants.RegistrationStatusCancelled, constants.ReviewStatusPending)
	if err := s.db.Create(&model.CheckInRecord{
		RegistrationID: checked.ID, ActivityID: a.ID,
		CheckInMethod: constants.CheckInMethodVoucher, CheckInTime: time.Now(), OperatorID: 5,
	}).Error; err != nil {
		t.Fatalf("create checkin fixture: %v", err)
	}

	stats, err := s.activity.Stats(a.ID, 5, constants.RoleOrganizer)
	if err != nil {
		t.Fatalf("stats failed: %v", err)
	}
	if stats["activity_id"] != a.ID {
		t.Errorf("activity_id = %v, want %d", stats["activity_id"], a.ID)
	}
	if stats["capacity"] != 10 {
		t.Errorf("capacity = %v, want 10", stats["capacity"])
	}
	if stats["registered_count"] != int64(2) {
		t.Errorf("registered_count = %v, want 2", stats["registered_count"])
	}
	if stats["checked_in_count"] != int64(1) {
		t.Errorf("checked_in_count = %v, want 1", stats["checked_in_count"])
	}
	if stats["checkin_rate"] != 50.0 {
		t.Errorf("checkin_rate = %v, want 50", stats["checkin_rate"])
	}

	// 管理员可看，其他组织者拒绝。
	if _, err := s.activity.Stats(a.ID, 1, constants.RoleAdmin); err != nil {
		t.Errorf("admin stats failed: %v", err)
	}
	_, err = s.activity.Stats(a.ID, 6, constants.RoleOrganizer)
	requireAppError(t, err, constants.CodeForbidden,
		"Activity[id="+itoa(a.ID)+"] stats forbidden: organizer not match")
}

// TestActivityGetRegisteredCount 详情接口的报名人数不计已取消。
func TestActivityGetRegisteredCount(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)
	s := newTestStack(t)
	a := s.mustActivity(t, 5, constants.ActivityStatusPublished, 10, future)
	s.mustRegistration(t, a.ID, 100, constants.RegistrationStatusRegistered, constants.ReviewStatusPending)
	s.mustRegistration(t, a.ID, 101, constants.RegistrationStatusCancelled, constants.ReviewStatusPending)

	_, count, err := s.activity.Get(a.ID)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if count != 1 {
		t.Errorf("registered count = %d, want 1", count)
	}
}
