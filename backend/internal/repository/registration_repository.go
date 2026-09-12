package repository

import (
	"errors"
	"fmt"

	"gbevent/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// RegistrationRepository 报名仓储。
type RegistrationRepository struct {
	db *gorm.DB
}

// NewRegistrationRepository 构造报名仓储。
func NewRegistrationRepository(db *gorm.DB) *RegistrationRepository {
	return &RegistrationRepository{db: db}
}

// Create 创建报名。
func (r *RegistrationRepository) Create(reg *model.Registration) error {
	return r.CreateTx(r.db, reg)
}

// CreateTx 在事务内创建报名。
func (r *RegistrationRepository) CreateTx(tx *gorm.DB, reg *model.Registration) error {
	if err := tx.Create(reg).Error; err != nil {
		return fmt.Errorf("create registration: %w", err)
	}
	return nil
}

// FindByID 按 ID 查询报名。
func (r *RegistrationRepository) FindByID(id uint64) (*model.Registration, error) {
	return r.findByID(r.db, id, false)
}

// FindByIDForUpdate 在事务内锁定报名行。
func (r *RegistrationRepository) FindByIDForUpdate(tx *gorm.DB, id uint64) (*model.Registration, error) {
	return r.findByID(tx, id, true)
}

func (r *RegistrationRepository) findByID(db *gorm.DB, id uint64, forUpdate bool) (*model.Registration, error) {
	var reg model.Registration
	q := db
	if forUpdate {
		q = db.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := q.First(&reg, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find registration by id: %w", err)
	}
	return &reg, nil
}

// FindByVoucher 按凭证号查询报名。
func (r *RegistrationRepository) FindByVoucher(voucher string) (*model.Registration, error) {
	var reg model.Registration
	if err := r.db.Where("voucher_no = ?", voucher).First(&reg).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find registration by voucher: %w", err)
	}
	return &reg, nil
}

// FindByActivityAndUser 查询某用户对某活动的报名。
func (r *RegistrationRepository) FindByActivityAndUser(activityID, userID uint64) (*model.Registration, error) {
	return r.findByActivityAndUser(r.db, activityID, userID)
}

// FindByActivityAndUserTx 在事务内查询某用户对某活动的报名。
func (r *RegistrationRepository) FindByActivityAndUserTx(tx *gorm.DB, activityID, userID uint64) (*model.Registration, error) {
	return r.findByActivityAndUser(tx, activityID, userID)
}

func (r *RegistrationRepository) findByActivityAndUser(db *gorm.DB, activityID, userID uint64) (*model.Registration, error) {
	var reg model.Registration
	if err := db.Where("activity_id = ? AND user_id = ?", activityID, userID).First(&reg).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find registration by activity and user: %w", err)
	}
	return &reg, nil
}

// List 分页查询报名，支持活动/状态/审核状态筛选。
func (r *RegistrationRepository) List(page, pageSize int, activityID uint64, status, reviewStatus string) ([]model.Registration, int64, error) {
	var list []model.Registration
	var total int64
	q := r.db.Model(&model.Registration{})
	if activityID > 0 {
		q = q.Where("activity_id = ?", activityID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if reviewStatus != "" {
		q = q.Where("review_status = ?", reviewStatus)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count registrations: %w", err)
	}
	if err := q.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("list registrations: %w", err)
	}
	return list, total, nil
}

// ListByUser 查询某用户的报名。
func (r *RegistrationRepository) ListByUser(userID uint64, page, pageSize int) ([]model.Registration, int64, error) {
	var list []model.Registration
	var total int64
	q := r.db.Model(&model.Registration{}).Where("user_id = ?", userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count user registrations: %w", err)
	}
	if err := q.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("list user registrations: %w", err)
	}
	return list, total, nil
}

// ListByActivity 导出某活动的全部报名。
func (r *RegistrationRepository) ListByActivity(activityID uint64) ([]model.Registration, error) {
	var list []model.Registration
	if err := r.db.Where("activity_id = ?", activityID).Order("id ASC").Find(&list).Error; err != nil {
		return nil, fmt.Errorf("list registrations by activity: %w", err)
	}
	return list, nil
}

// Update 更新报名。
func (r *RegistrationRepository) Update(reg *model.Registration) error {
	return r.UpdateTx(r.db, reg)
}

// UpdateTx 在事务内更新报名。
func (r *RegistrationRepository) UpdateTx(tx *gorm.DB, reg *model.Registration) error {
	if err := tx.Save(reg).Error; err != nil {
		return fmt.Errorf("update registration: %w", err)
	}
	return nil
}
