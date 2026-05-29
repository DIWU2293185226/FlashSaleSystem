// ═════════════════════════════════════════════════════════════════════
// 博客社交相关仓库 — tb_blog / tb_blog_comments / tb_follow（广播表，不分片）
// tb_blog 提供点赞计数的原子增减（IncrementLiked / DecrementLiked）
// tb_follow 提供关注关系查询（共同关注、关注列表、粉丝列表）
// ═════════════════════════════════════════════════════════════════════
package repository

import (
	"github.com/javaup/flashsale-system/internal/model"
	"gorm.io/gorm"
)

// BlogRepository tb_blog 表的数据访问（广播表，不分片）
type BlogRepository struct {
	dm *DatabaseManager
}

func NewBlogRepository(dm *DatabaseManager) *BlogRepository {
	return &BlogRepository{dm: dm}
}

// GetDB 返回广播数据库连接，供 service 层做关注推送 Feed 流的跨表查询
func (r *BlogRepository) GetDB() *gorm.DB {
	return r.dm.GetDB(broadcastDB)
}

func (r *BlogRepository) GetByID(id int64) (*model.Blog, error) {
	db := r.dm.GetDB(broadcastDB)
	var blog model.Blog
	err := db.First(&blog, id).Error
	return &blog, err
}

func (r *BlogRepository) Create(blog *model.Blog) error {
	db := r.dm.GetDB(broadcastDB)
	return db.Create(blog).Error
}

func (r *BlogRepository) Update(blog *model.Blog) error {
	db := r.dm.GetDB(broadcastDB)
	return db.Where("id = ?", blog.ID).Updates(blog).Error
}

func (r *BlogRepository) Delete(id int64) error {
	db := r.dm.GetDB(broadcastDB)
	return db.Delete(&model.Blog{}, id).Error
}

// ListByUserID 查询指定用户发布的博客列表，按时间倒序分页
func (r *BlogRepository) ListByUserID(userID int64, offset, limit int) ([]model.Blog, int64, error) {
	db := r.dm.GetDB(broadcastDB)
	var blogs []model.Blog
	var total int64
	query := db.Model(&model.Blog{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("create_time DESC").Offset(offset).Limit(limit).Find(&blogs).Error
	return blogs, total, err
}

// ListHot 查询热门博客，按点赞数(降序)+发布时间(降序)排序
// 热门排序策略：点赞数越多排越前，点赞数相同时最新的在前
func (r *BlogRepository) ListHot(offset, limit int) ([]model.Blog, int64, error) {
	db := r.dm.GetDB(broadcastDB)
	var blogs []model.Blog
	var total int64
	query := db.Model(&model.Blog{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("liked DESC, create_time DESC").Offset(offset).Limit(limit).Find(&blogs).Error
	return blogs, total, err
}

// IncrementLiked 点赞数 +1（原子更新，仅操作 liked 列，不会触发 ON UPDATE）
func (r *BlogRepository) IncrementLiked(id int64) error {
	db := r.dm.GetDB(broadcastDB)
	return db.Model(&model.Blog{}).Where("id = ?", id).
		UpdateColumn("liked", gorm.Expr("liked + 1")).Error
}

// DecrementLiked 点赞数 -1（原子更新，WHERE liked > 0 防止扣成负数）
func (r *BlogRepository) DecrementLiked(id int64) error {
	db := r.dm.GetDB(broadcastDB)
	return db.Model(&model.Blog{}).Where("id = ? AND liked > 0", id).
		UpdateColumn("liked", gorm.Expr("liked - 1")).Error
}

// BlogCommentsRepository tb_blog_comments 表的数据访问（广播表，不分片）
type BlogCommentsRepository struct {
	dm *DatabaseManager
}

func NewBlogCommentsRepository(dm *DatabaseManager) *BlogCommentsRepository {
	return &BlogCommentsRepository{dm: dm}
}

func (r *BlogCommentsRepository) Create(comment *model.BlogComments) error {
	db := r.dm.GetDB(broadcastDB)
	return db.Create(comment).Error
}

// ListByBlogID 查询某博客的评论列表，按发布时间倒序分页
func (r *BlogCommentsRepository) ListByBlogID(blogID int64, offset, limit int) ([]model.BlogComments, int64, error) {
	db := r.dm.GetDB(broadcastDB)
	var comments []model.BlogComments
	var total int64
	query := db.Model(&model.BlogComments{}).Where("blog_id = ?", blogID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("create_time DESC").Offset(offset).Limit(limit).Find(&comments).Error
	return comments, total, err
}

// FollowRepository tb_follow 表的数据访问（广播表，不分片）
// 关注关系是一种"多对多"的社交关系，通过 user_id = 关注者, follow_user_id = 被关注者 表达
type FollowRepository struct {
	dm *DatabaseManager
}

func NewFollowRepository(dm *DatabaseManager) *FollowRepository {
	return &FollowRepository{dm: dm}
}

func (r *FollowRepository) Create(follow *model.Follow) error {
	db := r.dm.GetDB(broadcastDB)
	return db.Create(follow).Error
}

// Delete 删除关注关系（取消关注）
func (r *FollowRepository) Delete(userID, followUserID int64) error {
	db := r.dm.GetDB(broadcastDB)
	return db.Where("user_id = ? AND follow_user_id = ?", userID, followUserID).
		Delete(&model.Follow{}).Error
}

// IsFollowing 判断 userID 是否已关注 targetID（用于"是否已关注"标记）
func (r *FollowRepository) IsFollowing(userID, targetID int64) (bool, error) {
	db := r.dm.GetDB(broadcastDB)
	var count int64
	err := db.Model(&model.Follow{}).
		Where("user_id = ? AND follow_user_id = ?", userID, targetID).
		Count(&count).Error
	return count > 0, err
}

// GetFollowees 查询 userID 关注的所有人（关注的博主列表）
func (r *FollowRepository) GetFollowees(userID int64) ([]model.Follow, error) {
	db := r.dm.GetDB(broadcastDB)
	var follows []model.Follow
	err := db.Where("user_id = ?", userID).Find(&follows).Error
	return follows, err
}

// GetFollowers 查询 userID 的所有粉丝
func (r *FollowRepository) GetFollowers(userID int64) ([]model.Follow, error) {
	db := r.dm.GetDB(broadcastDB)
	var follows []model.Follow
	err := db.Where("follow_user_id = ?", userID).Find(&follows).Error
	return follows, err
}

// GetCommon 查找两个用户的共同关注
// 通过 INNER JOIN 找出 userID 和 otherID 都关注的人
// 即：f1.user_id = userID AND f2.user_id = otherID AND f1.follow_user_id = f2.follow_user_id
func (r *FollowRepository) GetCommon(userID, otherID int64) ([]model.Follow, error) {
	db := r.dm.GetDB(broadcastDB)
	var common []model.Follow
	err := db.Raw(`
		SELECT f1.* FROM tb_follow f1
		INNER JOIN tb_follow f2 ON f1.follow_user_id = f2.follow_user_id
		WHERE f1.user_id = ? AND f2.user_id = ?
	`, userID, otherID).Scan(&common).Error
	return common, err
}
