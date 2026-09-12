package service

import (
	"log/slog"

	"gbevent/internal/constants"
	"gbevent/internal/repository"
	"gbevent/internal/util"
)

// NotificationService 通知业务逻辑。
type NotificationService struct {
	repo   *repository.NotificationRepository
	logger *slog.Logger
}

// NewNotificationService 构造通知服务。
func NewNotificationService(repo *repository.NotificationRepository, logger *slog.Logger) *NotificationService {
	return &NotificationService{repo: repo, logger: logger}
}

// ListMine 查询我的通知。
func (s *NotificationService) ListMine(userID uint64, page, pageSize int) ([]any, int64, error) {
	list, total, err := s.repo.ListByUser(userID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	out := make([]any, 0, len(list))
	for i := range list {
		n := list[i]
		out = append(out, map[string]any{
			"id": n.ID, "user_id": n.UserID, "notification_type": n.NotificationType,
			"title": n.Title, "content": n.Content, "is_read": n.IsRead, "created_at": n.CreatedAt,
			"type_text": util.NotificationTypeText(n.NotificationType),
		})
	}
	return out, total, nil
}

// MarkRead 标记单条已读。
func (s *NotificationService) MarkRead(id, userID uint64) error {
	if err := s.repo.MarkRead(id, userID); err != nil {
		return util.Wrap(err, "Notification[id=%d,user_id=%d] mark read failed", id, userID)
	}
	s.logger.Info(constants.LogNotificationMarkRead, "notification_id", id)
	return nil
}

// MarkAllRead 全部已读。
func (s *NotificationService) MarkAllRead(userID uint64) error {
	if err := s.repo.MarkAllRead(userID); err != nil {
		return util.Wrap(err, "Notification[user_id=%d] mark all read failed", userID)
	}
	s.logger.Info(constants.LogNotificationMarkRead, "user_id", userID, "all", true)
	return nil
}
