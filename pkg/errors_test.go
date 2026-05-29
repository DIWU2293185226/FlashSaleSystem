package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBaseCodeError(t *testing.T) {
	err := NewBaseCodeError(10001, "测试")
	assert.Equal(t, 10001, err.Code)
	assert.Equal(t, "测试", err.Msg)
	assert.Equal(t, "测试", err.Error())
}

func TestGetBaseCodeKnown(t *testing.T) {
	err := GetBaseCode(10001)
	assert.Equal(t, ErrSeckillVoucherNotExist, err)

	err = GetBaseCode(0)
	assert.Equal(t, ErrSuccess, err)

	err = GetBaseCode(20001)
	assert.Equal(t, ErrUserAlreadyPurchase, err)
}

func TestGetBaseCodeUnknown(t *testing.T) {
	err := GetBaseCode(99999)
	assert.Equal(t, 99999, err.Code)
	assert.Contains(t, err.Error(), "未知错误码")
}

func TestBaseCodeErrorInterface(t *testing.T) {
	var e error = ErrSeckillVoucherNotExist
	assert.Equal(t, "秒杀优惠券不存在", e.Error())
}
