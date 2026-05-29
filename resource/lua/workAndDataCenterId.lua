-- workAndDataCenterId.lua - Assign Snowflake worker/datacenter IDs
-- Keys: [1] worker_key
-- Args: [1] worker_id, [2] datacenter_id
-- Returns: {success, worker_id, datacenter_id}
local key = KEYS[1]
local worker_id = tonumber(ARGV[1])
local datacenter_id = tonumber(ARGV[2])
local max_workers = 31

-- Try to claim this worker ID
local claimed = redis.call('HSETNX', key, worker_id, datacenter_id)
if claimed == 1 then
    redis.call('EXPIRE', key, 60)
    return {1, worker_id, datacenter_id}
end

-- Worker ID already claimed, try the next one
for i = worker_id + 1, max_workers do
    claimed = redis.call('HSETNX', key, i, datacenter_id)
    if claimed == 1 then
        redis.call('EXPIRE', key, 60)
        return {1, i, datacenter_id}
    end
end

-- Try from 0
for i = 0, worker_id - 1 do
    claimed = redis.call('HSETNX', key, i, datacenter_id)
    if claimed == 1 then
        redis.call('EXPIRE', key, 60)
        return {1, i, datacenter_id}
    end
end

return {0, -1, -1}
