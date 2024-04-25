package repository

import (
	"context"
	stancev1 "github.com/MuxiKeStack/be-api/gen/proto/stance/v1"
	"github.com/MuxiKeStack/be-stance/domain"
	"github.com/MuxiKeStack/be-stance/pkg/logger"
	"github.com/MuxiKeStack/be-stance/repository/cache"
	"github.com/MuxiKeStack/be-stance/repository/dao"
)

type StanceRepository interface {
	Endorse(ctx context.Context, ubs domain.UserBizStance) error
	GetUserBizStance(ctx context.Context, uid int64, biz stancev1.Biz, bizId int64) (domain.UserBizStance, error)
	CountStance(ctx context.Context, biz stancev1.Biz, bizId int64) (domain.UserBizStance, error)
}

type CachedStanceRepository struct {
	dao   dao.StanceDAO
	cache cache.StanceCache
	l     logger.Logger
}

func NewCachedStanceRepository(dao dao.StanceDAO, cache cache.StanceCache, l logger.Logger) StanceRepository {
	return &CachedStanceRepository{dao: dao, cache: cache, l: l}
}

func (repo *CachedStanceRepository) Endorse(ctx context.Context, ubs domain.UserBizStance) error {
	// 要返回一个增量 increment 表示<biz,bizId>的支持数，反对数的变化，这里是不是能利用位运算
	increment, err := repo.dao.Upsert(ctx, repo.toEntity(ubs))
	if err != nil {
		return err
	}
	// todo 保持缓存一致,但是无法保证强一致，但是可以容忍
	// 更新缓存，要知道是增是减，这个要从dao里面返回出来
	return repo.cache.IncrBizStanceCountIfPresent(ctx, int32(ubs.Biz), ubs.BizId, increment.Support, increment.Oppose)
}

func (repo *CachedStanceRepository) GetUserBizStance(ctx context.Context, uid int64, biz stancev1.Biz, bizId int64) (domain.UserBizStance, error) {
	// 要聚合count和stance,count先从缓存中拿
	// 这个get，只有count数据是有效的
	ubs, err := repo.cache.GetBizStanceCount(ctx, int32(biz), bizId)
	if err != nil {
		// 不区分redis是否崩了，直接去数据库拿
		bsc, er := repo.dao.GetBizStanceCount(ctx, int32(biz), bizId)
		if er != nil && er != dao.ErrRecordNotFound {
			return domain.UserBizStance{}, er
		}
		ubs.SupportCnt = bsc.SupportCnt
		ubs.OpposeCnt = bsc.OpposeCnt
	}
	go func() {
		// 回写数量缓存
		er := repo.cache.SetBizStanceCount(ctx, int32(biz), bizId, ubs)
		if er != nil {
			repo.l.Error("回写Stance Count缓存失败",
				logger.Error(err),
				logger.String("biz", biz.String()),
				logger.Int64("bizId", bizId))
		}
	}()
	// 要从数据库查 拿到个人stance
	daoUbs, err := repo.dao.GetUserBizStance(ctx, uid, int32(biz), bizId)
	if err != nil && err != dao.ErrRecordNotFound {
		return domain.UserBizStance{}, err
	}
	ubs.Uid = uid
	ubs.Stance = stancev1.Stance(daoUbs.Stance)
	ubs.Biz = biz
	ubs.BizId = bizId
	return ubs, nil
}

func (repo *CachedStanceRepository) CountStance(ctx context.Context, biz stancev1.Biz, bizId int64) (domain.UserBizStance, error) {
	ubs, err := repo.cache.GetBizStanceCount(ctx, int32(biz), bizId)
	if err == nil {
		// 不区分redis是否崩了，直接去数据库拿
		return ubs, nil
	}
	bsc, er := repo.dao.GetBizStanceCount(ctx, int32(biz), bizId)
	if er != nil && er != dao.ErrRecordNotFound {
		return domain.UserBizStance{}, er
	}
	ubs.SupportCnt = bsc.SupportCnt
	ubs.OpposeCnt = bsc.OpposeCnt
	go func() {
		// 回写数量缓存
		er := repo.cache.SetBizStanceCount(ctx, int32(biz), bizId, ubs)
		if er != nil {
			repo.l.Error("回写Stance Count缓存失败",
				logger.Error(err),
				logger.String("biz", biz.String()),
				logger.Int64("bizId", bizId))
		}
	}()
	return ubs, nil
}

func (repo *CachedStanceRepository) toEntity(ubs domain.UserBizStance) dao.UserBizStance {
	return dao.UserBizStance{
		Uid:    ubs.Uid,
		Biz:    int32(ubs.Biz),
		BizId:  ubs.BizId,
		Stance: int32(ubs.Stance),
	}
}
