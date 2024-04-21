package cache

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/MuxiKeStack/be-stance/domain"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

var ErrKeyNotExists = redis.Nil

const (
	filedSupportCnt = "support_cnt"
	filedOpposeCnt  = "oppose_cnt"
)

//go:embed lua/stance_incr_cnt.lua
var luaScript string

type StanceCache interface {
	GetBizStanceCount(ctx context.Context, biz int32, bizId int64) (domain.UserBizStance, error)
	SetBizStanceCount(ctx context.Context, biz int32, bizId int64, ubs domain.UserBizStance) error
	IncrBizStanceCountIfPresent(ctx context.Context, biz int32, bizId int64, supportDelta int64, opposeDelta int64) error
}

type RedisStanceCache struct {
	cmd redis.Cmdable
}

func NewRedisStanceCache(cmd redis.Cmdable) StanceCache {
	return &RedisStanceCache{cmd: cmd}
}

func (cache *RedisStanceCache) GetBizStanceCount(ctx context.Context, biz int32, bizId int64) (domain.UserBizStance, error) {
	key := cache.bizStanceCountKey(biz, bizId)
	data, err := cache.cmd.HGetAll(ctx, key).Result()
	if err != nil {
		return domain.UserBizStance{}, err
	}
	if len(data) == 0 {
		return domain.UserBizStance{}, ErrKeyNotExists
	}
	supportCnt, _ := strconv.ParseInt(data[filedSupportCnt], 10, 64)
	opposeCnt, _ := strconv.ParseInt(data[filedOpposeCnt], 10, 64)
	return domain.UserBizStance{
		SupportCnt: supportCnt,
		OpposeCnt:  opposeCnt,
	}, nil
}

func (cache *RedisStanceCache) SetBizStanceCount(ctx context.Context, biz int32, bizId int64, ubs domain.UserBizStance) error {
	key := cache.bizStanceCountKey(biz, bizId)
	err := cache.cmd.HSet(ctx, key,
		filedSupportCnt, ubs.SupportCnt,
		filedOpposeCnt, ubs.OpposeCnt).Err()
	if err != nil {
		return err
	}
	return cache.cmd.Expire(ctx, key, time.Minute*15).Err()
}

func (cache *RedisStanceCache) IncrBizStanceCountIfPresent(ctx context.Context, biz int32, bizId int64, supportDelta int64,
	opposeDelta int64) error {
	key := cache.bizStanceCountKey(biz, bizId)
	return cache.cmd.Eval(ctx, luaScript, []string{key}, filedSupportCnt, supportDelta, filedOpposeCnt, opposeDelta).Err()
}

func (cache *RedisStanceCache) bizStanceCountKey(biz int32, bizId int64) string {
	return fmt.Sprintf("kstack:stance:biz_count:<%d,%d>", biz, bizId)
}
