package repository

import (
	"errors"
	"fmt"

	"gbevent/internal/model"

	"gorm.io/gorm"
)

// NotificationRepository 通知仓储。
type NotificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository 构造通知仓储。
func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Create 创建通知。
func (r *NotificationRepository) Create(n *model.Notification) error {
	return r.CreateTx(r.db, n)
}

// CreateTx 在事务内创建通知。
func (r *NotificationRepository) CreateTx(tx *gorm.DB, n *model.Notification) error {
	if err := tx.Create(n).Error; err != nil {
		return fmt.Errorf("create notification: %w", err)
	}
	return nil
}

// ListByUser 查询某用户通知列表。
func (r *NotificationRepository) ListByUser(userID uint64, page, pageSize int) ([]model.Notification, int64, error) {
	var list []model.Notification
	var total int64
	q := r.db.Model(&model.Notification{}).Where("user_id = ?", userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count notifications: %w", err)
	}
	if err := q.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("list notifications: %w", err)
	}
	return list, total, nil
}

// FindByID 按 ID 查询通知。
func (r *NotificationRepository) FindByID(id uint64) (*model.Notification, error) {
	var n model.Notification
	if err := r.db.First(&n, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find notification by id: %w", err)
	}
	return &n, nil
}

// MarkRead 标记已读。
func (r *NotificationRepository) MarkRead(id, userID uint64) error {
	if err := r.db.Model(&model.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).Update("is_read", true).Error; err != nil {
		return fmt.Errorf("mark notification read: %w", err)
	}
	return nil
}

// MarkAllRead 全部标记已读。
func (r *NotificationRepository) MarkAllRead(userID uint64) error {
	if err := r.db.Model(&model.Notification{}).
		Where("user_id = ?", userID).Update("is_read", true).Error; err != nil {
		return fmt.Errorf("mark all notifications read: %w", err)
	}
	return nil
}
