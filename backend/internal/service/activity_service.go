package service

import (
	"errors"
	"log/slog"
	"time"

	"gbevent/internal/constants"
	"gbevent/internal/model"
	"gbevent/internal/repository"
	"gbevent/internal/util"

	"gorm.io/gorm"
)

// ActivityService 活动业务逻辑。
type ActivityService struct {
	repo        *repository.ActivityRepository
	regRepo     *repository.RegistrationRepository
	notifyRepo  *repository.NotificationRepository
	checkinRepo *repository.CheckInRecordRepository
	logger      *slog.Logger
}

// NewActivityService 构造活动服务。
func NewActivityService(repo *repository.ActivityRepository, regRepo *repository.RegistrationRepository,
	notifyRepo *repository.NotificationRepository, checkinRepo *repository.CheckInRecordRepository,
	logger *slog.Logger) *ActivityService {
	return &ActivityService{repo: repo, regRepo: regRepo, notifyRepo: notifyRepo, checkinRepo: checkinRepo, logger: logger}
}

// Create 创建活动。
func (s *ActivityService) Create(organizerID uint64, title, description, coverImage, activityType, location string,
	startTime, endTime, signupDeadline time.Time, capacity int, status string) (*model.Activity, error) {
	if !constants.IsValidActivityType(activityType) {
		return nil, util.NewAppError(constants.CodeValidationFailed, "Activity[activity_type="+activityType+"] create: invalid type")
	}
	if status == "" {
		status = constants.ActivityStatusDraft
	}
	if !constants.IsValidActivityStatus(status) {
		return nil, util.NewAppError(constants.CodeValidationFailed, "Activity[status="+status+"] create: invalid status")
	}
	a := &model.Activity{
		Title:          title,
		Description:    description,
		CoverImage:     coverImage,
		ActivityType:   activityType,
		StartTime:      startTime,
		EndTime:        endTime,
		Location:       location,
		Capacity:       capacity,
		SignupDeadline: signupDeadline,
		Status:         status,
		OrganizerID:    organizerID,
	}
	if err := s.repo.Create(a); err != nil {
		s.logger.Error(constants.LogActivityCreateFailed, "error", err)
		return nil, util.Wrap(err, "Activity[organizer_id=%d] create failed", organizerID)
	}
	s.logger.Info(constants.LogActivityCreateSuccess, "activity_id", a.ID)
	return a, nil
}

// Update 更新活动（仅发布者或管理员）。
func (s *ActivityService) Update(id, operatorID uint64, operatorRole string, fields map[string]any) (*model.Activity, error) {
	a, err := s.repo.FindByID(id)
	if err != nil {
		return nil, util.Wrap(err, "Activity[id=%d] update find failed", id)
	}
	if operatorRole != constants.RoleAdmin && a.OrganizerID != operatorID {
		return nil, util.NewAppError(constants.CodeForbidden, "Activity[id="+itoa(id)+"] update forbidden: organizer not match")
	}
	if v, ok := fields["title"].(string); ok && v != "" {
		a.Title = v
	}
	if v, ok := fields["description"].(string); ok {
		a.Description = v
	}
	if v, ok := fields["cover_image"].(string); ok {
		a.CoverImage = v
	}
	if v, ok := fields["activity_type"].(string); ok && v != "" {
		if !constants.IsValidActivityType(v) {
			return nil, util.NewAppError(constants.CodeValidationFailed, "Activity[id="+itoa(id)+"] update: invalid activity_type="+v)
		}
		a.ActivityType = v
	}
	if v, ok := fields["location"].(string); ok {
		a.Location = v
	}
	if v, ok := fields["capacity"].(int); ok {
		a.Capacity = v
	}
	if err := s.repo.Update(a); err != nil {
		return nil, util.Wrap(err, "Activity[id=%d] update save failed", id)
	}
	s.logger.Info(constants.LogActivityUpdateSuccess, "activity_id", a.ID)
	return a, nil
}

