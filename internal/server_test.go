package internal

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/javaup/flashsale-system/internal/config"
	"github.com/javaup/flashsale-system/internal/handler"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewApp_SetupRouter_Ping(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: 8085,
			Mode: "test",
		},
		Database: config.DatabaseConfig{
			Driver:  "mysql",
			Sources: []config.DataSourceConfig{},
		},
		Redis: config.RedisConfig{
			Host: "127.0.0.1",
			Port: 6379,
		},
		Shard: config.ShardConfig{
			DbCount:    2,
			TableCount: 2,
		},
	}

	log := zerolog.Nop()
	app := &App{
		Config:        cfg,
		Log:           log,
		HealthHandler: handler.NewHealthHandler(),
	}

	r := app.SetupRouter()
	require.NotNil(t, r)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "pong")
}

func TestSetupRouter_CORS(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := zerolog.Nop()

	app := &App{
		Config: &config.Config{
			Server: config.ServerConfig{Mode: "test"},
		},
		Log: log,
	}
	app.HealthHandler = handler.NewHealthHandler()

	r := app.SetupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestSetupRouter_ModeRelease(t *testing.T) {
	log := zerolog.Nop()
	app := &App{
		Config: &config.Config{
			Server: config.ServerConfig{Mode: "release"},
		},
		Log: log,
	}
	_ = app.SetupRouter()
	assert.Equal(t, gin.ReleaseMode, gin.Mode())
}
