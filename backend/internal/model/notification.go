package model

import "time"

// Notification 消息通知实体。
type Notification struct {
	ID               uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID           uint64    `gorm:"not null;index" json:"user_id"`
	NotificationType string    `gorm:"size:30;not null;default:signup_success" json:"notification_type"`
	Title            string    `gorm:"size:200;not null" json:"title"`
	Content          string    `gorm:"type:text" json:"content"`
	IsRead           bool      `gorm:"not null;default:false" json:"is_read"`
	CreatedAt        time.Time `json:"created_at"`
}

// TableName 指定表名。
func (Notification) TableName() string { return "notifications" }
