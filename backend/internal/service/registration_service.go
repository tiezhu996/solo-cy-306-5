package service

import (
	"encoding/csv"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"gbevent/internal/constants"
	"gbevent/internal/model"
	"gbevent/internal/repository"
	"gbevent/internal/util"

	"gorm.io/gorm"
)

// RegistrationService 报名业务逻辑。
type RegistrationService struct {
	db          *gorm.DB
	repo        *repository.RegistrationRepository
	activitySvc *ActivityService
	notifyRepo  *repository.NotificationRepository
	logger      *slog.Logger
}

// NewRegistrationService 构造报名服务。
func NewRegistrationService(db *gorm.DB, repo *repository.RegistrationRepository, activitySvc *ActivityService,
	notifyRepo *repository.NotificationRepository, logger *slog.Logger) *RegistrationService {
	return &RegistrationService{db: db, repo: repo, activitySvc: activitySvc, notifyRepo: notifyRepo, logger: logger}
}

// Create 在线报名：校验名额与截止时间、防重复报名，生成凭证号并发送报名成功通知。
func (s *RegistrationService) Create(activityID, userID uint64, name, phone, remark string) (*model.Registration, error) {
	s.logger.Info(constants.LogRegistrationCreateStart, "activity_id", activityID, "user_id", userID)
	reg := &model.Registration{}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.activitySvc.CheckRegistrationLimitTx(tx, activityID); err != nil {
			return err
		}
		if _, err := s.repo.FindByActivityAndUserTx(tx, activityID, userID); err == nil {
			return util.NewAppError(constants.CodeDuplicateSignup, constants.MsgDuplicateSignup)
		} else if !errNotFound(err) {
			return err
		}
		reg.ActivityID = activityID
		reg.UserID = userID
		reg.Name = name
		reg.Phone = phone
		reg.Remark = remark
		reg.VoucherNo = util.GenerateVoucherNo()
		reg.Status = constants.RegistrationStatusRegistered
		reg.ReviewStatus = constants.ReviewStatusPending
		if err := s.repo.CreateTx(tx, reg); err != nil {
			s.logger.Error(constants.LogRegistrationCreateFailed, "error", err)
			return util.Wrap(err, "Registration[activity_id=%d,user_id=%d] create failed", activityID, userID)
		}
		if err := s.activitySvc.CreateSignupNotificationTx(tx, userID, constants.NotificationSignupSuccess,
			"报名成功", "您已成功报名活动，凭证号："+reg.VoucherNo); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.logger.Info(constants.LogRegistrationCreateSuccess, "registration_id", reg.ID, "voucher_no", reg.VoucherNo)
	return reg, nil
}

// Cancel 取消报名（registered -> cancelled）。
func (s *RegistrationService) Cancel(id, operatorID uint64, operatorRole string) (*model.Registration, error) {
	reg, err := s.repo.FindByID(id)
	if err != nil {
		return nil, util.Wrap(err, "Registration[id=%d] cancel find failed", id)
	}
	if operatorRole != constants.RoleAdmin && reg.UserID != operatorID {
		return nil, util.NewAppError(constants.CodeForbidden, "Registration[id="+itoa(id)+"] cancel forbidden: not owner")
	}
	if reg.Status != constants.RegistrationStatusRegistered {
		return nil, util.NewAppError(constants.CodeCancelConflict, constants.MsgCancelConflict)
	}
	reg.Status = constants.RegistrationStatusCancelled
	if err := s.repo.Update(reg); err != nil {
		return nil, util.Wrap(err, "Registration[id=%d] cancel save failed", id)
	}
	s.logger.Info(constants.LogRegistrationCancelSuccess, "registration_id", reg.ID)
	return reg, nil
}

