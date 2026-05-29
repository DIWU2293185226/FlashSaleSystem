package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestUserHandler_SendCode_MissingPhone(t *testing.T) {
	r := setupTestRouter()
	h := NewUserHandler(nil)
	r.POST("/user/code", h.SendCode)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/user/code", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.False(t, resp["success"].(bool))
}

func TestUserHandler_Login_InvalidBody(t *testing.T) {
	r := setupTestRouter()
	h := NewUserHandler(nil)
	r.POST("/user/login", h.Login)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/user/login", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_GetMe_Unauthorized(t *testing.T) {
	r := setupTestRouter()
	h := NewUserHandler(nil)
	r.GET("/user/me", h.GetMe)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/user/me", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUserHandler_GetByID_InvalidParam(t *testing.T) {
	r := setupTestRouter()
	h := NewUserHandler(nil)
	r.GET("/user/:id", h.GetByID)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/user/abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestShopHandler_GetByID_InvalidParam(t *testing.T) {
	r := setupTestRouter()
	h := NewShopHandler(nil)
	r.GET("/shop/:id", h.GetByID)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/shop/abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestVoucherHandler_GetByID_MissingParam(t *testing.T) {
	r := setupTestRouter()
	h := NewVoucherHandler(nil)
	r.POST("/voucher/get", h.GetByID)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/voucher/get", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestVoucherHandler_ListByShopID_InvalidParam(t *testing.T) {
	r := setupTestRouter()
	h := NewVoucherHandler(nil)
	r.GET("/voucher/list/:shopId", h.ListByShopID)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/voucher/list/abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestVoucherHandler_AddSeckill_InvalidBody(t *testing.T) {
	r := setupTestRouter()
	h := NewVoucherHandler(nil)
	r.POST("/voucher/seckill", h.AddSeckill)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/voucher/seckill", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code) // required fields missing
}

func TestBlogHandler_Create_Unauthorized(t *testing.T) {
	r := setupTestRouter()
	h := NewBlogHandler(nil)
	r.POST("/blog", h.Create)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/blog", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestBlogHandler_Like_Unauthorized(t *testing.T) {
	r := setupTestRouter()
	h := NewBlogHandler(nil)
	r.PUT("/blog/like/:id", h.Like)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/blog/like/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestBlogHandler_GetByID_InvalidParam(t *testing.T) {
	r := setupTestRouter()
	h := NewBlogHandler(nil)
	r.GET("/blog/:id", h.GetByID)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/blog/abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFollowHandler_Follow_Unauthorized(t *testing.T) {
	r := setupTestRouter()
	h := NewFollowHandler(nil)
	r.PUT("/follow/:id/:isFollow", h.Follow)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/follow/1/true", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestFollowHandler_IsFollowed_Unauthorized(t *testing.T) {
	r := setupTestRouter()
	h := NewFollowHandler(nil)
	r.GET("/follow/or/not/:id", h.IsFollowed)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/follow/or/not/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestFollowHandler_GetCommon_Unauthorized(t *testing.T) {
	r := setupTestRouter()
	h := NewFollowHandler(nil)
	r.GET("/follow/common/:id", h.GetCommon)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/follow/common/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUploadHandler_UploadBlog_NoFile(t *testing.T) {
	r := setupTestRouter()
	h := NewUploadHandler(nil)
	r.POST("/upload/blog", h.UploadBlog)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/upload/blog", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUploadHandler_DeleteBlog_MissingName(t *testing.T) {
	r := setupTestRouter()
	h := NewUploadHandler(nil)
	r.DELETE("/upload/blog/delete", h.DeleteBlog)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/upload/blog/delete", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_Logout(t *testing.T) {
	r := setupTestRouter()
	h := NewUserHandler(nil)
	r.POST("/user/logout", h.Logout)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/user/logout", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.True(t, resp["success"].(bool))
}

// ---------- Seckill handler tests ----------

func TestSeckillHandler_Seckill_Unauthorized(t *testing.T) {
	r := setupTestRouter()
	h := NewSeckillHandler(nil)
	r.POST("/voucher-order/seckill", h.Seckill)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/voucher-order/seckill", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSeckillHandler_CancelOrder_Unauthorized(t *testing.T) {
	r := setupTestRouter()
	h := NewSeckillHandler(nil)
	r.PUT("/voucher-order/cancel/:orderId", h.CancelOrder)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/voucher-order/cancel/1", nil)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSeckillHandler_GetOrder_InvalidParam(t *testing.T) {
	r := setupTestRouter()
	h := NewSeckillHandler(nil)
	r.GET("/voucher-order/:orderId", h.GetOrder)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/voucher-order/abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSeckillHandler_LoadStock_MissingBody(t *testing.T) {
	r := setupTestRouter()
	h := NewSeckillHandler(nil)
	r.POST("/voucher-order/load-stock", h.LoadStock)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/voucher-order/load-stock", nil)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSeckillHandler_GenerateToken_Unauthorized(t *testing.T) {
	r := setupTestRouter()
	h := NewSeckillHandler(nil)
	r.POST("/voucher-order/token", h.GenerateToken)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/voucher-order/token", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ---------- Subscribe handler tests ----------

func TestSubscribeHandler_Subscribe_Unauthorized(t *testing.T) {
	r := setupTestRouter()
	h := NewSubscribeHandler(nil)
	r.POST("/subscribe", h.Subscribe)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/subscribe", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSubscribeHandler_Unsubscribe_Unauthorized(t *testing.T) {
	r := setupTestRouter()
	h := NewSubscribeHandler(nil)
	r.DELETE("/subscribe/:voucherId", h.Unsubscribe)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/subscribe/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestReconcileHandler_CheckStock_InvalidParam(t *testing.T) {
	r := setupTestRouter()
	h := NewReconciliationHandler(nil)
	r.GET("/reconcile/check/:voucherId", h.CheckStock)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/reconcile/check/abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestReconcileHandler_FixStock_InvalidParam(t *testing.T) {
	r := setupTestRouter()
	h := NewReconciliationHandler(nil)
	r.POST("/reconcile/fix/:voucherId", h.FixStock)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/reconcile/fix/abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
