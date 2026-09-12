export const ErrorCode = {
  OK: 0,
  BAD_REQUEST: 40000,
  UNAUTHORIZED: 40100,
  FORBIDDEN: 40300,
  NOT_FOUND: 40400,
  CONFLICT: 40900,
  TOO_MANY_REQUESTS: 42900,
  VALIDATION_FAILED: 42200,
  INTERNAL_ERROR: 50000,
  ACTIVITY_FULL: 40901,
  DUPLICATE_SIGNUP: 40903,
  ALREADY_CHECKED_IN: 40904,
} as const

export const ErrorMessage: Record<number, string> = {
  [ErrorCode.UNAUTHORIZED]: '未登录或登录已过期',
  [ErrorCode.FORBIDDEN]: '没有权限执行该操作',
  [ErrorCode.NOT_FOUND]: '资源不存在',
  [ErrorCode.TOO_MANY_REQUESTS]: '请求过于频繁，请稍后再试',
  [ErrorCode.ACTIVITY_FULL]: '活动名额已满',
  [ErrorCode.DUPLICATE_SIGNUP]: '您已报名过该活动',
  [ErrorCode.ALREADY_CHECKED_IN]: '该报名已签到',
}
