package service

import (
	"log/slog"
	"time"

	"gbevent/internal/constants"
	"gbevent/internal/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedService 负责启动时幂等写入预置数据。
type SeedService struct {
	db     *gorm.DB
	logger *slog.Logger
}

// NewSeedService 构造种子服务。
func NewSeedService(db *gorm.DB, logger *slog.Logger) *SeedService {
	return &SeedService{db: db, logger: logger}
}

// Seed 当 users 表为空时写入管理员、组织者、普通用户与示例活动。
func (s *SeedService) Seed() error {
	var count int64
	if err := s.db.Model(&model.User{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	adminHash, _ := bcrypt.GenerateFromPassword([]byte("Admin@123"), bcrypt.DefaultCost)
	userHash, _ := bcrypt.GenerateFromPassword([]byte("User@123"), bcrypt.DefaultCost)
	users := []model.User{
		{Username: "admin", PasswordHash: string(adminHash), Nickname: "系统管理员", Role: constants.RoleAdmin, Email: "admin@gbevent.dev", Phone: "13800000001"},
		{Username: "organizer", PasswordHash: string(userHash), Nickname: "活动组织者", Role: constants.RoleOrganizer, Email: "org@gbevent.dev", Phone: "13800000002"},
		{Username: "user", PasswordHash: string(userHash), Nickname: "普通用户", Role: constants.RoleUser, Email: "user@gbevent.dev", Phone: "13800000003"},
	}
	for i := range users {
		if err := s.db.Create(&users[i]).Error; err != nil {
			return err
		}
	}
	now := time.Now()
	activities := []model.Activity{
		{Title: "Go 语言企业级开发实战讲座", Description: "深入讲解 Go 1.22 + Gin + GORM 的企业级工程实践。", ActivityType: constants.ActivityTypeLecture, StartTime: now.AddDate(0, 0, 7), EndTime: now.AddDate(0, 0, 7), Location: "线上直播", Capacity: 200, SignupDeadline: now.AddDate(0, 0, 6), Status: constants.ActivityStatusPublished, OrganizerID: 2},
		{Title: "新员工安全培训", Description: "面向新入职员工的安全意识与应急处理培训。", ActivityType: constants.ActivityTypeTraining, StartTime: now.AddDate(0, 0, 14), EndTime: now.AddDate(0, 0, 14), Location: "A 座 3 楼培训室", Capacity: 50, SignupDeadline: now.AddDate(0, 0, 13), Status: constants.ActivityStatusPublished, OrganizerID: 2},
		{Title: "秋季团队趣味运动会", Description: "团队协作趣味运动会，包含拔河、接力、跳绳等项目。", ActivityType: constants.ActivityTypeParty, StartTime: now.AddDate(0, 0, 30), EndTime: now.AddDate(0, 0, 30), Location: "城市体育公园", Capacity: 120, SignupDeadline: now.AddDate(0, 0, 28), Status: constants.ActivityStatusPublished, OrganizerID: 2},
		{Title: "黑客松编程竞赛（草稿）", Description: "24 小时黑客松编程竞赛，暂未发布。", ActivityType: constants.ActivityTypeCompetition, StartTime: now.AddDate(0, 0, 60), EndTime: now.AddDate(0, 0, 62), Location: "创新中心", Capacity: 80, SignupDeadline: now.AddDate(0, 0, 55), Status: constants.ActivityStatusDraft, OrganizerID: 2},
	}
	for i := range activities {
		if err := s.db.Create(&activities[i]).Error; err != nil {
			return err
		}
	}
	regs := []model.Registration{
		{ActivityID: 1, UserID: 3, Name: "张三", Phone: "13900000001", VoucherNo: "GB20260816000001", Status: constants.RegistrationStatusRegistered, ReviewStatus: constants.ReviewStatusApproved},
		{ActivityID: 2, UserID: 3, Name: "张三", Phone: "13900000001", VoucherNo: "GB20260816000002", Status: constants.RegistrationStatusCheckedIn, ReviewStatus: constants.ReviewStatusApproved},
		{ActivityID: 3, UserID: 3, Name: "张三", Phone: "13900000001", VoucherNo: "GB20260816000003", Status: constants.RegistrationStatusRegistered, ReviewStatus: constants.ReviewStatusPending},
	}
	for i := range regs {
		if err := s.db.Create(&regs[i]).Error; err != nil {
			return err
		}
	}
	s.logger.Info("seed data created")
	return nil
}
