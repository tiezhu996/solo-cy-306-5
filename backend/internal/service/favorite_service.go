package service

import (
	"log/slog"

	"gbevent/internal/constants"
	"gbevent/internal/model"
	"gbevent/internal/repository"
	"gbevent/internal/util"
)

// FavoriteService 收藏业务逻辑。
type FavoriteService struct {
	repo        *repository.FavoriteRepository
	activitySvc *ActivityService
	logger      *slog.Logger
}

// NewFavoriteService 构造收藏服务。
func NewFavoriteService(repo *repository.FavoriteRepository, activitySvc *ActivityService, logger *slog.Logger) *FavoriteService {
	return &FavoriteService{repo: repo, activitySvc: activitySvc, logger: logger}
}

// Add 收藏活动。
func (s *FavoriteService) Add(userID, activityID uint64) (*model.Favorite, error) {
	if _, _, err := s.activitySvc.Get(activityID); err != nil {
		return nil, err
	}
	exists, err := s.repo.Exists(userID, activityID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, util.NewAppError(constants.CodeConflict, "Favorite[user_id="+itoa(userID)+",activity_id="+itoa(activityID)+"] already favorited")
	}
	f := &model.Favorite{UserID: userID, ActivityID: activityID}
	if err := s.repo.Create(f); err != nil {
		return nil, util.Wrap(err, "Favorite[user_id=%d,activity_id=%d] create failed", userID, activityID)
	}
	s.logger.Info(constants.LogFavoriteAddSuccess, "favorite_id", f.ID)
	return f, nil
}

// Remove 取消收藏。
func (s *FavoriteService) Remove(userID, activityID uint64) error {
	if err := s.repo.DeleteByUserActivity(userID, activityID); err != nil {
		return util.Wrap(err, "Favorite[user_id=%d,activity_id=%d] remove failed", userID, activityID)
	}
	s.logger.Info(constants.LogFavoriteRemoveSuccess, "user_id", userID, "activity_id", activityID)
	return nil
}

// ListMine 查询我的收藏。
func (s *FavoriteService) ListMine(userID uint64, page, pageSize int) ([]model.Favorite, int64, error) {
	return s.repo.ListByUser(userID, page, pageSize)
}

// IsFavorited 判断是否已收藏。
func (s *FavoriteService) IsFavorited(userID, activityID uint64) (bool, error) {
	return s.repo.Exists(userID, activityID)
}
