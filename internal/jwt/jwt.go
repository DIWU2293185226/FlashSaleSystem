// ═════════════════════════════════════════════════════════════════════
// JWT 令牌管理 — 用户认证凭证
// 使用 HS256 签名算法，令牌中嵌入 userId/nickName/icon 供各业务直接使用
// 过期时间通过配置控制，默认 72 小时
// ═════════════════════════════════════════════════════════════════════
package jwt

import (
	"fmt"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/javaup/flashsale-system/internal/config"
)

// UserClaims JWT 令牌中携带的用户信息
// 减少每次请求都查数据库的频率，前端也可直接读取昵称和头像
type UserClaims struct {
	UserID   int64  `json:"userId"`
	NickName string `json:"nickName"`
	Icon     string `json:"icon"`
	jwtv5.RegisteredClaims
}

// Manager JWT 管理器，负责令牌的签发和验证
type Manager struct {
	secret   []byte // 签名密钥，配置为空时使用默认密钥
	expHours int    // 过期时间（小时）
}

// NewManager 根据配置创建 JWT 管理器
// 密钥和过期时间都有默认值兜底，防止配置遗漏导致服务异常
func NewManager(cfg *config.JWTConfig) *Manager {
	secret := cfg.Secret
	if secret == "" {
		secret = "flashsale-default-secret"
	}
	expHours := cfg.ExpirationHours
	if expHours <= 0 {
		expHours = 72
	}
	return &Manager{secret: []byte(secret), expHours: expHours}
}

// GenerateToken 为用户生成签名的 JWT 令牌
// 携带 userId/nickName/icon，供后续请求认证使用
func (m *Manager) GenerateToken(userID int64, nickName, icon string) (string, error) {
	claims := UserClaims{
		UserID:   userID,
		NickName: nickName,
		Icon:     icon,
		RegisteredClaims: jwtv5.RegisteredClaims{
			ExpiresAt: jwtv5.NewNumericDate(time.Now().Add(time.Duration(m.expHours) * time.Hour)),
			IssuedAt:  jwtv5.NewNumericDate(time.Now()),
		},
	}
	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// ParseToken 验证令牌并解析出用户 Claims
// 校验签名算法是否为 HMAC 系列，防止攻击者使用其他算法伪造令牌
func (m *Manager) ParseToken(tokenStr string) (*UserClaims, error) {
	token, err := jwtv5.ParseWithClaims(tokenStr, &UserClaims{}, func(token *jwtv5.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwtv5.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}
