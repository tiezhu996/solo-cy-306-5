package service

import (
	"gbevent/internal/constants"
	"gbevent/internal/model"
	"gbevent/internal/util"
)

// ActivityAccessPolicy 活动访问控制：集中承载"组织者本人或管理员"的判定，
// 供活动流程（更新/发布/结束/删除/统计）与报名审核、导出等场景复用。
type ActivityAccessPolicy struct{}

// NewActivityAccessPolicy 构造活动访问控制策略。
func NewActivityAccessPolicy() *ActivityAccessPolicy {
	return &ActivityAccessPolicy{}
}

// CanManage 判断操作者是否可管理该组织者的活动（组织者本人或管理员）。
func (p *ActivityAccessPolicy) CanManage(operatorID uint64, operatorRole string, organizerID uint64) bool {
	return operatorRole == constants.RoleAdmin || operatorID == organizerID
}

// RequireManager 校验操作者是否可管理该活动，不可管理时返回统一 403 业务错误。
// action 为操作名（update/publish/end/delete/stats），用于拼接错误文案。
func (p *ActivityAccessPolicy) RequireManager(a *model.Activity, operatorID uint64, operatorRole, action string) error {
	if p.CanManage(operatorID, operatorRole, a.OrganizerID) {
		return nil
	}
	return util.NewAppError(constants.CodeForbidden,
		"Activity[id="+itoa(a.ID)+"] "+action+" forbidden: organizer not match")
}

// IsOrganizer 判断操作者是否活动组织者或管理员。
func IsOrganizer(operatorID uint64, operatorRole string, organizerID uint64) bool {
	return NewActivityAccessPolicy().CanManage(operatorID, operatorRole, organizerID)
}
