-- seckillVoucher.lua - Core seckill atomic operation
-- Keys: [1] stock_key, [2] user_key, [3] trace_key, [4] voucher_key
-- Args: [1] user_id, [2] voucher_id, [3] trace_id
-- Returns: {success, msg}
local stock_key = KEYS[1]
local user_key = KEYS[2]
local trace_key = KEYS[3]
local voucher_key = KEYS[4]

local user_id = ARGV[1]
local voucher_id = ARGV[2]
local trace_id = ARGV[3]

-- 1. Check voucher exists and is valid
local voucher_json = redis.call('GET', voucher_key)
if not voucher_json then
    return {0, "秒杀优惠券不存在"}
end

-- 2. Check stock
local stock = redis.call('GET', stock_key)
if not stock then
    return {0, "秒杀优惠券库存不存在"}
end
if tonumber(stock) <= 0 then
    return {0, "秒杀优惠券库存不足"}
end

-- 3. Check one-person-one-order
local is_member = redis.call('SISMEMBER', user_key, user_id)
if is_member == 1 then
    return {0, "秒杀优惠券已领取"}
end

-- 4. Decrement stock
redis.call('DECR', stock_key)

-- 5. Record user
redis.call('SADD', user_key, user_id)

-- 6. Write trace log
redis.call('HSET', trace_key, 'user_id', user_id, 'voucher_id', voucher_id, 'trace_id', trace_id, 'status', 'success')
redis.call('EXPIRE', trace_key, 86400)

return {1, "success"}
