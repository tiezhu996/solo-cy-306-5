package service

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"gbevent/internal/constants"
	"gbevent/internal/repository"

	"gorm.io/gorm"
)

// TestSignupGuardCheckRegistrationLimit 名额与截止时间校验矩阵：
// 活动状态、截止时间与容量判断的错误码、文案必须与重构前逐项一致。
// 计数规则：未审核/已通过/已签到占用名额，已取消与审核被拒不占用。
func TestSignupGuardCheckRegistrationLimit(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)
	past := time.Now().Add(-time.Hour)

	type regSeed struct{ status, reviewStatus string }

	cases := []struct {
		name        string
		status      string
		capacity    int
		deadline    time.Time
		seed        []regSeed // 预置报名
		wantCode    int
		wantMessage string // 含 %d 时为活动 ID 占位
	}{
		{"draft not published", constants.ActivityStatusDraft, 0, future, nil,
			constants.CodeConflict, "Activity[id=%d] not published"},
		{"ended activity", constants.ActivityStatusEnded, 0, future, nil,
			constants.CodeActivityEnded, constants.MsgActivityEnded},
		{"deadline passed", constants.ActivityStatusPublished, 0, past, nil,
			constants.CodeActivityEnded, "Activity[id=%d] signup deadline passed"},
		{"capacity zero unlimited", constants.ActivityStatusPublished, 0, future,
			[]regSeed{{constants.RegistrationStatusRegistered, constants.ReviewStatusPending},
				{constants.RegistrationStatusRegistered, constants.ReviewStatusPending}}, 0, ""},
		{"capacity reached", constants.ActivityStatusPublished, 2, future,
			[]regSeed{{constants.RegistrationStatusRegistered, constants.ReviewStatusPending},
				{constants.RegistrationStatusCheckedIn, constants.ReviewStatusApproved}},
			constants.CodeActivityFull, constants.MsgActivityFull},
		{"capacity not reached", constants.ActivityStatusPublished, 2, future,
			[]regSeed{{constants.RegistrationStatusRegistered, constants.ReviewStatusPending}}, 0, ""},
		{"pending counts", constants.ActivityStatusPublished, 1, future,
			[]regSeed{{constants.RegistrationStatusRegistered, constants.ReviewStatusPending}},
			constants.CodeActivityFull, constants.MsgActivityFull},
		{"approved counts", constants.ActivityStatusPublished, 1, future,
			[]regSeed{{constants.RegistrationStatusRegistered, constants.ReviewStatusApproved}},
			constants.CodeActivityFull, constants.MsgActivityFull},
		{"cancelled not counted", constants.ActivityStatusPublished, 1, future,
			[]regSeed{{constants.RegistrationStatusCancelled, constants.ReviewStatusPending}}, 0, ""},
		{"checked in counts", constants.ActivityStatusPublished, 1, future,
			[]regSeed{{constants.RegistrationStatusCheckedIn, constants.ReviewStatusApproved}},
			constants.CodeActivityFull, constants.MsgActivityFull},
		{"rejected not counted", constants.ActivityStatusPublished, 1, future,
			[]regSeed{{constants.RegistrationStatusRegistered, constants.ReviewStatusRejected}}, 0, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := newTestStack(t)
			a := s.mustActivity(t, 5, tc.status, tc.capacity, tc.deadline)
			for i, seed := range tc.seed {
				s.mustRegistration(t, a.ID, uint64(100+i), seed.status, seed.reviewStatus)
			}

			err := s.guard.CheckRegistrationLimit(a.ID)
			if tc.wantCode == 0 {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
				return
			}
			want := tc.wantMessage
			if strings.Contains(want, "%d") {
				want = fmt.Sprintf(want, a.ID)
			}
			requireAppError(t, err, tc.wantCode, want)
		})
	}
}

// TestSignupGuardCheckRegistrationLimitTx 事务内校验与非事务版本结果一致。
func TestSignupGuardCheckRegistrationLimitTx(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)
	s := newTestStack(t)
	a := s.mustActivity(t, 5, constants.ActivityStatusPublished, 1, future)
	s.mustRegistration(t, a.ID, 100, constants.RegistrationStatusRegistered, constants.ReviewStatusPending)

	err := s.db.Transaction(func(tx *gorm.DB) error {
		return s.guard.CheckRegistrationLimitTx(tx, a.ID)
	})
	requireAppError(t, err, constants.CodeActivityFull, constants.MsgActivityFull)

	b := s.mustActivity(t, 5, constants.ActivityStatusPublished, 2, future)
	err = s.db.Transaction(func(tx *gorm.DB) error {
		return s.guard.CheckRegistrationLimitTx(tx, b.ID)
	})
	if err != nil {
		t.Fatalf("expected nil error in tx, got %v", err)
	}
}

// TestSignupGuardActivityNotFound 活动不存在时透传仓储层 ErrNotFound。
func TestSignupGuardActivityNotFound(t *testing.T) {
	s := newTestStack(t)
	err := s.guard.CheckRegistrationLimit(999)
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected wrapped ErrNotFound, got %v", err)
	}
}
