// ═════════════════════════════════════════════════════════════════════
// 用户相关仓库 — tb_user / tb_user_info / tb_user_phone 的分片 CRUD
// 三个表使用不同的分片键：user 按 id 分片，user_info 按 user_id 分片，
// user_phone 按 phone 哈希分片
// ═════════════════════════════════════════════════════════════════════
package repository

import (
	"github.com/javaup/flashsale-system/internal/model"
	"github.com/javaup/flashsale-system/internal/sharding"
)

// UserRepository tb_user 表的数据访问
// 按用户 ID 分片（分片键 = id）
type UserRepository struct {
	dm     *DatabaseManager
	router *sharding.Router
}

func NewUserRepository(dm *DatabaseManager, router *sharding.Router) *UserRepository {
	return &UserRepository{dm: dm, router: router}
}

// GetByID 按用户 ID 查询，根据 ID 路由到对应的分片
func (r *UserRepository) GetByID(id int64) (*model.User, error) {
	db := r.dm.GetDB(r.router.UserDB(id))
	table := r.router.UserTable(id)
	var user model.User
	err := db.Table(table).Where("id = ?", id).First(&user).Error
	return &user, err
}

// Create 新增用户，写入 ID 对应的分片
func (r *UserRepository) Create(user *model.User) error {
	db := r.dm.GetDB(r.router.UserDB(user.ID))
	table := r.router.UserTable(user.ID)
	return db.Table(table).Create(user).Error
}

// Update 更新用户信息，按 ID 路由到分片
func (r *UserRepository) Update(user *model.User) error {
	db := r.dm.GetDB(r.router.UserDB(user.ID))
	table := r.router.UserTable(user.ID)
	return db.Table(table).Where("id = ?", user.ID).Updates(user).Error
}

// UserInfoRepository tb_user_info 表的数据访问
// 按用户 ID 分片（分片键 = user_id）
type UserInfoRepository struct {
	dm     *DatabaseManager
	router *sharding.Router
}

func NewUserInfoRepository(dm *DatabaseManager, router *sharding.Router) *UserInfoRepository {
	return &UserInfoRepository{dm: dm, router: router}
}

func (r *UserInfoRepository) GetByUserID(userID int64) (*model.UserInfo, error) {
	db := r.dm.GetDB(r.router.UserInfoDB(userID))
	table := r.router.UserInfoTable(userID)
	var info model.UserInfo
	err := db.Table(table).Where("user_id = ?", userID).First(&info).Error
	return &info, err
}

func (r *UserInfoRepository) Create(info *model.UserInfo) error {
	db := r.dm.GetDB(r.router.UserInfoDB(info.UserID))
	table := r.router.UserInfoTable(info.UserID)
	return db.Table(table).Create(info).Error
}

func (r *UserInfoRepository) Update(info *model.UserInfo) error {
	db := r.dm.GetDB(r.router.UserInfoDB(info.UserID))
	table := r.router.UserInfoTable(info.UserID)
	return db.Table(table).Where("user_id = ?", info.UserID).Updates(info).Error
}

// UserPhoneRepository tb_user_phone 表的数据访问
// 按手机号哈希分片（分片键 = phone 字符串的 hash）
// 登录时通过手机号定位分片查询用户，不走 ID 路由
type UserPhoneRepository struct {
	dm     *DatabaseManager
	router *sharding.Router
}

func NewUserPhoneRepository(dm *DatabaseManager, router *sharding.Router) *UserPhoneRepository {
	return &UserPhoneRepository{dm: dm, router: router}
}

// GetByPhone 按手机号查询用户的 phone → user_id 映射关系
// 使用 hash 分片，确保相同手机号落到同一分片
func (r *UserPhoneRepository) GetByPhone(phone string) (*model.UserPhone, error) {
	db := r.dm.GetDB(r.router.UserPhoneDB(phone))
	table := r.router.UserPhoneTable(phone)
	var up model.UserPhone
	err := db.Table(table).Where("phone = ?", phone).First(&up).Error
	return &up, err
}

func (r *UserPhoneRepository) Create(up *model.UserPhone) error {
	db := r.dm.GetDB(r.router.UserPhoneDB(up.Phone))
	table := r.router.UserPhoneTable(up.Phone)
	return db.Table(table).Create(up).Error
}
