// ═════════════════════════════════════════════════════════════════════
// 博客社交服务 — 发布/点赞/热门/关注推送
// 点赞使用 Redis Set 做去重（SIsMember 判断 + SAdd/SRem 切换），DB 存计数兜底
// 关注推送使用 scroll 分页，避免传统分页在数据频繁插入时的重复/跳变问题
// ═════════════════════════════════════════════════════════════════════
package service

import (
	"strconv"

	"github.com/javaup/flashsale-system/internal/cache"
	"github.com/javaup/flashsale-system/internal/dto"
	"github.com/javaup/flashsale-system/internal/model"
	"github.com/javaup/flashsale-system/internal/repository"
	"github.com/javaup/flashsale-system/pkg"
)

// BlogService 博客社交业务逻辑
type BlogService struct {
	blogRepo   *repository.BlogRepository
	userRepo   *repository.UserRepository
	followRepo *repository.FollowRepository
	redis      *cache.RedisCache
}

func NewBlogService(
	blogRepo *repository.BlogRepository,
	userRepo *repository.UserRepository,
	followRepo *repository.FollowRepository,
	redis *cache.RedisCache,
) *BlogService {
	return &BlogService{
		blogRepo:   blogRepo,
		userRepo:   userRepo,
		followRepo: followRepo,
		redis:      redis,
	}
}

const blogLikedKey = "blog:liked:"

// Create 发布博客
func (s *BlogService) Create(blog *model.Blog, userID int64) *pkg.Result {
	blog.UserID = userID
	if err := s.blogRepo.Create(blog); err != nil {
		return pkg.FailWithMsg("发布博客失败")
	}
	return pkg.OKWithData(blog.ID)
}

// GetByID 博客详情（含博主信息、当前用户是否点赞）
func (s *BlogService) GetByID(id int64, currentUserID int64) *pkg.Result {
	blog, err := s.blogRepo.GetByID(id)
	if err != nil {
		return pkg.FailWithMsg("博客不存在")
	}
	user, _ := s.userRepo.GetByID(blog.UserID)
	blog.Icon = user.Icon
	blog.Name = user.NickName
	if currentUserID > 0 {
		isMember, _ := s.redis.SIsMember(bg, blogLikedKey+strconv.FormatInt(id, 10), currentUserID)
		blog.IsLike = isMember
	}
	return pkg.OKWithData(blog)
}

// Like 点赞/取消点赞切换（Redis Set 去重 + DB 计数兜底）
func (s *BlogService) Like(id int64, userID int64) *pkg.Result {
	key := blogLikedKey + strconv.FormatInt(id, 10)
	isMember, _ := s.redis.SIsMember(bg, key, userID)
	if isMember {
		_, _ = s.redis.SRem(bg, key, userID)
		_ = s.blogRepo.DecrementLiked(id)
	} else {
		_, _ = s.redis.SAdd(bg, key, userID)
		_ = s.blogRepo.IncrementLiked(id)
	}
	return pkg.OK()
}

// ListLikes 查询点赞用户列表
func (s *BlogService) ListLikes(id int64) *pkg.Result {
	key := blogLikedKey + strconv.FormatInt(id, 10)
	userIDs, err := s.redis.Client.SMembers(bg, key).Result()
	if err != nil {
		return pkg.FailWithMsg("查询点赞失败")
	}
	var users []dto.UserDTO
	for _, uidStr := range userIDs {
		uid, _ := strconv.ParseInt(uidStr, 10, 64)
		user, err := s.userRepo.GetByID(uid)
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

// ListHot 热门博客列表（按点赞数排序）
func (s *BlogService) ListHot(current int) *pkg.Result {
	if current <= 0 {
		current = 1
	}
	offset := (current - 1) * pkg.MaxPageSize
	blogs, total, err := s.blogRepo.ListHot(offset, pkg.MaxPageSize)
	if err != nil {
		return pkg.FailWithMsg("查询热门博客失败")
	}
	s.enrichBlogs(blogs)
	return pkg.OKWithDataTotal(blogs, total)
}

// ListByUserID 查询指定用户的博客列表
func (s *BlogService) ListByUserID(userID int64, current int) *pkg.Result {
	if current <= 0 {
		current = 1
	}
	offset := (current - 1) * pkg.MaxPageSize
	blogs, total, err := s.blogRepo.ListByUserID(userID, offset, pkg.MaxPageSize)
	if err != nil {
		return pkg.FailWithMsg("查询用户博客失败")
	}
	s.enrichBlogs(blogs)
	return pkg.OKWithDataTotal(blogs, total)
}

// ListFollowBlog 关注推送 Feed 流
// 查出当前用户关注的所有人 → 按时间倒序拉取他们的博客
// 支持 scroll 分页（通过 lastID 游标）
func (s *BlogService) ListFollowBlog(userID int64, lastID int64, offset int) *pkg.Result {
	follows, err := s.followRepo.GetFollowees(userID)
	if err != nil || len(follows) == 0 {
		return pkg.OKWithData([]model.Blog{})
	}
	userIDs := make([]int64, 0, len(follows))
	for _, f := range follows {
		userIDs = append(userIDs, f.FollowUserID)
	}
	db := s.blogRepo.GetDB()
	var blogs []model.Blog
	query := db.Where("user_id IN ?", userIDs).Order("create_time DESC")
	if lastID > 0 {
		query = query.Where("id < ?", lastID)
	}
	if err := query.Limit(pkg.MaxPageSize).Offset(offset).Find(&blogs).Error; err != nil {
		return pkg.FailWithMsg("查询关注博客失败")
	}
	s.enrichBlogs(blogs)
	return pkg.OKWithData(blogs)
}

// enrichBlogs 填充博客列表的作者信息（头像、昵称）
func (s *BlogService) enrichBlogs(blogs []model.Blog) {
	for i := range blogs {
		user, err := s.userRepo.GetByID(blogs[i].UserID)
		if err != nil {
			continue
		}
		blogs[i].Icon = user.Icon
		blogs[i].Name = user.NickName
	}
}
