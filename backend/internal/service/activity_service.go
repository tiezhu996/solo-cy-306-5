package service

import (
	"log/slog"
	"time"

	"gbevent/internal/constants"
	"gbevent/internal/model"
	"gbevent/internal/repository"
	"gbevent/internal/util"
)

// ActivityService 活动流程与查询：创建、编辑、发布、结束、删除、列表、日历与统计。
// 访问控制委托 ActivityAccessPolicy，报名名额/截止校验委托 SignupGuard，
// 通知写入委托 NotificationService。
type ActivityService struct {
	repo        *repository.ActivityRepository
	checkinRepo *repository.CheckInRecordRepository
	access      *ActivityAccessPolicy
	logger      *slog.Logger
}

// NewActivityService 构造活动服务。
func NewActivityService(repo *repository.ActivityRepository, checkinRepo *repository.CheckInRecordRepository,
	access *ActivityAccessPolicy, logger *slog.Logger) *ActivityService {
	return &ActivityService{repo: repo, checkinRepo: checkinRepo, access: access, logger: logger}
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
	if err := s.access.RequireManager(a, operatorID, operatorRole, "update"); err != nil {
		return nil, err
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
	if err := s.access.RequireManager(a, operatorID, operatorRole, "publish"); err != nil {
		return nil, err
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
	if err := s.access.RequireManager(a, operatorID, operatorRole, "end"); err != nil {
		return nil, err
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
	if err := s.access.RequireManager(a, operatorID, operatorRole, "delete"); err != nil {
		return err
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
	if err := s.access.RequireManager(a, operatorID, operatorRole, "stats"); err != nil {
		return nil, err
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