// Publish 发布活动（draft -> published）。
func (s *ActivityService) Publish(id, operatorID uint64, operatorRole string) (*model.Activity, error) {
	a, err := s.repo.FindByID(id)
	if err != nil {
		return nil, util.Wrap(err, "Activity[id=%d] publish find failed", id)
	}
	if operatorRole != constants.RoleAdmin && a.OrganizerID != operatorID {
		return nil, util.NewAppError(constants.CodeForbidden, "Activity[id="+itoa(id)+"] publish forbidden: organizer not match")
	}
	if a.Status != constants.ActivityStatusDraft {
		return nil, util.NewAppError(constants.CodeConflict, "Activity[id="+itoa(id)+"] publish conflict: status="+a.Status)
	}
	a.Status = constants.ActivityStatusPublished
	if err := s.repo.Update(a); err != nil {
		return nil, util.Wrap(err, "Activity[id=%d] publish save failed", id)
	}
	s.logger.Info(constants.LogActivityPublishSuccess, "activity_id", a.ID)
	return a, nil
}

// End 结束活动（published -> ended）。
func (s *ActivityService) End(id, operatorID uint64, operatorRole string) (*model.Activity, error) {
	a, err := s.repo.FindByID(id)
	if err != nil {
		return nil, util.Wrap(err, "Activity[id=%d] end find failed", id)
	}
	if operatorRole != constants.RoleAdmin && a.OrganizerID != operatorID {
		return nil, util.NewAppError(constants.CodeForbidden, "Activity[id="+itoa(id)+"] end forbidden: organizer not match")
	}
	if a.Status != constants.ActivityStatusPublished {
		return nil, util.NewAppError(constants.CodeConflict, "Activity[id="+itoa(id)+"] end conflict: status="+a.Status)
	}
	a.Status = constants.ActivityStatusEnded
	if err := s.repo.Update(a); err != nil {
		return nil, util.Wrap(err, "Activity[id=%d] end save failed", id)
	}
	s.logger.Info(constants.LogActivityEndSuccess, "activity_id", a.ID)
	return a, nil
}

// Delete 删除活动（仅发布者或管理员）。
func (s *ActivityService) Delete(id, operatorID uint64, operatorRole string) error {
	a, err := s.repo.FindByID(id)
	if err != nil {
		return util.Wrap(err, "Activity[id=%d] delete find failed", id)
	}
	if operatorRole != constants.RoleAdmin && a.OrganizerID != operatorID {
		return util.NewAppError(constants.CodeForbidden, "Activity[id="+itoa(id)+"] delete forbidden: organizer not match")
	}
	if err := s.repo.Delete(id); err != nil {
		return util.Wrap(err, "Activity[id=%d] delete failed", id)
	}
	s.logger.Info(constants.LogActivityDeleteSuccess, "activity_id", id)
	return nil
}

// List 分页查询活动。
func (s *ActivityService) List(page, pageSize int, activityType, status, keyword string) ([]model.Activity, int64, error) {
	return s.repo.List(page, pageSize, activityType, status, keyword)
}

// ListByOrganizer 查询组织者自己的活动。
func (s *ActivityService) ListByOrganizer(organizerID uint64, page, pageSize int, status string) ([]model.Activity, int64, error) {
	return s.repo.ListByOrganizer(organizerID, page, pageSize, status)
}

// Get 查询活动详情，附带报名人数与签到人数。
func (s *ActivityService) Get(id uint64) (*model.Activity, int64, error) {
	a, err := s.repo.FindByID(id)
	if err != nil {
		return nil, 0, util.Wrap(err, "Activity[id=%d] get failed", id)
	}
	count, err := s.repo.CountRegistered(id)
	if err != nil {
		return nil, 0, util.Wrap(err, "Activity[id=%d] count registered failed", id)
	}
	return a, count, nil
}

// Calendar 按月份返回活动日历数据。
func (s *ActivityService) Calendar(month string) ([]model.Activity, error) {
	start, err := time.Parse("2006-01", month)
	if err != nil {
		start = time.Now()
	}
	start = time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0)
	list, err := s.repo.CalendarCounts(start, end)
	if err != nil {
		return nil, util.Wrap(err, "Activity[month=%s] calendar failed", month)
	}
	return list, nil
}