// Review 审核报名（pending -> approved/rejected）。
func (s *RegistrationService) Review(id, operatorID uint64, operatorRole string, reviewStatus string) (*model.Registration, error) {
	var reg *model.Registration
	err := s.db.Transaction(func(tx *gorm.DB) error {
		cur, err := s.repo.FindByIDForUpdate(tx, id)
		if err != nil {
			return util.Wrap(err, "Registration[id=%d] review find failed", id)
		}
		act, _, err := s.activitySvc.Get(cur.ActivityID)
		if err != nil {
			return err
		}
		if !IsOrganizer(operatorID, operatorRole, act.OrganizerID) {
			return util.NewAppError(constants.CodeForbidden, "Registration[id="+itoa(id)+"] review forbidden: organizer not match")
		}
		if cur.ReviewStatus != constants.ReviewStatusPending {
			return util.NewAppError(constants.CodeReviewConflict, constants.MsgReviewConflict)
		}
		if !constants.IsValidReviewStatus(reviewStatus) {
			return util.NewAppError(constants.CodeValidationFailed, "Registration[id="+itoa(id)+"] review invalid status="+reviewStatus)
		}
		cur.ReviewStatus = reviewStatus
		if err := s.repo.UpdateTx(tx, cur); err != nil {
			s.logger.Error(constants.LogRegistrationReviewFailed, "error", err)
			return util.Wrap(err, "Registration[id=%d] review save failed", id)
		}
		title := "审核结果"
		content := "您的报名审核已" + util.ReviewStatusText(reviewStatus)
		if reviewStatus == constants.ReviewStatusApproved {
			content = "您的报名已通过审核，凭证号：" + cur.VoucherNo
		}
		if err := s.activitySvc.CreateSignupNotificationTx(tx, cur.UserID, constants.NotificationReviewResult, title, content); err != nil {
			return err
		}
		reg = cur
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.logger.Info(constants.LogRegistrationReviewSuccess, "registration_id", reg.ID, "review_status", reviewStatus)
	return reg, nil
}

// OfflineCreate 线下补录报名。
func (s *RegistrationService) OfflineCreate(activityID, operatorID uint64, name, phone, remark string) (*model.Registration, error) {
	if err := s.activitySvc.CheckRegistrationLimit(activityID); err != nil {
		return nil, err
	}
	reg := &model.Registration{
		ActivityID:   activityID,
		UserID:       operatorID,
		Name:         name,
		Phone:        phone,
		Remark:       remark,
		VoucherNo:    util.GenerateVoucherNo(),
		Status:       constants.RegistrationStatusRegistered,
		ReviewStatus: constants.ReviewStatusApproved,
	}
	if err := s.repo.Create(reg); err != nil {
		return nil, util.Wrap(err, "Registration[activity_id=%d] offline create failed", activityID)
	}
	s.logger.Info(constants.LogRegistrationCreateSuccess, "registration_id", reg.ID, "source", "offline")
	return reg, nil
}

// List 分页查询报名（组织者）。
func (s *RegistrationService) List(page, pageSize int, activityID uint64, status, reviewStatus string) ([]model.Registration, int64, error) {
	return s.repo.List(page, pageSize, activityID, status, reviewStatus)
}

// ListMine 查询我的报名。
func (s *RegistrationService) ListMine(userID uint64, page, pageSize int) ([]model.Registration, int64, error) {
	return s.repo.ListByUser(userID, page, pageSize)
}

// Get 查询报名详情。
func (s *RegistrationService) Get(id uint64) (*model.Registration, error) {
	return s.repo.FindByID(id)
}

// ExportCSV 导出报名名单 CSV。
func (s *RegistrationService) ExportCSV(activityID, operatorID uint64, operatorRole string) (string, string, error) {
	a, _, err := s.activitySvc.Get(activityID)
	if err != nil {
		return "", "", err
	}
	if !IsOrganizer(operatorID, operatorRole, a.OrganizerID) {
		return "", "", util.NewAppError(constants.CodeForbidden, "Registration[activity_id="+itoa(activityID)+"] export forbidden: organizer not match")
	}
	list, err := s.repo.ListByActivity(activityID)
	if err != nil {
		return "", "", err
	}
	var buf strings.Builder
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"ID", "活动ID", "报名人", "手机号", "凭证号", "状态", "审核状态", "备注", "报名时间"})
	for _, r := range list {
		_ = w.Write([]string{
			strconv.FormatUint(r.ID, 10),
			strconv.FormatUint(r.ActivityID, 10),
			r.Name, r.Phone, r.VoucherNo,
			util.RegistrationStatusText(r.Status),
			util.ReviewStatusText(r.ReviewStatus),
			r.Remark,
			r.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	w.Flush()
	filename := "registrations_" + strconv.FormatUint(activityID, 10) + "_" + time.Now().Format("20060102150405") + ".csv"
	return filename, buf.String(), nil
}
