-- tokenBucket.lua - Token bucket rate limiter
-- Keys: [1] token_key
-- Args: [1] rate, [2] capacity, [3] now_ms, [4] tokens_to_consume
local key = KEYS[1]
local rate = tonumber(ARGV[1])
local capacity = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local consume = tonumber(ARGV[4]) or 1

local bucket = redis.call('HMGET', key, 'tokens', 'last_refill')
local tokens = tonumber(bucket[1]) or capacity
local last_refill = tonumber(bucket[2]) or now

-- Refill tokens based on elapsed time
local elapsed = now - last_refill
local new_tokens = math.min(capacity, tokens + elapsed * rate / 1000)

if new_tokens >= consume then
    new_tokens = new_tokens - consume
    redis.call('HMSET', key, 'tokens', new_tokens, 'last_refill', now)
    redis.call('EXPIRE', key, math.ceil(capacity / rate * 2))
    return {1, "ok", new_tokens}
else
    return {0, "rate_limit_exceeded", new_tokens}
end
