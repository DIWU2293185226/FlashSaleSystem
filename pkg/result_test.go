package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOK(t *testing.T) {
	r := OK()
	assert.True(t, r.Success)
	assert.Empty(t, r.ErrorMsg)
	assert.Nil(t, r.Data)
	assert.Zero(t, r.Total)
}

func TestOKWithData(t *testing.T) {
	r := OKWithData("hello")
	assert.True(t, r.Success)
	assert.Equal(t, "hello", r.Data)
}

func TestOKWithDataTotal(t *testing.T) {
	r := OKWithDataTotal([]int{1, 2, 3}, 100)
	assert.True(t, r.Success)
	assert.EqualValues(t, 100, r.Total)
}

func TestFail(t *testing.T) {
	r := Fail()
	assert.False(t, r.Success)
	assert.Equal(t, "系统错误，请稍后重试!", r.ErrorMsg)
}

func TestFailWithMsg(t *testing.T) {
	r := FailWithMsg("自定义错误")
	assert.False(t, r.Success)
	assert.Equal(t, "自定义错误", r.ErrorMsg)
}

func TestFailWithError(t *testing.T) {
	r := FailWithError(assert.AnError)
	assert.False(t, r.Success)
	assert.Contains(t, r.ErrorMsg, "assert.AnError")
}

func TestFailWithCode(t *testing.T) {
	err := NewBaseCodeError(10001, "测试错误")
	r := FailWithCode(err)
	assert.False(t, r.Success)
	assert.Equal(t, "测试错误", r.ErrorMsg)
}
