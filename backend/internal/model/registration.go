package model

import "time"

// Registration 报名实体。
type Registration struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ActivityID   uint64    `gorm:"not null;index" json:"activity_id"`
	UserID       uint64    `gorm:"not null;index" json:"user_id"`
	Name         string    `gorm:"size:50;not null" json:"name"`
	Phone        string    `gorm:"size:20;not null" json:"phone"`
	Remark       string    `gorm:"size:255;not null;default:''" json:"remark"`
	VoucherNo    string    `gorm:"size:50;uniqueIndex;not null" json:"voucher_no"`
	Status       string    `gorm:"size:20;not null;default:registered" json:"status"`
	ReviewStatus string    `gorm:"size:20;not null;default:pending" json:"review_status"`
	CreatedAt    time.Time `json:"created_at"`
}

// TableName 指定表名。
func (Registration) TableName() string { return "registrations" }
