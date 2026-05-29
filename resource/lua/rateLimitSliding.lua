-- rateLimitSliding.lua - Sliding window rate limiter
-- Keys: [1] window_key
-- Args: [1] window_ms, [2] max_requests, [3] now_ms
local key = KEYS[1]
local window = tonumber(ARGV[1])
local max_req = tonumber(ARGV[2])
local now = tonumber(ARGV[3])

-- Remove expired entries
redis.call('ZREMRANGEBYSCORE', key, 0, now - window)

-- Count current requests
local count = redis.call('ZCARD', key)

if count >= max_req then
    return {0, "rate_limit_exceeded"}
end

-- Add current request
redis.call('ZADD', key, now, now .. ':' .. math.random())
redis.call('EXPIRE', key, math.ceil(window / 1000))

return {1, "ok"}
