// 通知类型枚举（与后端 backend/internal/constants/notification.go 保持一致）
export const NotificationType = {
  SIGNUP_SUCCESS: 'signup_success',
  REMINDER: 'reminder',
  REVIEW_RESULT: 'review_result',
  CHECKIN_SUCCESS: 'checkin_success',
} as const

export const NotificationTypeText: Record<string, string> = {
  [NotificationType.SIGNUP_SUCCESS]: '报名成功',
  [NotificationType.REMINDER]: '活动提醒',
  [NotificationType.REVIEW_RESULT]: '审核结果',
  [NotificationType.CHECKIN_SUCCESS]: '签到成功',
}

export const UserRole = {
  USER: 'user',
  ORGANIZER: 'organizer',
  ADMIN: 'admin',
} as const

export const UserRoleText: Record<string, string> = {
  [UserRole.USER]: '普通用户',
  [UserRole.ORGANIZER]: '组织者',
  [UserRole.ADMIN]: '管理员',
}
