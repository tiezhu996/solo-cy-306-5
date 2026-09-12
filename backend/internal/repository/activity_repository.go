package repository

import (
	"errors"
	"fmt"
	"time"

	"gbevent/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ActivityRepository 活动仓储。
type ActivityRepository struct {
	db *gorm.DB
}

// NewActivityRepository 构造活动仓储。
func NewActivityRepository(db *gorm.DB) *ActivityRepository {
	return &ActivityRepository{db: db}
}

// Create 创建活动。
func (r *ActivityRepository) Create(a *model.Activity) error {
	if err := r.db.Create(a).Error; err != nil {
		return fmt.Errorf("create activity: %w", err)
	}
	return nil
}

// FindByID 按 ID 查询活动。
func (r *ActivityRepository) FindByID(id uint64) (*model.Activity, error) {
	return r.findByID(r.db, id, false)
}

// FindByIDForUpdate 在事务内锁定活动行，用于报名名额校验等并发场景。
func (r *ActivityRepository) FindByIDForUpdate(tx *gorm.DB, id uint64) (*model.Activity, error) {
	return r.findByID(tx, id, true)
}

func (r *ActivityRepository) findByID(db *gorm.DB, id uint64, forUpdate bool) (*model.Activity, error) {
	var a model.Activity
	q := db
	if forUpdate {
		q = db.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := q.First(&a, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find activity by id: %w", err)
	}
	return &a, nil
}

// List 分页查询活动，支持类型/状态/关键词筛选。
func (r *ActivityRepository) List(page, pageSize int, activityType, status, keyword string) ([]model.Activity, int64, error) {
	var list []model.Activity
	var total int64
	q := r.db.Model(&model.Activity{})
	if activityType != "" {
		q = q.Where("activity_type = ?", activityType)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if keyword != "" {
		q = q.Where("title LIKE ?", "%"+keyword+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count activities: %w", err)
	}
	if err := q.Order("start_time ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("list activities: %w", err)
	}
	return list, total, nil
}

// ListByOrganizer 查询某组织者的活动。
func (r *ActivityRepository) ListByOrganizer(organizerID uint64, page, pageSize int, status string) ([]model.Activity, int64, error) {
	var list []model.Activity
	var total int64
	q := r.db.Model(&model.Activity{}).Where("organizer_id = ?", organizerID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count organizer activities: %w", err)
	}
	if err := q.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("list organizer activities: %w", err)
	}
	return list, total, nil
}

// CalendarCounts 按月统计每天的已发布活动数量。
func (r *ActivityRepository) CalendarCounts(start, end time.Time) ([]model.Activity, error) {
	var list []model.Activity
	if err := r.db.Where("status = ? AND start_time >= ? AND start_time < ?", "published", start, end).
		Order("start_time ASC").Find(&list).Error; err != nil {
		return nil, fmt.Errorf("calendar activities: %w", err)
	}
	return list, nil
}

// CountRegistered 统计活动已报名人数（未取消）。
func (r *ActivityRepository) CountRegistered(activityID uint64) (int64, error) {
	return r.countRegistered(r.db, activityID)
}

// CountRegisteredTx 在事务内统计活动已报名人数（未取消）。
func (r *ActivityRepository) CountRegisteredTx(tx *gorm.DB, activityID uint64) (int64, error) {
	return r.countRegistered(tx, activityID)
}

func (r *ActivityRepository) countRegistered(db *gorm.DB, activityID uint64) (int64, error) {
	var n int64
	if err := db.Model(&model.Registration{}).
		Where("activity_id = ? AND status <> ?", activityID, "cancelled").Count(&n).Error; err != nil {
		return 0, fmt.Errorf("count registrations: %w", err)
	}
	return n, nil
}

// Update 更新活动。
func (r *ActivityRepository) Update(a *model.Activity) error {
	if err := r.db.Save(a).Error; err != nil {
		return fmt.Errorf("update activity: %w", err)
	}
	return nil
}

// Delete 删除活动。
func (r *ActivityRepository) Delete(id uint64) error {
	if err := r.db.Delete(&model.Activity{}, id).Error; err != nil {
		return fmt.Errorf("delete activity: %w", err)
	}
	return nil
}
