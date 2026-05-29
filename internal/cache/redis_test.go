package cache

import (
	"testing"

	"github.com/javaup/flashsale-system/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestInitRedis_NilConfig(t *testing.T) {
	rc, err := InitRedis(nil)
	assert.Error(t, err)
	assert.Nil(t, rc)
	assert.Contains(t, err.Error(), "config is nil")
}

func TestInitRedis_InvalidAddr(t *testing.T) {
	rc, err := InitRedis(&config.RedisConfig{
		Host: "invalid-host-that-does-not-exist",
		Port: 99999,
	})
	assert.Error(t, err)
	assert.Nil(t, rc)
}
