package model

import "time"

// CheckInRecord 签到记录实体。
type CheckInRecord struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	RegistrationID uint64    `gorm:"not null;index" json:"registration_id"`
	ActivityID     uint64    `gorm:"not null;index" json:"activity_id"`
	CheckInMethod  string    `gorm:"size:20;not null;default:voucher" json:"check_in_method"`
	CheckInTime    time.Time `json:"check_in_time"`
	OperatorID     uint64    `gorm:"not null" json:"operator_id"`
	CreatedAt      time.Time `json:"created_at"`
}

// TableName 指定表名。
func (CheckInRecord) TableName() string { return "check_in_records" }
