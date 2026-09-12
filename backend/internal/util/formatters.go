package util

import (
	"fmt"
	"time"

	"gbevent/internal/constants"
)

// formatters.go 同时提供日期格式化、名额格式化、状态文本、活动类型文本、报名状态文本，
// 多个 handler/service 直接引用本文件，修改时牵一发动全身。

// FormatDateTime 格式化日期时间。
func FormatDateTime(t time.Time) string {
	return t.Format("2006-01-02 15:04")
}

// FormatDate 格式化日期。
func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// FormatCapacity 格式化名额文本，如 "120/200"。
func FormatCapacity(registered, capacity int) string {
	return fmt.Sprintf("%d/%d", registered, capacity)
}

// ActivityStatusText 活动状态文本。
func ActivityStatusText(status string) string {
	switch status {
	case constants.ActivityStatusDraft:
		return "草稿"
	case constants.ActivityStatusPublished:
		return "已发布"
	case constants.ActivityStatusEnded:
		return "已结束"
	default:
		return status
	}
}

// ActivityTypeText 活动类型文本。
func ActivityTypeText(t string) string {
	switch t {
	case constants.ActivityTypeLecture:
		return "讲座"
	case constants.ActivityTypeTraining:
		return "培训"
	case constants.ActivityTypeParty:
		return "聚会"
	case constants.ActivityTypeCompetition:
		return "比赛"
	default:
		return t
	}
}

// RegistrationStatusText 报名状态文本。
func RegistrationStatusText(status string) string {
	switch status {
	case constants.RegistrationStatusRegistered:
		return "已报名"
	case constants.RegistrationStatusCancelled:
		return "已取消"
	case constants.RegistrationStatusCheckedIn:
		return "已签到"
	default:
		return status
	}
}

// ReviewStatusText 审核状态文本。
func ReviewStatusText(status string) string {
	switch status {
	case constants.ReviewStatusPending:
		return "待审核"
	case constants.ReviewStatusApproved:
		return "已通过"
	case constants.ReviewStatusRejected:
		return "已拒绝"
	default:
		return status
	}
}

// NotificationTypeText 通知类型文本。
func NotificationTypeText(t string) string {
	switch t {
	case constants.NotificationSignupSuccess:
		return "报名成功"
	case constants.NotificationReminder:
		return "活动提醒"
	case constants.NotificationReviewResult:
		return "审核结果"
	case constants.NotificationCheckinSuccess:
		return "签到成功"
	default:
		return t
	}
}

// CheckInMethodText 签到方式文本。
func CheckInMethodText(m string) string {
	if m == constants.CheckInMethodScan {
		return "扫码签到"
	}
	return "凭证签到"
}
