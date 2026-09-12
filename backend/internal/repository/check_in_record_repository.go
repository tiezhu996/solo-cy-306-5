package repository

import (
	"fmt"

	"gbevent/internal/model"

	"gorm.io/gorm"
)

// CheckInRecordRepository 签到记录仓储。
type CheckInRecordRepository struct {
	db *gorm.DB
}

// NewCheckInRecordRepository 构造签到记录仓储。
func NewCheckInRecordRepository(db *gorm.DB) *CheckInRecordRepository {
	return &CheckInRecordRepository{db: db}
}

// Create 创建签到记录。
func (r *CheckInRecordRepository) Create(rec *model.CheckInRecord) error {
	return r.CreateTx(r.db, rec)
}

// CreateTx 在事务内创建签到记录。
func (r *CheckInRecordRepository) CreateTx(tx *gorm.DB, rec *model.CheckInRecord) error {
	if err := tx.Create(rec).Error; err != nil {
		return fmt.Errorf("create check-in record: %w", err)
	}
	return nil
}

// ListByActivity 查询某活动的签到记录。
func (r *CheckInRecordRepository) ListByActivity(activityID uint64) ([]model.CheckInRecord, error) {
	var list []model.CheckInRecord
	if err := r.db.Where("activity_id = ?", activityID).Order("check_in_time DESC").Find(&list).Error; err != nil {
		return nil, fmt.Errorf("list check-in records: %w", err)
	}
	return list, nil
}

// CountCheckedIn 统计某活动已签到人数。
func (r *CheckInRecordRepository) CountCheckedIn(activityID uint64) (int64, error) {
	var n int64
	if err := r.db.Model(&model.CheckInRecord{}).Where("activity_id = ?", activityID).Count(&n).Error; err != nil {
		return 0, fmt.Errorf("count checked-in: %w", err)
	}
	return n, nil
}
