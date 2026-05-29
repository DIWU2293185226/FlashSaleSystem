package internal

import (
	"fmt"
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

func TestSetupRouter_AllRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := zerolog.Nop()

	// Create a minimal app with nil-dependent handlers for route testing
	app := &App{
		Config: &config.Config{
			Server: config.ServerConfig{Mode: "test"},
			JWT:    config.JWTConfig{Secret: "test", ExpirationHours: 72},
		},
		Log:              log,
		HealthHandler:    handler.NewHealthHandler(),
		UserHandler:      handler.NewUserHandler(nil),
		ShopHandler:      handler.NewShopHandler(nil),
		ShopTypeHandler:  handler.NewShopTypeHandler(nil),
		VoucherHandler:   handler.NewVoucherHandler(nil),
		BlogHandler:      handler.NewBlogHandler(nil),
		FollowHandler:    handler.NewFollowHandler(nil),
		UploadHandler:    handler.NewUploadHandler(nil),
		SeckillHandler:   handler.NewSeckillHandler(nil),
		SubscribeHandler: handler.NewSubscribeHandler(nil),
		ReconcileHandler: handler.NewReconciliationHandler(nil),
	}

	r := app.SetupRouter()
	require.NotNil(t, r)

	// Verify health check
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test all registered routes
	routes := []struct {
		method string
		path   string
	}{
		// User routes
		{"POST", "/user/code"},
		{"POST", "/user/login"},
		{"POST", "/user/logout"},
		{"GET", "/user/me"},
		{"GET", "/user/1"},
		{"GET", "/user/info/1"},
		{"POST", "/user/sign"},
		{"GET", "/user/sign/count"},
		// Shop routes
		{"GET", "/shop/1"},
		{"POST", "/shop"},
		{"PUT", "/shop"},
		{"GET", "/shop/of/type"},
		{"GET", "/shop/of/name"},
		{"GET", "/shop/of/nearby"},
		// Shop type routes
		{"GET", "/shop-type/list"},
		// Voucher routes
		{"POST", "/voucher"},
		{"POST", "/voucher/seckill"},
		{"POST", "/voucher/get"},
		{"GET", "/voucher/list/1"},
		{"POST", "/voucher/update/seckill"},
		{"POST", "/voucher/update/seckill/stock"},
		// Blog routes
		{"POST", "/blog"},
		{"PUT", "/blog/like/1"},
		{"GET", "/blog/hot"},
		{"GET", "/blog/1"},
		{"GET", "/blog/likes/1"},
		{"GET", "/blog/of/user"},
		{"GET", "/blog/of/me"},
		{"GET", "/blog/of/follow"},
		// Follow routes
		{"PUT", "/follow/1/true"},
		{"GET", "/follow/or/not/1"},
		{"GET", "/follow/common/1"},
		// Upload routes
		{"POST", "/upload/blog"},
		{"DELETE", "/upload/blog/delete"},
		// Seckill voucher-order routes (Phase 3)
		{"POST", "/voucher-order/seckill"},
		{"GET", "/voucher-order/1"},
		{"PUT", "/voucher-order/cancel/1"},
		{"POST", "/voucher-order/load-stock"},
		{"GET", "/voucher-order/stock/1"},
		{"GET", "/voucher-order/voucher/1"},
		{"POST", "/voucher-order/token"},
		// Subscribe routes (Phase 4)
		{"POST", "/subscribe"},
		{"DELETE", "/subscribe/1"},
		{"GET", "/subscribe/status/1"},
		{"GET", "/subscribe/history"},
		// Reconcile routes (Phase 4)
		{"GET", "/reconcile/check/1"},
		{"POST", "/reconcile/fix/1"},
	}

	for _, route := range routes {
		t.Run(fmt.Sprintf("%s %s", route.method, route.path), func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(route.method, route.path, nil)
			// Set content-type for POST/PUT requests
			if route.method == "POST" || route.method == "PUT" {
				req.Header.Set("Content-Type", "application/json")
			}
			r.ServeHTTP(w, req)
			// Route should not return 404 (it might return 401 or 400, but not 404)
			assert.NotEqual(t, http.StatusNotFound, w.Code,
				"route %s %s should be registered", route.method, route.path)
		})
	}
}
