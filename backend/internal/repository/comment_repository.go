package repository

import (
	"errors"
	"fmt"

	"gbevent/internal/model"

	"gorm.io/gorm"
)

// CommentRepository 评论仓储。
type CommentRepository struct {
	db *gorm.DB
}

// NewCommentRepository 构造评论仓储。
func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

// Create 创建评论。
func (r *CommentRepository) Create(c *model.Comment) error {
	if err := r.db.Create(c).Error; err != nil {
		return fmt.Errorf("create comment: %w", err)
	}
	return nil
}

// ListByActivity 查询某活动评论列表。
func (r *CommentRepository) ListByActivity(activityID uint64) ([]model.Comment, error) {
	var list []model.Comment
	if err := r.db.Where("activity_id = ?", activityID).Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, fmt.Errorf("list comments: %w", err)
	}
	return list, nil
}

// ListByUser 查询某用户的评论。
func (r *CommentRepository) ListByUser(userID uint64) ([]model.Comment, error) {
	var list []model.Comment
	if err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, fmt.Errorf("list user comments: %w", err)
	}
	return list, nil
}

// AvgRating 计算某活动平均评分。
func (r *CommentRepository) AvgRating(activityID uint64) (float64, error) {
	var avg float64
	if err := r.db.Model(&model.Comment{}).Where("activity_id = ?", activityID).
		Select("COALESCE(AVG(rating), 0)").Scan(&avg).Error; err != nil {
		return 0, fmt.Errorf("avg rating: %w", err)
	}
	return avg, nil
}

// FindByID 按 ID 查询评论。
func (r *CommentRepository) FindByID(id uint64) (*model.Comment, error) {
	var c model.Comment
	if err := r.db.First(&c, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find comment by id: %w", err)
	}
	return &c, nil
}

// Delete 删除评论。
func (r *CommentRepository) Delete(id uint64) error {
	if err := r.db.Delete(&model.Comment{}, id).Error; err != nil {
		return fmt.Errorf("delete comment: %w", err)
	}
	return nil
}
