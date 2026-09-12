package model

import "time"

// Favorite 收藏实体。
type Favorite struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint64    `gorm:"not null;uniqueIndex:uk_fav_user_activity" json:"user_id"`
	ActivityID uint64    `gorm:"not null;uniqueIndex:uk_fav_user_activity" json:"activity_id"`
	CreatedAt  time.Time `json:"created_at"`
}

// TableName 指定表名。
func (Favorite) TableName() string { return "favorites" }
