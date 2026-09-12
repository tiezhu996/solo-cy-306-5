package model

import "time"

// User 用户实体：可创建 Activity、Registration、Comment、Favorite 与 Notification。
type User struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Username     string    `gorm:"size:50;uniqueIndex;not null" json:"username"`
	PasswordHash string    `gorm:"size:100;not null" json:"-"`
	Nickname     string    `gorm:"size:50;not null;default:''" json:"nickname"`
	Avatar       string    `gorm:"size:255;not null;default:''" json:"avatar"`
	Role         string    `gorm:"size:20;not null;default:user" json:"role"`
	Email        string    `gorm:"size:100;not null;default:''" json:"email"`
	Phone        string    `gorm:"size:20;not null;default:''" json:"phone"`
	CreatedAt    time.Time `json:"created_at"`
}

// TableName 指定表名。
func (User) TableName() string { return "users" }
