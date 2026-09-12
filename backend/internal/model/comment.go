package model

import "time"

// Comment 活动评论实体。
type Comment struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ActivityID uint64    `gorm:"not null;index" json:"activity_id"`
	UserID     uint64    `gorm:"not null" json:"user_id"`
	Rating     int       `gorm:"not null;default:5" json:"rating"`
	Content    string    `gorm:"type:text" json:"content"`
	CreatedAt  time.Time `json:"created_at"`
}

// TableName 指定表名。
func (Comment) TableName() string { return "comments" }
