// ═════════════════════════════════════════════════════════════════════
// 用户服务 — 登录/注册/签到
// 支持两种登录方式：验证码登录（自动注册新用户）、密码登录
// 签到使用 Redis Bitmap 存储，每月一个 key，BITFIELD 读取
// 连续签到天数通过位运算从今天往前遍历统计
// ═════════════════════════════════════════════════════════════════════
package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/javaup/flashsale-system/internal/cache"
	"github.com/javaup/flashsale-system/internal/dto"
	"github.com/javaup/flashsale-system/internal/jwt"
	"github.com/javaup/flashsale-system/internal/model"
	"github.com/javaup/flashsale-system/internal/repository"
	"github.com/javaup/flashsale-system/pkg"
)

var bg = context.Background()

// UserService 用户业务逻辑
// 整合 User/UserInfo/UserPhone 三个仓库的操作
type UserService struct {
	userRepo      *repository.UserRepository
	userInfoRepo  *repository.UserInfoRepository
	userPhoneRepo *repository.UserPhoneRepository
	redis         *cache.RedisCache
	jwtManager    *jwt.Manager
}

func NewUserService(
	userRepo *repository.UserRepository,
	userInfoRepo *repository.UserInfoRepository,
	userPhoneRepo *repository.UserPhoneRepository,
	redis *cache.RedisCache,
	jwtManager *jwt.Manager,
) *UserService {
	return &UserService{
		userRepo:      userRepo,
		userInfoRepo:  userInfoRepo,
		userPhoneRepo: userPhoneRepo,
		redis:         redis,
		jwtManager:    jwtManager,
	}
}

const codePrefix = "login:code:"

// SendCode 生成 6 位验证码并存入 Redis（2 分钟过期）
func (s *UserService) SendCode(phone string) *pkg.Result {
	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	if err := s.redis.Set(bg, codePrefix+phone, code, 120*time.Second); err != nil {
		return pkg.FailWithMsg("验证码发送失败")
	}
	return pkg.OKWithData(code)
}

// Login 登录入口：有密码则走密码登录，否则走验证码登录
func (s *UserService) Login(form *dto.LoginFormDTO) *pkg.Result {
	if form.Password != "" {
		return s.loginByPassword(form.Phone, form.Password)
	}
	return s.loginByCode(form.Phone, form.Code)
}

func (s *UserService) loginByPassword(phone, password string) *pkg.Result {
	up, err := s.userPhoneRepo.GetByPhone(phone)
	if err != nil {
		return pkg.FailWithMsg("手机号未注册")
	}
	user, err := s.userRepo.GetByID(up.UserID)
	if err != nil {
		return pkg.FailWithMsg("用户不存在")
	}
	if user.Password != password {
		return pkg.FailWithMsg("密码错误")
	}
	token, err := s.jwtManager.GenerateToken(user.ID, user.NickName, user.Icon)
	if err != nil {
		return pkg.FailWithMsg("登录失败")
	}
	return pkg.OKWithData(token)
}

// loginByCode 验证码登录：首次登录自动注册新用户
// 自动注册时创建 User/UserPhone/UserInfo 三条记录
func (s *UserService) loginByCode(phone, code string) *pkg.Result {
	stored, err := s.redis.Get(bg, codePrefix+phone)
	if err != nil || stored != code {
		return pkg.FailWithMsg("验证码错误")
	}
	_ = s.redis.Del(bg, codePrefix+phone)

	up, err := s.userPhoneRepo.GetByPhone(phone)
	var user *model.User
	if err != nil {
		// 新用户自动注册
		user = &model.User{
			Phone:    phone,
			NickName: "user_" + phone[len(phone)-4:],
		}
		if err := s.userRepo.Create(user); err != nil {
			return pkg.FailWithMsg("注册失败")
		}
		up = &model.UserPhone{
			UserID: user.ID,
			Phone:  phone,
		}
		_ = s.userPhoneRepo.Create(up)
		_ = s.userInfoRepo.Create(&model.UserInfo{
			UserID: user.ID,
		})
	} else {
		user, err = s.userRepo.GetByID(up.UserID)
		if err != nil {
			return pkg.FailWithMsg("用户不存在")
		}
	}

	token, err := s.jwtManager.GenerateToken(user.ID, user.NickName, user.Icon)
	if err != nil {
		return pkg.FailWithMsg("登录失败")
	}
	return pkg.OKWithData(token)
}

// GetByID 查询用户信息（返回 DTO，不含手机号密码等敏感字段）
func (s *UserService) GetByID(id int64) *pkg.Result {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return pkg.FailWithMsg("用户不存在")
	}
	return pkg.OKWithData(&dto.UserDTO{
		ID:       user.ID,
		NickName: user.NickName,
		Icon:     user.Icon,
	})
}

// GetUserInfo 查询用户详细信息（含城市、简介、等级等）
func (s *UserService) GetUserInfo(userID int64) *pkg.Result {
	info, err := s.userInfoRepo.GetByUserID(userID)
	if err != nil {
		return pkg.FailWithMsg("用户信息不存在")
	}
	user, _ := s.userRepo.GetByID(userID)
	result := map[string]interface{}{
		"id":        userID,
		"nickName":  user.NickName,
		"icon":      user.Icon,
		"city":      info.City,
		"introduce": info.Introduce,
		"fans":      info.Fans,
		"followee":  info.Followee,
		"gender":    info.Gender,
		"birthday":  info.Birthday,
		"credits":   info.Credits,
		"level":     info.Level,
	}
	return pkg.OKWithData(result)
}

// Sign 签到：使用 Redis Bitmap 记录，key 格式 sign:{userID}:{YYYYMM}
// 每天的签到状态对应一个 bit 位
func (s *UserService) Sign(userID int64) *pkg.Result {
	now := time.Now()
	key := fmt.Sprintf("sign:%d:%s", userID, now.Format("200601"))
	day := now.Day()
	_, err := s.redis.Client.SetBit(bg, key, int64(day-1), 1).Result()
	if err != nil {
		return pkg.FailWithMsg("签到失败")
	}
	return pkg.OK()
}

// SignCount 统计当月连续签到天数
// 使用 BITFIELD 一次性读取整个月的 bitmap 为 uint
// 从今天往前遍历，遇到第一个未签到的 bit 即停止
func (s *UserService) SignCount(userID int64) *pkg.Result {
	now := time.Now()
	key := fmt.Sprintf("sign:%d:%s", userID, now.Format("200601"))

	val, err := s.redis.Client.BitField(bg, key, "GET", "u31", 0).Result()
	if err != nil || len(val) == 0 {
		return pkg.OKWithData(0)
	}
	bits := uint64(val[0])
	count := 0
	today := now.Day()
	for i := today - 1; i >= 0; i-- {
		if bits&(1<<i) != 0 {
			count++
		} else {
			break
		}
	}
	return pkg.OKWithData(count)
}
