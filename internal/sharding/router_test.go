package sharding

import (
	"testing"

	"github.com/javaup/flashsale-system/internal/config"
	"github.com/stretchr/testify/assert"
)

func newTestRouter() *Router {
	return NewRouter(&config.ShardConfig{DbCount: 2, TableCount: 2})
}

func TestDBForMod(t *testing.T) {
	r := newTestRouter()
	assert.Equal(t, "hmdp_0", r.DBForMod(0))
	assert.Equal(t, "hmdp_1", r.DBForMod(1))
	assert.Equal(t, "hmdp_0", r.DBForMod(2))
}

func TestTableForDivMod(t *testing.T) {
	r := newTestRouter()
	// (0 / 2) % 2 = 0
	assert.Equal(t, "tb_user_0", r.TableForDivMod("tb_user", 0))
	// (1 / 2) % 2 = 0
	assert.Equal(t, "tb_user_0", r.TableForDivMod("tb_user", 1))
	// (2 / 2) % 2 = 1
	assert.Equal(t, "tb_user_1", r.TableForDivMod("tb_user", 2))
	// (3 / 2) % 2 = 1
	assert.Equal(t, "tb_user_1", r.TableForDivMod("tb_user", 3))
	// (4 / 2) % 2 = 0
	assert.Equal(t, "tb_user_0", r.TableForDivMod("tb_user", 4))
}

func TestTableForMod(t *testing.T) {
	r := newTestRouter()
	assert.Equal(t, "tb_order_0", r.TableForMod("tb_order", 0))
	assert.Equal(t, "tb_order_1", r.TableForMod("tb_order", 1))
	assert.Equal(t, "tb_order_0", r.TableForMod("tb_order", 2))
}

func TestUserTable(t *testing.T) {
	r := newTestRouter()
	assert.Equal(t, "hmdp_0", r.UserDB(0))
	assert.Equal(t, "hmdp_1", r.UserDB(1))
	assert.Equal(t, "tb_user_0", r.UserTable(0))
	assert.Equal(t, "tb_user_0", r.UserTable(1))
	assert.Equal(t, "tb_user_1", r.UserTable(2))
}

func TestVoucherOrderTable(t *testing.T) {
	r := newTestRouter()
	assert.Equal(t, "hmdp_0", r.VoucherOrderDB(0))
	assert.Equal(t, "hmdp_1", r.VoucherOrderDB(1))
	// voucher_id % 2
	assert.Equal(t, "tb_voucher_order_0", r.VoucherOrderTable(0))
	assert.Equal(t, "tb_voucher_order_1", r.VoucherOrderTable(1))
}

func TestSeckillVoucherTable(t *testing.T) {
	r := newTestRouter()
	// voucher_id % 2 for db
	assert.Equal(t, "hmdp_0", r.SeckillVoucherDB(0))
	assert.Equal(t, "hmdp_1", r.SeckillVoucherDB(1))
	// (voucher_id / 2) % 2 for table
	assert.Equal(t, "tb_seckill_voucher_0", r.SeckillVoucherTable(0))
	assert.Equal(t, "tb_seckill_voucher_0", r.SeckillVoucherTable(1))
	assert.Equal(t, "tb_seckill_voucher_1", r.SeckillVoucherTable(2))
}

func TestUserPhoneTable(t *testing.T) {
	r := newTestRouter()
	// Test that hash-based routing is deterministic
	phone1 := "13838411438"
	phone2 := "13686869696"

	table1 := r.UserPhoneTable(phone1)
	table2 := r.UserPhoneTable(phone1)
	assert.Equal(t, table1, table2, "hash must be deterministic")

	db1 := r.UserPhoneDB(phone1)
	db2 := r.UserPhoneDB(phone1)
	assert.Equal(t, db1, db2, "hash must be deterministic")

	// Different phones may route to different shards
	t.Logf("phone %s -> db=%s, table=%s", phone1, db1, table1)
	t.Logf("phone %s -> db=%s, table=%s", phone2, r.UserPhoneDB(phone2), r.UserPhoneTable(phone2))
}

func TestUserInfoTable(t *testing.T) {
	r := newTestRouter()
	// user_id % 2 for db
	assert.Equal(t, "hmdp_0", r.UserInfoDB(0))
	assert.Equal(t, "hmdp_1", r.UserInfoDB(1))
	// (user_id / 2) % 2 for table
	assert.Equal(t, "tb_user_info_0", r.UserInfoTable(0))
	assert.Equal(t, "tb_user_info_0", r.UserInfoTable(1))
	assert.Equal(t, "tb_user_info_1", r.UserInfoTable(2))
}

func TestNewRouter_Defaults(t *testing.T) {
	r := NewRouter(&config.ShardConfig{DbCount: 4, TableCount: 4})
	assert.Equal(t, "hmdp_3", r.DBForMod(3))
	assert.Equal(t, "tb_test_3", r.TableForMod("tb_test", 3))
}
