// ═════════════════════════════════════════════════════════════════════
// 关注服务 — 关注/取关/共同关注
// ═════════════════════════════════════════════════════════════════════
package service

import (
	"github.com/javaup/flashsale-system/internal/dto"
	"github.com/javaup/flashsale-system/internal/model"
	"github.com/javaup/flashsale-system/internal/repository"
	"github.com/javaup/flashsale-system/pkg"
)

// FollowService 关注业务逻辑
type FollowService struct {
	followRepo *repository.FollowRepository
	userRepo   *repository.UserRepository
}

func NewFollowService(followRepo *repository.FollowRepository, userRepo *repository.UserRepository) *FollowService {
	return &FollowService{followRepo: followRepo, userRepo: userRepo}
}

// Follow 关注/取关切换
// isFollow=true: 关注, false: 取消关注
// 不允许关注自己
func (s *FollowService) Follow(userID, followUserID int64, isFollow bool) *pkg.Result {
	if userID == followUserID {
		return pkg.FailWithMsg("不能关注自己")
	}
	if isFollow {
		follow := &model.Follow{
			UserID:       userID,
			FollowUserID: followUserID,
		}
		if err := s.followRepo.Create(follow); err != nil {
			return pkg.FailWithMsg("关注失败")
		}
	} else {
		if err := s.followRepo.Delete(userID, followUserID); err != nil {
			return pkg.FailWithMsg("取消关注失败")
		}
	}
	return pkg.OK()
}

// IsFollowed 检查是否已关注
func (s *FollowService) IsFollowed(userID, targetID int64) *pkg.Result {
	following, err := s.followRepo.IsFollowing(userID, targetID)
	if err != nil {
		return pkg.FailWithMsg("查询失败")
	}
	return pkg.OKWithData(following)
}

// GetCommon 查找两个用户的共同关注
func (s *FollowService) GetCommon(userID, targetID int64) *pkg.Result {
	common, err := s.followRepo.GetCommon(userID, targetID)
	if err != nil {
		return pkg.FailWithMsg("查询共同关注失败")
	}
	var users []dto.UserDTO
	for _, f := range common {
		user, err := s.userRepo.GetByID(f.FollowUserID)
		if err != nil {
			continue
		}
		users = append(users, dto.UserDTO{
			ID:       user.ID,
			NickName: user.NickName,
			Icon:     user.Icon,
		})
	}
	if users == nil {
		users = []dto.UserDTO{}
	}
	return pkg.OKWithData(users)
}
