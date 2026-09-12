// 活动状态与类型枚举（与后端 backend/internal/constants/activity.go 保持一致）
export const ActivityStatus = {
  DRAFT: 'draft',
  PUBLISHED: 'published',
  ENDED: 'ended',
} as const

export type ActivityStatusType = (typeof ActivityStatus)[keyof typeof ActivityStatus]

export const ActivityStatusText: Record<string, string> = {
  [ActivityStatus.DRAFT]: '草稿',
  [ActivityStatus.PUBLISHED]: '已发布',
  [ActivityStatus.ENDED]: '已结束',
}

export const ActivityType = {
  LECTURE: 'lecture',
  TRAINING: 'training',
  PARTY: 'party',
  COMPETITION: 'competition',
} as const

export type ActivityTypeType = (typeof ActivityType)[keyof typeof ActivityType]

export const ActivityTypeText: Record<string, string> = {
  [ActivityType.LECTURE]: '讲座',
  [ActivityType.TRAINING]: '培训',
  [ActivityType.PARTY]: '聚会',
  [ActivityType.COMPETITION]: '比赛',
}

export const ActivityTypeOptions = Object.values(ActivityType).map((v) => ({
  label: ActivityTypeText[v],
  value: v,
}))

export const ActivityStatusOptions = Object.values(ActivityStatus).map((v) => ({
  label: ActivityStatusText[v],
  value: v,
}))
