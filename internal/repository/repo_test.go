package repository

import (
	"testing"

	"github.com/javaup/flashsale-system/internal/config"
	"github.com/javaup/flashsale-system/internal/sharding"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func testRouter() *sharding.Router {
	return sharding.NewRouter(&config.ShardConfig{DbCount: 2, TableCount: 2})
}

func TestNewUserRepository(t *testing.T) {
	dm := &DatabaseManager{Sources: map[string]*gorm.DB{}}
	repo := NewUserRepository(dm, testRouter())
	assert.NotNil(t, repo)
}

func TestNewVoucherRepository(t *testing.T) {
	dm := &DatabaseManager{Sources: map[string]*gorm.DB{}}
	repo := NewVoucherRepository(dm, testRouter())
	assert.NotNil(t, repo)
}

func TestNewSeckillVoucherRepository(t *testing.T) {
	dm := &DatabaseManager{Sources: map[string]*gorm.DB{}}
	repo := NewSeckillVoucherRepository(dm, testRouter())
	assert.NotNil(t, repo)
}

func TestNewVoucherOrderRepository(t *testing.T) {
	dm := &DatabaseManager{Sources: map[string]*gorm.DB{}}
	repo := NewVoucherOrderRepository(dm, testRouter())
	assert.NotNil(t, repo)
}

func TestNewUserInfoRepository(t *testing.T) {
	dm := &DatabaseManager{Sources: map[string]*gorm.DB{}}
	repo := NewUserInfoRepository(dm, testRouter())
	assert.NotNil(t, repo)
}

func TestNewUserPhoneRepository(t *testing.T) {
	dm := &DatabaseManager{Sources: map[string]*gorm.DB{}}
	repo := NewUserPhoneRepository(dm, testRouter())
	assert.NotNil(t, repo)
}

func TestNewShopRepository(t *testing.T) {
	dm := &DatabaseManager{Sources: map[string]*gorm.DB{}}
	repo := NewShopRepository(dm, testRouter())
	assert.NotNil(t, repo)
}

func TestNewShopTypeRepository(t *testing.T) {
	dm := &DatabaseManager{Sources: map[string]*gorm.DB{}}
	repo := NewShopTypeRepository(dm)
	assert.NotNil(t, repo)
}

func TestNewBlogRepository(t *testing.T) {
	dm := &DatabaseManager{Sources: map[string]*gorm.DB{}}
	repo := NewBlogRepository(dm)
	assert.NotNil(t, repo)
}

func TestNewBlogCommentsRepository(t *testing.T) {
	dm := &DatabaseManager{Sources: map[string]*gorm.DB{}}
	repo := NewBlogCommentsRepository(dm)
	assert.NotNil(t, repo)
}

func TestNewFollowRepository(t *testing.T) {
	dm := &DatabaseManager{Sources: map[string]*gorm.DB{}}
	repo := NewFollowRepository(dm)
	assert.NotNil(t, repo)
}

func TestNewRollbackFailureLogRepository(t *testing.T) {
	dm := &DatabaseManager{Sources: map[string]*gorm.DB{}}
	repo := NewRollbackFailureLogRepository(dm)
	assert.NotNil(t, repo)
}
