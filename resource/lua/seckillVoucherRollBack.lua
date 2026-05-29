-- seckillVoucherRollBack.lua - Rollback seckill operation
-- Keys: [1] stock_key, [2] user_key, [3] trace_key
-- Args: [1] user_id, [2] voucher_id
local stock_key = KEYS[1]
local user_key = KEYS[2]
local trace_key = KEYS[3]

local user_id = ARGV[1]
local voucher_id = ARGV[2]

-- 1. Restore stock
redis.call('INCR', stock_key)

-- 2. Remove user from purchased set
redis.call('SREM', user_key, user_id)

-- 3. Update trace status
redis.call('HSET', trace_key, 'status', 'rolled_back')

return {1, "rollback_success"}