// Stats 统计活动报名/签到情况。
func (s *ActivityService) Stats(activityID, operatorID uint64, operatorRole string) (map[string]any, error) {
	a, err := s.repo.FindByID(activityID)
	if err != nil {
		return nil, util.Wrap(err, "Activity[id=%d] stats find failed", activityID)
	}
	if operatorRole != constants.RoleAdmin && a.OrganizerID != operatorID {
		return nil, util.NewAppError(constants.CodeForbidden, "Activity[id="+itoa(activityID)+"] stats forbidden: organizer not match")
	}
	registered, err := s.repo.CountRegistered(activityID)
	if err != nil {
		return nil, err
	}
	checked, err := s.checkinRepo.CountCheckedIn(activityID)
	if err != nil {
		return nil, err
	}
	rate := 0.0
	if registered > 0 {
		rate = float64(checked) / float64(registered) * 100
	}
	s.logger.Info(constants.LogCheckinRateStats, "activity_id", activityID, "rate", rate)
	return map[string]any{
		"activity_id":      activityID,
		"capacity":         a.Capacity,
		"registered_count": registered,
		"checked_in_count": checked,
		"checkin_rate":     round2(rate),
	}, nil
}

// CheckRegistrationLimit 校验报名名额与截止时间（供 RegistrationService 使用）。
func (s *ActivityService) CheckRegistrationLimit(activityID uint64) error {
	a, err := s.repo.FindByID(activityID)
	if err != nil {
		return util.Wrap(err, "Activity[id=%d] check limit failed", activityID)
	}
	return s.checkRegistrationLimit(a, func(activityID uint64) (int64, error) {
		return s.repo.CountRegistered(activityID)
	})
}

// CheckRegistrationLimitTx 在事务内校验报名名额与截止时间。
func (s *ActivityService) CheckRegistrationLimitTx(tx *gorm.DB, activityID uint64) error {
	a, err := s.repo.FindByIDForUpdate(tx, activityID)
	if err != nil {
		return util.Wrap(err, "Activity[id=%d] check limit failed", activityID)
	}
	return s.checkRegistrationLimit(a, func(activityID uint64) (int64, error) {
		return s.repo.CountRegisteredTx(tx, activityID)
	})
}

func (s *ActivityService) checkRegistrationLimit(a *model.Activity, countFn func(uint64) (int64, error)) error {
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

// CreateSignupNotification 生成报名/审核/签到通知。
func (s *ActivityService) CreateSignupNotification(userID uint64, notifType, title, content string) error {
	n := &model.Notification{UserID: userID, NotificationType: notifType, Title: title, Content: content}
	if err := s.notifyRepo.Create(n); err != nil {
		s.logger.Error(constants.LogNotificationCreate, "error", err)
		return util.Wrap(err, "Notification[user_id=%d] create failed", userID)
	}
	s.logger.Info(constants.LogNotificationCreate, "user_id", userID, "type", notifType)
	return nil
}

// CreateSignupNotificationTx 在事务内生成报名/审核/签到通知。
func (s *ActivityService) CreateSignupNotificationTx(tx *gorm.DB, userID uint64, notifType, title, content string) error {
	n := &model.Notification{UserID: userID, NotificationType: notifType, Title: title, Content: content}
	if err := s.notifyRepo.CreateTx(tx, n); err != nil {
		s.logger.Error(constants.LogNotificationCreate, "error", err)
		return util.Wrap(err, "Notification[user_id=%d] create failed", userID)
	}
	s.logger.Info(constants.LogNotificationCreate, "user_id", userID, "type", notifType)
	return nil
}

// IsOrganizer 判断操作者是否活动组织者或管理员。
func IsOrganizer(operatorID uint64, operatorRole string, organizerID uint64) bool {
	return operatorRole == constants.RoleAdmin || operatorID == organizerID
}

// errNotFound 判断是否为未找到错误。
func errNotFound(err error) bool {
	return errors.Is(err, repository.ErrNotFound)
}

func itoa(v uint64) string {
	return fmtUint(v)
}

func fmtUint(v uint64) string {
	if v == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for v > 0 {
		i--
		b[i] = byte('0' + v%10)
		v /= 10
	}
	return string(b[i:])
}

func round2(f float64) float64 {
	return float64(int(f*100+0.5)) / 100
}
