package repository

import (
	"errors"
	"fmt"

	"gbevent/internal/model"

	"gorm.io/gorm"
)

// FavoriteRepository 收藏仓储。
type FavoriteRepository struct {
	db *gorm.DB
}

// NewFavoriteRepository 构造收藏仓储。
func NewFavoriteRepository(db *gorm.DB) *FavoriteRepository {
	return &FavoriteRepository{db: db}
}

// Create 添加收藏。
func (r *FavoriteRepository) Create(f *model.Favorite) error {
	if err := r.db.Create(f).Error; err != nil {
		return fmt.Errorf("create favorite: %w", err)
	}
	return nil
}

// DeleteByUserActivity 取消收藏。
func (r *FavoriteRepository) DeleteByUserActivity(userID, activityID uint64) error {
	if err := r.db.Where("user_id = ? AND activity_id = ?", userID, activityID).
		Delete(&model.Favorite{}).Error; err != nil {
		return fmt.Errorf("delete favorite: %w", err)
	}
	return nil
}

// Exists 判断是否已收藏。
func (r *FavoriteRepository) Exists(userID, activityID uint64) (bool, error) {
	var n int64
	if err := r.db.Model(&model.Favorite{}).
		Where("user_id = ? AND activity_id = ?", userID, activityID).Count(&n).Error; err != nil {
		return false, fmt.Errorf("check favorite exists: %w", err)
	}
	return n > 0, nil
}

// ListByUser 查询某用户收藏列表。
func (r *FavoriteRepository) ListByUser(userID uint64, page, pageSize int) ([]model.Favorite, int64, error) {
	var list []model.Favorite
	var total int64
	q := r.db.Model(&model.Favorite{}).Where("user_id = ?", userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count favorites: %w", err)
	}
	if err := q.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("list favorites: %w", err)
	}
	return list, total, nil
}

// FindByID 按 ID 查询收藏。
func (r *FavoriteRepository) FindByID(id uint64) (*model.Favorite, error) {
	var f model.Favorite
	if err := r.db.First(&f, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find favorite by id: %w", err)
	}
	return &f, nil
}
