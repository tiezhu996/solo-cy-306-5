package constants

// NotificationType 消息通知类型枚举。
const (
	NotificationSignupSuccess  = "signup_success"
	NotificationReminder       = "reminder"
	NotificationReviewResult   = "review_result"
	NotificationCheckinSuccess = "checkin_success"
)

// NotificationTypeValues 全部通知类型值。
var NotificationTypeValues = []string{NotificationSignupSuccess, NotificationReminder, NotificationReviewResult, NotificationCheckinSuccess}

// UserRole 用户角色枚举。
const (
	RoleUser      = "user"
	RoleOrganizer = "organizer"
	RoleAdmin     = "admin"
)

// UserRoleValues 全部角色值。
var UserRoleValues = []string{RoleUser, RoleOrganizer, RoleAdmin}
