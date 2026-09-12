package service

import (
	"log/slog"
	"time"

	"gbevent/internal/constants"
	"gbevent/internal/model"
	"gbevent/internal/repository"
	"gbevent/internal/util"

	"gorm.io/gorm"
)

// CheckInRecordService 签到业务逻辑。
type CheckInRecordService struct {
	db          *gorm.DB
	repo        *repository.CheckInRecordRepository
	regRepo     *repository.RegistrationRepository
	activitySvc *ActivityService
	notifyRepo  *repository.NotificationRepository
	logger      *slog.Logger
}

// NewCheckInRecordService 构造签到服务。
func NewCheckInRecordService(db *gorm.DB, repo *repository.CheckInRecordRepository, regRepo *repository.RegistrationRepository,
	activitySvc *ActivityService, notifyRepo *repository.NotificationRepository, logger *slog.Logger) *CheckInRecordService {
	return &CheckInRecordService{db: db, repo: repo, regRepo: regRepo, activitySvc: activitySvc, notifyRepo: notifyRepo, logger: logger}
}

// CheckInByVoucher 凭证号签到。
func (s *CheckInRecordService) CheckInByVoucher(activityID, operatorID uint64, voucher string) (*model.CheckInRecord, error) {
	reg, err := s.regRepo.FindByVoucher(voucher)
	if err != nil {
		s.logger.Warn(constants.LogCheckinVoucherNotFound, "voucher", voucher)
		return nil, util.NewAppError(constants.CodeInvalidVoucher, constants.MsgInvalidVoucher)
	}
	if reg.ActivityID != activityID {
		return nil, util.NewAppError(constants.CodeInvalidVoucher, "CheckInRecord[activity_id="+itoa(activityID)+"] voucher not match this activity")
	}
	return s.doCheckIn(reg, operatorID, constants.CheckInMethodVoucher)
}

// CheckInByScan 扫码签到（二维码内容为报名 ID）。
func (s *CheckInRecordService) CheckInByScan(activityID, operatorID uint64, qrContent string) (*model.CheckInRecord, error) {
	regID, err := util.SscanUint64(qrContent)
	if err != nil || regID == 0 {
		s.logger.Warn(constants.LogCheckinScanFailed, "qr_content", qrContent)
		return nil, util.NewAppError(constants.CodeInvalidVoucher, "CheckInRecord[qr="+qrContent+"] scan failed: invalid qr content")
	}
	reg, err := s.regRepo.FindByID(regID)
	if err != nil {
		return nil, util.Wrap(err, "CheckInRecord[qr=%s] scan find registration failed", qrContent)
	}
	if reg.ActivityID != activityID {
		return nil, util.NewAppError(constants.CodeInvalidVoucher, "CheckInRecord[activity_id="+itoa(activityID)+"] scan not match this activity")
	}
	return s.doCheckIn(reg, operatorID, constants.CheckInMethodScan)
}

// doCheckIn 执行签到并发送签到成功通知。
func (s *CheckInRecordService) doCheckIn(reg *model.Registration, operatorID uint64, method string) (*model.CheckInRecord, error) {
	var rec *model.CheckInRecord
	err := s.db.Transaction(func(tx *gorm.DB) error {
		cur, err := s.regRepo.FindByIDForUpdate(tx, reg.ID)
		if err != nil {
			return util.Wrap(err, "Registration[id=%d] checkin find failed", reg.ID)
		}
		if cur.Status == constants.RegistrationStatusCheckedIn {
			return util.NewAppError(constants.CodeAlreadyCheckedIn, constants.MsgAlreadyCheckedIn)
		}
		if cur.Status != constants.RegistrationStatusRegistered {
			return util.NewAppError(constants.CodeCancelConflict, "Registration[id="+itoa(cur.ID)+"] cannot checkin: status="+cur.Status)
		}
		rec = &model.CheckInRecord{
			RegistrationID: cur.ID,
			ActivityID:     cur.ActivityID,
			CheckInMethod:  method,
			CheckInTime:    time.Now(),
			OperatorID:     operatorID,
		}
		if err := s.repo.CreateTx(tx, rec); err != nil {
			s.logger.Error(constants.LogRegistrationCheckinFailed, "error", err)
			return util.Wrap(err, "CheckInRecord[registration_id=%d] create failed", cur.ID)
		}
		cur.Status = constants.RegistrationStatusCheckedIn
		if err := s.regRepo.UpdateTx(tx, cur); err != nil {
			return util.Wrap(err, "Registration[id=%d] checkin update failed", cur.ID)
		}
		if err := s.activitySvc.CreateSignupNotificationTx(tx, cur.UserID, constants.NotificationCheckinSuccess,
			"签到成功", "您已完成签到，凭证号："+cur.VoucherNo); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.logger.Info(constants.LogRegistrationCheckinSuccess, "registration_id", reg.ID, "method", method)
	return rec, nil
}

// ListByActivity 查询某活动签到记录。
func (s *CheckInRecordService) ListByActivity(activityID uint64) ([]model.CheckInRecord, error) {
	return s.repo.ListByActivity(activityID)
}
