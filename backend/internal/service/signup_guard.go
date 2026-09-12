package service

import (
	"time"

	"gbevent/internal/constants"
	"gbevent/internal/model"
	"gbevent/internal/repository"
	"gbevent/internal/util"

	"gorm.io/gorm"
)

// SignupGuard 报名校验：集中承载活动状态、报名截止时间与名额校验，
// 供在线报名与线下补录复用；Tx 版本在事务内锁定活动行，保证并发下名额不超卖。
type SignupGuard struct {
	activityRepo *repository.ActivityRepository
}

// NewSignupGuard 构造报名校验器。
func NewSignupGuard(activityRepo *repository.ActivityRepository) *SignupGuard {
	return &SignupGuard{activityRepo: activityRepo}
}

// CheckRegistrationLimit 校验报名名额与截止时间。
func (g *SignupGuard) CheckRegistrationLimit(activityID uint64) error {
	a, err := g.activityRepo.FindByID(activityID)
	if err != nil {
		return util.Wrap(err, "Activity[id=%d] check limit failed", activityID)
	}
	return g.checkRegistrationLimit(a, func(activityID uint64) (int64, error) {
		return g.activityRepo.CountRegistered(activityID)
	})
}

// CheckRegistrationLimitTx 在事务内校验报名名额与截止时间。
func (g *SignupGuard) CheckRegistrationLimitTx(tx *gorm.DB, activityID uint64) error {
	a, err := g.activityRepo.FindByIDForUpdate(tx, activityID)
	if err != nil {
		return util.Wrap(err, "Activity[id=%d] check limit failed", activityID)
	}
	return g.checkRegistrationLimit(a, func(activityID uint64) (int64, error) {
		return g.activityRepo.CountRegisteredTx(tx, activityID)
	})
}

func (g *SignupGuard) checkRegistrationLimit(a *model.Activity, countFn func(uint64) (int64, error)) error {
	if a.Status == constants.ActivityStatusEnded {
		return util.NewAppError(constants.CodeActivityEnded, constants.MsgActivityEnded)
	}
	if a.Status != constants.ActivityStatusPublished {
		return util.NewAppError(constants.CodeConflict, "Activity[id="+itoa(a.ID)+"] not published")
	}
	if time.Now().After(a.SignupDeadline) {
		return util.NewAppError(constants.CodeActivityEnded, "Activity[id="+itoa(a.ID)+"] signup deadline passed")
	}
	count, err := countFn(a.ID)
	if err != nil {
		return err
	}
	if a.Capacity > 0 && count >= int64(a.Capacity) {
		return util.NewAppError(constants.CodeActivityFull, constants.MsgActivityFull)
	}
	return nil
}
