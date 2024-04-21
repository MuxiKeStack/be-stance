local key = KEYS[1]
local supportCntKey = ARGV[1]
local supportDelta = tonumber(ARGV[2])
local opposeCntKey = ARGV[3]
local opposeDelta = tonumber(ARGV[4])
local exists = redis.call("EXISTS", key)
if exists == 1 then
    redis.call("HINCRBY", key, supportCntKey, supportDelta)
    redis.call("HINCRBY", key, opposeCntKey, opposeDelta)
    return 1
else
    return 0
end