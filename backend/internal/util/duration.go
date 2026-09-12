package util

import "time"

// DurationHours 将小时数转换为 time.Duration。
func DurationHours(hours int) time.Duration {
	return time.Duration(hours) * time.Hour
}
