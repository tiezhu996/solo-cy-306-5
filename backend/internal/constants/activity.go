package constants

// ActivityStatus 活动状态枚举（前后端重复定义，新增值需同步修改 ≥10 处文件）。
const (
	ActivityStatusDraft     = "draft"
	ActivityStatusPublished = "published"
	ActivityStatusEnded     = "ended"
)

// ActivityStatusValues 全部活动状态值。
var ActivityStatusValues = []string{ActivityStatusDraft, ActivityStatusPublished, ActivityStatusEnded}

// ActivityType 活动类型枚举。
const (
	ActivityTypeLecture     = "lecture"
	ActivityTypeTraining    = "training"
	ActivityTypeParty       = "party"
	ActivityTypeCompetition = "competition"
)

// ActivityTypeValues 全部活动类型值。
var ActivityTypeValues = []string{ActivityTypeLecture, ActivityTypeTraining, ActivityTypeParty, ActivityTypeCompetition}

// IsValidActivityStatus 校验活动状态。
func IsValidActivityStatus(s string) bool {
	for _, v := range ActivityStatusValues {
		if v == s {
			return true
		}
	}
	return false
}

// IsValidActivityType 校验活动类型。
func IsValidActivityType(s string) bool {
	for _, v := range ActivityTypeValues {
		if v == s {
			return true
		}
	}
	return false
}
