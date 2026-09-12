// 报名状态枚举（与后端 backend/internal/constants/registration.go 保持一致）
export const RegistrationStatus = {
  REGISTERED: 'registered',
  CANCELLED: 'cancelled',
  CHECKED_IN: 'checked_in',
} as const

export const RegistrationStatusText: Record<string, string> = {
  [RegistrationStatus.REGISTERED]: '已报名',
  [RegistrationStatus.CANCELLED]: '已取消',
  [RegistrationStatus.CHECKED_IN]: '已签到',
}

export const ReviewStatus = {
  PENDING: 'pending',
  APPROVED: 'approved',
  REJECTED: 'rejected',
} as const

export const ReviewStatusText: Record<string, string> = {
  [ReviewStatus.PENDING]: '待审核',
  [ReviewStatus.APPROVED]: '已通过',
  [ReviewStatus.REJECTED]: '已拒绝',
}

export const CheckInMethodText: Record<string, string> = {
  voucher: '凭证签到',
  scan: '扫码签到',
}
