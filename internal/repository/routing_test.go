package repository

import (
	"testing"

	"github.com/javaup/flashsale-system/internal/config"
	"github.com/javaup/flashsale-system/internal/sharding"
	"github.com/stretchr/testify/assert"
)

func TestUserRouting(t *testing.T) {
	r := sharding.NewRouter(&config.ShardConfig{DbCount: 2, TableCount: 2})
	tests := []struct {
		id      int64
		wantDB  string
		wantTbl string
	}{
		{1, "hmdp_1", "tb_user_0"},
		{2, "hmdp_0", "tb_user_1"},
		{3, "hmdp_1", "tb_user_1"},
		{4, "hmdp_0", "tb_user_0"},
		{0, "hmdp_0", "tb_user_0"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.wantDB, r.UserDB(tt.id))
		assert.Equal(t, tt.wantTbl, r.UserTable(tt.id))
	}
}

func TestVoucherRouting(t *testing.T) {
	r := sharding.NewRouter(&config.ShardConfig{DbCount: 2, TableCount: 2})
	assert.Equal(t, "hmdp_1", r.VoucherDB(1))
	assert.Equal(t, "tb_voucher_0", r.VoucherTable(1))
}

func TestSeckillVoucherRouting(t *testing.T) {
	r := sharding.NewRouter(&config.ShardConfig{DbCount: 2, TableCount: 2})
	assert.Equal(t, "hmdp_1", r.SeckillVoucherDB(1))
	assert.Equal(t, "tb_seckill_voucher_0", r.SeckillVoucherTable(1))
}

func TestVoucherOrderRouting(t *testing.T) {
	r := sharding.NewRouter(&config.ShardConfig{DbCount: 2, TableCount: 2})
	assert.Equal(t, "hmdp_1", r.VoucherOrderDB(1))             // user_id % 2
	assert.Equal(t, "tb_voucher_order_1", r.VoucherOrderTable(1)) // voucher_id % 2
	assert.Equal(t, "tb_voucher_order_0", r.VoucherOrderTable(2))
}

func TestUserPhoneRouting(t *testing.T) {
	r := sharding.NewRouter(&config.ShardConfig{DbCount: 2, TableCount: 2})
	db := r.UserPhoneDB("13800138000")
	table := r.UserPhoneTable("13800138000")
	assert.Contains(t, []string{"hmdp_0", "hmdp_1"}, db)
	assert.Contains(t, []string{"tb_user_phone_0", "tb_user_phone_1"}, table)
}

func TestBroadcastDBConstant(t *testing.T) {
	assert.Equal(t, "hmdp_0", broadcastDB)
}
