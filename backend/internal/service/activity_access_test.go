package service

import (
	"testing"

	"gbevent/internal/constants"
	"gbevent/internal/model"
)

// TestAccessPolicyCanManage 管理员与组织者的权限判定矩阵，
// 并与既有 IsOrganizer 函数逐项保持一致。
func TestAccessPolicyCanManage(t *testing.T) {
	policy := NewActivityAccessPolicy()
	cases := []struct {
		name         string
		operatorID   uint64
		operatorRole string
		organizerID  uint64
		want         bool
	}{
		{"admin manages anyone", 1, constants.RoleAdmin, 99, true},
		{"admin manages self", 1, constants.RoleAdmin, 1, true},
		{"organizer owner", 5, constants.RoleOrganizer, 5, true},
		{"organizer not owner", 5, constants.RoleOrganizer, 6, false},
		{"user owner", 3, constants.RoleUser, 3, true},
		{"user not owner", 3, constants.RoleUser, 5, false},
	}
	for _, tc := range cases {
		if got := policy.CanManage(tc.operatorID, tc.operatorRole, tc.organizerID); got != tc.want {
			t.Errorf("%s: CanManage = %v, want %v", tc.name, got, tc.want)
		}
		// 与包级 IsOrganizer（既有调用方仍在用）结果必须一致。
		if got := IsOrganizer(tc.operatorID, tc.operatorRole, tc.organizerID); got != tc.want {
			t.Errorf("%s: IsOrganizer = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// TestAccessPolicyRequireManager 校验各操作的拒绝错误码与文案逐字一致。
func TestAccessPolicyRequireManager(t *testing.T) {
	policy := NewActivityAccessPolicy()
	a := &model.Activity{ID: 7, OrganizerID: 5}

	// 放行：组织者本人与管理员。
	if err := policy.RequireManager(a, 5, constants.RoleOrganizer, "update"); err != nil {
		t.Errorf("owner should be allowed: %v", err)
	}
	if err := policy.RequireManager(a, 1, constants.RoleAdmin, "update"); err != nil {
		t.Errorf("admin should be allowed: %v", err)
	}

	// 拒绝：非本人非管理员，文案按操作名拼接。
	actions := []string{"update", "publish", "end", "delete", "stats"}
	for _, action := range actions {
		err := policy.RequireManager(a, 6, constants.RoleOrganizer, action)
		requireAppError(t, err, constants.CodeForbidden,
			"Activity[id=7] "+action+" forbidden: organizer not match")
	}
}
