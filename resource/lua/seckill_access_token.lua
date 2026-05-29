-- seckill_access_token.lua - Generate and validate seckill access token
-- Keys: [1] token_hash_key
-- Args: [1] action, [2] voucher_id, [3] user_id, [4] token_value, [5] ttl_seconds
-- action: "generate" or "validate"
local key = KEYS[1]
local action = ARGV[1]

if action == "generate" then
    local voucher_id = ARGV[2]
    local user_id = ARGV[3]
    local token = ARGV[4]
    local ttl = tonumber(ARGV[5]) or 300

    local field = user_id .. ":" .. voucher_id
    redis.call('HSET', key, field, token)
    redis.call('EXPIRE', key, ttl)
    return {1, token}

elseif action == "validate" then
    local voucher_id = ARGV[2]
    local user_id = ARGV[3]
    local token = ARGV[4]

    local field = user_id .. ":" .. voucher_id
    local stored = redis.call('HGET', key, field)
    if stored == token then
        redis.call('HDEL', key, field)
        return {1, "valid"}
    end
    return {0, "invalid"}
end

return {0, "unknown_action"}
