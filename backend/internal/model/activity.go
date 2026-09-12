package model

import "time"

// Activity 活动实体。
type Activity struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Title          string    `gorm:"size:200;not null" json:"title"`
	Description    string    `gorm:"type:text" json:"description"`
	CoverImage     string    `gorm:"size:255;not null;default:''" json:"cover_image"`
	ActivityType   string    `gorm:"size:20;not null;default:lecture" json:"activity_type"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	Location       string    `gorm:"size:255;not null;default:''" json:"location"`
	Capacity       int       `gorm:"not null;default:0" json:"capacity"`
	SignupDeadline time.Time `json:"signup_deadline"`
	Status         string    `gorm:"size:20;not null;default:draft;index" json:"status"`
	OrganizerID    uint64    `gorm:"not null;index" json:"organizer_id"`
	CreatedAt      time.Time `json:"created_at"`
}

// TableName 指定表名。
func (Activity) TableName() string { return "activities" }
