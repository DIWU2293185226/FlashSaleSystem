package jwt

import (
	"testing"

	"github.com/javaup/flashsale-system/internal/config"
	"github.com/stretchr/testify/assert"
)

func testManager() *Manager {
	return NewManager(&config.JWTConfig{
		Secret:          "test-secret",
		ExpirationHours: 72,
	})
}

func TestGenerateToken(t *testing.T) {
	m := testManager()
	token, err := m.GenerateToken(1001, "testuser", "icon.png")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestParseToken_Valid(t *testing.T) {
	m := testManager()
	token, _ := m.GenerateToken(1001, "testuser", "icon.png")

	claims, err := m.ParseToken(token)
	assert.NoError(t, err)
	assert.Equal(t, int64(1001), claims.UserID)
	assert.Equal(t, "testuser", claims.NickName)
	assert.Equal(t, "icon.png", claims.Icon)
}

func TestParseToken_Invalid(t *testing.T) {
	m := testManager()
	_, err := m.ParseToken("invalid-token")
	assert.Error(t, err)
}

func TestParseToken_WrongSecret(t *testing.T) {
	m1 := testManager()
	token, _ := m1.GenerateToken(1001, "user", "icon.png")

	m2 := NewManager(&config.JWTConfig{Secret: "different-secret"})
	_, err := m2.ParseToken(token)
	assert.Error(t, err)
}

func TestParseToken_DifferentUsers(t *testing.T) {
	m := testManager()
	token1, _ := m.GenerateToken(1001, "user1", "icon1.png")
	token2, _ := m.GenerateToken(1002, "user2", "icon2.png")

	claims1, _ := m.ParseToken(token1)
	assert.Equal(t, int64(1001), claims1.UserID)
	assert.Equal(t, "user1", claims1.NickName)

	claims2, _ := m.ParseToken(token2)
	assert.Equal(t, int64(1002), claims2.UserID)
	assert.Equal(t, "user2", claims2.NickName)
}

func TestNewManager_Defaults(t *testing.T) {
	m := NewManager(&config.JWTConfig{})
	assert.NotNil(t, m)
	assert.Equal(t, []byte("flashsale-default-secret"), m.secret)
	assert.Equal(t, 72, m.expHours)
}

func TestNewManager_CustomConfig(t *testing.T) {
	m := NewManager(&config.JWTConfig{
		Secret:          "custom",
		ExpirationHours: 24,
	})
	assert.NotNil(t, m)
	token, err := m.GenerateToken(1, "u", "i")
	assert.NoError(t, err)
	claims, err := m.ParseToken(token)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), claims.UserID)
}
