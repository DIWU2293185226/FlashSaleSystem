package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeTestConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yaml")
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)
	return path
}

func TestLoadDefaultConfig(t *testing.T) {
	// Use the project's real config file as reference
	cfg, err := Load("../../resource/config.yaml")
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, 8085, cfg.Server.Port)
	assert.Equal(t, "debug", cfg.Server.Mode)
	assert.Len(t, cfg.Database.Sources, 2)
	assert.Equal(t, 2, cfg.Shard.DbCount)
	assert.Equal(t, 2, cfg.Shard.TableCount)
}

func TestLoadDataSourceDSN(t *testing.T) {
	ds := &DataSourceConfig{
		User:     "root",
		Password: "123456",
		Host:     "127.0.0.1",
		Port:     3308,
		Database: "hmdp_0",
	}
	expected := "root:123456@tcp(127.0.0.1:3308)/hmdp_0?charset=utf8mb4&parseTime=True&loc=Local"
	assert.Equal(t, expected, ds.DSN())
}

func TestLoadRedisAddr(t *testing.T) {
	rc := &RedisConfig{Host: "127.0.0.1", Port: 6379}
	assert.Equal(t, "127.0.0.1:6379", rc.Addr())
}

func TestLoadMinimalConfig(t *testing.T) {
	content := `
server:
  port: 9090
database:
  driver: mysql
  sources:
    - name: hmdp_0
      host: 127.0.0.1
      port: 3306
      user: test
      password: test
      database: hmdp_0
redis:
  host: 127.0.0.1
  port: 6379
kafka:
  brokers:
    - localhost:9092
`
	path := writeTestConfig(t, content)
	cfg, err := Load(path)
	require.NoError(t, err)
	assert.Equal(t, 9090, cfg.Server.Port)
	assert.Equal(t, "127.0.0.1:6379", cfg.Redis.Addr())
	assert.Equal(t, 2, cfg.Shard.DbCount) // default
}

func TestLoadFileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/config.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read config file")
}

func TestMustLoad(t *testing.T) {
	assert.NotPanics(t, func() {
		cfg := MustLoad("../../resource/config.yaml")
		assert.NotNil(t, cfg)
	})

	assert.Panics(t, func() {
		MustLoad("/nonexistent.yaml")
	})
}
